//go:build full

package script

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestBridgeFailuresAreStickyAndRollback(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		options    Options
	}{
		{name: "database", body: `try{db.query('SELECT * FROM missing')}catch(e){} db.exec('UPDATE users SET profile=?',['changed'])`},
		{name: "unsafe number", body: `db.exec('UPDATE users SET id=?',[9007199254740993])`},
		{name: "undefined", body: `db.exec('UPDATE users SET profile=?',[undefined])`},
		{name: "object", body: `db.exec('UPDATE users SET profile=?',[{}])`},
		{name: "abort alias", body: `db.exec('ABORT')`},
		{name: "SQLite comment escape", body: `db.exec('/* outer /* inner */ COMMIT -- */ SELECT 1');throw Error("after control")`},
		{name: "nested comment commit", body: `db.exec('/* outer /* inner */ ignored */ COMMIT')`},
		{name: "prepare transaction", body: `db.exec("PREPARE TRANSACTION 'example'")`},
		{name: "throwing parameter", body: `const p=[];Object.defineProperty(p,0,{get(){throw Error("private value")}});try{db.exec('SELECT ?',p)}catch(e){}`},
		{name: "manual commit", body: `db.exec(' /* leading */ COMMIT')`},
		{name: "multiple statements", body: `db.exec('SELECT 1; COMMIT')`},
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

func TestSingleStatementControl(t *testing.T) {
	for _, q := range []string{"COMMIT", "-- comment\nROLLBACK", "/* comment */ BEGIN", "SET autocommit=1", "SELECT 1; COMMIT", "/*! COMMIT */", "SELECT ';'"} {
		if _, err := singleStatement(q); err == nil {
			t.Fatalf("accepted %q", q)
		}
	}
	for _, q := range []string{"SELECT 1;", "/* comment */ SELECT 1", "UPDATE users SET profile=?"} {
		if _, err := singleStatement(q); err != nil {
			t.Fatalf("%q: %v", q, err)
		}
	}
}

func TestControlCommandsRejectedBeforeDriver(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, query := range []string{"ABORT", "PREPARE TRANSACTION 'example'", "EXECUTE IMMEDIATE 'COMMIT'", "CALL commit_procedure()", "RESET QUERY CACHE", "/* outer /* inner */ ignored */ COMMIT", "/* outer /* inner */ COMMIT -- */ SELECT 1"} {
		source := `export function up(db){db.exec(` + strconv.Quote(query) + `)}`
		p, err := Compile("V1__control.js", source, "up", Options{})
		if err != nil {
			t.Fatal(err)
		}
		i, err := p.Instantiate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		err = i.Execute(context.Background(), db, "postgres")
		want := "manual transaction"
		if strings.Contains(query, "/*") {
			want = "nested SQL comments"
		}
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("%q: %v", query, err)
		}
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
