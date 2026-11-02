package config

import (
	"net/url"
	"strings"
	"testing"
)

func TestOracleDriverDSN(t *testing.T) {
	raw := "oracle://user:dummy@db:1521/service?SSL=true&TIMEOUT=10"
	got, err := DriverDSN("oracle", raw)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.User.String() != "user:dummy" || parsed.Host != "db:1521" || strings.Trim(parsed.Path, "/") != "service" || parsed.Query().Get("FAST LOGIN") != "false" || parsed.Query().Get("SSL") != "true" || parsed.Query().Get("TIMEOUT") != "10" {
		t.Fatal("Oracle DSN values changed unexpectedly")
	}
	for _, raw := range []string{"oracle://user:dummy@db:1521/service?FAST+LOGIN=true", "oracle://user:dummy@db:1521/service?fast+login=false"} {
		got, err := DriverDSN("oracle", raw)
		if err != nil || got != raw {
			t.Fatal("explicit fast-login setting was not preserved")
		}
	}
	for _, raw := range []string{"oracle://user:dummy@db:1521/%zz", "oracle://user:dummy@db:1521/service?invalid=%zz"} {
		_, err := DriverDSN("oracle", raw)
		if err == nil || strings.Contains(err.Error(), "dummy") || strings.Contains(err.Error(), raw) {
			t.Fatal("invalid DSN must fail without exposing credentials")
		}
	}
	got, err = DriverDSN("mysql", "user:dummy@tcp(db:3306)/test")
	if err != nil || got != "user:dummy@tcp(db:3306)/test" {
		t.Fatal("non-Oracle DSN changed")
	}
}
