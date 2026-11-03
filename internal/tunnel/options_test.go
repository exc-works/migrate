package tunnel

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func validOptions() Options {
	return Options{Host: "bastion.example", User: "example", Auth: AuthOptions{Type: AuthPasswordMode}}
}

func TestOptions(t *testing.T) {
	s, err := validOptions().Normalize()
	if err != nil || s.Port != 22 || s.Connect != 10*time.Second || s.Keepalive != 15*time.Second {
		t.Fatalf("defaults: %+v %v", s, err)
	}
	for _, mutate := range []func(*Options){
		func(o *Options) { o.Host = "" }, func(o *Options) { o.Port = -1 }, func(o *Options) { o.Port = 65536 },
		func(o *Options) { o.Auth.Type = "other" }, func(o *Options) { o.Auth.Type = AuthPrivateKey },
		func(o *Options) { o.Auth.Type = AuthAgent; o.Auth.PasswordEnv = "EXAMPLE_SECRET" },
		func(o *Options) { o.Auth.PasswordEnv = "${EXAMPLE_SECRET}" },
		func(o *Options) { o.ConnectTimeout = "0s" }, func(o *Options) { o.KeepaliveTimeout = "-1s" },
		func(o *Options) {
			o.Proxy = &ProxyOptions{Type: ProxySOCKS5, Address: "example:1080", Username: "example"}
		},
		func(o *Options) { o.Proxy = &ProxyOptions{Type: ProxyHTTP, Address: "example:0"} },
		func(o *Options) { o.Proxy = &ProxyOptions{Type: "https", Address: "example:443"} },
	} {
		o := validOptions()
		mutate(&o)
		if _, err := o.Normalize(); err == nil {
			t.Errorf("accepted invalid options: %+v", o)
		}
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	for input, want := range map[string]string{"~": home, "~/keys/example": filepath.Join(home, "keys", "example"), "key": filepath.Join("config", "key")} {
		got, err := ExpandPath(input, "config")
		if err != nil || got != want {
			t.Errorf("%q: %q %v", input, got, err)
		}
	}
	if _, err := ExpandPath("~other/key", ""); err == nil {
		t.Fatal("accepted another user's home")
	}
}

func TestSecret(t *testing.T) {
	ctx := context.Background()
	const variable = "MIGRATE_TEST_EXAMPLE_SECRET"
	t.Setenv(variable, " example value ")
	got, err := Secret(ctx, variable, "example", nil)
	if err != nil || string(got) != " example value " {
		t.Fatalf("secret was trimmed: %q %v", got, err)
	}
	t.Setenv(variable, "")
	if _, err := Secret(ctx, variable, "example", func(context.Context, string) ([]byte, error) {
		t.Fatal("prompted on explicit empty env")
		return nil, nil
	}); err == nil {
		t.Fatal("accepted empty secret")
	}
	if err := os.Unsetenv(variable); err != nil {
		t.Fatal(err)
	}
	if _, err := Secret(ctx, variable, "example", nil); err == nil {
		t.Fatal("missing noninteractive secret accepted")
	}
	got, err = Secret(ctx, variable, "example", func(context.Context, string) ([]byte, error) { return []byte("example"), nil })
	if err != nil || string(got) != "example" {
		t.Fatal(err)
	}
	_, err = Secret(ctx, variable, "example", func(context.Context, string) ([]byte, error) { return nil, errors.New("sensitive example") })
	if err == nil || strings.Contains(err.Error(), "sensitive") {
		t.Fatalf("raw prompt error leaked: %v", err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := Secret(canceled, variable, "example", nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
