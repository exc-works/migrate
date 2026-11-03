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
			if d.Name() == dialect.NameClickHouse {
				runClickHouseScriptScenario(t, db, d)
			} else {
				runScriptScenario(t, db, d)
			}
		})
	}
}

func runScriptScenario(t *testing.T, db *sql.DB, d dialect.Dialect) {
	t.Helper()
	ctx := context.Background()
	if d.Name() == dialect.NameOracle {
		query := `SELECT TO_CLOB('text') AS "text",DBMS_LOB.SUBSTR(TO_BLOB(HEXTORAW('00FF')),2,1) AS "payload",CAST(1.25 AS BINARY_FLOAT) AS "ratio",CAST('1234567890123456789012345678.123456789' AS NUMBER(38,10)) AS "amount" FROM dual`
		p, err := script.Compile("V0__oracle_scalars.js", `export function up(db){const r=db.query(`+strconv.Quote(query)+`)[0];if(r.text!=='text'||r.payload.length!==2||r.payload[1]!==255||r.ratio!==1.25||r.amount!=='1234567890123456789012345678.123456789')throw Error('Oracle scalar fidelity')}`, "up", script.Options{})
		if err != nil {
			t.Fatal(err)
		}
		invocation, err := p.Instantiate(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err = invocation.Execute(ctx, db, d.Name()); err != nil {
			t.Fatal(err)
		}
	}
	if d.Name() == dialect.NameMSSQL {
		runMSSQLScalarScenario(t, db)
		runMSSQLNativeScriptScenario(t, db, d)
	}
	table := randomName("script_users")
	amountType, binaryType := "DECIMAL(30,10)", "BLOB"
	idType := "BIGINT"
	profileType, ratioType := "TEXT", "DOUBLE PRECISION"
	amountExpected := "12345678901234567890.1234567890"
	if d.Name() == "mysql" || d.Name() == "mariadb" || d.Name() == "tidb" {
		idType = "BIGINT UNSIGNED"
	}
	insertParams := "?,?,?,?,?"
	binaryParam := "[0,255]"
	binaryCheck := "r.payload.length!==2||r.payload[1]!==255"
	if d.Name() == "postgres" || d.Name() == "redshift" {
		binaryType = "BYTEA"
		insertParams = "$1,$2,$3,$4,$5"
	}
	if d.Name() == "redshift" {
		binaryType = "VARBYTE"
		insertParams = "$1,$2,$3,FROM_HEX($4),$5"
		binaryParam = "'00ff'"
		binaryCheck = "r.payload!=='00ff'"
	}
	// SQLite NUMERIC affinity can store decimals as floats. TEXT preserves exact
	// decimal input; the runtime rejects float-backed declared numeric columns.
	if d.Name() == "sqlite" {
		amountType = "TEXT"
	}
	if d.Name() == "oracle" {
		idType, amountType, profileType, binaryType, ratioType = "NUMBER(20)", "NUMBER(30,10)", "VARCHAR2(200)", "RAW(16)", "BINARY_DOUBLE"
		insertParams = ":1,:2,:3,:4,:5"
		// Oracle NUMBER preserves the value, without retaining trailing zero scale.
		amountExpected = "12345678901234567890.123456789"
	}
	if d.Name() == dialect.NameMSSQL {
		profileType, binaryType, ratioType = "NVARCHAR(200)", "VARBINARY(16)", "FLOAT"
		insertParams = "@p1,@p2,@p3,@p4,@p5"
	}
	create := fmt.Sprintf("CREATE TABLE %s (id %s PRIMARY KEY, amount %s, profile %s, payload %s, ratio %s);", table, idType, amountType, profileType, binaryType, ratioType)
	insert := fmt.Sprintf("INSERT INTO %s VALUES (%s)", table, insertParams)
	selectSQL := "SELECT id,amount,profile,payload,ratio FROM " + table
	if d.Name() == "redshift" {
		// pgx does not know Redshift-specific OIDs. Materialize VARBYTE as explicit
		// hex text instead of guessing the encoding of unknown driver values.
		selectSQL = "SELECT id,amount,profile,FROM_VARBYTE(payload,'hex') AS payload,ratio FROM " + table
	}
	if d.Name() == "oracle" {
		selectSQL = `SELECT id AS "id",amount AS "amount",profile AS "profile",payload AS "payload",ratio AS "ratio" FROM ` + table
	}
	update := "UPDATE " + table + " SET profile="
	if d.Name() == "postgres" || d.Name() == "redshift" {
		update += "$1 WHERE id=$2"
	} else if d.Name() == "oracle" {
		update += ":1 WHERE id=:2"
	} else if d.Name() == dialect.NameMSSQL {
		update += "@p1 WHERE id=@p2"
	} else {
		update += "? WHERE id=?"
	}
	mergeCheck := ""
	if d.Name() == dialect.NameMSSQL {
		merge := "WITH s AS (SELECT CAST(@p1 AS BIGINT) AS id, CAST(@p2 AS NVARCHAR(200)) AS profile) MERGE " + table + " AS t USING s ON t.id=s.id WHEN MATCHED THEN UPDATE SET profile=s.profile;"
		replaceUpdate := "UPDATE " + table + " SET profile=REPLACE(profile,@p1,@p2)"
		replaceQuery := "SELECT profile,REPLACE(@p1,@p2,@p3) AS value FROM " + table
		mergeCheck = `db.exec(` + strconv.Quote(replaceUpdate) + `,['old','transformed']);const transformed=db.query(` + strconv.Quote(replaceQuery) + `,['old','old','expected'])[0];if(transformed.profile!=='transformed'||transformed.value!=='expected')throw Error('REPLACE function');const merged=db.exec(` + strconv.Quote(merge) + `,['9007199254740993','old']);if(merged.rowsAffected!=='1')throw Error('MERGE count');`
		parent, view := table+"_parent", table+"_view"
		statements := []string{
			"CREATE OR ALTER VIEW " + view + " AS SELECT CAST(1 AS INT) AS id",
			"CREATE OR ALTER VIEW " + view + " AS SELECT CAST(2 AS INT) AS id",
			"DROP VIEW " + view,
			"ALTER TABLE " + table + " ADD email NVARCHAR(20) NULL",
			"ALTER TABLE " + table + " ALTER COLUMN email NVARCHAR(200) NULL",
			"ALTER TABLE " + table + " ADD CONSTRAINT " + table + "_email CHECK (email IS NULL)",
			"ALTER TABLE " + table + " DROP CONSTRAINT " + table + "_email",
			"ALTER TABLE " + table + " DROP COLUMN email",
			"CREATE TABLE " + parent + " (id BIGINT PRIMARY KEY)",
			"ALTER TABLE " + table + " ADD parent_id BIGINT NULL CONSTRAINT " + table + "_parent_default DEFAULT NULL",
		}
		for _, action := range []string{"ON DELETE CASCADE ON UPDATE NO ACTION", "ON DELETE NO ACTION ON UPDATE CASCADE", "ON DELETE SET NULL ON UPDATE SET NULL", "ON DELETE SET DEFAULT ON UPDATE SET DEFAULT"} {
			statements = append(statements,
				"ALTER TABLE "+table+" ADD CONSTRAINT "+table+"_fk FOREIGN KEY(parent_id) REFERENCES "+parent+"(id) "+action,
				"ALTER TABLE "+table+" DROP CONSTRAINT "+table+"_fk",
			)
		}
		statements = append(statements,
			"ALTER TABLE "+table+" DROP CONSTRAINT "+table+"_parent_default",
			"ALTER TABLE "+table+" DROP COLUMN parent_id",
			"DROP TABLE "+parent,
		)
		for _, statement := range statements {
			mergeCheck += `db.exec(` + strconv.Quote(statement) + `);`
		}
	}
	items := []source.Migration{
		{Filename: "V1__create.sql", Source: "-- +migrate Up\n" + create + "\n-- +migrate Down\nDROP TABLE " + table + ";\n"},
		{Filename: "V2__seed.mjs", Source: `export function up(db){const seeded=db.exec(` + strconv.Quote(insert) + `,['9007199254740993','12345678901234567890.1234567890','old',` + binaryParam + `,1.25]);if(db.dialect==='mssql'&&seeded.rowsAffected!=='1')throw Error('driver count');` + mergeCheck + `const r=db.query(` + strconv.Quote(selectSQL) + `)[0];if(r.id!=='9007199254740993'||r.amount!==` + strconv.Quote(amountExpected) + `||` + binaryCheck + `||r.ratio!==1.25)throw Error('value fidelity');db.exec(` + strconv.Quote(update) + `,['new',r.id]);}export function down(db){db.exec('DELETE FROM ` + table + `')}`},
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

	if d.Name() == dialect.NameMySQL || d.Name() == dialect.NameMariaDB {
		// Kill only this explicitly reserved test connection. The driver may hide
		// the underlying network error behind its own invalid-connection sentinel.
		func() {
			conn, err := db.Conn(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			p, err := script.Compile("V0__lost_connection.js", `export function up(db){const id=db.query('SELECT CONNECTION_ID() AS id')[0].id;if(!/^[0-9]+$/.test(id))throw Error('invalid connection id');try{db.exec('KILL CONNECTION '+id)}catch(e){}}`, "up", script.Options{})
			if err != nil {
				t.Fatal(err)
			}
			i, err := p.Instantiate(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if err = i.Execute(ctx, conn, d.Name()); err == nil {
				t.Fatal("caught native connection loss returned success")
			}
		}()
	}

	// Real-driver recovery must repair the database's transaction state explicitly.
	if d.Name() == dialect.NamePostgres || d.Name() == dialect.NameMSSQL {
		save, rollback, release := "SAVEPOINT script_recovery", "ROLLBACK TO SAVEPOINT script_recovery", "RELEASE SAVEPOINT script_recovery"
		if d.Name() == dialect.NameMSSQL {
			save, rollback, release = "SAVE TRANSACTION script_recovery", "ROLLBACK TRANSACTION script_recovery", ""
		}
		recoveredItems := append([]source.Migration(nil), items...)
		body := `export function up(db){db.exec(` + strconv.Quote(save) + `);let recovered=false;try{db.query('SELECT * FROM missing_script_recovery_table')}catch(e){db.exec(` + strconv.Quote(rollback) + `);recovered=true}if(!recovered)throw Error('expected SQL failure');`
		if release != "" {
			body += `db.exec(` + strconv.Quote(release) + `);`
		}
		body += `return db.exec(` + strconv.Quote(update) + `,['recovered','9007199254740993'])}export const down=db=>db.exec(` + strconv.Quote(update) + `,['new-ts','9007199254740993']);`
		recoveredItems = append(recoveredItems, source.Migration{Filename: "V4__recovery.js", Source: body})
		recoveryCfg := cfg
		recoveryCfg.MigrationSource = source.StringSource{Migrations: recoveredItems}
		recovery, err := migrate.NewService(ctx, recoveryCfg)
		if err != nil {
			t.Fatal(err)
		}
		if err = recovery.Up(); err != nil {
			t.Fatal(err)
		}
		assertScriptProfile(t, db, table, "recovered")
		statuses, err = recovery.Status()
		if err != nil || statuses[len(statuses)-1].Status != migrate.StatusApplied {
			t.Fatalf("%+v %v", statuses, err)
		}
		if err = recovery.Down("3", false); err != nil {
			t.Fatal(err)
		}
		assertScriptProfile(t, db, table, "new-ts")
	}

	if d.Name() == dialect.NamePostgres {
		cancelledItems := append([]source.Migration(nil), items...)
		body := `export const config={transaction:false};export function up(db){db.exec("SET statement_timeout='50ms'");db.exec('BEGIN');db.exec(` + strconv.Quote(update) + `,['temporary','9007199254740993']);db.exec('SAVEPOINT cancelled');try{db.query('SELECT pg_sleep(1)')}catch(e){try{db.exec('ROLLBACK TO SAVEPOINT cancelled')}catch(e){}try{db.exec(` + strconv.Quote(update) + `,['must-not-run','9007199254740993'])}catch(e){}try{db.exec('COMMIT')}catch(e){}}}`
		cancelledItems = append(cancelledItems, source.Migration{Filename: "V4__cancelled.js", Source: body})
		cancelledCfg := cfg
		cancelledCfg.MigrationSource = source.StringSource{Migrations: cancelledItems}
		cancelled, err := migrate.NewService(ctx, cancelledCfg)
		if err != nil {
			t.Fatal(err)
		}
		if err = cancelled.Up(); err == nil {
			t.Fatal("caught native cancellation recorded success")
		}
		assertScriptProfile(t, db, table, "new-ts")
		statuses, err = cancelled.Status()
		if err != nil || statuses[len(statuses)-1].Status != migrate.StatusPending {
			t.Fatalf("%+v %v", statuses, err)
		}
	}

	// Fail after a real update; an uncaught DB error must roll the update back.
	broken := append([]source.Migration(nil), items...)
	broken = append(broken, source.Migration{Filename: "V4__failure.js", Source: `export function up(db){db.exec(` + strconv.Quote(update) + `,['temporary','9007199254740993']);db.query('SELECT * FROM missing_script_table')}`})
	failedCfg := cfg
	failedCfg.MigrationSource = source.StringSource{Migrations: broken}
	failed, err := migrate.NewService(ctx, failedCfg)
	if err != nil {
		t.Fatal(err)
	}
	if err = failed.Up(); err == nil {
		t.Fatal("uncaught DB failure recorded success")
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
