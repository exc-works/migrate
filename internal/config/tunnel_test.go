package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/exc-works/migrate/internal/tunnel"
)

func TestReadTunnel(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MIGRATE_TEST_SSH_HOST", "bastion.example")
	cfg := DefaultTemplate()
	cfg.SSHTunnel = &tunnel.Options{Host: "${MIGRATE_TEST_SSH_HOST}", User: "example", Auth: tunnel.AuthOptions{Type: tunnel.AuthPrivateKey, KeyFile: "keys/example"}}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(dir, "config.json")
	if err := os.WriteFile(name, data, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := Read(name, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.SSHTunnel.Host != "bastion.example" || got.SSHTunnel.Port != 22 || got.SSHTunnel.Auth.KeyFile != filepath.Join(dir, "keys", "example") {
		t.Fatalf("wrong resolution: %+v", got.SSHTunnel)
	}
}
