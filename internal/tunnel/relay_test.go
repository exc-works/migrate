package tunnel

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/exc-works/migrate/internal/testssh"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func relayOptions(t *testing.T, s *testssh.Server) Options {
	t.Helper()
	host, port, _ := net.SplitHostPort(s.Address)
	p, _ := strconv.Atoi(port)
	t.Setenv("MIGRATE_SSH_EXAMPLE_SECRET", "example")
	return Options{Host: host, Port: p, User: "example", Auth: AuthOptions{Type: AuthPasswordMode, PasswordEnv: "MIGRATE_SSH_EXAMPLE_SECRET"}, KnownHostsFile: s.KnownHosts, KeepaliveInterval: "20ms", KeepaliveTimeout: "50ms"}
}

func TestRelayAgent(t *testing.T) {
	s := testssh.New(t)
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	ring := agent.NewKeyring()
	if err := ring.Add(agent.AddedKey{PrivateKey: key}); err != nil {
		t.Fatal(err)
	}
	s.Config.PasswordCallback = nil
	s.Config.PublicKeyCallback = func(_ ssh.ConnMetadata, _ ssh.PublicKey) (*ssh.Permissions, error) { return nil, nil }
	s.Serve()
	socket := filepath.Join(t.TempDir(), "s")
	l, err := net.Listen("unix", socket)
	if err != nil {
		t.Skipf("Unix agent socket unavailable: %v", err)
	}
	defer l.Close()
	go func() {
		c, e := l.Accept()
		if e == nil {
			defer c.Close()
			_ = agent.ServeAgent(ring, c)
		}
	}()
	t.Setenv("SSH_AUTH_SOCK", socket)
	o := relayOptions(t, s)
	o.Auth = AuthOptions{Type: AuthAgent}
	r, err := Start(context.Background(), o, "db.internal.example:5432", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
}

func TestRelayPendingChannelTimeout(t *testing.T) {
	s := testssh.New(t)
	release := make(chan struct{})
	defer close(release)
	s.Forward = func(string) (net.Conn, error) { <-release; return nil, io.EOF }
	s.Serve()
	o := relayOptions(t, s)
	o.ConnectTimeout = "50ms"
	r, err := Start(context.Background(), o, "db.internal.example:5432", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	c, err := net.Dial("tcp", r.Address())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	select {
	case <-r.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("pending channel open survived timeout")
	}
	if r.Err() == nil {
		t.Fatal("channel timeout did not invalidate transport")
	}
}

func TestRelayRoundtripAndClose(t *testing.T) {
	s := testssh.New(t)
	s.Forward = func(target string) (net.Conn, error) {
		if target != "db.internal.example:5432" {
			t.Errorf("wrong remote target: %s", target)
		}
		a, b := net.Pipe()
		go func() { defer b.Close(); io.Copy(b, b) }()
		return a, nil
	}
	s.Serve()
	r, err := Start(context.Background(), relayOptions(t, s), "db.internal.example:5432", nil)
	if err != nil {
		t.Fatal(err)
	}
	c, err := net.Dial("tcp", r.Address())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(c, "example"); err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 7)
	if _, err := io.ReadFull(c, data); err != nil || string(data) != "example" {
		t.Fatalf("roundtrip: %q %v", data, err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := net.DialTimeout("tcp", r.Address(), 100*time.Millisecond); err == nil {
		t.Fatal("listener remained open")
	}
	if _, err := c.Read(data); err == nil {
		t.Fatal("active connection survived close")
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRelayHostKeyAndAuthFailures(t *testing.T) {
	for _, kind := range []string{"hostkey", "password"} {
		t.Run(kind, func(t *testing.T) {
			s := testssh.New(t)
			s.Serve()
			o := relayOptions(t, s)
			if kind == "hostkey" {
				o.KnownHostsFile = filepath.Join(t.TempDir(), "known_hosts")
				if err := os.WriteFile(o.KnownHostsFile, nil, 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Setenv(o.Auth.PasswordEnv, "dummy-invalid")
			}
			if r, err := Start(context.Background(), o, "db.internal.example:5432", nil); err == nil {
				r.Close()
				t.Fatal("invalid authentication accepted")
			} else if strings.Contains(err.Error(), "dummy-invalid") {
				t.Fatal("secret leaked")
			}
		})
	}
}

func TestRelayTransportFailure(t *testing.T) {
	for _, kind := range []string{"disconnect", "keepalive", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			s := testssh.New(t)
			s.IgnoreKeepalive = kind == "keepalive"
			s.Serve()
			o := relayOptions(t, s)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			r, err := Start(ctx, o, "db.internal.example:5432", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			switch kind {
			case "disconnect":
				s.Close()
			case "cancel":
				cancel()
			}
			select {
			case <-r.Done():
			case <-time.After(2 * time.Second):
				t.Fatal("transport failure did not close relay")
			}
			if r.Err() == nil || r.Context().Err() == nil {
				t.Fatal("transport failure was not terminal")
			}
		})
	}
}

func TestRelayPrivateKey(t *testing.T) {
	for _, encrypted := range []bool{false, true} {
		t.Run(strconv.FormatBool(encrypted), func(t *testing.T) {
			s := testssh.New(t)
			_, key, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				t.Fatal(err)
			}
			signer, err := ssh.NewSignerFromKey(key)
			if err != nil {
				t.Fatal(err)
			}
			s.Config.PublicKeyCallback = func(_ ssh.ConnMetadata, k ssh.PublicKey) (*ssh.Permissions, error) {
				if string(k.Marshal()) != string(signer.PublicKey().Marshal()) {
					return nil, io.EOF
				}
				return nil, nil
			}
			s.Serve()
			o := relayOptions(t, s)
			o.Auth = AuthOptions{Type: AuthPrivateKey, KeyFile: filepath.Join(t.TempDir(), "key"), PassphraseEnv: "MIGRATE_KEY_EXAMPLE_SECRET"}
			var block *pem.Block
			if encrypted {
				t.Setenv(o.Auth.PassphraseEnv, "example")
				block, err = ssh.MarshalPrivateKeyWithPassphrase(key, "example", []byte("example"))
			} else {
				block, err = ssh.MarshalPrivateKey(key, "example")
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(o.Auth.KeyFile, pem.EncodeToMemory(block), 0600); err != nil {
				t.Fatal(err)
			}
			r, err := Start(context.Background(), o, "db.internal.example:5432", nil)
			if err != nil {
				t.Fatal(err)
			}
			r.Close()
		})
	}
}
