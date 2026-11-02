//go:build full

package script

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const maxSafeInteger = int64(1<<53 - 1)

var decimalPattern = regexp.MustCompile(`^[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?$`)
var integerPattern = regexp.MustCompile(`^[+-]?[0-9]+$`)

func columnType(name string) string {
	name = strings.ToUpper(strings.TrimSpace(name))
	if index := strings.IndexByte(name, '('); index >= 0 {
		name = name[:index]
	}
	return strings.TrimSpace(strings.TrimSuffix(name, " UNSIGNED"))
}

func decimalType(name string) bool { return name == "DECIMAL" || name == "NUMERIC" || name == "NUMBER" }
func integerType(name string) bool {
	switch name {
	case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "INTEGER", "BIGINT", "INT2", "INT4", "INT8", "SERIAL", "BIGSERIAL":
		return true
	}
	return false
}
func binaryType(name string) bool {
	switch name {
	case "BLOB", "TINYBLOB", "MEDIUMBLOB", "LONGBLOB", "BINARY", "VARBINARY", "BYTEA":
		return true
	}
	return false
}
func textType(name string) bool {
	switch name {
	case "TEXT", "TINYTEXT", "MEDIUMTEXT", "LONGTEXT", "CHAR", "VARCHAR", "CHARACTER VARYING", "NCHAR", "NVARCHAR", "CLOB", "JSON", "JSONB", "UUID", "DATE", "TIME", "TIMESTAMP", "TIMESTAMPTZ", "DATETIME", "ENUM", "SET":
		return true
	}
	return false
}

// resultValue uses driver metadata for byte-backed text and exact numeric columns.
// Its byte count bounds returned data, not the driver's internal allocations.
func resultValue(value any, typeName string) (any, int64, error) {
	name := columnType(typeName)
	switch v := value.(type) {
	case nil:
		return nil, 1, nil
	case int64:
		if name == "BOOL" || name == "BOOLEAN" {
			if v != 0 && v != 1 {
				return nil, 0, errors.New("invalid boolean result")
			}
			return v == 1, 1, nil
		}
		out := strconv.FormatInt(v, 10)
		return out, int64(len(out)), nil
	case uint64:
		out := strconv.FormatUint(v, 10)
		return out, int64(len(out)), nil
	case bool:
		return v, 1, nil
	case float64:
		if decimalType(name) {
			return nil, 0, errors.New("driver returned an inexact decimal")
		}
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, 0, errors.New("non-finite result")
		}
		return v, 8, nil
	case time.Time:
		out := v.Format(time.RFC3339Nano)
		return out, int64(len(out)), nil
	case string:
		return resultText(v, name)
	case []byte:
		if binaryType(name) {
			out := append([]byte{}, v...)
			return out, int64(len(out)), nil
		}
		if decimalType(name) || integerType(name) || textType(name) || name == "BOOL" || name == "BOOLEAN" {
			return resultText(string(v), name)
		}
		return nil, 0, errors.New("ambiguous byte result; use an explicit SQL cast")
	default:
		return nil, 0, errors.New("unsupported result type")
	}
}

func resultText(value, name string) (any, int64, error) {
	if !utf8.ValidString(value) {
		return nil, 0, errors.New("invalid UTF-8 result")
	}
	if decimalType(name) && !decimalPattern.MatchString(value) {
		return nil, 0, errors.New("invalid decimal result")
	}
	if integerType(name) && !integerPattern.MatchString(value) {
		return nil, 0, errors.New("invalid integer result")
	}
	if name == "BOOL" || name == "BOOLEAN" {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, 0, errors.New("invalid boolean result")
		}
		return v, 1, nil
	}
	return value, int64(len(value)), nil
}

// parameterValue accepts only explicit SQL values. Objects and undefined are
// rejected by the VM bridge before this conversion; arrays mean binary bytes.
func parameterValue(value any) (any, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		if !utf8.ValidString(v) {
			return nil, errors.New("invalid UTF-8 parameter")
		}
		return v, nil
	case bool:
		return v, nil
	case int64:
		if v < -maxSafeInteger || v > maxSafeInteger {
			return nil, errors.New("unsafe integer parameter; pass a decimal string")
		}
		return v, nil
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("non-finite parameter")
		}
		if math.Trunc(v) == v && math.Abs(v) > float64(maxSafeInteger) {
			return nil, errors.New("unsafe integer parameter; pass a decimal string")
		}
		return v, nil
	case []byte:
		return append([]byte{}, v...), nil
	case []any:
		out := make([]byte, len(v))
		for index, item := range v {
			b, err := parameterValue(item)
			if err != nil {
				return nil, errors.New("invalid binary parameter")
			}
			switch n := b.(type) {
			case int64:
				if n < 0 || n > 255 {
					return nil, errors.New("invalid binary parameter")
				}
				out[index] = byte(n)
			case float64:
				if n < 0 || n > 255 || math.Trunc(n) != n {
					return nil, errors.New("invalid binary parameter")
				}
				out[index] = byte(n)
			default:
				return nil, errors.New("invalid binary parameter")
			}
		}
		return out, nil
	default:
		return nil, errors.New("unsupported parameter type")
	}
}
