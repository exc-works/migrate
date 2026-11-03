// Package testssh provides a disposable SSH forwarding server for transport tests.
package testssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type Server struct {
	Address         string
	KnownHosts      string
	Forward         func(string) (net.Conn, error)
	IgnoreKeepalive bool
	Config          *ssh.ServerConfig
	mu              sync.Mutex
	active          map[net.Conn]bool
	closed          bool
	listener        net.Listener
}

func New(t testing.TB) *Server {
	t.Helper()
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	config := &ssh.ServerConfig{
		PasswordCallback: func(_ ssh.ConnMetadata, p []byte) (*ssh.Permissions, error) {
			if string(p) != "example" {
				return nil, errors.New("example authentication failed")
			}
			return nil, nil
		},
	}
	config.AddHostKey(signer)
	l, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", "0"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "known_hosts")
	line := knownhosts.Line([]string{l.Addr().String()}, signer.PublicKey())
	if err := os.WriteFile(path, []byte(line+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s := &Server{
		Address:    l.Addr().String(),
		KnownHosts: path,
		Config:     config,
		Forward:    func(target string) (net.Conn, error) { return net.Dial("tcp", target) },
		active:     make(map[net.Conn]bool),
		listener:   l,
	}
	t.Cleanup(s.Close)
	return s
}

// Serve starts after test-specific auth and forwarding settings are installed.
func (s *Server) Serve() {
	go func() {
		for {
			c, e := s.listener.Accept()
			if e != nil {
				return
			}
			s.mu.Lock()
			if s.closed {
				s.mu.Unlock()
				c.Close()
				return
			}
			s.active[c] = true
			s.mu.Unlock()
			go s.handle(c)
		}
	}()
}

func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.listener.Close()
	for c := range s.active {
		c.Close()
	}
}

func (s *Server) handle(raw net.Conn) {
	defer func() { raw.Close(); s.mu.Lock(); delete(s.active, raw); s.mu.Unlock() }()
	conn, channels, requests, err := ssh.NewServerConn(raw, s.Config)
	if err != nil {
		return
	}
	defer conn.Close()
	go func() {
		for req := range requests {
			if !s.IgnoreKeepalive {
				_ = req.Reply(false, nil)
			}
		}
	}()
	for ch := range channels {
		if ch.ChannelType() != "direct-tcpip" {
			ch.Reject(ssh.UnknownChannelType, "unsupported")
			continue
		}
		var target struct {
			Host       string
			Port       uint32
			Origin     string
			OriginPort uint32
		}
		if ssh.Unmarshal(ch.ExtraData(), &target) != nil {
			ch.Reject(ssh.ConnectionFailed, "invalid target")
			continue
		}
		destination := net.JoinHostPort(target.Host, strconv.FormatUint(uint64(target.Port), 10))
		remote, err := s.Forward(destination)
		if err != nil {
			ch.Reject(ssh.ConnectionFailed, "unreachable")
			continue
		}
		channel, requests, err := ch.Accept()
		if err != nil {
			remote.Close()
			continue
		}
		go ssh.DiscardRequests(requests)
		go func() {
			defer remote.Close()
			defer channel.Close()
			done := make(chan struct{})
			go func() { io.Copy(channel, remote); channel.Close(); close(done) }()
			io.Copy(remote, channel)
			remote.Close()
			<-done
		}()
	}
}
