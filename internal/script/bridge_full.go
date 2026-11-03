//go:build full

package script

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	dbdialect "github.com/exc-works/migrate/internal/dialect"
)

// Execute binds only a narrow, synchronous database capability to the entrypoint.
func (i *Invocation) Execute(ctx context.Context, db Database, dialect string) error {
	capability, supported := Capability(dialect)
	if !supported {
		return errors.New("script dialect is unsupported")
	}
	dialect = capability.Name
	if db == nil {
		return errors.New("script database is required")
	}
	var failure error
	remember := func(err error) error {
		if failure == nil {
			failure = err
		}
		return failure
	}
	err := i.Run(ctx, func(runCtx context.Context) any {
		return map[string]any{
			"dialect": dialect,
			"exec": func(query, params goja.Value) (any, error) {
				if failure != nil {
					return nil, failure
				}
				text, args, err := i.decodeCall(query, params, dialect)
				if err != nil {
					return nil, remember(err)
				}
				if err = runCtx.Err(); err != nil {
					return nil, remember(err)
				}
				result, err := db.ExecContext(runCtx, text, args...)
				if err != nil {
					return nil, remember(errors.New("database execution failed"))
				}
				count, err := result.RowsAffected()
				if err != nil {
					return nil, remember(errors.New("database affected-row count unavailable"))
				}
				return map[string]any{"rowsAffected": strconv.FormatInt(count, 10)}, nil
			},
			"query": func(query, params goja.Value) (any, error) {
				if failure != nil {
					return nil, failure
				}
				text, args, err := i.decodeCall(query, params, dialect)
				if err != nil {
					return nil, remember(err)
				}
				if err = runCtx.Err(); err != nil {
					return nil, remember(err)
				}
				rows, err := readRows(runCtx, db, text, args, i.program.options, dialect)
				if err != nil {
					return nil, remember(err)
				}
				return rows, nil
			},
		}
	})
	return errors.Join(err, failure)
}

// Accessors and Export can throw inside reflected Go callbacks. Convert those
// exceptions into ordinary errors so callers always record a sticky failure.
func (i *Invocation) decodeCall(query, params goja.Value, dialectName string) (text string, args []any, err error) {
	if exception := i.vm.Try(func() { text, args, err = decodeCall(query, params) }); exception != nil {
		return "", nil, errors.New("SQL parameter access failed")
	}
	if err == nil && dialectName == dbdialect.NameMSSQL {
		text, err = prepareMSSQL(text)
	}
	return
}

func decodeCall(query, params goja.Value) (string, []any, error) {
	if query == nil {
		return "", nil, errors.New("SQL must be a string")
	}
	text, ok := query.Export().(string)
	if !ok {
		return "", nil, errors.New("SQL must be a string")
	}
	text, err := singleStatement(text)
	if err != nil {
		return "", nil, err
	}
	if params == nil || goja.IsUndefined(params) {
		return text, nil, nil
	}
	array, ok := params.(*goja.Object)
	if !ok || array.ClassName() != "Array" {
		return "", nil, errors.New("SQL parameters must be an array")
	}
	count := array.Get("length").ToInteger()
	if count > 65535 {
		return "", nil, errors.New("too many SQL parameters")
	}
	args := make([]any, int(count))
	for index := range args {
		v := array.Get(strconv.Itoa(index))
		if v == nil || goja.IsUndefined(v) {
			return "", nil, errors.New("undefined SQL parameter")
		}
		args[index], err = parameterValue(v.Export())
		if err != nil {
			return "", nil, err
		}
	}
	return text, args, nil
}

