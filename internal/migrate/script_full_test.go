//go:build full

package migrate

import (
	"bytes"
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/source"
	_ "modernc.org/sqlite"
)

func mixedService(t *testing.T, items []source.Migration, dry bool) (*Service, *sql.DB, *bytes.Buffer) {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "mixed.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	out := &bytes.Buffer{}
	svc, err := NewService(context.Background(), Config{DB: db, Dialect: dialect.SQLiteDialect{}, MigrationSource: source.StringSource{Migrations: items}, DryRun: dry, DryRunOutput: out})
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.Create(); err != nil {
		t.Fatal(err)
	}
	return svc, db, out
}

func mixedSources() []source.Migration {
	return []source.Migration{
		{Filename: "V1__create.sql", Source: migrationSQL("CREATE TABLE users(id INTEGER PRIMARY KEY, profile TEXT);", "DROP TABLE users;")},
		{Filename: "V2__seed.js", Source: `export function up(db){db.exec('INSERT INTO users VALUES (?,?)',['9007199254740993','old'])}export function down(db){db.exec('DELETE FROM users')}`},
		{Filename: "V3__transform.ts", Source: `export function up(db:any):void{const row=db.query('SELECT id,profile FROM users')[0];db.exec('UPDATE users SET profile=? WHERE id=?',[row.profile+'-new',row.id])}export function down(db:any):void{db.exec('UPDATE users SET profile=?',['old'])}`},
	}
}

func TestMixedLifecycleAndHashes(t *testing.T) {
	for _, extension := range []string{"js", "mjs"} {
		t.Run(extension, func(t *testing.T) { testMixedLifecycleAndHashes(t, extension) })
	}
}

func testMixedLifecycleAndHashes(t *testing.T, extension string) {
	t.Helper()
	items := mixedSources()
	items[1].Filename = "V2__seed." + extension
	svc, db, _ := mixedService(t, items, false)
	for range 2 {
		if err := svc.Up(); err != nil {
			t.Fatal(err)
		}
	}
	var profile, id string
	if err := db.QueryRow("SELECT id,profile FROM users").Scan(&id, &profile); err != nil || id != "9007199254740993" || profile != "old-new" {
		t.Fatalf("%s %s %v", id, profile, err)
	}
	statuses, err := svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	for index, st := range statuses {
		want := 64
		if index == 0 {
			want = 32
		}
		if st.Status != StatusApplied || len(st.Migration.Hash) != want {
			t.Fatalf("%+v", st)
		}
	}
	if err = svc.Down("2", false); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "old" {
		t.Fatalf("%s %v", profile, err)
	}
	if err = svc.Down("", true); err != nil {
		t.Fatal(err)
	}
	statuses, err = svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range statuses {
		if st.Status != StatusPending {
			t.Fatalf("%+v", st)
		}
	}
}

func TestPreflightAndMetadataNeverPartiallyApply(t *testing.T) {
	for _, sourceText := range []string{`export function up(`, `export const up=1`, `export async function* up(){yield 1}`, `export const config={transaction:"false"};export function up(){}`} {
		t.Run(sourceText[:12], func(t *testing.T) {
			items := mixedSources()
			items[1].Source = sourceText
			svc, db, _ := mixedService(t, items, false)
			if _, err := svc.Status(); err != nil {
				t.Fatal(err)
			}
			if err := svc.Up(); err == nil {
				t.Fatal("expected preflight failure")
			}
			var count int
			if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='users'").Scan(&count); err != nil || count != 0 {
				t.Fatalf("partial plan execution: %d %v", count, err)
			}
		})
	}
}

func TestIrreversibleRangeDoesNotPartiallyRollback(t *testing.T) {
	items := mixedSources()
	items[1].Source = `export function up(db){db.exec('INSERT INTO users VALUES (?,?)',['1','old'])}`
	svc, db, _ := mixedService(t, items, false)
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	if err := svc.Down("", true); err == nil {
		t.Fatal("expected irreversible migration")
	}
	var profile string
	if err := db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "old-new" {
		t.Fatalf("partial rollback: %s %v", profile, err)
	}
}

func TestScriptsNotEvaluatedByDryRunOrBaseline(t *testing.T) {
	items := mixedSources()
	items[1].Filename = "V2__seed.mjs"
	items[1].Source = `throw Error('must not run');export function up(){}`
	svc, db, out := mixedService(t, items, true)
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "script up V2__seed.mjs (not executed)") {
		t.Fatalf("%s", out.String())
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM migration_schema").Scan(&count); err != nil || count != 0 {
		t.Fatalf("%d %v", count, err)
	}
	if err := svc.Baseline(); err != nil {
		t.Fatal(err)
	}
	statuses, err := svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range statuses {
		if st.Status != StatusBaseline {
			t.Fatalf("%+v", st)
		}
	}
}

func TestCrossFormatDuplicateVersion(t *testing.T) {
	for _, filename := range []string{"V1__duplicate.js", "V1__duplicate.mjs", "V2__duplicate.mjs", "V3__duplicate.mjs"} {
		t.Run(filename, func(t *testing.T) {
			items := append(mixedSources(), source.Migration{Filename: filename, Source: `export function up(){}`})
			svc, _, _ := mixedService(t, items, false)
			if err := svc.Up(); err == nil || !strings.Contains(err.Error(), "duplicate version") {
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestScriptTamperingBlocksExecution(t *testing.T) {
	items := mixedSources()
	items[1].Filename = "V2__seed.mjs"
	svc, db, _ := mixedService(t, items, false)
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	items[1].Source += "\n// changed"
	svc.cfg.MigrationSource = source.StringSource{Migrations: items}
	if err := svc.Up(); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("%v", err)
	}
	var profile string
	if err := db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "old-new" {
		t.Fatalf("%s %v", profile, err)
	}
}

func TestDisabledOutOfOrderScriptIsNotInitialized(t *testing.T) {
	items := mixedSources()
	svc, _, _ := mixedService(t, items, false)
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	items = append(items, source.Migration{Filename: "V0__older.js", Source: `throw Error('must not initialize');export function up(){}`})
	svc.cfg.MigrationSource = source.StringSource{Migrations: items}
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	svc.cfg.MigrateOutOfOrder = true
	if err := svc.Up(); err == nil {
		t.Fatal("enabled older script was not initialized")
	}
}

func TestScriptAndHistoryFailureRollbackTogether(t *testing.T) {
	items := []source.Migration{
		{Filename: "V1__create.sql", Source: migrationSQL("CREATE TABLE users(profile TEXT);\nINSERT INTO users VALUES ('old');", "DROP TABLE users;")},
		{Filename: "V2__update.js", Source: `export function up(db){db.exec('UPDATE users SET profile=?',['changed'])}`},
	}
	svc, db, _ := mixedService(t, items, false)
	if _, err := db.Exec(`CREATE TRIGGER reject_history BEFORE INSERT ON migration_schema WHEN NEW.version='2' BEGIN SELECT RAISE(ABORT,'history failure'); END;`); err != nil {
		t.Fatal(err)
	}
	if err := svc.Up(); err == nil {
		t.Fatal("expected history failure")
	}
	var profile string
	var count int
	if err := db.QueryRow("SELECT profile FROM users").Scan(&profile); err != nil || profile != "old" {
		t.Fatalf("%s %v", profile, err)
	}
	if err := db.QueryRow("SELECT count(*) FROM migration_schema").Scan(&count); err != nil || count != 1 {
		t.Fatalf("%d %v", count, err)
	}
}
