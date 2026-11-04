package script

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	dbdialect "github.com/exc-works/migrate/internal/dialect"
	_ "modernc.org/sqlite"
)

func bridgeDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "bridge.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	if _, err = db.Exec(`CREATE TABLE users(id INTEGER PRIMARY KEY,profile TEXT); INSERT INTO users VALUES(9007199254740993,'{"name":"Ada"}')`); err != nil {
		t.Fatal(err)
	}
	return db
}

func bridgeRun(t *testing.T, db *sql.DB, source string, options Options) error {
	t.Helper()
	p, err := Compile("V1__bridge.js", source, "up", options)
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = i.Execute(context.Background(), tx, "sqlite"); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}

func TestBridgeQueryTransformWrite(t *testing.T) {
	db := bridgeDB(t)
	err := bridgeRun(t, db, `export function up(db){const rows=db.query('SELECT id,profile FROM users');if(rows[0].id!=='9007199254740993')throw Error('precision');const p=JSON.parse(rows[0].profile);p.displayName=p.name;delete p.name;const result=db.exec('UPDATE users SET profile=? WHERE id=?',[JSON.stringify(p),rows[0].id]);if(result.rowsAffected!=='1')throw Error('count');}`, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var id, p string
	if err = db.QueryRow("SELECT id,profile FROM users").Scan(&id, &p); err != nil || id != "9007199254740993" || p != `{"displayName":"Ada"}` {
		t.Fatalf("%s %s %v", id, p, err)
	}
}

func TestFatalBridgeFailuresRollback(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		options    Options
	}{
		{name: "uncaught database", body: `db.query('SELECT * FROM missing');db.exec('UPDATE users SET profile=?',['must not execute'])`},
		{name: "caught unsafe number", body: `try{db.exec('UPDATE users SET id=?',[9007199254740993])}catch(e){}`},
		{name: "caught row cap", body: `try{db.query('SELECT id FROM users UNION ALL SELECT id FROM users')}catch(e){}`, options: Options{MaxRows: 1}},
		{name: "caught byte cap", body: `try{db.query('SELECT profile FROM users')}catch(e){}`, options: Options{MaxBytes: 2}},
		{name: "unsafe number", body: `db.exec('UPDATE users SET id=?',[9007199254740993])`},
		{name: "undefined", body: `db.exec('UPDATE users SET profile=?',[undefined])`},
		{name: "object", body: `db.exec('UPDATE users SET profile=?',[{}])`},
		{name: "throwing parameter", body: `const p=[];Object.defineProperty(p,0,{get(){throw Error("private value")}});try{db.exec('SELECT ?',p)}catch(e){}`},
		{name: "row cap", body: `db.query('SELECT id FROM users UNION ALL SELECT id FROM users')`, options: Options{MaxRows: 1}},
		{name: "byte cap", body: `db.query('SELECT profile FROM users')`, options: Options{MaxBytes: 2}},
		{name: "duplicate columns", body: `db.query('SELECT id,id FROM users')`},
		{name: "timeout", body: `db.query('WITH RECURSIVE x(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM x WHERE n<100000000) SELECT SUM(n) FROM x')`, options: Options{Timeout: 25 * time.Millisecond}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := bridgeDB(t)
			err := bridgeRun(t, db, `export function up(db){db.exec('UPDATE users SET profile=?',['temporary']);`+tc.body+`}`, tc.options)
			if err == nil {
				t.Fatal("expected failure")
			}
			var p string
			if e := db.QueryRow("SELECT profile FROM users").Scan(&p); e != nil || p != `{"name":"Ada"}` {
				t.Fatalf("partial write: %q %v", p, e)
			}
		})
	}
}

func TestBridgeExactMetadataAndCloseFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows := sqlmock.NewRowsWithColumnDefinition(sqlmock.NewColumn("amount").OfType("DECIMAL", "")).AddRow([]byte("12345678901234567890.1234567890"))
	mock.ExpectQuery("SELECT amount").WillReturnRows(rows).RowsWillBeClosed()
	p, err := Compile("V1__decimal.js", `export function up(db){const r=db.query('SELECT amount')[0];if(r.amount!=='12345678901234567890.1234567890')throw Error('precision')}`, "up", Options{})
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = i.Execute(context.Background(), db, "postgres"); err != nil {
		t.Fatal(err)
	}
	rows = sqlmock.NewRows([]string{"id"}).AddRow(1).CloseError(errors.New("private driver detail"))
	mock.ExpectQuery("SELECT id").WillReturnRows(rows).RowsWillBeClosed()
	p, err = Compile("V1__close.js", `export function up(db){db.query('SELECT id')}`, "up", Options{})
	if err != nil {
		t.Fatal(err)
	}
	i, err = p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = i.Execute(context.Background(), db, "postgres"); err == nil || strings.Contains(err.Error(), "private driver detail") {
		t.Fatalf("%v", err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestBridgeSQLPassesUnchanged(t *testing.T) {
	for _, dialect := range dbdialect.Names() {
		for _, query := range []string{"-- native SQL\nSELECT ';' AS value;", "SELECT 1; SELECT 2;", "DECLARE @value INT; SET @value=1; SELECT @value", "EXEC native_procedure", "BEGIN; COMMIT;", "/*! native comment */ SELECT 1"} {
			t.Run(dialect+"/"+query, func(t *testing.T) {
				db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				mock.ExpectExec(query).WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("native"))
				p, err := Compile("V1__native.js", `export function up(db){db.exec(`+strconv.Quote(query)+`);db.query(`+strconv.Quote(query)+`)}`, "up", Options{})
				if err != nil {
					t.Fatal(err)
				}
				i, err := p.Instantiate(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if err = i.Execute(context.Background(), db, dialect); err != nil {
					t.Fatal(err)
				}
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestBridgeEmptySQLFailsBeforeDriver(t *testing.T) {
	db := bridgeDB(t)
	for _, text := range []string{"", " \n\t"} {
		if err := bridgeRun(t, db, `export function up(db){try{db.exec(`+strconv.Quote(text)+`)}catch(e){}}`, Options{}); err == nil {
			t.Fatal("empty SQL escaped sticky failure")
		}
	}
}

func TestOracleMissingTypeNamesUseColumnMetadata(t *testing.T) {
	for _, tc := range []struct {
		dialect      string
		numericFloat bool
		bad          bool
	}{
		{"oracle", false, false}, {"postgres", false, true}, {"oracle", true, true},
	} {
		t.Run(fmt.Sprintf("%s/float=%t", tc.dialect, tc.numericFloat), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			amount := any("1234567890123456789012345678.1234567890")
			if tc.numericFloat {
				amount = float64(0.1)
			}
			rows := sqlmock.NewRowsWithColumnDefinition(sqlmock.NewColumn("payload").OfType("", []byte{}), sqlmock.NewColumn("amount").OfType("", float64(0)).WithPrecisionAndScale(38, 10)).AddRow([]byte{0, 255}, amount)
			mock.ExpectQuery("SELECT payload,amount").WillReturnRows(rows).RowsWillBeClosed()
			p, err := Compile("V1__oracle.js", `export function up(db){const r=db.query('SELECT payload,amount')[0];if(r.payload.length!==2||r.payload[1]!==255||r.amount!=='1234567890123456789012345678.1234567890')throw Error('fidelity')}`, "up", Options{})
			if err != nil {
				t.Fatal(err)
			}
			i, err := p.Instantiate(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			err = i.Execute(context.Background(), db, tc.dialect)
			if (err != nil) != tc.bad {
				t.Fatalf("%v", err)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestClickHouseAffectedRowsUseDriverResult(t *testing.T) {
	for _, tc := range []struct {
		name        string
		count       int64
		unavailable bool
	}{
		{name: "zero", count: 0}, {name: "count", count: 17}, {name: "large", count: 9007199254740993}, {name: "unavailable", unavailable: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			result := sqlmock.NewResult(0, tc.count)
			body := fmt.Sprintf(`if(db.exec('INSERT INTO users VALUES (?)',['exact']).rowsAffected!==%q)throw Error('count')`, strconv.FormatInt(tc.count, 10))
			if tc.unavailable {
				result = sqlmock.NewErrorResult(errors.New("private count failure"))
				body = `try{db.exec('INSERT INTO users VALUES (?)',['exact'])}catch(e){}db.exec('INSERT INTO users VALUES (?)',['must not run'])`
			}
			mock.ExpectExec("INSERT INTO users").WithArgs("exact").WillReturnResult(result)
			p, err := Compile("V1__clickhouse.js", `export function up(db){`+body+`}`, "up", Options{})
			if err != nil {
				t.Fatal(err)
			}
			i, err := p.Instantiate(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			err = i.Execute(context.Background(), db, "clickhouse")
			if (err != nil) != tc.unavailable || err != nil && strings.Contains(err.Error(), "private count failure") {
				t.Fatalf("%v", err)
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestBridgeDialectAliasesUseCanonicalNameAndNativeParameters(t *testing.T) {
	for _, canonical := range dbdialect.Names() {
		for _, input := range append([]string{canonical}, dbdialect.Aliases(canonical)...) {
			t.Run(input, func(t *testing.T) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()
				placeholder := "?"
				switch canonical {
				case dbdialect.NamePostgres, dbdialect.NameRedshift:
					placeholder = "$1"
				case dbdialect.NameMSSQL:
					placeholder = "@p1"
				case dbdialect.NameOracle:
					placeholder = ":1"
				}
				query := "SELECT " + placeholder
				mock.ExpectExec(regexp.QuoteMeta(query)).WithArgs("9007199254740993").WillReturnResult(sqlmock.NewResult(0, 1))
				source := fmt.Sprintf(`export function up(db){if(db.dialect!==%q)throw Error('alias');if(db.exec(%q,['9007199254740993']).rowsAffected!=='1')throw Error('count')}`, canonical, query)
				p, err := Compile("V1__alias.js", source, "up", Options{})
				if err != nil {
					t.Fatal(err)
				}
				i, err := p.Instantiate(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if err = i.Execute(context.Background(), db, " "+strings.ToUpper(input)+" "); err != nil {
					t.Fatal(err)
				}
				if err = mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestMSSQLMetadataConversionsAreDialectScoped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, tc := range []struct {
		dialect, dbType string
		value           any
		check           string
	}{
		{"sqlserver", "TIMESTAMP", []byte{0, 255}, "r.value.length===2&&r.value[1]===255"},
		{"azuresql", "ROWVERSION", []byte{0, 255}, "r.value.length===2&&r.value[1]===255"},
		{"postgres", "TIMESTAMP", []byte("2026-11-03 01:00:00"), "r.value==='2026-11-03 01:00:00'"},
		{"mssql", "MONEY", []byte("922337203685477.5807"), "r.value==='922337203685477.5807'"},
		{"sqlserver", "SMALLMONEY", []byte("-214748.3648"), "r.value==='-214748.3648'"},
		{"postgres", "MONEY", "$1.25", "r.value==='$1.25'"},
	} {
		mock.ExpectQuery("SELECT value").WillReturnRows(sqlmock.NewRowsWithColumnDefinition(sqlmock.NewColumn("value").OfType(tc.dbType, tc.value)).AddRow(tc.value)).RowsWillBeClosed()
		p, err := Compile("V1__timestamp.js", `export function up(db){const r=db.query('SELECT value')[0];if(!(`+tc.check+`))throw Error('conversion')}`, "up", Options{})
		if err != nil {
			t.Fatal(err)
		}
		i, err := p.Instantiate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if err = i.Execute(context.Background(), db, tc.dialect); err != nil {
			t.Fatal(err)
		}
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMSSQLMoneyRejectsInexactDriverValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, name := range []string{"MONEY", "SMALLMONEY"} {
		mock.ExpectQuery("SELECT amount").WillReturnRows(sqlmock.NewRowsWithColumnDefinition(sqlmock.NewColumn("amount").OfType(name, float64(0))).AddRow(float64(0.1))).RowsWillBeClosed()
		p, err := Compile("V1__money.js", `export function up(db){try{db.query('SELECT amount')}catch(e){}}`, "up", Options{})
		if err != nil {
			t.Fatal(err)
		}
		i, err := p.Instantiate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if err = i.Execute(context.Background(), db, "azuresql"); err == nil {
			t.Fatalf("accepted inexact %s", name)
		}
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
