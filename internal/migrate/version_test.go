package migrate

import "testing"

func TestSplitFilename(t *testing.T) {
	version, err := SplitFilename("V20260101010101__init.sql")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "20260101010101" {
		t.Fatalf("unexpected version: %s", version)
	}
}

func TestCompareVersion(t *testing.T) {
	if !CompareVersion("2", "10") {
		t.Fatalf("expected numeric compare")
	}
	if !CompareVersion("a1", "a2") {
		t.Fatalf("expected lexical compare")
	}
}
