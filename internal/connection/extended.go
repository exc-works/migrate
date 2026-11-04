package connection

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/exc-works/migrate/internal/config"
	"github.com/exc-works/migrate/internal/dialect"
	"github.com/exc-works/migrate/internal/tunnel"
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/microsoft/go-mssqldb/msdsn"
	goora "github.com/sijms/go-ora/v3"
	"github.com/sijms/go-ora/v3/configurations"
)

// fixedDialer keeps driver DNS and reconnects on the configured SSH target.
type fixedDialer struct {
	host   string
	target string
	dial   func(context.Context, string, string) (net.Conn, error)
}

func (d *fixedDialer) HostName() string { return d.host }
func (d *fixedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if address != d.target || (network != "tcp" && network != "tcp4" && network != "tcp6") {
		return nil, errors.New("SSH database redirects and alternate targets are unsupported")
	}
	return d.dial(ctx, network, address)
}

func singleTCPAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	n, e := strconv.Atoi(port)
	if err != nil || e != nil || n < 1 || n > 65535 || strings.TrimSpace(host) == "" || strings.ContainsAny(host, " ,/\\\t\r\n") {
		return errors.New("SSH requires one valid TCP host and port")
	}
	return nil
}

func mssqlConfig(dsn string) (msdsn.Config, string, error) {
	cfg, err := msdsn.Parse(dsn)
	if err != nil {
		return msdsn.Config{}, "", errors.New("invalid MSSQL connection configuration")
	}
	if cfg.Instance != "" || cfg.FailOverPartner != "" || (cfg.Parameters[msdsn.MultiSubnetFailover] != "" && cfg.MultiSubnetFailover) {
		return msdsn.Config{}, "", errors.New("SSH MSSQL requires one TCP endpoint without instance discovery or failover")
	}
	for _, protocol := range cfg.Protocols {
		if protocol != "tcp" {
			return msdsn.Config{}, "", errors.New("SSH MSSQL supports only TCP")
		}
	}
	if cfg.Port == 0 {
		cfg.Port = 1433
	}
	target := net.JoinHostPort(cfg.Host, strconv.FormatUint(cfg.Port, 10))
	if err := singleTCPAddress(target); err != nil {
		return msdsn.Config{}, "", err
	}
	cfg.Protocols = []string{"tcp"}
	cfg.MultiSubnetFailover = false
	cfg.LogFlags = 0
	return cfg, target, nil
}

func openMSSQL(ctx context.Context, o Options) (*Connection, error) {
	cfg, target, err := mssqlConfig(o.DataSourceName)
	if err != nil {
		return nil, err
	}
	r, err := tunnel.Start(ctx, *o.SSHTunnel, target, o.Prompt, o.HostKeyPrompt)
	if err != nil {
		return nil, err
	}
	c := mssql.NewConnectorConfig(cfg)
	c.Dialer = &fixedDialer{
		host:   cfg.Host,
		target: target,
		dial:   relayDialer(r, target, cfg.DialTimeout),
	}
	return &Connection{
		DB:    sql.OpenDB(c),
		ctx:   r.Context(),
		relay: r,
	}, nil
}

func oracleConfig(dsn string) (string, *configurations.ConnectionConfig, string, error) {
	dsn, err := config.DriverDSN(dialect.DriverOracle, dsn)
	if err != nil {
		return "", nil, "", err
	}
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme != "oracle" || u.User == nil || u.Hostname() == "" {
		return "", nil, "", errors.New("SSH Oracle requires a URL with one host, port and credentials")
	}
	if err := singleTCPAddress(u.Host); err != nil {
		return "", nil, "", err
	}
	// Reject descriptors/additional servers before parsing (which may load a wallet).
	for key := range u.Query() {
		if strings.EqualFold(key, "connStr") || strings.EqualFold(key, "server") {
			return "", nil, "", errors.New("SSH Oracle does not support descriptors or additional servers")
		}
	}
	cfg, err := configurations.ParseConfig(dsn)
	if err != nil {
		return "", nil, "", errors.New("invalid Oracle connection configuration")
	}
	if len(cfg.Servers) != 1 || cfg.UnixAddress != "" || cfg.EnableOOB || (cfg.AuthType != configurations.Normal && cfg.AuthType != configurations.TCPS) {
		return "", nil, "", errors.New("SSH Oracle requires one TCP/TCPS endpoint without OOB or external authentication")
	}
	server := cfg.Servers[0]
	if cfg.Protocol != "tcp" && cfg.Protocol != "tcps" {
		return "", nil, "", errors.New("SSH Oracle supports only TCP/TCPS")
	}
	target := server.NetworkAddr()
	if err := singleTCPAddress(target); err != nil {
		return "", nil, "", err
	}
	return dsn, cfg, target, nil
}

func openOracle(ctx context.Context, o Options) (*Connection, error) {
	dsn, cfg, target, err := oracleConfig(o.DataSourceName)
	if err != nil {
		return nil, err
	}
	r, err := tunnel.Start(ctx, *o.SSHTunnel, target, o.Prompt, o.HostKeyPrompt)
	if err != nil {
		return nil, err
	}
	c, ok := goora.NewConnector(dsn).(*goora.OracleConnector)
	if !ok {
		_ = r.Close()
		return nil, errors.New("Oracle connector does not support SSH routing")
	}
	c.Dialer(&fixedDialer{
		host:   cfg.Servers[0].Addr,
		target: target,
		dial:   relayDialer(r, target, cfg.ConnectTimeout),
	})
	return &Connection{
		DB:    sql.OpenDB(c),
		ctx:   r.Context(),
		relay: r,
	}, nil
}
