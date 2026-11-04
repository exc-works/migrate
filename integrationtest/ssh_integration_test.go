//go:build integration

package integrationtest

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	public "github.com/exc-works/migrate"
	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/migrate"
	"github.com/exc-works/migrate/internal/source"
	"github.com/exc-works/migrate/internal/testssh"
	"github.com/exc-works/migrate/internal/tunnel"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/ssh/knownhosts"
)

func TestSSHMigrationWorkflows(t *testing.T) {
	for _, flavor := range selectedFlavors() {
		if flavor != dialect.NamePostgres && flavor != dialect.NameMySQL && flavor != dialect.NameMariaDB {
			continue
		}
		t.Run(flavor, func(t *testing.T) {
			d, _ := dialect.FromName(flavor)
			_, dsn, cleanup := startContainerDBWithDSN(t, flavor, d)
			defer cleanup()
			target, tunneled := sshDatabaseDSN(t, flavor, dsn)
			for _, route := range []string{"direct", tunnel.ProxySOCKS5, tunnel.ProxyHTTP} {
				t.Run(route, func(t *testing.T) {
					s := testssh.New(t)
					s.Forward = func(address string) (net.Conn, error) {
						if address != sshTarget(t, flavor, tunneled) {
							t.Errorf("wrong SSH target: %s", address)
						}
						return net.Dial("tcp", target)
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
					// Building the CLI can take minutes on a cold CI cache; finish it
					// before starting the independent cancellation/loss checks below.
					if flavor == dialect.NamePostgres && route == tunnel.ProxyHTTP {
						runSSHCLI(t, tunneled, o)
					}
					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()
					conn, err := public.OpenConnection(ctx, public.ConnectionOptions{Dialect: flavor, DataSourceName: tunneled, SSHTunnel: &o})
					if err != nil {
						t.Fatal(err)
					}
					defer conn.Close()
					if err := conn.DB.PingContext(conn.Context()); err != nil {
						t.Fatalf("SSH database ping: %v", err)
					}
					table := randomName("ssh_table")
					entries := append([]source.Migration{{Filename: "V1__ssh.sql", Source: "-- +migrate Up\nCREATE TABLE " + table + " (id INT);\n-- +migrate Down\nDROP TABLE " + table + ";"}}, sshScriptMigrations(table)...)
					svc, err := migrate.NewService(conn.Context(), migrate.Config{Dialect: d, DB: conn.DB, SchemaName: randomName("ssh_history"), MigrationSource: source.StringSource{Migrations: entries}})
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
					if err != nil || len(statuses) != len(entries) || statuses[0].Status != migrate.StatusApplied {
						t.Fatalf("status failed: %v", err)
					}
					if err := svc.Down("", true); err != nil {
						t.Fatal(err)
					}
					if err := conn.Context().Err(); err != nil {
						t.Fatalf("connection expired before query cancellation: %v", err)
					}
					query := "SELECT pg_sleep(5)"
					if flavor != dialect.NamePostgres {
						query = "SELECT SLEEP(5)"
					}
					queryCtx, queryCancel := context.WithTimeout(conn.Context(), 100*time.Millisecond)
					start := time.Now()
					_, err = conn.DB.ExecContext(queryCtx, query)
					queryCancel()
					if err == nil || time.Since(start) > 2*time.Second {
						t.Fatalf("query cancellation failed: %v", err)
					}
					if err := conn.Context().Err(); err != nil {
						t.Fatalf("connection expired before transport loss: %v", err)
					}
					if err := conn.DB.PingContext(conn.Context()); err != nil {
						t.Fatalf("route unavailable before transport loss: %v", err)
					}
					s.Close()
					select {
					case <-conn.Context().Done():
					case <-time.After(time.Second):
						t.Fatal("transport loss did not cancel connection")
					}
					if err := conn.DB.PingContext(context.Background()); err == nil {
						t.Fatal("failed transport reconnected")
					}
				})
			}
		})
	}
}

func runSSHCLI(t *testing.T, dsn string, o tunnel.Options) {
	t.Helper()
	dir := t.TempDir()
	binary := filepath.Join(dir, "migrate")
	args := []string{"build", "-o", binary, "./cmd/migrate"}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", args...)
	build.Dir = ".."
	if data, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v %s", err, data)
	}
	cfg := config.DefaultTemplate()
	cfg.DataSourceName = dsn
	cfg.SSHTunnel = &o
	cfg.SchemaName = randomName("ssh_cli_history")
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	migrations := filepath.Join(dir, "migrations")
	if err := os.Mkdir(migrations, 0700); err != nil {
		t.Fatal(err)
	}
	table := randomName("ssh_cli_table")
	if err := os.WriteFile(filepath.Join(migrations, "V1__cli.sql"), []byte("-- +migrate Up\nCREATE TABLE "+table+" (id INT);\n-- +migrate Down\nDROP TABLE "+table+";"), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) []byte {
		command := exec.CommandContext(ctx, binary, append([]string{"--non-interactive", "--working-dir", dir, "--config", "config.json"}, args...)...)
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("CLI %s failed: %v %s", args[0], err, output)
		}
		return output
	}
	run("create")
	run("up", "--dry-run")
	if data := run("status", "--output", "json"); !json.Valid(data) || !strings.Contains(string(data), "pending") {
		t.Fatalf("dry-run changed state or polluted JSON: %s", data)
	}
	run("up")
	if data := run("status", "--output", "json"); !json.Valid(data) || !strings.Contains(string(data), "applied") {
		t.Fatalf("invalid CLI status: %s", data)
	}
	run("down", "--all")
}

