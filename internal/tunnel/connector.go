package tunnel

import (
	"context"
	"errors"
	"io"
	"net"
)

// DialContext returns an ordinary TCP connection only after its remote SSH
// channel is established. This keeps driver dial deadlines meaningful.
func (r *Relay) DialContext(ctx context.Context) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, r.settings.Connect)
	defer cancel()
	if r.ctx.Err() != nil {
		return nil, errors.New("SSH route is closed")
	}
	remote, err := r.client.DialContext(ctx, "tcp", r.target)
	if err != nil {
		if ctx.Err() != nil && r.ctx.Err() == nil {
			r.shutdown(errors.New("SSH database channel canceled or timed out; inspect migration state before retrying"))
		}
		return nil, networkError(ctx, "SSH database channel failed")
	}
	keep := false
	defer func() {
		if !keep {
			_ = remote.Close()
		}
	}()
	listener, err := net.Listen("tcp4", net.JoinHostPort("127.0.0.1", "0"))
	if err != nil {
		return nil, errors.New("cannot create database loopback connection")
	}
	defer listener.Close()
	stop := context.AfterFunc(ctx, func() { _ = listener.Close() })
	defer stop()
	dialer := net.Dialer{}
	local, err := dialer.DialContext(ctx, "tcp4", listener.Addr().String())
	if err != nil {
		return nil, networkError(ctx, "database loopback connection failed")
	}
	defer func() {
		if !keep {
			_ = local.Close()
		}
	}()
	var peer net.Conn
	for {
		peer, err = listener.Accept()
		if err != nil {
			return nil, networkError(ctx, "database loopback accept failed")
		}
		if peer.RemoteAddr().String() == local.LocalAddr().String() {
			break
		}
		_ = peer.Close()
	}
	if ctx.Err() != nil {
		_ = peer.Close()
		return nil, ctx.Err()
	}
	r.mu.Lock()
	if r.ctx.Err() != nil {
		r.mu.Unlock()
		_ = peer.Close()
		return nil, errors.New("SSH route is closed")
	}
	r.active[peer] = struct{}{}
	r.active[remote] = struct{}{}
	r.wg.Add(1)
	r.mu.Unlock()
	keep = true
	go func() {
		defer r.wg.Done()
		defer r.release(peer)
		defer r.release(remote)
		done := make(chan struct{})
		go func() { _, _ = io.Copy(peer, remote); _ = peer.Close(); _ = remote.Close(); close(done) }()
		_, _ = io.Copy(remote, peer)
		_ = peer.Close()
		_ = remote.Close()
		<-done
	}()
	return local, nil
}

// OwnsPeer recognizes cancellation connections addressed to an existing local
// peer, without permitting a driver to select a different database destination.
func (r *Relay) OwnsPeer(address string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for c := range r.active {
		if c.LocalAddr().String() == address {
			return true
		}
	}
	return false
}
