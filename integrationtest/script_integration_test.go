//go:build integration && full

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/script"
	"github.com/exc-works/migrate/internal/source"
)

func TestScriptMigrationsIntegration(t *testing.T) {
	for _, flavor := range selectedFlavors() {
		if !script.SupportedDialect(flavor) {
			continue
		}
		t.Run(flavor, func(t *testing.T) {
			t.Parallel()
			db, d, cleanup := startDB(t, flavor)
			defer cleanup()
			runScriptScenario(t, db, d)
		})
	}
}

func runScriptScenario(t *testing.T, db *sql.DB, d dialect.Dialect) {
	t.Helper()
	ctx := context.Background()
	table := randomName("script_users")
	amountType, binaryType := "DECIMAL(30,10)", "BLOB"
	insertParams := "?,?,?,?"
	if d.Name() == "postgres" {
		binaryType = "BYTEA"
		insertParams = "$1,$2,$3,$4"
	}
	// SQLite NUMERIC affinity can store decimals as floats. TEXT preserves exact
	// decimal input; the runtime rejects float-backed declared numeric columns.
	if d.Name() == "sqlite" {
		amountType = "TEXT"
	}
	create := fmt.Sprintf("CREATE TABLE %s (id BIGINT PRIMARY KEY, amount %s, profile TEXT, payload %s);", table, amountType, binaryType)
	insert := fmt.Sprintf("INSERT INTO %s VALUES (%s)", table, insertParams)
	selectSQL := "SELECT id,amount,profile,payload FROM " + table
	update := "UPDATE " + table + " SET profile="
	if d.Name() == "postgres" {
		update += "$1 WHERE id=$2"
	} else {
		update += "? WHERE id=?"
	}
	items := []source.Migration{
		{Filename: "V1__create.sql", Source: "-- +migrate Up\n" + create + "\n-- +migrate Down\nDROP TABLE " + table + ";\n"},
		{Filename: "V2__seed.js", Source: `export function up(db){db.exec(` + strconv.Quote(insert) + `,['9007199254740993','12345678901234567890.1234567890','old',[0,255]]);const r=db.query(` + strconv.Quote(selectSQL) + `)[0];if(r.id!=='9007199254740993'||r.amount!=='12345678901234567890.1234567890'||r.payload.length!==2||r.payload[1]!==255)throw Error('value fidelity');db.exec(` + strconv.Quote(update) + `,['new',r.id]);}export function down(db){db.exec('DELETE FROM ` + table + `')}`},
		{Filename: "V3__transform.ts", Source: `export function up(db:any){const r=db.query(` + strconv.Quote(selectSQL) + `)[0];db.exec(` + strconv.Quote(update) + `,[r.profile+'-ts',r.id]);}export function down(db:any){db.exec(` + strconv.Quote(update) + `,['new','9007199254740993']);}`},
	}
	cfg := migrate.Config{DB: db, Dialect: d, SchemaName: randomName("script_history"), MigrationSource: source.StringSource{Migrations: items}}
	svc, err := migrate.NewService(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.Create(); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = svc.Up(); err != nil {
			t.Fatal(err)
		}
	}
	assertScriptProfile(t, db, table, "new-ts")
	statuses, err := svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range statuses {
		if st.Status != migrate.StatusApplied {
			t.Fatalf("%+v", st)
		}
	}

	// Fail after a real update; even a caught DB error must roll the update back.
	broken := append([]source.Migration(nil), items...)
	broken = append(broken, source.Migration{Filename: "V4__failure.js", Source: `export function up(db){db.exec(` + strconv.Quote(update) + `,['temporary','9007199254740993']);try{db.query('SELECT * FROM missing_script_table')}catch(e){}}`})
	failedCfg := cfg
	failedCfg.MigrationSource = source.StringSource{Migrations: broken}
	failed, err := migrate.NewService(ctx, failedCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = failed.Up(); err == nil {
		t.Fatal("caught DB failure recorded success")
	}
	assertScriptProfile(t, db, table, "new-ts")

	// Verify VM timeout after a write rolls back on each supported real driver.
	broken[len(broken)-1].Source = `export function up(db){db.exec(` + strconv.Quote(update) + `,['temporary','9007199254740993']);while(true){}}`
	failedCfg.ScriptOptions = script.Options{Timeout: 30 * time.Millisecond}
	failed, err = migrate.NewService(ctx, failedCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = failed.Up(); err == nil {
		t.Fatal("expected timeout")
	}
	assertScriptProfile(t, db, table, "new-ts")

	// Transaction control is rejected before it can release framework ownership.
	if d.Name() == "postgres" {
		for _, control := range []string{"ABORT", "/* outer /* inner */ ignored */ COMMIT"} {
			broken[len(broken)-1].Source = `export function up(db){db.exec(` + strconv.Quote(update) + `,['temporary','9007199254740993']);db.exec(` + strconv.Quote(control) + `);throw Error('after control')}`
			failedCfg.ScriptOptions = script.Options{}
			failed, err = migrate.NewService(ctx, failedCfg)
			if err != nil {
				t.Fatal(err)
			}
			if err = failed.Up(); err == nil || !strings.Contains(err.Error(), "manual transaction") {
				t.Fatalf("expected bridge transaction-control rejection: %v", err)
			}
			assertScriptProfile(t, db, table, "new-ts")
		}
	}

	tampered := append([]source.Migration(nil), items...)
	tampered[1].Source += "\n// changed"
	badCfg := cfg
	badCfg.MigrationSource = source.StringSource{Migrations: tampered}
	bad, err := migrate.NewService(ctx, badCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = bad.Up(); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("%v", err)
	}

	baselineCfg := cfg
	baselineCfg.SchemaName = randomName("script_baseline")
	baselineItems := append([]source.Migration(nil), items...)
	baselineItems[1].Source = `throw Error('must not initialize');export function up(){}`
	baselineCfg.MigrationSource = source.StringSource{Migrations: baselineItems}
	baseline, err := migrate.NewService(ctx, baselineCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = baseline.Create(); err != nil {
		t.Fatal(err)
	}
	if err = baseline.Baseline(); err != nil {
		t.Fatal(err)
	}
	statuses, err = baseline.Status()
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range statuses {
		if st.Status != migrate.StatusBaseline {
			t.Fatalf("%+v", st)
		}
	}

	if err = svc.Down("2", false); err != nil {
		t.Fatal(err)
	}
	assertScriptProfile(t, db, table, "new")
	if err = svc.Down("", true); err != nil {
		t.Fatal(err)
	}
	statuses, err = svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range statuses {
		if st.Status != migrate.StatusPending {
			t.Fatalf("%+v", st)
		}
	}
}

func assertScriptProfile(t *testing.T, db *sql.DB, table, want string) {
	t.Helper()
	var got string
	if err := db.QueryRow("SELECT profile FROM " + table).Scan(&got); err != nil || got != want {
		t.Fatalf("profile=%q want=%q error=%v", got, want, err)
	}
}
