package config

import "testing"

func TestReplaceEnvVar(t *testing.T) {
	t.Setenv("APP_USER", "alice")
	out, err := ReplaceEnvVar("user=${APP_USER} pass=${APP_PASS:default}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "user=alice pass=default" {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestReplaceEnvVarMissing(t *testing.T) {
	_, err := ReplaceEnvVar("x=${NOT_SET}")
	if err == nil {
		t.Fatalf("expected error")
	}
}
