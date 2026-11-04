# SSH tunnels and proxies

Connect through one SSH bastion, optionally reached through one SOCKS5 or HTTP
CONNECT proxy. Connections are managed in Go, without an
external ssh process or OpenSSH configuration.

## Configure the route

Keep the original database address and TLS settings in `DATABASE_DSN`. The bastion
must be able to reach that database address. Add this to your project configuration:

```json
{
  "dialect": "postgres",
  "data_source_name": "${DATABASE_DSN}",
  "ssh_tunnel": {
    "host": "${SSH_HOST}",
    "port": 22,
    "user": "${SSH_USER}",
    "auth": {
      "type": "password",
      "password_env": "MIGRATE_SSH_PASSWORD"
    },
    "known_hosts_file": "~/.ssh/known_hosts"
  }
}
```

On an unknown host key, interactive CLI mode shows the SSH address, key algorithm
and SHA256 fingerprint. Verify the fingerprint independently, then explicitly choose
“Trust and save”; rejection is the default. Approval appends the record and creates
a missing file/directory with 0600/0700 permissions, preserving existing entries.
Changed or revoked keys and invalid files fail without a trust prompt. Noninteractive
mode never approves or writes unknown keys; provision trusted records before CI runs.

## Choose authentication

| `auth.type` | Fields | Behavior |
| --- | --- | --- |
| `password` | `password_env` | Read the password from the named environment variable |
| `private_key` | `key_file`, optional `passphrase_env` | Read a local key; use the named variable for an encrypted key's passphrase |
| `agent` | No password/key fields | Use the Unix `SSH_AUTH_SOCK` socket |

Secret fields contain variable names such as `MIGRATE_SSH_PASSWORD`, not secret
values or `${...}` templates. Unset secrets may prompt without echo when a controlling
terminal is available; explicitly empty values fail. Authentication failure does
not retry input. On Windows, choose password or key-file authentication.

In CI, inject the required secrets and run `migrate --non-interactive up`.
Prompts go to stderr so stdout JSON remains usable. Library callers opt into
secret prompting through `ConnectionOptions.Prompt`. The separate `HostKeyPrompt`
receives `SSHHostKeyInfo` and returns true only after fingerprint verification to
approve saving a new key. Nil callbacks never prompt or trust unknown keys; callbacks
must honor cancellation. See [Use from Go](./go-library).

## Add an optional proxy

Place a `proxy` object inside `ssh_tunnel`:

```json
{
  "type": "socks5",
  "address": "127.0.0.1:1080"
}
```

Use `"type": "http"` for HTTP CONNECT. The address is host:port, not a URL.
Optional authentication requires both `username` and `password_env`. SOCKS5 supports
username/password and HTTP CONNECT supports Basic. HTTP proxy authentication occurs
before SSH encryption and has no TLS protection; use it only on trusted networks.

Proxy failure does not fall back to a direct connection. System proxy variables,
HTTPS proxy endpoints, NTLM/Kerberos, MFA, proxy chains and SSH jumps are unsupported.

## Paths and timeouts

SSH file paths accept `~` and `~/...`; relative paths resolve from the migration
config file directory. The known-hosts default is `~/.ssh/known_hosts`.
Other nonsecret string settings accept environment templates; `port` is a JSON number.

| Setting | Default |
| --- | --- |
| `connect_timeout` | `10s` |
| `keepalive_interval` | `15s` |
| `keepalive_timeout` | `5s` |

Durations must be positive. Waiting for host-key approval pauses the network handshake
timeout; the caller context can still cancel it. Driver connection deadlines also
bound remote channel setup. A channel-open timeout invalidates the whole route. Transport loss cancels
execution without reconnecting or replaying SQL; inspect data/history before recovery
because DDL or COMMIT may already have completed.

## Supported database routes

PostgreSQL, MySQL, MariaDB, MSSQL, Oracle, ClickHouse, TiDB and Redshift support a
single TCP target. Multi-host DSNs, Unix sockets, foreign-endpoint redirects and
SQLite tunnel settings fail explicitly.

- TiDB uses MySQL routing; Redshift uses PostgreSQL routing.
- MSSQL requires TCP host/port (default 1433), without named-instance discovery,
  failover or explicit MultiSubnetFailover.
- Oracle requires a single-host `oracle://` URL with a port, using TCP/TCPS;
  descriptors, extra SERVER entries, OOB and external auth are unsupported.
- ClickHouse supports Native and HTTP(S). Database `http_proxy` options are rejected;
  `HTTP_PROXY`/`HTTPS_PROXY` are disabled for database requests. Preserve the database
  TLS identity in the original DSN.

`version`, `capabilities` and `new` do not start SSH. `status`, `baseline` and
dry-run may connect to read history. CLI-owned pools and tunnels close on success,
failure or cancellation.
