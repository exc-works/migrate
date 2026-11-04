package tunnel

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/exc-works/migrate/internal/testssh"
	"github.com/gofrs/flock"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func testHostKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func serverHostKey(t *testing.T, s *testssh.Server) ssh.PublicKey {
	t.Helper()
	data, err := os.ReadFile(s.KnownHosts)
	if err != nil {
		t.Fatal(err)
	}
	_, _, key, _, _, err := ssh.ParseKnownHosts(data)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestAskHostKeyPersistsAndReusesTrust(t *testing.T) {
	s := testssh.New(t)
	key := serverHostKey(t, s)
	s.Serve()
	o := relayOptions(t, s)
	o.ConnectTimeout = "250ms"
	o.KnownHostsFile = filepath.Join(t.TempDir(), "ssh", "known_hosts")
	calls := 0
	confirm := func(ctx context.Context, info HostKeyInfo) (bool, error) {
		calls++
		if info.Address != s.Address || info.Algorithm != key.Type() || info.Fingerprint != ssh.FingerprintSHA256(key) {
			t.Errorf("incorrect SSH identity: %+v", info)
		}
		timer := time.NewTimer(500 * time.Millisecond)
		defer timer.Stop()
		select {
		case <-timer.C:
			return true, nil
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}
	r, err := Start(context.Background(), o, "db.internal.example:5432", nil, confirm)
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	if calls != 1 {
		t.Fatalf("confirmations: %d", calls)
	}
	check, err := knownhosts.New(o.KnownHostsFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := check(s.Address, &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: o.Port}, key); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(o.KnownHostsFile); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0600) {
		t.Fatalf("known_hosts permissions: %v", err)
	}
	if info, err := os.Stat(filepath.Dir(o.KnownHostsFile)); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm() != 0700) {
		t.Fatalf("directory permissions: %v", err)
	}
	r, err = Start(context.Background(), o, "db.internal.example:5432", nil, func(context.Context, HostKeyInfo) (bool, error) {
		t.Error("trusted host prompted again")
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
}

func TestAskHostKeyPersistencePausesHandshakeTimeout(t *testing.T) {
	for _, operation := range []string{"save", "cancel"} {
		t.Run(operation, func(t *testing.T) {
			s := testssh.New(t)
			s.Serve()
			o := relayOptions(t, s)
			o.ConnectTimeout = "250ms"
			o.KnownHostsFile = filepath.Join(t.TempDir(), "known_hosts")
			lock := flock.New(o.KnownHostsFile + ".lock")
			if err := lock.Lock(); err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			asked := make(chan struct{})
			result := make(chan error, 1)
			go func() {
				r, err := Start(ctx, o, "db.internal.example:5432", nil, func(context.Context, HostKeyInfo) (bool, error) {
					close(asked)
					return true, nil
				})
				if r != nil {
					r.Close()
				}
				result <- err
			}()
			select {
			case <-asked:
			case err := <-result:
				t.Fatalf("handshake ended before confirmation: %v", err)
			case <-ctx.Done():
				t.Fatal("host confirmation did not start")
			}
			// Trust persistence can wait on another process or a slow filesystem.
			// This local work must not consume the network budget after approval.
			select {
			case err := <-result:
				t.Fatalf("handshake ended while saving host trust: %v", err)
			case <-time.After(500 * time.Millisecond):
			}
			if operation == "cancel" {
				cancel()
			} else if err := lock.Unlock(); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-result:
				if operation == "cancel" {
					if !errors.Is(err, context.Canceled) {
						t.Fatalf("trust lock wait ignored caller cancellation: %v", err)
					}
					if _, err := os.Stat(o.KnownHostsFile); !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("canceled lock wait persisted host trust: %v", err)
					}
				} else if err != nil {
					t.Fatalf("handshake failed after saving host trust: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("handshake did not finish after releasing or canceling the trust lock")
			}
		})
	}
}

func TestAskHostKeyHandshakeTimeoutResumesAfterTrust(t *testing.T) {
	s := testssh.New(t)
	authenticating := make(chan struct{}, 1)
	release := make(chan struct{})
	defer close(release)
	s.Config.PasswordCallback = func(ssh.ConnMetadata, []byte) (*ssh.Permissions, error) {
		authenticating <- struct{}{}
		<-release
		return nil, nil
	}
	s.Serve()
	o := relayOptions(t, s)
	o.ConnectTimeout = "250ms"
	o.KnownHostsFile = filepath.Join(t.TempDir(), "known_hosts")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r, err := Start(ctx, o, "db.internal.example:5432", nil, func(context.Context, HostKeyInfo) (bool, error) {
		return true, nil
	})
	if r != nil {
		r.Close()
	}
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "SSH handshake timed out") {
		t.Fatalf("authentication did not enforce the resumed network timeout: %v", err)
	}
	if ctx.Err() != nil {
		t.Fatal("handshake waited for caller timeout instead of its network timeout")
	}
	select {
	case <-authenticating:
	default:
		t.Fatal("handshake timed out before reaching authentication")
	}
	check, err := knownhosts.New(o.KnownHostsFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := check(s.Address, &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: o.Port}, serverHostKey(t, s)); err != nil {
		t.Fatalf("host trust was not saved before authentication: %v", err)
	}
}

func TestAskHostKeyRefusesWithoutAuthenticationOrWrite(t *testing.T) {
	for _, mode := range []string{"decline", "cancel", "strict", "changed", "revoked", "malformed", "changed_during_prompt", "save_failure"} {
		t.Run(mode, func(t *testing.T) {
			s := testssh.New(t)
			key := serverHostKey(t, s)
			var authCalls atomic.Int32
			s.Config.PasswordCallback = func(ssh.ConnMetadata, []byte) (*ssh.Permissions, error) { authCalls.Add(1); return nil, nil }
			s.Serve()
			o := relayOptions(t, s)
			o.KnownHostsFile = filepath.Join(t.TempDir(), "known_hosts")
			var original string
			switch mode {
			case "changed":
				original = knownhosts.Line([]string{s.Address}, testHostKey(t)) + "\n"
			case "revoked":
				original = "@revoked " + knownhosts.Line([]string{s.Address}, key) + "\n"
			case "malformed":
				original = "not a valid host record\n"
			}
			if original != "" {
				if err := os.WriteFile(o.KnownHostsFile, []byte(original), 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			confirm := HostKeyPrompt(func(_ context.Context, info HostKeyInfo) (bool, error) {
				calls++
				switch mode {
				case "cancel":
					cancel()
					return true, nil
				case "changed_during_prompt":
					original = knownhosts.Line([]string{s.Address}, testHostKey(t)) + "\n"
					if err := os.WriteFile(o.KnownHostsFile, []byte(original), 0600); err != nil {
						t.Error(err)
						return false, err
					}
					return true, nil
				case "save_failure":
					if err := os.Mkdir(o.KnownHostsFile, 0700); err != nil {
						t.Error(err)
						return false, err
					}
					return true, nil
				}
				return false, nil
			})
			if mode == "strict" {
				confirm = nil
			}
			r, err := Start(ctx, o, "db.internal.example:5432", nil, confirm)
			if err == nil {
				r.Close()
				t.Fatal("untrusted host connected")
			}
			if authCalls.Load() != 0 {
				t.Fatal("credentials sent before host trust")
			}
			if strings.Contains(err.Error(), o.KnownHostsFile) {
				t.Fatal("raw file error leaked")
			}
			if mode == "changed" || mode == "revoked" || mode == "malformed" || mode == "strict" {
				if calls != 0 {
					t.Fatal("invalid host prompted")
				}
			}
			if mode == "save_failure" {
				return
			}
			data, readErr := os.ReadFile(o.KnownHostsFile)
			if original == "" {
				if !errors.Is(readErr, os.ErrNotExist) {
					t.Fatal("unapproved trust file created")
				}
			} else if readErr != nil || string(data) != original {
				t.Fatalf("existing records changed: %v", readErr)
			}
		})
	}
}

func TestRememberHostKeyConcurrentAppendAndLockCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "known_hosts")
	prefix := "# existing entries\n" + knownhosts.Line([]string{"other.internal.example:22"}, testHostKey(t))
	if err := os.WriteFile(path, []byte(prefix), 0600); err != nil {
		t.Fatal(err)
	}
	key := testHostKey(t)
	remote := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 22}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			if err := rememberHostKey(context.Background(), path, "db.internal.example:22", remote, key); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != prefix+"\n"+knownhosts.Line([]string{"db.internal.example:22"}, key)+"\n" {
		t.Fatal("append lost or duplicated existing records")
	}
	lock := flock.New(path + ".lock")
	if err := lock.Lock(); err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := rememberHostKey(ctx, path, "new.internal.example:22", remote, testHostKey(t)); err == nil {
		t.Fatal("locked file ignored cancellation")
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(data) {
		t.Fatalf("canceled writer changed records: %v", err)
	}
}

func TestAskHostKeyRemainsTrustedForRekey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "known_hosts")
	check, err := loadHostKeys(path, true)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	verify := hostKeyCallback(ctx, path, check, func(context.Context, HostKeyInfo) (bool, error) { calls++; return true, nil })
	key := testHostKey(t)
	remote := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 22}
	if err := verify("ssh.internal.example:22", remote, key); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := verify("ssh.internal.example:22", remote, key); err != nil {
		t.Fatalf("rekey lost trust: %v", err)
	}
	if err := verify("ssh.internal.example:22", remote, testHostKey(t)); err == nil {
		t.Fatal("changed rekey accepted")
	}
	if calls != 1 {
		t.Fatalf("rekey prompted again: %d", calls)
	}
}
