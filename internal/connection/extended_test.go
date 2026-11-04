package connection

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/exc-works/migrate/internal/testssh"
	"github.com/exc-works/migrate/internal/tunnel"
)

func TestAdditionalConfigs(t *testing.T) {
	ms, target, err := mssqlConfig("sqlserver://example:example@db.internal.example:1433?encrypt=true&TrustServerCertificate=false&dial+timeout=2")
	if err != nil || target != "db.internal.example:1433" || ms.TLSConfig == nil || ms.TLSConfig.ServerName != "db.internal.example" || ms.TLSConfig.InsecureSkipVerify || ms.DialTimeout != 2*time.Second {
		t.Fatalf("MSSQL identity/timeout: %s %v", target, err)
	}
	dsn, ora, target, err := oracleConfig("oracle://example:example@db.internal.example:1521/service?SSL=true&SSL+VERIFY=true&CONNECTION+TIMEOUT=2")
	if err != nil || target != "db.internal.example:1521" || !ora.SSL || !ora.SSLVerify || ora.Protocol != "tcps" || ora.FastLogin || ora.ConnectTimeout != 2*time.Second {
		t.Fatalf("Oracle identity/timeout: %s %v", target, err)
	}
	if !strings.Contains(dsn, "FAST+LOGIN=false") {
		t.Fatal("Oracle reconnect normalization lost")
	}
	for _, dsn := range []string{"clickhouse://example:example@db.internal.example:9440/example?secure=true", "https://example:example@db.internal.example:8443/example?secure=true"} {
		cfg, target, err := clickhouseConfig(dsn)
		if err != nil || cfg.TLS == nil || cfg.TLS.ServerName != "db.internal.example" || cfg.TLS.InsecureSkipVerify {
			t.Fatalf("ClickHouse TLS: %s %v", target, err)
		}
	}
}

func TestAdditionalInvalidConfigurations(t *testing.T) {
	for _, dsn := range []string{"sqlserver://example:example@db.internal.example/instance", "server=db.internal.example;port=1433;failoverpartner=other.example", "server=db.internal.example;protocol=np", "sqlserver://db.internal.example:1433?MultiSubnetFailover=true"} {
		if _, _, err := mssqlConfig(dsn); err == nil {
			t.Errorf("accepted MSSQL alternate route: %s", dsn)
		}
	}
	for _, dsn := range []string{"oracle://example:example@db.internal.example:1521/service?SERVER=other.example:1521", "oracle://example:example@db.internal.example:1521/service?connStr=example", "oracle://example:example@db.internal.example:1521/service?UNIX+SOCKET=example.sock", "oracle://example:example@db.internal.example:1521/service?ENABLE_OOB=true", "oracle://example:example@db.internal.example:1521/service?AUTH+TYPE=KERBEROS", "oracle://db.internal.example:1521/service", "oracle://example:example@db.internal.example:99999/service"} {
		if _, _, _, err := oracleConfig(dsn); err == nil {
			t.Errorf("accepted Oracle alternate route: %s", dsn)
		}
	}
	for _, dsn := range []string{"clickhouse://one.example:9000,two.example:9000/example", "http://db.internal.example:8123?http_proxy=http%3A%2F%2Fproxy.example%3A8080", "ftp://db.internal.example:9000/example", "clickhouse://db.internal.example:0/example"} {
		if _, _, err := clickhouseConfig(dsn); err == nil {
			t.Errorf("accepted ClickHouse alternate route: %s", dsn)
		}
	}
}

func TestFixedEndpointRoute(t *testing.T) {
	calls := 0
	d := &fixedDialer{host: "db.internal.example", target: "db.internal.example:1433", dial: func(context.Context, string, string) (net.Conn, error) {
		calls++
		return nil, errors.New("example dial failure")
	}}
	if d.HostName() != "db.internal.example" {
		t.Fatal("remote DNS identity lost")
	}
	d.DialContext(context.Background(), "udp", d.target)
	d.DialContext(context.Background(), "tcp", "other.example:1433")
	if calls != 0 {
		t.Fatal("alternate target/network reached dialer")
	}
	d.DialContext(context.Background(), "tcp", d.target)
	d.DialContext(context.Background(), "tcp", d.target)
	if calls != 2 {
		t.Fatal("fixed-target pooling/reconnect was blocked")
	}
}

