//go:build integration

package integrationtest

import (
	"context"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	public "github.com/exc-works/migrate"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/source"
	"github.com/exc-works/migrate/internal/testssh"
	"github.com/exc-works/migrate/internal/tunnel"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/microsoft/go-mssqldb/msdsn"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestAdditionalSSHMigrationWorkflows(t *testing.T) {
	for _, flavor := range selectedFlavors() {
		if flavor != dialect.NameOracle && flavor != dialect.NameMSSQL && flavor != dialect.NameClickHouse && flavor != dialect.NameTiDB && flavor != dialect.NameRedshift {
			continue
		}
		t.Run(flavor, func(t *testing.T) {
			d, _ := dialect.FromName(flavor)
			dsn := additionalSSHDSN(t, flavor, d)
			target := additionalSSHTarget(t, flavor, dsn)
			for _, route := range []string{"direct", tunnel.ProxySOCKS5, tunnel.ProxyHTTP} {
				t.Run(route, func(t *testing.T) {
					runAdditionalSSHWorkflow(t, flavor, dsn, target, route, d)
				})
			}
		})
	}
}

func additionalSSHDSN(t *testing.T, flavor string, d dialect.Dialect) string {
	t.Helper()
	variables := map[string]string{dialect.NameOracle: envIntegrationOracle, dialect.NameRedshift: envIntegrationRedshift, dialect.NameMSSQL: envIntegrationMSSQL, dialect.NameClickHouse: envIntegrationCH, dialect.NameTiDB: envIntegrationTiDB}
	if value := strings.TrimSpace(os.Getenv(variables[flavor])); value != "" {
		return value
	}
	if flavor == dialect.NameOracle || flavor == dialect.NameRedshift {
		return requireExternalDSN(t, flavor, variables[flavor], isFlavorExplicitlySelected(flavor))
	}
	if !isFlavorExplicitlySelected(flavor) {
		t.Skip("optional SSH database requires an explicit selection or external DSN")
	}
	_, dsn, _ := startContainerDBWithDSN(t, flavor, d)
	return dsn
}

func additionalSSHTarget(t *testing.T, flavor, dsn string) string {
	t.Helper()
	switch flavor {
	case dialect.NameMSSQL:
		cfg, err := msdsn.Parse(dsn)
		if err != nil {
			t.Fatal("invalid MSSQL fixture")
		}
		port := cfg.Port
		if port == 0 {
			port = 1433
		}
		return net.JoinHostPort(cfg.Host, strconv.FormatUint(port, 10))
	case dialect.NameOracle:
		u, err := url.Parse(dsn)
		if err != nil {
			t.Fatal("invalid Oracle fixture")
		}
		return u.Host
	case dialect.NameClickHouse:
		cfg, err := clickhouse.ParseDSN(dsn)
		if err != nil || len(cfg.Addr) != 1 {
			t.Fatal("invalid ClickHouse fixture")
		}
		return cfg.Addr[0]
	case dialect.NameTiDB:
		cfg, err := mysql.ParseDSN(dsn)
		if err != nil {
			t.Fatal("invalid TiDB fixture")
		}
		return cfg.Addr
	default:
		cfg, err := pgx.ParseConfig(dsn)
		if err != nil {
			t.Fatal("invalid Redshift fixture")
		}
		return net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	}
}

func runAdditionalSSHWorkflow(t *testing.T, flavor, dsn, target, route string, d dialect.Dialect) {
	t.Helper()
	s := testssh.New(t)
	s.Forward = func(address string) (net.Conn, error) {
		if address != target {
			t.Errorf("SSH destination changed: %s", address)
		}
		return net.DialTimeout("tcp", target, 5*time.Second)
	}
	s.Serve()
	o := workflowSSHOptions(t, s)
	if route != "direct" {
		o.Host = "ssh.internal.example"
		data, err := os.ReadFile(s.KnownHosts)
		if err != nil {
			t.Fatal(err)
		}
		fields := strings.Fields(string(data))
		fields[0] = knownhosts.Normalize(net.JoinHostPort(o.Host, strconv.Itoa(o.Port)))
		if err := os.WriteFile(s.KnownHosts, []byte(strings.Join(fields, " ")+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		o.Proxy = &tunnel.ProxyOptions{Type: route, Address: workflowProxy(t, route, net.JoinHostPort(o.Host, strconv.Itoa(o.Port)), s.Address), Username: "example", PasswordEnv: "MIGRATE_WORKFLOW_EXAMPLE_SECRET"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := public.OpenConnection(ctx, public.ConnectionOptions{Dialect: flavor, DataSourceName: dsn, SSHTunnel: &o})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.DB.SetMaxIdleConns(0)
	for i := 0; i < 2; i++ {
		if err := conn.DB.PingContext(conn.Context()); err != nil {
			t.Fatalf("SSH physical connection: %v", err)
		}
	}
	conn.DB.SetMaxIdleConns(1)
	table := randomName("ssh_extra")
	create := "CREATE TABLE " + table + " (id INT)"
	if flavor == dialect.NameClickHouse {
		create = "CREATE TABLE " + table + " (id Int32) ENGINE=MergeTree ORDER BY id"
	}
	entries := append([]source.Migration{{Filename: "V1__ssh.sql", Source: "-- +migrate Up\n" + create + ";\n-- +migrate Down\nDROP TABLE " + table + ";"}}, sshScriptMigrations(table, flavor)...)
	history := randomName("ssh_history")
	defer func() {
		cleanupCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		conn.DB.ExecContext(cleanupCtx, "DROP TABLE "+history)
	}()
	svc, err := migrate.NewService(conn.Context(), migrate.Config{Dialect: d, DB: conn.DB, SchemaName: history, MigrationSource: source.StringSource{Migrations: entries}})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Create(); err != nil {
		t.Fatal(err)
	}
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	if err := svc.Up(); err != nil {
		t.Fatal(err)
	}
	statuses, err := svc.Status()
	if err != nil || len(statuses) != len(entries) {
		t.Fatalf("SSH status: %v", err)
	}
	for _, status := range statuses {
		if status.Status != migrate.StatusApplied {
			t.Fatalf("unexpected SSH migration state: %s", status.Status)
		}
	}
	if got := queryCount(t, conn.DB, "SELECT COUNT(*) FROM "+table); got != len(entries)-1 {
		t.Fatalf("script writes: got %d", got)
	}
	if err := svc.Down("", true); err != nil {
		t.Fatal(err)
	}
	if err := conn.DB.PingContext(conn.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.DB.ExecContext(conn.Context(), "DROP TABLE "+history); err != nil {
		t.Fatal(err)
	}
	s.Close()
	select {
	case <-conn.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("SSH transport loss did not cancel the connection")
	}
	if err := conn.DB.PingContext(context.Background()); err == nil {
		t.Fatal("closed SSH route reconnected")
	}
}

func TestClickHouseHTTPSSH(t *testing.T) {
	if !isFlavorExplicitlySelected(dialect.NameClickHouse) {
		t.Skip("ClickHouse HTTP container requires explicit selection")
	}
	if os.Getenv(envIntegrationCH) != "" {
		t.Skip("external ClickHouse DSN does not imply an HTTP endpoint")
	}
	d, _ := dialect.FromName(dialect.NameClickHouse)
	_, dsn, _ := startContainerDBWithDSN(t, dialect.NameClickHouse, d, clickhouse.HTTP.String())
	runAdditionalSSHWorkflow(t, dialect.NameClickHouse, dsn, additionalSSHTarget(t, dialect.NameClickHouse, dsn), tunnel.ProxyHTTP, d)
}
