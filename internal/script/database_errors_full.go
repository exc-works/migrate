//go:build full

package script

import (
	"errors"
	"strings"

	clickhouseproto "github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	"github.com/go-sql-driver/mysql"
	"github.com/sijms/go-ora/v3/network"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	postgresQueryCancelled    = "57014"
	mysqlQueryInterrupted     = 1317
	mysqlQueryTimeout         = 3024
	mariaDBStatementTimeout   = 1969
	oracleQueryCancelled      = 1013
	clickHouseTimeoutExceeded = 159
	clickHouseQueryCancelled  = 394
	clickHouseClientCancelled = 735
	sqlServerQueryTimeout     = -2
	sqlServerFatalSeverity    = 20
)

// Native servers need not wrap a Go context/network sentinel when cancelling a
// statement. Inspect structured codes before replacing the error with a public
// message. Ordinary SQL errors (including constraint violations) stay catchable.
func nativeDatabaseTermination(err error) bool {
	if err == nil {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, cause := range joined.Unwrap() {
			if nativeDatabaseTermination(cause) {
				return true
			}
		}
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok && nativeDatabaseTermination(wrapped.Unwrap()) {
		return true
	}
	var state interface{ SQLState() string }
	if errors.As(err, &state) {
		code := state.SQLState()
		if code == postgresQueryCancelled || strings.HasPrefix(code, "08") {
			return true
		}
		switch code {
		case "57P01", "57P02", "57P03", "57P04", "57P05":
			return true
		}
	}
	var my *mysql.MySQLError
	if errors.As(err, &my) {
		switch my.Number {
		case mysqlQueryInterrupted, mysqlQueryTimeout, mariaDBStatementTimeout:
			return true
		}
	}
	var ora *network.OracleError
	if errors.As(err, &ora) && ora.ErrCode == oracleQueryCancelled {
		return true
	}
	var server interface {
		SQLErrorNumber() int32
		SQLErrorClass() uint8
	}
	if errors.As(err, &server) && (server.SQLErrorNumber() == sqlServerQueryTimeout || server.SQLErrorClass() >= sqlServerFatalSeverity) {
		return true
	}
	var ch *clickhouseproto.Exception
	if errors.As(err, &ch) {
		switch ch.Code {
		case clickHouseTimeoutExceeded, clickHouseQueryCancelled, clickHouseClientCancelled:
			return true
		}
		for i := range ch.Nested {
			if nativeDatabaseTermination(&ch.Nested[i]) {
				return true
			}
		}
	}
	var sq *sqlite.Error
	return errors.As(err, &sq) && sq.Code()&0xff == sqlite3.SQLITE_INTERRUPT
}

// database/sql can report either a driver Next error or an automatic EOF-close
// error via Rows.Err. Only structured native SQL exceptions are safe to classify
// as statement failures there; an unclassified stream/cleanup error stays fatal.
func nativeSQLStatementError(err error) bool {
	if err == nil {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		causes := joined.Unwrap()
		if len(causes) == 0 {
			return false
		}
		for _, cause := range causes {
			if !nativeSQLStatementError(cause) {
				return false
			}
		}
		return true
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return nativeSQLStatementError(wrapped.Unwrap())
	}
	if state, ok := err.(interface{ SQLState() string }); ok {
		return len(state.SQLState()) == 5
	}
	switch err.(type) {
	case *mysql.MySQLError, *network.OracleError, *clickhouseproto.Exception, *sqlite.Error:
		return true
	}
	_, ok := err.(interface {
		SQLErrorNumber() int32
		SQLErrorClass() uint8
	})
	return ok
}
