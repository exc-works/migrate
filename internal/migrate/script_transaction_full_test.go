//go:build full

package migrate

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/exc-works/migrate/internal/script"
	"github.com/exc-works/migrate/internal/source"
)

func nontransactionSources(body, down string) []source.Migration {
	return []source.Migration{
		{Filename: "V1__create.sql", Source: migrationSQL("CREATE TABLE users(profile TEXT); INSERT INTO users VALUES ('old');", "DROP TABLE users;")},
		{Filename: "V2__native.js", Source: `export const config={transaction:false};export function up(db){` + body + `}export function down(db){` + down + `}`},
	}
}

func TestNontransactionalScriptManualTransactionsReuseHealthyConnection(t *testing.T) {
	items := nontransactionSources(`db.exec("CREATE TEMP TABLE session_marker(id INT); INSERT INTO session_marker VALUES(1);");db.exec('BEGIN');db.exec("-- native SQL\nUPDATE users SET profile='changed;native';");if(db.query('SELECT profile FROM users')[0].profile!=='changed;native')throw Error('connection changed');db.exec('COMMIT')`, `db.exec('BEGIN');db.exec("UPDATE users SET profile='old'");db.exec('COMMIT')`)
	svc, db, _ := mixedService(t, items, false)
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	var profile string
	if err := db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "changed;native" {
		t.Fatalf("%q %v", profile, err)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_temp_master WHERE name='session_marker'").Scan(&count); err != nil || count != 1 {
		t.Fatalf("healthy connection was discarded: %d %v", count, err)
	}
	statuses, err := svc.Status()
	if err != nil || statuses[1].Status != StatusApplied {
		t.Fatalf("%+v %v", statuses, err)
	}
	if err = svc.Down("1", false); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "old" {
		t.Fatalf("down: %q %v", profile, err)
	}
	statuses, err = svc.Status()
	if err != nil || statuses[1].Status != StatusPending {
		t.Fatalf("down: %+v %v", statuses, err)
	}
}

func TestNontransactionalScriptFailuresExposePartialChanges(t *testing.T) {
	for _, failure := range []string{"script", "database", "history", "timeout", "unfinished transaction"} {
		t.Run(failure, func(t *testing.T) {
			body := `db.exec("CREATE TEMP TABLE failed_session(id INT)");db.exec("UPDATE users SET profile='changed'");`
			want := "changed"
			switch failure {
			case "script":
				body += `throw Error('private value')`
			case "database":
				body += `db.query('SELECT * FROM missing')`
			case "timeout":
				body += `while(true){}`
			case "unfinished transaction":
				body = `db.exec('BEGIN');` + body + `throw Error('private value')`
				want = "old"
			}
			svc, db, _ := mixedService(t, nontransactionSources(body, `db.exec("UPDATE users SET profile='old'")`), false)
			if failure == "timeout" {
				svc.cfg.ScriptOptions = script.Options{Timeout: 50 * time.Millisecond}
			}
			if failure == "history" {
				if _, err := db.Exec(`CREATE TRIGGER reject_native_history BEFORE INSERT ON migration_schema WHEN NEW.version='2' BEGIN SELECT RAISE(ABORT,'history failure'); END;`); err != nil {
					t.Fatal(err)
				}
			}
			err := svc.Up()
			if err == nil || !strings.Contains(err.Error(), "partial changes") || strings.Contains(err.Error(), "private value") {
				t.Fatalf("%v", err)
			}
			var profile string
			if err = db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != want {
				t.Fatalf("%q want %q: %v", profile, want, err)
			}
			var count int
			if err = db.QueryRow("SELECT count(*) FROM sqlite_temp_master WHERE name='failed_session'").Scan(&count); err != nil || count != 0 {
				t.Fatalf("failed connection reused: %d %v", count, err)
			}
			statuses, err := svc.Status()
			if err != nil || statuses[1].Status != StatusPending {
				t.Fatalf("%+v %v", statuses, err)
			}
		})
	}
}

