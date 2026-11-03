package tunnel

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
)

// DialServer opens only the configured route. It never falls back to direct TCP.
func DialServer(ctx context.Context, s Settings, prompt Prompt) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, s.Connect)
	defer cancel()
	dialer := &net.Dialer{}
	target := net.JoinHostPort(s.Host, fmt.Sprint(s.Port))
	if s.Proxy == nil {
		conn, err := dialer.DialContext(ctx, "tcp", target)
		if err != nil {
			return nil, networkError(ctx, "SSH TCP connection failed")
		}
		return conn, nil
	}
	p := s.Proxy
	var secret []byte
	if p.Username != "" {
		var err error
		secret, err = Secret(ctx, p.PasswordEnv, "Proxy password", prompt)
		if err != nil {
			return nil, err
		}
	}
	if p.Type == ProxySOCKS5 {
		var auth *proxy.Auth
		if p.Username != "" {
			auth = &proxy.Auth{
				User:     p.Username,
				Password: string(secret),
			}
		}
		d, err := proxy.SOCKS5("tcp", p.Address, auth, dialer)
		if err != nil {
			return nil, errors.New("invalid SOCKS5 configuration")
		}
		cd, ok := d.(proxy.ContextDialer)
		if !ok {
			return nil, errors.New("SOCKS5 context dialing unavailable")
		}
		conn, err := cd.DialContext(ctx, "tcp", target)
		if err != nil {
			return nil, networkError(ctx, "SOCKS5 connection or authentication failed")
		}
		return conn, nil
	}
	if p.Type != ProxyHTTP {
		return nil, errors.New("unsupported proxy type")
	}
	conn, err := dialer.DialContext(ctx, "tcp", p.Address)
	if err != nil {
		return nil, networkError(ctx, "HTTP proxy TCP connection failed")
	}
	ok := false
	defer func() {
		if !ok {
			_ = conn.Close()
		}
	}()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, errors.New("cannot set HTTP proxy deadline")
	}
	req := &http.Request{
		Method: http.MethodConnect,
		URL: &url.URL{
			Opaque: target,
		},
		Host:   target,
		Header: make(http.Header),
	}
	if p.Username != "" {
		req.Header.Set("Proxy-Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(p.Username+":"+string(secret))))
	}
	if err := req.Write(conn); err != nil {
		return nil, networkError(ctx, "HTTP CONNECT request failed")
	}
	// Limit header parsing, retaining any SSH banner read ahead by the HTTP reader.
	reader := bufio.NewReader(io.LimitReader(conn, 64<<10))
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, networkError(ctx, "invalid HTTP CONNECT response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("HTTP CONNECT rejected; check proxy access and authentication")
	}
	if !stop() {
		return nil, networkError(ctx, "HTTP CONNECT canceled")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, errors.New("cannot clear HTTP proxy deadline")
	}
	ok = true
	return &bufferedConn{
		Conn:   conn,
		reader: io.MultiReader(reader, conn),
	}, nil
}

type bufferedConn struct {
	net.Conn
	reader io.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) { return c.reader.Read(p) }

func networkError(ctx context.Context, stage string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%s: %w", stage, err)
	}
	return errors.New(stage)
}
