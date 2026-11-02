package script

import (
	"context"
	"database/sql"
)

// Database is implemented by sql.Tx; the runtime never exposes it directly to JS.
type Database interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
