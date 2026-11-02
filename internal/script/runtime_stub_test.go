//go:build !full

package script

import (
	"errors"
	"testing"
)

func TestDefaultBuildHasNoScriptRuntime(t *testing.T) {
	if formats := Formats(); len(formats) != 1 || formats[0] != "sql" {
		t.Fatalf("%v", formats)
	}
	if _, err := Compile("V1__test.js", `export function up(){}`, "up", Options{}); !errors.Is(err, ErrRequiresFull) {
		t.Fatalf("%v", err)
	}
}
