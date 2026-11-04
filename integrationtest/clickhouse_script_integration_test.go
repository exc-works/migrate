//go:build integration

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

func runClickHouseScriptScenario(t *testing.T, db *sql.DB, d dialect.Dialect) {
	t.Helper()
	table, history := randomName("script_users"), randomName("script_history")
	create := fmt.Sprintf("CREATE TABLE %s (id UInt64, amount Decimal(30,10), profile String) ENGINE=MergeTree ORDER BY id", table)
	seed := "INSERT INTO " + table + " VALUES (toUInt64(?),toDecimal128(?,10),?)"
	query := "SELECT id,amount,profile FROM " + table
	update := "ALTER TABLE " + table + " UPDATE profile=? WHERE id=toUInt64(?) SETTINGS mutations_sync=2"
	remove := "ALTER TABLE " + table + " DELETE WHERE 1 SETTINGS mutations_sync=2"
	items := []source.Migration{
		{Filename: "V1__create.sql", Source: "-- +migrate Up\n" + create + ";\n-- +migrate Down\nDROP TABLE " + table + ";\n"},
		{Filename: "V2__seed.js", Source: `export function up(db){if(db.exec(` + strconv.Quote(seed) + `,['18446744073709551615','12345678901234567890.123456789','old']).rowsAffected!=='0')throw Error('count');const r=db.query(` + strconv.Quote(query) + `)[0];if(r.id!=='18446744073709551615'||r.amount!=='12345678901234567890.123456789')throw Error('precision')}export function down(db){db.exec(` + strconv.Quote(remove) + `)}`},
		{Filename: "V3__transform.ts", Source: `export function up(db:any){const r=db.query(` + strconv.Quote(query) + `)[0];db.exec(` + strconv.Quote(update) + `,[r.profile+'-new',r.id])}export function down(db:any){db.exec(` + strconv.Quote(update) + `,['old','18446744073709551615'])}`},
	}
	cfg := migrate.Config{DB: db, Dialect: d, SchemaName: history, MigrationSource: source.StringSource{Migrations: items}}
	svc, err := migrate.NewService(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.Create(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TABLE IF EXISTS " + table); db.Exec("DROP TABLE IF EXISTS " + history) })
	// Use the actual driver's non-standard Go values (small integer widths,
	// big.Int, Decimal driver.Valuer and nullable wrappers), not sqlmock coercions.
	scalarQuery := "SELECT toInt8(-8) AS i8,toInt16(-16) AS i16,toInt32(-32) AS i32,toUInt8(8) AS u8,toUInt16(16) AS u16,toUInt32(4294967295) AS u32,toFloat32(1.25) AS f32,toUInt128('340282366920938463463374607431768211455') AS huge,CAST(NULL AS Nullable(UInt8)) AS missing,CAST('text' AS LowCardinality(String)) AS label"
	p, err := script.Compile("V0__scalars.js", `export function up(db){const r=db.query(`+strconv.Quote(scalarQuery)+`)[0];if(r.i8!=='-8'||r.i16!=='-16'||r.i32!=='-32'||r.u8!=='8'||r.u16!=='16'||r.u32!=='4294967295'||r.f32!==1.25||r.huge!=='340282366920938463463374607431768211455'||r.missing!==null||r.label!=='text')throw Error('scalar fidelity')}`, "up", script.Options{})
	if err != nil {
		t.Fatal(err)
	}
	invocation, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = invocation.Execute(context.Background(), db, d.Name()); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = svc.Up(); err != nil {
			t.Fatal(err)
		}
	}
	assertScriptProfile(t, db, table, "old-new")
	statuses, err := svc.Status()
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range statuses {
		if st.Status != migrate.StatusApplied {
			t.Fatalf("%+v", st)
		}
	}
	// An uncaught failure cannot undo already successful writes. No success
	// history may be recorded, and statements after the error must not run.
	failedItems := append([]source.Migration(nil), items...)
	failedItems = append(failedItems, source.Migration{Filename: "V4__failure.js", Source: `export function up(db){db.exec(` + strconv.Quote(update) + `,['partial','18446744073709551615']);db.query('SELECT * FROM missing_script_table');db.exec(` + strconv.Quote(update) + `,['must-not-run','18446744073709551615'])}`})
	cfg.MigrationSource = source.StringSource{Migrations: failedItems}
	failed, err := migrate.NewService(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = failed.Up(); err == nil || !strings.Contains(err.Error(), "partial changes") {
		t.Fatalf("%v", err)
	}
	assertScriptProfile(t, db, table, "partial")
	statuses, err = failed.Status()
	if err != nil || statuses[len(statuses)-1].Status != migrate.StatusPending {
		t.Fatalf("%+v %v", statuses, err)
	}
	// A VM timeout has identical partial-state semantics.
	failedItems[len(failedItems)-1].Source = `export function up(db){db.exec(` + strconv.Quote(update) + `,['timeout-partial','18446744073709551615']);while(true){}}`
	cfg.ScriptOptions = script.Options{Timeout: 2 * time.Second}
	failed, err = migrate.NewService(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = failed.Up(); err == nil || !strings.Contains(err.Error(), "partial changes") {
		t.Fatalf("%v", err)
	}
	assertScriptProfile(t, db, table, "timeout-partial")
	if err = svc.Down("2", false); err != nil {
		t.Fatal(err)
	}
	assertScriptProfile(t, db, table, "old")
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
