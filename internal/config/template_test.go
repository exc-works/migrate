package config

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/exc-works/migrate/internal/tunnel"
)

func TestTemplateDestination(t *testing.T) {
	for _, tc := range []struct{ name, format, wantName, wantFormat string }{
		{"", "", DefaultYAMLFilename, FormatYAML},
		{"", FormatYAML, DefaultYAMLFilename, FormatYAML},
		{"", FormatJSON, DefaultJSONFilename, FormatJSON},
		{"example.yaml", "", "example.yaml", FormatYAML},
		{"example.yml", "", "example.yml", FormatYAML},
		{"example.json", "", "example.json", FormatJSON},
		{"legacy.conf", "", "legacy.conf", FormatJSON},
	} {
		name, format, err := TemplateDestination(tc.name, tc.format)
		if err != nil || name != tc.wantName || format != tc.wantFormat {
			t.Fatalf("%+v: %s %s %v", tc, name, format, err)
		}
	}
	for _, tc := range []struct{ name, format string }{{"example.json", FormatYAML}, {"example.yml", FormatJSON}, {"", "unknown"}, {"legacy.conf", FormatYAML}} {
		if _, _, err := TemplateDestination(tc.name, tc.format); err == nil {
			t.Fatalf("accepted %+v", tc)
		}
	}
}

func TestTemplateMatrixRoundTrips(t *testing.T) {
	for key, value := range map[string]string{
		"DATABASE_DSN": "single-example", "TENANT_A_DSN": "a-example", "TENANT_B_DSN": "b-example",
		"SSH_HOST": "bastion.example", "SSH_USER": "example", "SSH_KEY_FILE": "keys/example", "SSH_PROXY_ADDRESS": "proxy.example:1080",
	} {
		t.Setenv(key, value)
	}
	for _, multi := range []bool{false, true} {
		for _, auth := range []string{"", tunnel.AuthAgent, tunnel.AuthPrivateKey, tunnel.AuthPasswordMode} {
			for _, proxy := range []string{"", tunnel.ProxySOCKS5, tunnel.ProxyHTTP} {
				cfg, err := BuildTemplate(TemplateOptions{Multi: multi, SSH: auth != "", SSHAuth: auth, ProxyType: proxy})
				if auth == "" && proxy != "" {
					if err == nil {
						t.Fatal("accepted proxy without SSH")
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				for _, format := range []string{FormatYAML, FormatJSON} {
					data, err := RenderTemplate(cfg, format)
					if err != nil {
						t.Fatal(err)
					}
					if format == FormatJSON && !json.Valid(data) {
						t.Fatal("JSON template is not strict JSON")
					}
					if format == FormatYAML && !strings.Contains(string(data), "#") {
						t.Fatal("YAML template has no comments")
					}
					filename := "config." + format
					var decoded FileConfig
					if err := decodeConfig(filename, data, &decoded); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(cfg, decoded) {
						t.Fatalf("%s template changes configuration", format)
					}
					path := writeConfigTestFile(t, t.TempDir(), filename, string(data))
					if _, err := Read(path, ""); err != nil {
						t.Fatalf("generated configuration cannot be read: %v", err)
					}
				}
			}
		}
	}
}

func TestTemplateEscapingAndProxyAuthentication(t *testing.T) {
	keyPath := "keys/example \"quoted\" # comment\nnext"
	cfg, err := BuildTemplate(TemplateOptions{SSH: true, SSHAuth: tunnel.AuthPrivateKey, SSHKeyFile: keyPath, SSHPassphraseEnv: "MIGRATE_EXAMPLE_PASSPHRASE", ProxyType: tunnel.ProxySOCKS5, ProxyAddress: "proxy.example:1080", ProxyUsername: "example", ProxyPasswordEnv: "MIGRATE_EXAMPLE_PROXY_PASSWORD"})
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{FormatYAML, FormatJSON} {
		data, err := RenderTemplate(cfg, format)
		if err != nil {
			t.Fatal(err)
		}
		var decoded FileConfig
		if err := decodeConfig("config."+format, data, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(cfg, decoded) {
			t.Fatal("template values changed during escaping")
		}
	}
}

func TestTemplateRejectsInvalidCombinations(t *testing.T) {
	for _, options := range []TemplateOptions{
		{SSHHost: "bastion.example"},
		{SSH: true, SSHAuth: "unknown"},
		{SSH: true, SSHPort: -1},
		{SSH: true, SSHAuth: tunnel.AuthAgent, SSHKeyFile: "keys/example"},
		{SSH: true, SSHAuth: tunnel.AuthPrivateKey, SSHPasswordEnv: "MIGRATE_EXAMPLE_PASSWORD"},
		{SSH: true, ProxyType: "unknown"},
		{SSH: true, ProxyType: tunnel.ProxyHTTP, ProxyUsername: "example"},
		{SSH: true, ProxyType: tunnel.ProxyHTTP, ProxyPasswordEnv: "MIGRATE_EXAMPLE_PASSWORD"},
		{SSH: true, ProxyAddress: "proxy.example:1080"},
		{SSH: true, SSHAuth: tunnel.AuthPasswordMode, SSHPasswordEnv: "invalid-name"},
	} {
		if _, err := BuildTemplate(options); err == nil {
			t.Fatalf("accepted %+v", options)
		}
	}
}
