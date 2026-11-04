package tunnel

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// HostKeyInfo identifies the unknown SSH server key shown for explicit approval.
type HostKeyInfo struct {
	Address     string
	Algorithm   string
	Fingerprint string
}

// HostKeyPrompt is opt-in. Returning true approves persisting this server key.
// Implementations must honor cancellation. A nil callback rejects unknown hosts.
type HostKeyPrompt func(context.Context, HostKeyInfo) (bool, error)

type hostKeyError string

func (e hostKeyError) Error() string { return string(e) }

func loadHostKeys(path string, allowMissing bool) (ssh.HostKeyCallback, error) {
	check, err := knownhosts.New(path)
	if errors.Is(err, os.ErrNotExist) && allowMissing {
		return knownhosts.New()
	}
	if err != nil {
		return nil, hostKeyError("cannot load SSH known_hosts file")
	}
	return check, nil
}

func unknownHostKey(err error) bool {
	var keyError *knownhosts.KeyError
	return errors.As(err, &keyError) && len(keyError.Want) == 0
}

func hostKeyCallback(ctx context.Context, path string, check ssh.HostKeyCallback, confirm HostKeyPrompt) ssh.HostKeyCallback {
	var mu sync.Mutex
	return func(address string, remote net.Addr, key ssh.PublicKey) error {
		mu.Lock()
		defer mu.Unlock()
		err := check(address, remote, key)
		if err == nil {
			return nil
		}
		if !unknownHostKey(err) {
			return hostKeyError("SSH server host key changed, is revoked or cannot be verified; check known_hosts")
		}
		if confirm == nil {
			return hostKeyError("SSH server host key is unknown; prepopulate known_hosts or enable interactive confirmation")
		}
		info := HostKeyInfo{
			Address:     address,
			Algorithm:   key.Type(),
			Fingerprint: ssh.FingerprintSHA256(key),
		}
		approved, err := confirm(ctx, info)
		if err != nil {
			return hostKeyError("SSH host trust confirmation canceled or unavailable")
		}
		if !approved {
			return hostKeyError("SSH server host key was not trusted")
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := rememberHostKey(ctx, path, address, remote, key); err != nil {
			return err
		}
		// Retain trusted state for later SSH rekeys after startup's context ends.
		check, err = loadHostKeys(path, false)
		if err != nil {
			return err
		}
		return check(address, remote, key)
	}
}

func rememberHostKey(ctx context.Context, path, address string, remote net.Addr, key ssh.PublicKey) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return hostKeyError("cannot create SSH known_hosts directory")
	}
	lock := flock.New(path+".lock", flock.SetPermissions(0600))
	defer lock.Close()
	locked, err := lock.TryLockContext(ctx, 25*time.Millisecond)
	if err != nil || !locked {
		return hostKeyError("cannot lock SSH known_hosts file")
	}
	// Another process may have trusted or changed this host while we asked.
	check, err := loadHostKeys(path, true)
	if err != nil {
		return err
	}
	if err := check(address, remote, key); err == nil {
		return nil
	} else if !unknownHostKey(err) {
		return hostKeyError("SSH server host key changed while awaiting confirmation; check known_hosts")
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return hostKeyError("cannot save SSH server host key")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return hostKeyError("cannot save SSH server host key")
	}
	line := knownhosts.Line([]string{address}, key) + "\n"
	if info.Size() > 0 {
		var last [1]byte
		if _, err := file.ReadAt(last[:], info.Size()-1); err != nil {
			return hostKeyError("cannot read SSH known_hosts file")
		}
		if last[0] != '\n' {
			line = "\n" + line
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := file.WriteString(line); err != nil {
		return hostKeyError("cannot save SSH server host key")
	}
	if err := file.Sync(); err != nil {
		return hostKeyError("cannot save SSH server host key")
	}
	if err := file.Close(); err != nil {
		return hostKeyError("cannot close SSH known_hosts file")
	}
	return nil
}
