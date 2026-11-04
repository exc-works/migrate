package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/exc-works/migrate/internal/config"
)

func executeInitializer(args ...string) error {
	var out bytes.Buffer
	command := newRootCommand(&out, &out)
	command.SetArgs(append([]string{"new", "config"}, args...))
	return command.Execute()
}

func TestInitializerNeedsNoEnvironmentAndPreservesExistingFiles(t *testing.T) {
	for _, key := range []string{"TENANT_A_DSN", "TENANT_B_DSN", "SSH_HOST", "SSH_USER", "SSH_PROXY_ADDRESS"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "config.yml")
	args := []string{path, "--multi", "--ssh-tunnel", "--proxy", "socks5"}
	if err := executeInitializer(args...); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(before), "#") || !strings.Contains(string(before), "${TENANT_A_DSN}") {
		t.Fatalf("commented template: %v", err)
	}
	if err := executeInitializer(path); err == nil {
		t.Fatal("initializer overwrote existing file")
	}
	if err := executeInitializer(path, "--force", "--ssh-tunnel", "--ssh-auth", "unknown"); err == nil {
		t.Fatal("invalid flags accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("invalid flags modified existing file")
	}
	if err := executeInitializer(path, "--force"); err != nil {
		t.Fatal(err)
	}
	after, _ = os.ReadFile(path)
	if bytes.Equal(before, after) {
		t.Fatal("force did not replace template")
	}
}

func TestInitializerFlagValidationBeforeFileCreation(t *testing.T) {
	for _, args := range [][]string{
		{"--proxy", "http"},
		{"--ssh-host", "bastion.example"},
		{"--proxy-address", "proxy.example:1080"},
		{"--ssh-tunnel", "--ssh-port", "0"},
		{"--ssh-tunnel", "--ssh-auth", ""},
		{"--ssh-tunnel", "--proxy", ""},
		{"--format", ""},
		{"--format", config.FormatJSON},
	} {
		path := filepath.Join(t.TempDir(), "config.yml")
		if err := executeInitializer(append([]string{path}, args...)...); err == nil {
			t.Fatalf("accepted %v", args)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("invalid flags created a file: %v", err)
		}
	}
}
