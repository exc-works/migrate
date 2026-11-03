// Package connection constructs caller-owned databases and optional SSH routes.
package connection

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/tunnel"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type Options struct {
	Dialect        string
	DataSourceName string
	SSHTunnel      *tunnel.Options
	Prompt         tunnel.Prompt
}

// Connection must outlive its Service. Close releases the DB before its tunnel.
type Connection struct {
	DB    *sql.DB
	ctx   context.Context
	relay *tunnel.Relay
	once  sync.Once
	err   error
}

func (c *Connection) Context() context.Context { return c.ctx }
func (c *Connection) Close() error {
	c.once.Do(func() {
		c.err = c.DB.Close()
		if c.relay != nil {
			c.err = errors.Join(c.err, c.relay.Close())
		}
	})
	return c.err
}

// Open constructs a pool; callers must PingContext before treating it as ready.
// Other direct drivers must be registered by the caller as with sql.Open.
func Open(ctx context.Context, o Options) (*Connection, error) {
	if ctx == nil {
		return nil, errors.New("connection context is required")
	}
	d, err := dialect.FromName(o.Dialect)
	if err != nil {
		return nil, err
	}
	if o.SSHTunnel == nil {
		dsn, err := config.DriverDSN(d.DriverName(), o.DataSourceName)
		if err != nil {
			return nil, err
		}
		db, err := sql.Open(d.DriverName(), dsn)
		if err != nil {
			return nil, errors.New("cannot configure database connection")
		}
		return &Connection{
			DB:  db,
			ctx: ctx,
		}, nil
	}
	canonical, _ := dialect.CanonicalName(o.Dialect)
	switch canonical {
	case dialect.NamePostgres:
		return openPostgres(ctx, o)
	case dialect.NameMySQL, dialect.NameMariaDB:
		return openMySQL(ctx, o)
	default:
		return nil, errors.New("SSH tunnels support only PostgreSQL, MySQL and MariaDB single TCP targets")
	}
}

func postgresConfig(dsn string) (*pgx.ConnConfig, string, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, "", errors.New("invalid PostgreSQL connection configuration")
	}
	if strings.HasPrefix(cfg.Host, "/") || cfg.Host == "" {
		return nil, "", errors.New("SSH requires a PostgreSQL TCP target")
	}
	for _, fallback := range cfg.Fallbacks {
		if fallback.Host != cfg.Host || fallback.Port != cfg.Port {
			return nil, "", errors.New("SSH does not support PostgreSQL multi-host targets")
		}
	}
	return cfg, net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port))), nil
}

func openPostgres(ctx context.Context, o Options) (*Connection, error) {
	cfg, target, err := postgresConfig(o.DataSourceName)
	if err != nil {
		return nil, err
	}
	r, err := tunnel.Start(ctx, *o.SSHTunnel, target, o.Prompt)
	if err != nil {
		return nil, err
	}
	// Keep Host/TLSConfig/Fallbacks unchanged, but resolve and dial only our relay.
	cfg.LookupFunc = func(ctx context.Context, host string) ([]string, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if host != cfg.Host {
			return nil, errors.New("unexpected PostgreSQL route")
		}
		return []string{host}, nil
	}
	cfg.DialFunc = relayDialer(r, target, cfg.ConnectTimeout)
	return &Connection{
		DB:    stdlib.OpenDB(*cfg),
		ctx:   r.Context(),
		relay: r,
	}, nil
}

func mysqlConfig(dsn string) (*mysql.Config, string, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, "", errors.New("invalid MySQL connection configuration")
	}
	if cfg.Net != "tcp" && cfg.Net != "tcp4" && cfg.Net != "tcp6" {
		return nil, "", errors.New("SSH requires a MySQL TCP target")
	}
	host, port, err := net.SplitHostPort(cfg.Addr)
	if err != nil || host == "" || strings.Contains(host, ",") {
		return nil, "", errors.New("SSH requires one MySQL TCP target")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return nil, "", errors.New("invalid MySQL TCP port")
	}
	return cfg, cfg.Addr, nil
}

func openMySQL(ctx context.Context, o Options) (*Connection, error) {
	cfg, target, err := mysqlConfig(o.DataSourceName)
	if err != nil {
		return nil, err
	}
	r, err := tunnel.Start(ctx, *o.SSHTunnel, target, o.Prompt)
	if err != nil {
		return nil, err
	}
	cfg.DialFunc = relayDialer(r, target, cfg.Timeout)
	cfg.Logger = quietMySQLLogger{}
	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		_ = r.Close()
		return nil, errors.New("cannot configure MySQL SSH connection")
	}
	return &Connection{
		DB:    sql.OpenDB(connector),
		ctx:   r.Context(),
		relay: r,
	}, nil
}

type quietMySQLLogger struct{}

func (quietMySQLLogger) Print(...any) {}

func relayDialer(r *tunnel.Relay, target string, timeout time.Duration) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" && network != "tcp4" && network != "tcp6" {
			return nil, errors.New("unexpected database network")
		}
		// pgx CancelRequest uses the peer address of its established TCP socket.
		if address != target && address != r.Address() && !r.OwnsPeer(address) {
			return nil, errors.New("unexpected database target")
		}
		if r.Context().Err() != nil {
			return nil, errors.New("SSH route is closed; inspect migration state before retrying")
		}
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		conn, err := r.DialContext(ctx)
		if err != nil {
			return nil, errors.New("database SSH route unavailable")
		}
		return conn, nil
	}
}
