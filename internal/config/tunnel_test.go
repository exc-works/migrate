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

func TestTunnelNonsecretTemplates(t *testing.T) {
	t.Setenv("MIGRATE_AUTH_KIND", "password")
	t.Setenv("MIGRATE_PROXY_KIND", "socks5")
	o := &tunnel.Options{
		Host:              "example",
		User:              "example",
		Auth:              tunnel.AuthOptions{Type: "${MIGRATE_AUTH_KIND}", PasswordEnv: "MIGRATE_EXAMPLE_SECRET"},
		ConnectTimeout:    "${MIGRATE_EXAMPLE_CONNECT:10s}",
		KeepaliveInterval: "${MIGRATE_EXAMPLE_INTERVAL:15s}",
		KeepaliveTimeout:  "${MIGRATE_EXAMPLE_WAIT:5s}",
		Proxy:             &tunnel.ProxyOptions{Type: "${MIGRATE_PROXY_KIND}", Address: "example:1080"},
	}
	if err := resolveTunnelConfig(o, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if o.Auth.Type != tunnel.AuthPasswordMode || o.Proxy.Type != tunnel.ProxySOCKS5 || o.ConnectTimeout != "10s" || o.KeepaliveInterval != "15s" || o.KeepaliveTimeout != "5s" || o.Auth.PasswordEnv != "MIGRATE_EXAMPLE_SECRET" {
		t.Fatalf("incomplete expansion: %+v", o)
	}
}
