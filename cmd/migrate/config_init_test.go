package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/exc-works/migrate/internal/config"
	"go.yaml.in/yaml/v3"
)

func executeInitializer(args ...string) error {
	var out bytes.Buffer
	command := newRootCommand(&out, &out)
	command.SetArgs(append([]string{"new", "config"}, args...))
	return command.Execute()
}

func TestInitializerNeedsNoEnvironmentAndPreservesExistingFiles(t *testing.T) {
	for _, key := range []string{"SHARD_0_DSN", "SHARD_1_DSN", "SSH_HOST", "SSH_USER", "SSH_PROXY_ADDRESS"} {
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
	if err != nil || !strings.Contains(string(before), "#") || !strings.Contains(string(before), "${SHARD_0_DSN}") {
		t.Fatalf("commented template: %v", err)
	}
	if err := executeInitializer(path); err == nil {
		t.Fatal("initializer overwrote existing file")
	}
	if err := executeInitializer(path, "--force", "--ssh-tunnel", "--ssh-auth", "unknown"); err == nil {
		t.Fatal("invalid flags accepted")
	}
	if err := executeInitializer(path, "--force", "--multi", "--shards", "0"); err == nil {
		t.Fatal("invalid shard count accepted")
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
		{"--shards", "3"},
		{"--multi", "--shards", "0"},
		{"--multi", "--shards", "-1"},
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

func TestInitializerGeneratesRequestedShards(t *testing.T) {
	for _, extension := range []string{"yml", "json"} {
		for _, count := range []int{0, 1, 4, 27} {
			path := filepath.Join(t.TempDir(), "config."+extension)
			args := []string{path, "--multi", "--ssh-tunnel", "--proxy", "socks5"}
			want := config.DefaultShardCount
			if count > 0 {
				args = append(args, "--shards", strconv.Itoa(count))
				want = count
			}
			if err := executeInitializer(args...); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var cfg config.FileConfig
			if extension == "json" {
				err = json.Unmarshal(data, &cfg)
			} else {
				err = yaml.Unmarshal(data, &cfg)
			}
			if err != nil || len(cfg.DataSources) != want || cfg.SSHTunnel == nil || cfg.SSHTunnel.Proxy == nil {
				t.Fatalf("wrong shard template: count=%d format=%s err=%v", count, extension, err)
			}
			if cfg.DataSources["shard_0"].DataSourceName != "${SHARD_0_DSN}" {
				t.Fatal("missing shard_0 or its environment reference")
			}
			for name, source := range cfg.DataSources {
				if source.DataSourceName != "${"+strings.ToUpper(name)+"_DSN}" || source.SSHTunnel != nil {
					t.Fatalf("source does not use its own DSN and shared SSH: %s", name)
				}
			}
		}
	}
}
