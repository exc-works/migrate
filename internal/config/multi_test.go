package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/exc-works/migrate/internal/tunnel"
	"go.yaml.in/yaml/v3"
)

func writeConfigTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConfigurationDiscovery(t *testing.T) {
	dir := t.TempDir()
	writeConfigTestFile(t, dir, DefaultJSONFilename, `{"data_source_name":"json"}`)
	check := func(want string) {
		t.Helper()
		cfg, err := Read("", dir)
		if err != nil || cfg.DataSourceName != want || cfg.WorkingDirectory != dir {
			t.Fatalf("discovery: cfg=%+v err=%v, want %s", cfg, err, want)
		}
	}
	check("json")
	writeConfigTestFile(t, dir, DefaultYAMLAlternativeFilename, "data_source_name: yaml\n")
	check("yaml")
	writeConfigTestFile(t, dir, DefaultYAMLFilename, "# preferred file\ndata_source_name: yml\n")
	check("yml")
	cfg, err := Read(DefaultJSONFilename, dir)
	if err != nil || cfg.DataSourceName != "json" {
		t.Fatalf("explicit JSON selection: %+v %v", cfg, err)
	}
	if _, err := Read("missing.json", dir); !os.IsNotExist(err) {
		t.Fatalf("explicit missing path: %v", err)
	}
	writeConfigTestFile(t, dir, DefaultYAMLFilename, "data_source_name: [broken\n")
	if _, err := Read("", dir); err == nil || os.IsNotExist(err) {
		t.Fatalf("invalid preferred file must not fall back: %v", err)
	}
}

func TestDiscoveryDoesNotHideUnreadableOrDanglingFile(t *testing.T) {
	for _, kind := range []string{"directory", "dangling"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			writeConfigTestFile(t, dir, DefaultJSONFilename, `{"data_source_name":"json"}`)
			preferred := filepath.Join(dir, DefaultYAMLFilename)
			if kind == "directory" {
				if err := os.Mkdir(preferred, 0700); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Symlink("missing-example-target", preferred); err != nil {
				t.Skipf("symlinks unavailable: %v", err)
			}
			if _, err := Read("", dir); err == nil || os.IsNotExist(err) {
				t.Fatalf("must not fall back or use new-version defaults: %v", err)
			}
		})
	}
}

