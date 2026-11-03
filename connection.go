package migrate

import (
	"context"

	"github.com/exc-works/migrate/internal/connection"
	"github.com/exc-works/migrate/internal/tunnel"
)

// ConnectionOptions configures a pool and an optional explicit SSH tunnel.
type ConnectionOptions = connection.Options

// Connection owns a DB pool and its optional SSH transport. Call Close after use.
type Connection = connection.Connection

// SSHTunnelOptions configures a single SSH server and optional upstream proxy.
type SSHTunnelOptions = tunnel.Options
type SSHAuthOptions = tunnel.AuthOptions
type SSHProxyOptions = tunnel.ProxyOptions

// SecretPrompt is opt-in. A nil prompt never reads a terminal.
type SecretPrompt = tunnel.Prompt

// OpenConnection constructs a pool without pinging. Use Connection.Context when
// constructing the migration service so SSH transport failure cancels execution.
// The SQLite driver must be registered by the caller.
func OpenConnection(ctx context.Context, options ConnectionOptions) (*Connection, error) {
	return connection.Open(ctx, options)
}
