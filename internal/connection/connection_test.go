package connection

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/exc-works/migrate/internal/testssh"
	"github.com/exc-works/migrate/internal/tunnel"
	"github.com/go-sql-driver/mysql"
)

func TestPostgresConfig(t *testing.T) {
	cfg, target, err := postgresConfig("host=db.internal.example port=5432 sslmode=verify-full")
	if err != nil || target != "db.internal.example:5432" || cfg.TLSConfig.ServerName != "db.internal.example" {
		t.Fatalf("TLS identity lost: %s %v", target, err)
	}
	for _, dsn := range []string{"host=one.example,two.example sslmode=disable", "host=" + string([]byte{'/'}) + "example sslmode=disable", "postgres://example:%invalid@host/db"} {
		if _, _, err := postgresConfig(dsn); err == nil {
			t.Errorf("accepted invalid TCP target: %s", dsn)
		} else if strings.Contains(err.Error(), "invalid@") {
			t.Fatal("DSN leaked")
		}
	}
}

func TestMySQLConfig(t *testing.T) {
	if err := mysql.RegisterTLSConfig("example", &tls.Config{ServerName: "db.internal.example", MinVersion: tls.VersionTLS12}); err != nil {
		t.Fatal(err)
	}
	defer mysql.DeregisterTLSConfig("example")
	cfg, target, err := mysqlConfig("example@tcp(db.internal.example:3306)/example?tls=example&readTimeout=1s")
	if err != nil || target != "db.internal.example:3306" || cfg.TLS.ServerName != "db.internal.example" || cfg.ReadTimeout == 0 {
		t.Fatalf("config lost: %s %v", target, err)
	}
	if _, _, err := mysqlConfig("example@unix(example.sock)/example"); err == nil {
		t.Fatal("accepted socket")
	}
}

func TestRelayDialer(t *testing.T) {
	s := testssh.New(t)
	s.Serve()
	host, port, _ := net.SplitHostPort(s.Address)
	p, _ := strconv.Atoi(port)
	t.Setenv("MIGRATE_TEST_SSH_EXAMPLE_SECRET", "example")
	o := tunnel.Options{Host: host, Port: p, User: "example", Auth: tunnel.AuthOptions{Type: tunnel.AuthPasswordMode, PasswordEnv: "MIGRATE_TEST_SSH_EXAMPLE_SECRET"}, KnownHostsFile: s.KnownHosts}
	conn, err := Open(context.Background(), Options{Dialect: "postgres", DataSourceName: "host=db.internal.example sslmode=verify-full", SSHTunnel: &o})
	if err != nil {
		t.Fatal(err)
	}
	dial := relayDialer(conn.relay, "db.internal.example:5432", 0)
	for _, target := range []string{"db.internal.example:5432", conn.relay.Address()} {
		c, err := dial(context.Background(), "tcp", target)
		if err != nil {
			t.Fatal(err)
		}
		c.Close()
	}
	if _, err := dial(context.Background(), "tcp", "unexpected.example:5432"); err == nil {
		t.Fatal("allowed route bypass")
	}
	conn.Close()
	if _, err := dial(context.Background(), "tcp", "db.internal.example:5432"); err == nil {
		t.Fatal("dialed failed route")
	}
	conn.Close()
}

func TestUnsupportedTunnelBeforeNetwork(t *testing.T) {
	_, err := Open(context.Background(), Options{Dialect: "sqlite", DataSourceName: "example", SSHTunnel: &tunnel.Options{}})
	if err == nil || !strings.Contains(err.Error(), "only PostgreSQL") {
		t.Fatal(err)
	}
}