func TestJSONAndYAMLShareConfigurationAndSSHInheritance(t *testing.T) {
	t.Setenv("MIGRATE_MULTI_EXAMPLE_DSN", "example://db/query?text=a#b")
	t.Setenv("MIGRATE_MULTI_EXAMPLE_HOST", "bastion.example")
	input := FileConfig{
		Dialect: "postgres",
		SSHTunnel: &tunnel.Options{
			Host: "${MIGRATE_MULTI_EXAMPLE_HOST}", User: "example",
			Auth:  tunnel.AuthOptions{Type: tunnel.AuthPrivateKey, KeyFile: "keys/example"},
			Proxy: &tunnel.ProxyOptions{Type: tunnel.ProxySOCKS5, Address: "proxy.example:1080"},
		},
		DataSources: map[string]DataSourceConfig{
			"b": {DataSourceName: "${MIGRATE_MULTI_EXAMPLE_DSN}"},
			"a": {DataSourceName: "${MIGRATE_MULTI_EXAMPLE_DSN}"},
			"c": {DataSourceName: "override", SSHTunnel: &tunnel.Options{Host: "other.example", User: "other", Auth: tunnel.AuthOptions{Type: tunnel.AuthAgent}}},
		},
		ScriptTimeout: "1m", ScriptMaxRows: 24, ScriptMaxBytes: 4096,
	}
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	yamlBytes, err := yaml.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	jsonPath := writeConfigTestFile(t, dir, "config.json", string(jsonBytes))
	yamlPath := writeConfigTestFile(t, dir, "config.yml", "# 注释中的 ${UNUSED_EXAMPLE_VARIABLE} 不需要定义\n"+string(yamlBytes))
	jsonConfig, err := Read(jsonPath, "")
	if err != nil {
		t.Fatal(err)
	}
	yamlConfig, err := Read(yamlPath, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(jsonConfig, yamlConfig) {
		t.Fatal("YAML and JSON configurations differ")
	}
	targets := yamlConfig.Targets()
	if len(targets) != 3 || targets[0].Name != "a" || targets[1].Name != "b" || targets[2].Name != "c" {
		t.Fatalf("target ordering: %+v", targets)
	}
	for _, target := range targets {
		if target.Config.DataSources != nil || target.Config.ScriptMaxRows != 24 || target.Config.SchemaName != "migration_schema" {
			t.Fatalf("incomplete effective configuration: %+v", target)
		}
	}
	if targets[0].Config.SSHTunnel.Auth.KeyFile != filepath.Join(dir, "keys/example") || targets[0].Config.SSHTunnel.Port != 22 {
		t.Fatal("inherited SSH defaults or relative paths not resolved")
	}
	if targets[2].Config.SSHTunnel.Proxy != nil || targets[2].Config.SSHTunnel.Auth.KeyFile != "" {
		t.Fatal("SSH override was merged rather than replaced")
	}
	targets[0].Config.SSHTunnel.Host = "changed.example"
	targets[0].Config.SSHTunnel.Proxy.Address = "changed.example:1080"
	if targets[1].Config.SSHTunnel.Host != "bastion.example" || targets[1].Config.SSHTunnel.Proxy.Address != "proxy.example:1080" || yamlConfig.SSHTunnel.Host != "${MIGRATE_MULTI_EXAMPLE_HOST}" {
		t.Fatal("effective SSH options share mutable state")
	}
}

func TestInvalidYAMLIsSanitized(t *testing.T) {
	for _, input := range []string{
		"data_source_name: [private-example-value\n",
		"data_source_name: first\ndata_source_name: private-example-value\n",
		"data_source_name: first\n---\ndata_source_name: private-example-value\n",
		"ssh_tunnel: private-example-value\ndata_source_name: first\n",
		"- private-example-value\n",
		"null\n",
		"",
	} {
		t.Run(input, func(t *testing.T) {
			path := writeConfigTestFile(t, t.TempDir(), "config.yaml", input)
			_, err := Read(path, "")
			if err == nil {
				t.Fatal("accepted invalid YAML")
			}
			if strings.Contains(err.Error(), "private-example-value") || strings.Contains(err.Error(), filepath.Dir(path)) {
				t.Fatalf("YAML error leaked configuration content: %v", err)
			}
		})
	}
}

func TestMultiConfigurationValidation(t *testing.T) {
	for _, input := range []string{
		`{"data_source_name":"single","data_sources":{"a":{"data_source_name":"multi"}}}`,
		`{"data_sources":{}}`,
		`{"data_sources":{" ":{"data_source_name":"example"}}}`,
		`{"data_sources":{"a":{}}}`,
		`{"data_sources":{"a":{"data_source_name":"${MIGRATE_UNSET_EXAMPLE}"}}}`,
	} {
		t.Run(input, func(t *testing.T) {
			unsetEnvForTest(t, "MIGRATE_UNSET_EXAMPLE")
			path := writeConfigTestFile(t, t.TempDir(), "config.json", input)
			if _, err := Read(path, ""); err == nil {
				t.Fatal("accepted invalid multi configuration")
			}
		})
	}
	dir := t.TempDir()
	path := writeConfigTestFile(t, dir, "config.yml", `data_sources:
  a:
    data_source_name: example
    ssh_tunnel: null
`)
	cfg, err := Read(path, "")
	if err != nil || cfg.Targets()[0].Config.SSHTunnel != nil {
		t.Fatalf("direct connection: %+v %v", cfg, err)
	}
}

func TestExplicitLegacyFilenameUsesJSON(t *testing.T) {
	path := writeConfigTestFile(t, t.TempDir(), "legacy.conf", `{"data_source_name":"example"}`)
	cfg, err := Read(path, "")
	if err != nil || cfg.DataSourceName != "example" {
		t.Fatalf("legacy extension: %+v %v", cfg, err)
	}
}