func TestAdditionalDriverRoutingAndTimeout(t *testing.T) {
	for _, tc := range []struct{ dialect, dsn, target string }{
		{"mssql", "sqlserver://example:example@db.internal.example:1433?encrypt=disable", "db.internal.example:1433"},
		{"oracle", "oracle://example:example@db.internal.example:1521/service", "db.internal.example:1521"},
		{"clickhouse", "clickhouse://example:example@db.internal.example:9000/example", "db.internal.example:9000"},
		{"clickhouse", "http://example:example@db.internal.example:8123/example", "db.internal.example:8123"},
		{"tidb", "example@tcp(db.internal.example:4000)/example", "db.internal.example:4000"},
		{"redshift", "postgres://example:example@db.internal.example:5439/example?sslmode=disable", "db.internal.example:5439"},
	} {
		t.Run(tc.dialect+tc.target, func(t *testing.T) {
			s := testssh.New(t)
			release := make(chan struct{})
			defer close(release)
			targets := make(chan string, 1)
			s.Forward = func(target string) (net.Conn, error) { targets <- target; <-release; return nil, io.EOF }
			s.Serve()
			host, port, _ := net.SplitHostPort(s.Address)
			p, _ := strconv.Atoi(port)
			t.Setenv("MIGRATE_EXT_EXAMPLE_SECRET", "example")
			o := tunnel.Options{Host: host, Port: p, User: "example", Auth: tunnel.AuthOptions{Type: tunnel.AuthPasswordMode, PasswordEnv: "MIGRATE_EXT_EXAMPLE_SECRET"}, KnownHostsFile: s.KnownHosts, ConnectTimeout: "100ms"}
			conn, err := Open(context.Background(), Options{Dialect: tc.dialect, DataSourceName: tc.dsn, SSHTunnel: &o})
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := conn.DB.PingContext(ctx); err == nil {
				t.Fatal("stalled remote channel connected")
			}
			select {
			case target := <-targets:
				if target != tc.target {
					t.Fatalf("target %q", target)
				}
			default:
				t.Fatal("driver did not reach SSH endpoint")
			}
			if conn.Context().Err() == nil {
				t.Fatal("timed-out remote channel remained reusable")
			}
		})
	}
}

func TestClickHouseHTTPDisablesEnvironmentProxy(t *testing.T) {
	cfg, _, err := clickhouseConfig("http://db.internal.example:8123/example")
	if err != nil {
		t.Fatal(err)
	}
	installClickHouseRoute(cfg, nil, "db.internal.example:8123")
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
	rt, err := cfg.TransportFunc(transport)
	if err != nil || rt != transport || transport.Proxy != nil || cfg.Protocol != clickhouse.HTTP {
		t.Fatal("HTTP proxy path was not disabled")
	}
}

func TestClickHouseNativeTLSOverSSH(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	hostname := server.Certificate().DNSNames[0]
	target := net.JoinHostPort(hostname, "9440")
	s := testssh.New(t)
	s.Forward = func(address string) (net.Conn, error) {
		if address != target {
			t.Errorf("wrong TLS destination: %s", address)
		}
		return net.Dial("tcp", server.Listener.Addr().String())
	}
	s.Serve()
	host, port, _ := net.SplitHostPort(s.Address)
	p, _ := strconv.Atoi(port)
	t.Setenv("MIGRATE_TLS_EXAMPLE_SECRET", "example")
	o := tunnel.Options{Host: host, Port: p, User: "example", Auth: tunnel.AuthOptions{Type: tunnel.AuthPasswordMode, PasswordEnv: "MIGRATE_TLS_EXAMPLE_SECRET"}, KnownHostsFile: s.KnownHosts}
	r, err := tunnel.Start(context.Background(), o, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, tc := range []struct {
		name string
		ok   bool
	}{{hostname, true}, {"wrong.internal.example", false}} {
		u := url.URL{Scheme: "clickhouse", Host: target, Path: "example", RawQuery: "secure=true&dial_timeout=1s"}
		cfg, _, err := clickhouseConfig(u.String())
		if err != nil {
			t.Fatal(err)
		}
		cfg.TLS.RootCAs = roots
		cfg.TLS.ServerName = tc.name
		installClickHouseRoute(cfg, r, target)
		conn, err := cfg.DialContext(context.Background(), target)
		if !tc.ok {
			if err == nil {
				conn.Close()
				t.Fatal("accepted mismatched database TLS certificate")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		secured, ok := conn.(*tls.Conn)
		if !ok || !secured.ConnectionState().HandshakeComplete || secured.ConnectionState().ServerName != hostname {
			t.Fatal("database TLS handshake or hostname was lost")
		}
		conn.Close()
	}
}
