//go:build integration

package integrationtest

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	public "github.com/exc-works/migrate"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/testssh"
)

func TestSSHPostgresTLSIdentity(t *testing.T) {
	selected := false
	for _, flavor := range selectedFlavors() {
		if flavor == dialect.NamePostgres {
			selected = true
		}
	}
	if !selected {
		t.Skip("PostgreSQL not selected")
	}
	d, _ := dialect.FromName(dialect.NamePostgres)
	_, dsn, cleanup := startContainerDBWithDSN(t, dialect.NamePostgres, d)
	defer cleanup()
	target, tunneled := sshDatabaseDSN(t, dialect.NamePostgres, dsn)
	bridge, root, names := postgresTLSBridge(t, target)
	s := testssh.New(t)
	s.Forward = func(string) (net.Conn, error) { return net.Dial("tcp", bridge) }
	s.Serve()
	o := workflowSSHOptions(t, s)
	u, err := url.Parse(tunneled)
	if err != nil {
		t.Fatal("invalid fixture URL")
	}
	query := u.Query()
	query.Set("sslmode", "verify-full")
	query.Set("sslrootcert", root)
	u.RawQuery = query.Encode()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := public.OpenConnection(ctx, public.ConnectionOptions{Dialect: dialect.NamePostgres, DataSourceName: u.String(), SSHTunnel: &o})
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.DB.PingContext(conn.Context()); err != nil {
		conn.Close()
		t.Fatalf("verified TLS over SSH failed: %v", err)
	}
	conn.Close()
	select {
	case name := <-names:
		if name != "db.internal.example" {
			t.Fatalf("TLS server name changed: %s", name)
		}
	case <-ctx.Done():
		t.Fatal("no TLS handshake")
	}
	_, port, _ := net.SplitHostPort(u.Host)
	u.Host = net.JoinHostPort("wrong.internal.example", port)
	conn, err = public.OpenConnection(ctx, public.ConnectionOptions{Dialect: dialect.NamePostgres, DataSourceName: u.String(), SSHTunnel: &o})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.DB.PingContext(conn.Context()); err == nil {
		t.Fatal("TLS hostname mismatch accepted")
	}
}

func postgresTLSBridge(t *testing.T, target string) (string, string, <-chan string) {
	t.Helper()
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), DNSNames: []string{"db.internal.example"}, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(t.TempDir(), "root.crt")
	if err := os.WriteFile(root, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", "0"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	names := make(chan string, 10)
	go func() {
		for {
			raw, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer raw.Close()
				_ = raw.SetDeadline(time.Now().Add(5 * time.Second))
				var request [8]byte
				if _, err := io.ReadFull(raw, request[:]); err != nil {
					return
				}
				if request != [8]byte{0, 0, 0, 8, 4, 210, 22, 47} {
					t.Error("missing PostgreSQL SSL request")
					return
				}
				if _, err := raw.Write([]byte{'S'}); err != nil {
					return
				}
				secured := tls.Server(raw, &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}})
				if err := secured.Handshake(); err != nil {
					return
				}
				names <- secured.ConnectionState().ServerName
				_ = raw.SetDeadline(time.Time{})
				remote, err := net.Dial("tcp", target)
				if err != nil {
					return
				}
				defer remote.Close()
				done := make(chan struct{})
				go func() { io.Copy(secured, remote); secured.Close(); close(done) }()
				io.Copy(remote, secured)
				remote.Close()
				<-done
			}()
		}
	}()
	return l.Addr().String(), root, names
}