// The bridge deliberately accepts no embedded semicolons. A trailing terminator
// is optional; use bound parameters for values containing semicolons. SQL files
// remain the escape hatch for compound SQL and stored routine definitions.
func singleStatement(query string) (string, error) {
	query = strings.TrimSpace(query)
	query = strings.TrimSpace(strings.TrimSuffix(query, ";"))
	if query == "" || strings.Contains(query, ";") || strings.Contains(query, "/*!") || strings.Contains(query, "/*M!") || strings.Contains(query, "/*m!") {
		return "", errors.New("expected one SQL statement; bind values containing semicolons")
	}
	prefix := query
	// Do not interpret SQL comments: newline and nesting rules differ by engine.
	// A directly leading keyword is unambiguous across all supported dialects.
	if strings.HasPrefix(prefix, "--") || strings.HasPrefix(prefix, "#") || strings.HasPrefix(prefix, "/*") {
		return "", errors.New("leading SQL comments are unsupported; place comments in the script")
	}
	n := 0
	for n < len(prefix) && (prefix[n] >= 'a' && prefix[n] <= 'z' || prefix[n] >= 'A' && prefix[n] <= 'Z') {
		n++
	}
	if n == 0 {
		return "", errors.New("expected SQL statement")
	}
	// Use a positive set: aliases and dynamic execution must not be forwarded
	// merely because they were absent from a transaction-command denylist.
	switch strings.ToUpper(prefix[:n]) {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "WITH", "VALUES", "TABLE",
		"REPLACE", "MERGE", "CREATE", "ALTER", "DROP", "TRUNCATE", "COMMENT",
		"REFRESH", "REINDEX", "EXPLAIN", "SHOW", "DESCRIBE", "DESC":
		return query, nil
	default:
		return "", errors.New("manual transaction, session control or dynamic SQL is unsupported; use a direct data/schema statement")
	}
}

func readRows(ctx context.Context, db Database, query string, args []any, options Options, dialectName string) (out []map[string]any, err error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.New("database query failed")
	}
	defer func() {
		if rows.Close() != nil {
			err = errors.Join(err, errors.New("database result close failed"))
		}
	}()
	columns, err := rows.Columns()
	if err != nil {
		return nil, errors.New("database columns unavailable")
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.New("database column types unavailable")
	}
	seen := make(map[string]bool, len(columns))
	for _, name := range columns {
		if seen[name] {
			return nil, errors.New("duplicate result columns; use aliases")
		}
		seen[name] = true
	}
	out = make([]map[string]any, 0)
	var size int64
	for rows.Next() {
		if len(out) >= options.MaxRows {
			return nil, errors.New("query row limit exceeded")
		}
		values := make([]any, len(columns))
		targets := make([]any, len(columns))
		for index := range values {
			targets[index] = &values[index]
		}
		if err = rows.Scan(targets...); err != nil {
			return nil, errors.New("database scan failed")
		}
		row := make(map[string]any, len(columns))
		for index, name := range columns {
			var n int64
			row[name], n, err = resultValue(values[index], resultColumnType(types[index], dialectName))
			if err != nil {
				return nil, fmt.Errorf("result column %d: %w", index+1, err)
			}
			size += n + int64(len(name))
			if size > options.MaxBytes {
				return nil, errors.New("query byte limit exceeded")
			}
		}
		out = append(out, row)
	}
	if rows.Err() != nil {
		return nil, errors.New("database iteration failed")
	}
	return out, nil
}

// go-ora v3.0.1 omits database type names. Its documented scan metadata still
// distinguishes Oracle binary/text values, and numeric precision identifies NUMBER.
// Keep this fallback scoped to Oracle; other ambiguous byte results still fail.
func resultColumnType(column *sql.ColumnType, dialectName string) string {
	name := column.DatabaseTypeName()
	if dialectName == dbdialect.NameMSSQL {
		switch strings.ToUpper(name) {
		case "TIMESTAMP", "ROWVERSION":
			return "BINARY"
		case "MONEY", "SMALLMONEY":
			return "DECIMAL"
		}
	}
	if name != "" || dialectName != dbdialect.NameOracle {
		return name
	}
	if _, _, numeric := column.DecimalSize(); numeric {
		return "NUMBER"
	}
	switch column.ScanType() {
	case reflect.TypeFor[[]byte]():
		return "RAW"
	case reflect.TypeFor[string]():
		return "VARCHAR2"
	}
	return name
}
