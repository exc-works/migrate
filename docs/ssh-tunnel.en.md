# SSH tunnels and proxies

Both builds support a single Go-owned SSH tunnel for PostgreSQL, MySQL and MariaDB.
Keep the original database address and TLS settings in the DSN. An optional SOCKS5
or HTTP CONNECT proxy reaches the SSH server; no external ssh executable or OpenSSH
config is used. See the [configuration example](ssh-tunnel.md).

Auth types: `password` uses `password_env`; `private_key` uses `key_file` and optional
`passphrase_env`; `agent` uses a Unix `SSH_AUTH_SOCK` socket. Windows users should use
password or key-file authentication. Secret fields name environment variables, not
literal credentials or `${...}` templates. Missing secrets prompt without echo when
a controlling terminal is available. Explicitly empty secrets fail. Library users
must opt into prompting with a callback. Authentication failures do not retry input.
For CI, supply secrets and run `migrate --non-interactive up`.

Proxy `type` is `socks5` or `http`, and `address` is host:port. Optional proxy auth
requires both `username` and `password_env`. SOCKS5 supports username/password and
HTTP CONNECT supports Basic. HTTP proxy authentication precedes SSH encryption and
has no TLS protection; use those credentials only on trusted networks. HTTPS proxy
endpoints, NTLM/Kerberos, MFA, proxy chains and SSH jumps are unsupported. Proxy
failure never falls back to direct TCP; system proxy variables are not discovered.

SSH paths accept `~`/`~/...`; relative paths resolve from the migration config file.
The default known-hosts file is `~/.ssh/known_hosts`. Provision trusted entries before
running: unknown/changed host keys fail, and migrate never edits the file. Other
nonsecret string settings accept existing environment templates; port is a number.

Positive durations: `connect_timeout` defaults to `10s`, `keepalive_interval` to
`15s`, and `keepalive_timeout` to `5s`. Driver dial deadlines cover remote channel
setup. Channel-open timeout invalidates the whole SSH route to release the pending
request. Existing query/context deadlines apply to normal TCP connections exposed
to drivers. Transport loss cancels execution without reconnecting or replaying SQL.
DDL or COMMIT might already have completed; inspect database/history before recovery.
Multi-host DSNs, Unix sockets and other tunneled dialects fail explicitly; direct
connections retain their existing support.

Prompts go to stderr. Input supports backspace, Ctrl+U and Ctrl+C/Ctrl+D cancellation
and restores echo. Version/capabilities/new do not start SSH; status/baseline/dry-run
may connect for existing history reads. CLI-created resources close on every exit.
Native PTY behavior is tested on Unix; Windows console input is cross-build checked.

For Go callers, `migrate.OpenConnection` returns an owned pool and optional tunnel.
Ping `Connection.DB` with `Connection.Context()`, pass that DB/context to NewService,
and close Connection after migration execution. `ConnectionOptions.Prompt` is opt-in.
PG/MySQL drivers are registered by the helper; other direct drivers require caller
registration as with sql.Open. The existing caller-supplied DB API is unchanged.
