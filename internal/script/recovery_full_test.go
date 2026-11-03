//go:build full

package script

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type recoveryDB struct {
	calls   int
	failure error
}

func (d *recoveryDB) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	d.calls++
	if d.calls == 1 && d.failure != nil {
		return nil, d.failure
	}
	return sqlmock.NewResult(0, 1), nil
}
func (d *recoveryDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	d.calls++
	return nil, d.failure
}
func runRecovery(t *testing.T, db Database, body string) error {
	t.Helper()
	p, err := Compile("V1__recovery.js", body, "up", Options{})
	if err != nil {
		t.Fatal(err)
	}
	i, err := p.Instantiate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return i.Execute(context.Background(), db, "sqlite")
}

func TestCaughtSQLFailuresAllowRecovery(t *testing.T) {
	for _, method := range []string{"exec", "query"} {
		t.Run(method, func(t *testing.T) {
			db := &recoveryDB{failure: errors.New("private driver detail")}
			err := runRecovery(t, db, `export function up(db){try{db.`+method+`('failing statement')}catch(e){if(String(e).includes('private driver detail'))throw Error('leak');db.exec('ROLLBACK')}}`)
			if err != nil || db.calls != 2 {
				t.Fatalf("calls=%d error=%v", db.calls, err)
			}
		})
	}
}
func TestConnectionAndCancellationFailuresRemainFatal(t *testing.T) {
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded, driver.ErrBadConn, sql.ErrConnDone, sql.ErrTxDone, io.EOF, io.ErrUnexpectedEOF, net.ErrClosed, &net.OpError{Op: "read", Net: "tcp", Err: errors.New("private driver detail")}} {
		for _, method := range []string{"exec", "query"} {
			t.Run(method+"/"+cause.Error(), func(t *testing.T) {
				db := &recoveryDB{failure: fmt.Errorf("private driver detail: %w", cause)}
				err := runRecovery(t, db, `export function up(db){try{db.`+method+`('first')}catch(e){}try{db.exec('must not execute')}catch(e){}}`)
				if err == nil || db.calls != 1 || strings.Contains(err.Error(), "private driver detail") {
					t.Fatalf("calls=%d error=%v", db.calls, err)
				}
			})
		}
	}
}
func TestUncaughtSQLFailureStillAborts(t *testing.T) {
	db := &recoveryDB{failure: errors.New("private driver detail")}
	err := runRecovery(t, db, `export function up(db){db.exec('failing statement');db.exec('must not execute')}`)
	if err == nil || db.calls != 1 || strings.Contains(err.Error(), "private driver detail") {
		t.Fatalf("calls=%d error=%v", db.calls, err)
	}
}

func TestSynchronousReturnValuesAreIgnored(t *testing.T) {
	for _, body := range []string{
		`export function up(){return 1}`,
		`export function up(){return null}`,
		`export const up=db=>db.exec('UPDATE fixture SET value=1')`,
		`export const up=()=>({get value(){throw Error('must not inspect')},get then(){throw Error('must not inspect')}})`,
		`export function up(){const value={};value.self=value;return value}`,
		`export function up(){return new Proxy({},{get(){throw Error('must not inspect')}})}`,
		`export function up(){const p=Proxy.revocable({},{});p.revoke();return p.proxy}`,
		`export function up(){return function*(){yield 1}}`,
	} {
		t.Run(body, func(t *testing.T) {
			if err := runRecovery(t, &recoveryDB{}, body); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCaughtIterationErrorAllowsRecovery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT fixture").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(1).RowError(0, errors.New("private driver detail"))).RowsWillBeClosed()
	mock.ExpectExec("ROLLBACK").WillReturnResult(sqlmock.NewResult(0, 0))
	err = runRecovery(t, db, `export function up(db){try{db.query('SELECT fixture')}catch(e){db.exec('ROLLBACK')}}`)
	if err != nil {
		t.Fatal(err)
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
