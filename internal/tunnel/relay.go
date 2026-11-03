package tunnel

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Relay owns one SSH transport and a loopback TCP forward to a fixed DB target.
type Relay struct {
	listener net.Listener
	client   *ssh.Client
	agent    io.Closer
	ctx      context.Context
	cancel   context.CancelFunc
	settings Settings
	target   string
	done     chan struct{}
	once     sync.Once
	wg       sync.WaitGroup
	mu       sync.Mutex
	active   map[net.Conn]struct{}
	err      error
}

func Start(ctx context.Context, o Options, target string, prompt Prompt) (*Relay, error) {
	s, err := o.Normalize()
	if err != nil {
		return nil, err
	}
	host, port, err := net.SplitHostPort(target)
	if err != nil || host == "" || port == "" {
		return nil, errors.New("SSH relay requires a TCP database target")
	}
	path, err := ExpandPath(s.KnownHostsFile, ".")
	if err != nil {
		return nil, err
	}
	verify, err := knownhosts.New(path)
	if err != nil {
		return nil, errors.New("cannot load SSH known_hosts file")
	}
	auth, agentConn, err := authentication(ctx, s, prompt)
	if err != nil {
		return nil, err
	}
	keepAgent := false
	defer func() {
		if !keepAgent && agentConn != nil {
			_ = agentConn.Close()
		}
	}()
	// Resolve human input before starting network handshake deadlines.
	raw, err := DialServer(ctx, s, prompt)
	if err != nil {
		return nil, err
	}
	handshakeCtx, cancel := context.WithTimeout(ctx, s.Connect)
	defer cancel()
	stop := context.AfterFunc(handshakeCtx, func() {
		_ = raw.Close()
		if agentConn != nil {
			_ = agentConn.Close()
		}
	})
	defer stop()
	deadline, _ := handshakeCtx.Deadline()
	if err := raw.SetDeadline(deadline); err != nil {
		_ = raw.Close()
		return nil, errors.New("cannot set SSH handshake deadline")
	}
	config := &ssh.ClientConfig{
		User:            s.User,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: verify,
	}
	conn, channels, requests, err := ssh.NewClientConn(raw, net.JoinHostPort(s.Host, fmt.Sprint(s.Port)), config)
	if err != nil {
		_ = raw.Close()
		return nil, networkError(handshakeCtx, "SSH handshake, host verification or authentication failed")
	}
	if !stop() || handshakeCtx.Err() != nil {
		_ = conn.Close()
		return nil, networkError(handshakeCtx, "SSH handshake canceled")
	}
	if err := raw.SetDeadline(time.Time{}); err != nil {
		_ = conn.Close()
		return nil, errors.New("cannot clear SSH handshake deadline")
	}
	client := ssh.NewClient(conn, channels, requests)
	listener, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", "0"))
	if err != nil {
		_ = client.Close()
		return nil, errors.New("cannot start SSH loopback relay")
	}
	runCtx, runCancel := context.WithCancel(ctx)
	r := &Relay{
		listener: listener,
		client:   client,
		agent:    agentConn,
		ctx:      runCtx,
		cancel:   runCancel,
		settings: s,
		target:   target,
		done:     make(chan struct{}),
		active:   make(map[net.Conn]struct{}),
	}
	keepAgent = true
	r.wg.Add(2)
	go r.accept()
	go r.keepalive()
	go func() {
		_ = client.Wait()
		r.shutdown(errors.New("SSH transport disconnected; inspect migration state before retrying"))
	}()
	go func() {
		select {
		case <-ctx.Done():
			r.shutdown(ctx.Err())
		case <-r.done:
		}
	}()
	return r, nil
}

func (r *Relay) Address() string          { return r.listener.Addr().String() }
func (r *Relay) Context() context.Context { return r.ctx }
func (r *Relay) Done() <-chan struct{}    { return r.done }
func (r *Relay) Err() error               { r.mu.Lock(); defer r.mu.Unlock(); return r.err }

func (r *Relay) Close() error {
	r.shutdown(nil)
	r.wg.Wait()
	return nil
}

func (r *Relay) shutdown(err error) {
	r.once.Do(func() {
		r.mu.Lock()
		r.err = err
		r.cancel()
		_ = r.listener.Close()
		_ = r.client.Close()
		if r.agent != nil {
			_ = r.agent.Close()
		}
		for c := range r.active {
			_ = c.Close()
		}
		r.mu.Unlock()
		close(r.done)
	})
}

func (r *Relay) track(c net.Conn) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ctx.Err() != nil {
		_ = c.Close()
		return false
	}
	r.active[c] = struct{}{}
	return true
}

func (r *Relay) release(c net.Conn) {
	_ = c.Close()
	r.mu.Lock()
	delete(r.active, c)
	r.mu.Unlock()
}

func (r *Relay) accept() {
	defer r.wg.Done()
	for {
		local, err := r.listener.Accept()
		if err != nil {
			if r.ctx.Err() == nil {
				r.shutdown(errors.New("SSH relay listener failed"))
			}
			return
		}
		if !r.track(local) {
			continue
		}
		r.wg.Add(1)
		go r.forward(local)
	}
}

func (r *Relay) forward(local net.Conn) {
	defer r.wg.Done()
	defer r.release(local)
	ctx, cancel := context.WithTimeout(r.ctx, r.settings.Connect)
	remote, err := r.client.DialContext(ctx, "tcp", r.target)
	if err != nil {
		// A timed-out channel open cannot be individually canceled by x/crypto;
		// closing the transport also releases its pending internal goroutine.
		if ctx.Err() != nil && r.ctx.Err() == nil {
			r.shutdown(errors.New("SSH database channel timed out; inspect migration state before retrying"))
		}
		cancel()
		return
	}
	cancel()
	if !r.track(remote) {
		return
	}
	defer r.release(remote)
	done := make(chan struct{})
	go func() { _, _ = io.Copy(local, remote); _ = local.Close(); _ = remote.Close(); close(done) }()
	_, _ = io.Copy(remote, local)
	_ = local.Close()
	_ = remote.Close()
	<-done
}

func (r *Relay) keepalive() {
	defer r.wg.Done()
	ticker := time.NewTicker(r.settings.Keepalive)
	defer ticker.Stop()
	for {
		select {
		case <-r.done:
			return
		case <-ticker.C:
		}
		result := make(chan error, 1)
		go func() { _, _, err := r.client.SendRequest("keepalive@openssh.com", true, nil); result <- err }()
		timer := time.NewTimer(r.settings.KeepaliveWait)
		select {
		case <-r.done:
			timer.Stop()
			return
		case err := <-result:
			timer.Stop()
			if err != nil {
				r.shutdown(errors.New("SSH keepalive failed; inspect migration state before retrying"))
				return
			}
		case <-timer.C:
			r.shutdown(errors.New("SSH keepalive timed out; inspect migration state before retrying"))
			return
		}
	}
}
