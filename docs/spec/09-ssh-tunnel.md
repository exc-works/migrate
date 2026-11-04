# SSH tunnels and proxy connections

## Scope

Implement a Go-owned local TCP relay through one SSH server, optionally reached
through one SOCKS5 or HTTP CONNECT proxy. Do not read OpenSSH config, invoke ssh,
support proxy chains, or implement keyboard-interactive/MFA, NTLM or Kerberos.
Keep SQL/script migrations and the existing caller-owned DB API compatible.

## Configuration and credentials

- Add optional `ssh_tunnel` settings: host, port (22), user, auth, known_hosts_file,
  proxy, connect_timeout (10s), keepalive_interval (15s), keepalive_timeout (5s).
- Auth supports password, private_key and agent. Secrets are referenced by environment
  variable name, never stored as literal configuration values. Password and encrypted
  key passphrase may be read without echo from a terminal when unset. An explicitly
  empty environment value is an error. Noninteractive mode fails instead of prompting.
- Proxy settings contain type, address and optional username/password_env. Support
  SOCKS5 username/password and HTTP CONNECT Basic, without automatic direct fallback.
  Proxy credentials are separate from SSH credentials. No implicit proxy environment
  discovery; explicit configuration fields support existing environment templates.
- Expand bare `~` and `~/` locally in SSH file paths, rejecting other tilde forms.
  Relative SSH paths are relative to the migration configuration file directory.
  Default known_hosts_file is `~/.ssh/known_hosts`. In interactive CLI mode, ask
  before trusting an unknown plain host key: display the original SSH address, key
  algorithm and SHA256 fingerprint, with rejection as the default. Append only
  after explicit approval; create a missing file/directory with 0600/0700 permissions.
  Preserve existing records and serialize/recheck concurrent additions. Changed,
  revoked or invalid keys fail without offering approval. Noninteractive mode and
  library callers with no HostKeyPrompt reject unknown keys and never write trust.
  Existing trusted host certificates continue through the known_hosts verifier.
- Validate settings before opening network connections. Do not trim passwords.
  Errors and logs must not expose DSNs, secrets, key material or raw causes.

## Network and driver contract

- Keep proxy dialing, SSH authentication and database routing separate. Build the
  proxy connection first, then authenticate with ssh.NewClientConn against the SSH
  target identity. Proxy-side DNS resolves the SSH host; SSH-side DNS resolves the DB.
- Bind the relay only to loopback using a kernel-assigned port. Drivers see ordinary
  TCP connections, not SSH channels that lack deadline support.
- Preserve original database TLS identity and authentication. PG uses per-instance
  pgx config and MySQL uses a per-instance connector; avoid global dialer registries.
- Support single TCP targets for PostgreSQL/MySQL/MariaDB, MSSQL, Oracle, ClickHouse,
  TiDB (MySQL route) and Redshift (PG route). SQLite has no TCP target. Reject
  multi-host, Unix socket and foreign-endpoint redirects explicitly. Preserve direct
  connections for all existing dialects and the Oracle FAST LOGIN normalization.
- MSSQL uses its connector dialer with remote DNS and the original TLS identity;
  require TCP host/port without named-instance discovery, failover or explicit
  MultiSubnetFailover. Oracle uses its connector dialer for one TCP/TCPS URL, without
  descriptors, extra servers, OOB or external authentication.
- ClickHouse Native and HTTP(S) use per-instance dial hooks. Native custom dialing
  must perform TLS itself; HTTP(S) retains transport TLS. Preserve original host
  verification. Reject database HTTP proxies and disable HTTP_PROXY/HTTPS_PROXY
  discovery so database traffic cannot bypass the configured SSH route.
- Every PG connection, including cancellation connections, must use the relay.
  Preserve connect timeouts and prevent DNS lookup or TLS fallback bypassing routing.
- SSH handshake, proxy negotiation and channel opens have bounded timeouts and
  cancellation. Pause the network handshake timeout during host-key confirmation
  and saving approved trust, including file-lock waits and filesystem synchronization;
  the caller's context still cancels it. Resume the timeout before SSH authentication.
  Preserve confirmed trust during SSH rekeys.
  Keepalive failure or transport loss is terminal; close active
  channels, reject new connections and cancel the command. Never reconnect or replay SQL.
- Local client closure must close its remote channel; shutdown must unblock accepts,
  pending opens and copy loops. Keep the SSH transport alive until DB shutdown during
  normal cleanup; cancellation interrupts network operations before waiting for them.
- CLI owns resources it creates and closes them on success, error and cancellation.
  Library callers supplying a DB retain ownership. Libraries never prompt by default;
  credential input and host-key approval are separate opt-in callbacks from the caller/CLI.
- Commands needing no DB (version, capabilities, new) must not start SSH. Dry-run and
  status retain existing history reads and may connect. Authentication prompts do not
  contaminate stdout/JSON; cancellation restores terminal echo.
- Delegate hidden terminal input, editing and terminal restoration to Huh. Keep
  project code limited to credential policy, controlling-terminal selection and
  output-error cancellation; do not maintain custom key parsing or console polling.

## Delivery and verification

Stack: contract -> settings/credentials -> proxy dialing -> SSH relay -> CLI/drivers,
workflow tests and user documentation. Each PR targets its direct predecessor and
is buildable with its own meaningful tests. Review current head/base with `@codex review`;
fix valid findings and explain findings requiring no change. Propagate parent fixes
to descendants and obtain current checks/reviews. Never merge or release automatically.

Verify proxy auth/rejection/cancellation, generated SSH host keys and mismatches,
password/key/agent authentication, ask confirmation/refusal/cancellation/persistence,
concurrent trust updates, relay lifecycle, transport loss, and actual PG/MySQL/
MariaDB, MSSQL, ClickHouse Native/HTTP and TiDB migrations; Oracle/Redshift workflows
use configured external fixtures. Exercise remote DNS, TLS hostname verification, driver timeouts,
query cancellation and no reconnect after failure. Generate test keys at runtime;
commit no private keys. Container DB tests use testcontainers and existing CI matrix.
Run all five merge gates; report skipped external tests honestly.

## Implementation stack

The layers were split further to keep database routing, terminal interaction and
workflow verification independently reviewable. Each targets its predecessor:

| Layer | PR |
| --- | --- |
| Contract | [#29](https://github.com/exc-works/migrate/pull/29) |
| Settings and secrets | [#30](https://github.com/exc-works/migrate/pull/30) |
| Proxy dialing | [#31](https://github.com/exc-works/migrate/pull/31) |
| SSH authentication and relay | [#32](https://github.com/exc-works/migrate/pull/32) |
| Driver routing and owned connection API | [#33](https://github.com/exc-works/migrate/pull/33) |
| CLI interaction and cleanup | [#34](https://github.com/exc-works/migrate/pull/34) |
| Native workflows and documentation | [#35](https://github.com/exc-works/migrate/pull/35) |
| Remaining network database drivers and workflows | [#39](https://github.com/exc-works/migrate/pull/39) |

Review head/base identities, findings, repairs and re-review triggers are recorded
in each PR discussion. Current CI/review status remains authoritative there.
