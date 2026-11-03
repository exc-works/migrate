package connection

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/exc-works/migrate/internal/tunnel"
)

func clickhouseConfig(dsn string) (*clickhouse.Options, string, error) {
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "clickhouse" && u.Scheme != "http" && u.Scheme != "https") {
		return nil, "", errors.New("SSH ClickHouse requires a clickhouse, http or https URL")
	}
	cfg, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, "", errors.New("invalid ClickHouse connection configuration")
	}
	if len(cfg.Addr) != 1 || cfg.HTTPProxyURL != nil {
		return nil, "", errors.New("SSH ClickHouse requires one address without a database HTTP proxy")
	}
	target := cfg.Addr[0]
	if err := singleTCPAddress(target); err != nil {
		return nil, "", err
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 30 * time.Second
	}
	if cfg.TLS != nil {
		cfg.TLS = cfg.TLS.Clone()
		if cfg.TLS.ServerName == "" {
			cfg.TLS.ServerName, _, _ = net.SplitHostPort(target)
		}
	}
	cfg.Debug = false
	return cfg, target, nil
}

func openClickHouse(ctx context.Context, o Options) (*Connection, error) {
	cfg, target, err := clickhouseConfig(o.DataSourceName)
	if err != nil {
		return nil, err
	}
	r, err := tunnel.Start(ctx, *o.SSHTunnel, target, o.Prompt)
	if err != nil {
		return nil, err
	}
	installClickHouseRoute(cfg, r, target)
	return &Connection{
		DB:    clickhouse.OpenDB(cfg),
		ctx:   r.Context(),
		relay: r,
	}, nil
}

func installClickHouseRoute(cfg *clickhouse.Options, r *tunnel.Relay, target string) {
	dial := relayDialer(r, target, cfg.DialTimeout)
	cfg.DialContext = func(ctx context.Context, address string) (net.Conn, error) {
		if address != target {
			return nil, errors.New("SSH ClickHouse redirects and alternate targets are unsupported")
		}
		conn, err := dial(ctx, "tcp", address)
		if err != nil {
			return nil, err
		}
		// The Native driver bypasses its TLS dialer when DialContext is set.
		// HTTP/HTTPS instead performs TLS in http.Transport, after this callback.
		if cfg.Protocol == clickhouse.Native && cfg.TLS != nil {
			tlsCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
			defer cancel()
			secured := tls.Client(conn, cfg.TLS.Clone())
			if err := secured.HandshakeContext(tlsCtx); err != nil {
				_ = conn.Close()
				return nil, errors.New("ClickHouse TLS handshake failed")
			}
			return secured, nil
		}
		return conn, nil
	}
	// Do not let HTTP_PROXY/HTTPS_PROXY redirect database traffic outside SSH.
	cfg.TransportFunc = func(transport *http.Transport) (http.RoundTripper, error) {
		transport.Proxy = nil
		return transport, nil
	}
}
