//go:build full

package script

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestResultValueFidelity(t *testing.T) {
	instant := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.FixedZone("offset", 8*60*60))
	for _, tc := range []struct {
		name, dbType string
		input, want  any
		bad          bool
	}{
		{name: "NULL", input: nil, want: nil},
		{name: "small integer", input: int64(1), want: "1"},
		{name: "big integer", dbType: "BIGINT", input: int64(9007199254740993), want: "9007199254740993"},
		{name: "unsigned integer", dbType: "BIGINT UNSIGNED", input: uint64(math.MaxUint64), want: "18446744073709551615"},
		{name: "MySQL unsigned bytes", dbType: "UNSIGNED BIGINT", input: []byte("18446744073709551615"), want: "18446744073709551615"},
		{name: "MySQL float bytes", dbType: "DOUBLE", input: []byte("1.25"), want: float64(1.25)},
		{name: "Postgres float text", dbType: "DOUBLE PRECISION", input: "1.25", want: float64(1.25)},
		{name: "non-finite float text", dbType: "FLOAT", input: []byte("NaN"), bad: true},
		{name: "decimal", dbType: "DECIMAL(30,10)", input: []byte("12345678901234567890.1234567890"), want: "12345678901234567890.1234567890"},
		{name: "decimal exponent", dbType: "NUMERIC", input: "1.23e-20", want: "1.23e-20"},
		{name: "integer float", dbType: "BIGINT", input: float64(9007199254740992), bad: true},
		{name: "fraction in integer column", dbType: "BIGINT", input: float64(1.25), bad: true},
		{name: "decimal float", dbType: "NUMERIC", input: float64(0.1), bad: true},
		{name: "invalid decimal", dbType: "DECIMAL", input: "not numeric", bad: true},
		{name: "invalid integer", dbType: "INT", input: []byte("1.2"), bad: true},
		{name: "boolean", dbType: "BOOL", input: []byte("t"), want: true},
		{name: "boolean integer", dbType: "BOOLEAN", input: int64(0), want: false},
		{name: "invalid boolean", dbType: "BOOLEAN", input: int64(2), bad: true},
		{name: "text bytes", dbType: "VARCHAR", input: []byte("hello"), want: "hello"},
		{name: "json", dbType: "JSONB", input: []byte(`{"a":1}`), want: `{"a":1}`},
		{name: "binary", dbType: "BYTEA", input: []byte{0, 255}, want: []byte{0, 255}},
		{name: "binary string", dbType: "BLOB", input: string([]byte{0, 255}), want: []byte{0, 255}},
		{name: "empty binary", dbType: "BLOB", input: []byte{}, want: []byte{}},
		{name: "unknown bytes", input: []byte("ambiguous"), bad: true},
		{name: "invalid text", dbType: "TEXT", input: []byte{255}, bad: true},
		{name: "date", input: instant, want: "2026-01-01T19:04:05.123456789Z"},
		{name: "float", input: float64(1.5), want: float64(1.5)},
		{name: "NaN", input: math.NaN(), bad: true},
		{name: "infinite", input: math.Inf(1), bad: true},
		{name: "unknown object", input: struct{}{}, bad: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, n, err := resultValue(tc.input, tc.dbType)
			if (err != nil) != tc.bad {
				t.Fatalf("%v", err)
			}
			if !tc.bad && !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
			if !tc.bad && n < 0 {
				t.Fatal("negative byte count")
			}
		})
	}
}

func TestParameterValueFidelity(t *testing.T) {
	for _, tc := range []struct {
		name        string
		input, want any
		bad         bool
	}{
		{name: "null"},
		{name: "exact decimal", input: "12345678901234567890.1234567890", want: "12345678901234567890.1234567890"},
		{name: "integer", input: int64(17), want: int64(17)},
		{name: "largest safe", input: maxSafeInteger, want: maxSafeInteger},
		{name: "unsafe", input: int64(9007199254740993), bad: true},
		{name: "rounded unsafe", input: float64(9007199254740993), bad: true},
		{name: "unsafe negative", input: int64(-9007199254740993), bad: true},
		{name: "floating", input: float64(1.25), want: float64(1.25)},
		{name: "NaN", input: math.NaN(), bad: true},
		{name: "infinite", input: math.Inf(-1), bad: true},
		{name: "boolean", input: true, want: true},
		{name: "binary array", input: []any{int64(0), float64(255)}, want: []byte{0, 255}},
		{name: "invalid binary", input: []any{float64(1.2)}, bad: true},
		{name: "negative binary", input: []any{int64(-1)}, bad: true},
		{name: "large binary", input: []any{int64(256)}, bad: true},
		{name: "object", input: map[string]any{"a": 1}, bad: true},
		{name: "nested array", input: []any{[]any{int64(1)}}, bad: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parameterValue(tc.input)
			if (err != nil) != tc.bad {
				t.Fatalf("%v", err)
			}
			if !tc.bad && !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestBinaryValuesAreCopied(t *testing.T) {
	original := []byte{1, 2}
	v, _, err := resultValue(original, "BLOB")
	if err != nil {
		t.Fatal(err)
	}
	v.([]byte)[0] = 9
	if original[0] != 1 {
		t.Fatal("result aliases driver buffer")
	}
	v, err = parameterValue(original)
	if err != nil {
		t.Fatal(err)
	}
	v.([]byte)[0] = 9
	if original[0] != 1 {
		t.Fatal("parameter aliases caller buffer")
	}
}

func TestTimestampHistoricalOffsetPreservesInstant(t *testing.T) {
	want := time.Date(1900, 1, 1, 12, 0, 0, 123, time.FixedZone("historical", 9*60+21))
	value, _, err := resultValue(want, "TIMESTAMPTZ")
	if err != nil {
		t.Fatal(err)
	}
	got, err := time.Parse(time.RFC3339Nano, value.(string))
	if err != nil || !got.Equal(want) {
		t.Fatalf("instant changed: %s %v", value, err)
	}
}

func TestOracleResultTypes(t *testing.T) {
	for _, tc := range []struct {
		name        string
		value, want any
	}{
		{"NUMBER", "123456789012345678901234567890.12345678", "123456789012345678901234567890.12345678"},
		{"VARCHAR2", []byte("text"), "text"},
		{"NVARCHAR2", []byte("中文"), "中文"},
		{"OCICLOBLOCATOR", []byte("large text"), "large text"},
		{"OCIBLOBLOCATOR", []byte{0, 255}, []byte{0, 255}},
		{"RAW", []byte{0, 255}, []byte{0, 255}},
		{"LONGRAW", []byte{0, 255}, []byte{0, 255}},
		{"IBFLOAT", float32(1.25), float64(1.25)},
		{"IBDOUBLE", float64(1.25), float64(1.25)},
		{"BINARY_FLOAT", []byte("1.25"), float64(1.25)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, _, err := resultValue(tc.value, tc.name)
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, error=%v; want %#v", got, err, tc.want)
			}
		})
	}
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"NUMBER", float64(0.1)}, {"NUMBER", float32(0.1)}, {"IBFLOAT", float32(math.Inf(1))}, {"IBFLOAT", float32(math.NaN())},
	} {
		if _, _, err := resultValue(tc.value, tc.name); err == nil {
			t.Fatalf("accepted inexact/non-finite %s", tc.name)
		}
	}
}
