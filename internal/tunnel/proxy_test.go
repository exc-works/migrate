package tunnel

import (
	"bufio"
	"context"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func proxyListener(t *testing.T, handler func(net.Conn)) string {
	t.Helper()
	l, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", "0"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	go func() {
		c, e := l.Accept()
		if e == nil {
			defer c.Close()
			handler(c)
		}
	}()
	return l.Addr().String()
}

func TestHTTPProxy(t *testing.T) {
	for _, reject := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "rejected"}[reject], func(t *testing.T) {
			t.Setenv("MIGRATE_PROXY_EXAMPLE_SECRET", "example")
			address := proxyListener(t, func(c net.Conn) {
				r, err := http.ReadRequest(bufio.NewReader(c))
				if err != nil {
					t.Error(err)
					return
				}
				if r.Method != "CONNECT" || r.Host != "unresolvable.example:22" {
					t.Errorf("wrong CONNECT target: %+v", r)
				}
				if r.Header.Get("Proxy-Authorization") != "Basic "+base64.StdEncoding.EncodeToString([]byte("example:example")) {
					t.Error("missing proxy auth")
				}
				if reject {
					_, _ = io.WriteString(c, "HTTP/1.1 407 example\r\n\r\n")
					return
				}
				_, _ = io.WriteString(c, "HTTP/1.1 200 OK\r\n\r\nSSH-example\r\n")
			})
			o := validOptions()
			o.Host = "unresolvable.example"
			o.Proxy = &ProxyOptions{Type: ProxyHTTP, Address: address, Username: "example", PasswordEnv: "MIGRATE_PROXY_EXAMPLE_SECRET"}
			s, err := o.Normalize()
			if err != nil {
				t.Fatal(err)
			}
			c, err := DialServer(context.Background(), s, nil)
			if reject {
				if err == nil {
					c.Close()
					t.Fatal("accepted rejected CONNECT")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			banner, err := bufio.NewReader(c).ReadString('\n')
			if err != nil || banner != "SSH-example\r\n" {
				t.Fatalf("lost read-ahead: %q %v", banner, err)
			}
		})
	}
}

func TestSOCKSProxy(t *testing.T) {
	address := proxyListener(t, func(c net.Conn) {
		var header [2]byte
		if _, e := io.ReadFull(c, header[:]); e != nil {
			return
		}
		methods := make([]byte, int(header[1]))
		_, _ = io.ReadFull(c, methods)
		_, _ = c.Write([]byte{5, 0})
		var request [5]byte
		if _, e := io.ReadFull(c, request[:]); e != nil {
			t.Error(e)
			return
		}
		if request[3] != 3 {
			t.Error("SOCKS target was locally resolved")
			return
		}
		target := make([]byte, int(request[4])+2)
		_, _ = io.ReadFull(c, target)
		if string(target[:len(target)-2]) != "unresolvable.example" {
			t.Error("wrong remote DNS name")
		}
		_, _ = c.Write([]byte{5, 0, 0, 1, 127, 0, 0, 1, 0, 22})
		_, _ = io.WriteString(c, "example")
	})
	o := validOptions()
	o.Host = "unresolvable.example"
	o.Proxy = &ProxyOptions{Type: ProxySOCKS5, Address: address}
	s, _ := o.Normalize()
	c, err := DialServer(context.Background(), s, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	data, err := io.ReadAll(c)
	if err != nil || string(data) != "example" {
		t.Fatalf("%q %v", data, err)
	}
}

func TestProxyCancellation(t *testing.T) {
	for _, kind := range []string{ProxyHTTP, ProxySOCKS5} {
		t.Run(kind, func(t *testing.T) {
			address := proxyListener(t, func(c net.Conn) { _, _ = io.Copy(io.Discard, c) })
			o := validOptions()
			o.Proxy = &ProxyOptions{Type: kind, Address: address}
			s, _ := o.Normalize()
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			start := time.Now()
			c, err := DialServer(ctx, s, nil)
			if c != nil {
				c.Close()
			}
			if err == nil || time.Since(start) > time.Second || !strings.Contains(err.Error(), "failed") && kind == ProxySOCKS5 {
				t.Fatalf("cancellation failed: %v", err)
			}
		})
	}
}