func TestNontransactionalDownHistoryFailureKeepsAppliedHistory(t *testing.T) {
	svc, db, _ := mixedService(t, nontransactionSources(`db.exec("UPDATE users SET profile='changed'")`, `db.exec("UPDATE users SET profile='old'")`), false)
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER reject_native_down BEFORE DELETE ON migration_schema WHEN OLD.version='2' BEGIN SELECT RAISE(ABORT,'history failure'); END;`); err != nil {
		t.Fatal(err)
	}
	if err := svc.Down("1", false); err == nil || !strings.Contains(err.Error(), "partial changes") {
		t.Fatalf("%v", err)
	}
	var profile string
	if err := db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "old" {
		t.Fatalf("%q %v", profile, err)
	}
	statuses, err := svc.Status()
	if err != nil || statuses[1].Status != StatusApplied {
		t.Fatalf("%+v %v", statuses, err)
	}
}

func TestNontransactionalPanicDiscardsConnection(t *testing.T) {
	svc, db, _ := mixedService(t, nil, false)
	p, err := script.Compile("V1__panic.js", `export const config={transaction:false};export function up(){}`, "up", script.Options{})
	if err != nil {
		t.Fatal(err)
	}
	invocation, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	err = svc.withStepTransaction(scriptStep{invocation: invocation}, func(runner execContextRunner) error {
		if _, err := runner.ExecContext(context.Background(), "CREATE TEMP TABLE failed_session(id INT)"); err != nil {
			return err
		}
		panic("private panic value")
	})
	if err == nil || !strings.Contains(err.Error(), "partial changes") || strings.Contains(err.Error(), "private panic value") {
		t.Fatalf("%v", err)
	}
	var count int
	if err = db.QueryRow("SELECT count(*) FROM sqlite_temp_master WHERE name='failed_session'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("panicked connection reused: %d %v", count, err)
	}
}

func TestManagedRecoveryAndReturnValuesShareHistoryTransaction(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprintf("fail=%t", fail), func(t *testing.T) {
			body := `db.exec('SAVEPOINT recovery');try{db.exec('INSERT INTO missing_recovery_table VALUES(1)')}catch(e){db.exec('ROLLBACK TO SAVEPOINT recovery')}db.exec('RELEASE SAVEPOINT recovery');`
			if fail {
				body += `db.exec("UPDATE users SET profile='changed'");throw Error('after recovery')`
			} else {
				body += `return db.exec("UPDATE users SET profile='changed'")`
			}
			items := nontransactionSources(body, `return db.exec("UPDATE users SET profile='old'")`)
			items[1].Source = strings.Replace(items[1].Source, "transaction:false", "transaction:true", 1)
			svc, db, _ := mixedService(t, items, false)
			err := svc.Up()
			if (err != nil) != fail {
				t.Fatalf("%v", err)
			}
			want, state := "changed", StatusApplied
			if fail {
				want, state = "old", StatusPending
			}
			var profile string
			if err = db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != want {
				t.Fatalf("%q %v", profile, err)
			}
			statuses, err := svc.Status()
			if err != nil || statuses[1].Status != state {
				t.Fatalf("%+v %v", statuses, err)
			}
			if !fail {
				if err = svc.Down("1", false); err != nil {
					t.Fatal(err)
				}
				if err = db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "old" {
					t.Fatalf("%q %v", profile, err)
				}
				statuses, err = svc.Status()
				if err != nil || statuses[1].Status != StatusPending {
					t.Fatalf("%+v %v", statuses, err)
				}
			}
		})
	}
}

func TestNontransactionalRecoveryCanRollbackAndContinue(t *testing.T) {
	body := `db.exec('BEGIN');db.exec("UPDATE users SET profile='temporary'");let recovered=false;try{db.query('SELECT * FROM missing_recovery_table')}catch(e){db.exec('ROLLBACK');recovered=true}if(!recovered||db.query('SELECT profile FROM users')[0].profile!=='old')throw Error('rollback did not run');return db.exec("UPDATE users SET profile='recovered'")`
	down := `db.exec('BEGIN');db.exec("UPDATE users SET profile='temporary'");try{db.exec('INSERT INTO missing_recovery_table VALUES(1)')}catch(e){db.exec('ROLLBACK')}return db.exec("UPDATE users SET profile='old'")`
	svc, db, _ := mixedService(t, nontransactionSources(body, down), false)
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	var profile string
	if err := db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "recovered" {
		t.Fatalf("%q %v", profile, err)
	}
	statuses, err := svc.Status()
	if err != nil || statuses[1].Status != StatusApplied {
		t.Fatalf("%+v %v", statuses, err)
	}
	if err = svc.Down("1", false); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "old" {
		t.Fatalf("%q %v", profile, err)
	}
	statuses, err = svc.Status()
	if err != nil || statuses[1].Status != StatusPending {
		t.Fatalf("%+v %v", statuses, err)
	}
}
