//go:build integration && full

package integrationtest

import (
	"context"
	"database/sql"
	"github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/source"
	"strconv"
	"strings"
	"testing"

	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/script"
)

func runMSSQLScalarScenario(t *testing.T, db *sql.DB) {
	t.Helper()
	query := `SELECT CAST(1 AS BIT) AS flag,CAST('922337203685477.5807' AS MONEY) AS amount,CAST('-214748.3648' AS SMALLMONEY) AS small_amount,CAST('00112233-4455-6677-8899-aabbccddeeff' AS UNIQUEIDENTIFIER) AS guid,CONVERT(VARCHAR(36),CAST('00112233-4455-6677-8899-aabbccddeeff' AS UNIQUEIDENTIFIER)) AS guid_text,CAST('2026-11-03T01:02:03.1234567+08:00' AS DATETIMEOFFSET(7)) AS instant,CAST(N'中文' AS NVARCHAR(20)) AS label,CAST('<root/>' AS XML) AS xml,CAST(0x00FF AS VARBINARY(2)) AS payload`
	program, err := script.Compile("V0__mssql_scalars.js", `export function up(db){const r=db.query(`+strconv.Quote(query)+`)[0];if(r.flag!==true||r.amount!=='922337203685477.5807'||r.small_amount!=='-214748.3648'||r.guid.length!==16||r.guid_text.toLowerCase()!=='00112233-4455-6677-8899-aabbccddeeff'||r.instant!=='2026-11-02T17:02:03.1234567Z'||r.label!=='中文'||(r.xml!=='<root/>'&&r.xml!=='<root />')||r.payload.length!==2||r.payload[1]!==255)throw Error('MSSQL scalar fidelity')}`, "up", script.Options{})
	if err != nil {
		t.Fatal(err)
	}
	invocation, err := program.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = invocation.Execute(context.Background(), db, dialect.NameMSSQL); err != nil {
		t.Fatal(err)
	}
}

func runMSSQLNativeScriptScenario(t *testing.T, db *sql.DB, d dialect.Dialect) {
	t.Helper()
	table, proc := randomName("native_users"), randomName("native_proc")
	ctx := context.Background()
	defer db.ExecContext(ctx, "DROP PROCEDURE IF EXISTS "+proc)
	defer db.ExecContext(ctx, "DROP TABLE IF EXISTS "+table)
	createProc := "CREATE OR ALTER PROCEDURE " + proc + " AS BEGIN SET NOCOUNT ON; UPDATE " + table + " SET profile=N'procedure'; END;"
	batch := "-- native T-SQL batch\nDECLARE @value NVARCHAR(50)=N'batch;value'; UPDATE " + table + " SET profile=@value,session_id=@@SPID;"
	items := []source.Migration{
		{Filename: "V1__native.sql", Source: "-- +migrate Up\nCREATE TABLE " + table + "(profile NVARCHAR(50),session_id INT); INSERT INTO " + table + " VALUES(N'old',0);\n-- +migrate Down\nDROP TABLE " + table + ";"},
		{Filename: "V2__native.ts", Source: `export const config={transaction:false};export function up(db:any){db.exec('BEGIN TRANSACTION');db.exec(` + strconv.Quote(batch) + `);if(db.query('SELECT profile FROM ` + table + `')[0].profile!=='batch;value')throw Error('batch');db.exec('COMMIT TRANSACTION');db.exec(` + strconv.Quote(createProc) + `);db.exec('EXEC ` + proc + `');}export function down(db:any){db.exec('BEGIN TRANSACTION');db.exec("UPDATE ` + table + ` SET profile=N'old'");db.exec('COMMIT TRANSACTION');db.exec('DROP PROCEDURE ` + proc + `');}`},
	}
	cfg := migrate.Config{DB: db, Dialect: d, SchemaName: randomName("native_history"), MigrationSource: source.StringSource{Migrations: items}}
	svc, err := migrate.NewService(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.Create(); err != nil {
		t.Fatal(err)
	}
	if err = svc.Up(); err != nil {
		t.Fatal(err)
	}
	assertScriptProfile(t, db, table, "procedure")
	var scriptSession, reusedSession int
	if err = db.QueryRowContext(ctx, "SELECT session_id,@@SPID FROM "+table).Scan(&scriptSession, &reusedSession); err != nil || scriptSession != reusedSession {
		t.Fatalf("healthy MSSQL session was replaced: %d != %d: %v", scriptSession, reusedSession, err)
	}
	failedItems := append([]source.Migration(nil), items...)
	failedItems = append(failedItems, source.Migration{Filename: "V3__partial.js", Source: `export const config={transaction:false};export function up(db){db.exec("UPDATE ` + table + ` SET profile=N'partial'");throw Error('after write')}`})
	failedCfg := cfg
	failedCfg.MigrationSource = source.StringSource{Migrations: failedItems}
	failed, err := migrate.NewService(ctx, failedCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = failed.Up(); err == nil || !strings.Contains(err.Error(), "partial changes") {
		t.Fatalf("%v", err)
	}
	assertScriptProfile(t, db, table, "partial")
	statuses, err := failed.Status()
	if err != nil || statuses[len(statuses)-1].Status != migrate.StatusPending {
		t.Fatalf("%+v %v", statuses, err)
	}
	if err = svc.Down("1", false); err != nil {
		t.Fatal(err)
	}
	assertScriptProfile(t, db, table, "old")
}