func workflowSSHOptions(t *testing.T, s *testssh.Server) tunnel.Options {
	t.Helper()
	host, port, _ := net.SplitHostPort(s.Address)
	p, _ := strconv.Atoi(port)
	t.Setenv("MIGRATE_WORKFLOW_EXAMPLE_SECRET", "example")
	return tunnel.Options{Host: host, Port: p, User: "example", Auth: tunnel.AuthOptions{Type: tunnel.AuthPasswordMode, PasswordEnv: "MIGRATE_WORKFLOW_EXAMPLE_SECRET"}, KnownHostsFile: s.KnownHosts}
}

func sshDatabaseDSN(t *testing.T, flavor, dsn string) (string, string) {
	t.Helper()
	if flavor == dialect.NamePostgres {
		c, err := pgx.ParseConfig(dsn)
		if err != nil {
			t.Fatal("invalid generated PG config")
		}
		port := strconv.Itoa(int(c.Port))
		u := url.URL{Scheme: "postgres", Host: net.JoinHostPort("db.internal.example", port), Path: c.Database, User: url.UserPassword(c.User, c.Password), RawQuery: "sslmode=disable"}
		return net.JoinHostPort(c.Host, port), u.String()
	}
	c, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid generated MySQL config")
	}
	target := c.Addr
	_, port, _ := net.SplitHostPort(target)
	c.Addr = net.JoinHostPort("db.internal.example", port)
	return target, c.FormatDSN()
}

func sshTarget(t *testing.T, flavor, dsn string) string {
	t.Helper()
	if flavor == dialect.NamePostgres {
		c, err := pgx.ParseConfig(dsn)
		if err != nil {
			t.Fatal("invalid PG target")
		}
		return net.JoinHostPort(c.Host, strconv.Itoa(int(c.Port)))
	}
	c, err := mysql.ParseDSN(dsn)
	if err != nil {
		t.Fatal("invalid MySQL target")
	}
	return c.Addr
}

func workflowProxy(t *testing.T, kind, target, upstream string) string {
	t.Helper()
	l, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", "0"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			local, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer local.Close()
				if kind == tunnel.ProxyHTTP {
					r, err := http.ReadRequest(bufio.NewReader(local))
					if err != nil {
						return
					}
					if r.Method != "CONNECT" || r.Host != target || r.Header.Get("Proxy-Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte("example:example")) {
						t.Error("invalid HTTP proxy request")
						return
					}
					_, _ = io.WriteString(local, "HTTP/1.1 200 OK\r\n\r\n")
				} else {
					var head [2]byte
					if _, err := io.ReadFull(local, head[:]); err != nil {
						return
					}
					methods := make([]byte, head[1])
					if _, err := io.ReadFull(local, methods); err != nil {
						return
					}
					_, _ = local.Write([]byte{5, 2})
					if _, err := io.ReadFull(local, head[:]); err != nil {
						return
					}
					user := make([]byte, head[1])
					_, _ = io.ReadFull(local, user)
					var size [1]byte
					if _, err := io.ReadFull(local, size[:]); err != nil {
						return
					}
					secret := make([]byte, size[0])
					_, _ = io.ReadFull(local, secret)
					if string(user) != "example" || string(secret) != "example" {
						t.Error("invalid SOCKS authentication")
						return
					}
					_, _ = local.Write([]byte{1, 0})
					var req [5]byte
					if _, err := io.ReadFull(local, req[:]); err != nil {
						return
					}
					if req[3] != 3 {
						t.Error("SSH host resolved locally")
						return
					}
					name := make([]byte, req[4])
					_, _ = io.ReadFull(local, name)
					var port [2]byte
					_, _ = io.ReadFull(local, port[:])
					if net.JoinHostPort(string(name), strconv.Itoa(int(binary.BigEndian.Uint16(port[:])))) != target {
						t.Error("wrong SOCKS target")
						return
					}
					_, _ = local.Write([]byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 1})
				}
				remote, err := net.Dial("tcp", upstream)
				if err != nil {
					return
				}
				defer remote.Close()
				done := make(chan struct{})
				go func() { io.Copy(local, remote); local.Close(); close(done) }()
				io.Copy(remote, local)
				remote.Close()
				<-done
			}()
		}
	}()
	return l.Addr().String()
}
