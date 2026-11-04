# migrate

SQL migration tool with PostgreSQL/MySQL/MariaDB/Oracle/SQLite/MSSQL/ClickHouse/TiDB/Redshift support.
JavaScript, TypeScript and MJS migrations support local and npm imports for PostgreSQL/MySQL/MariaDB/SQLite, plus experimental TiDB/Redshift/Oracle/ClickHouse/MSSQL support.
Use `migrate capabilities` to inspect verification status and transaction limits.

## Documentation

- [Documentation website](https://exc-works.github.io/migrate/) / [中文文档站](https://exc-works.github.io/migrate/zh/)
- [Quick start](https://exc-works.github.io/migrate/quick-start) / [快速开始](https://exc-works.github.io/migrate/zh/quick-start)
- [CLI reference](https://exc-works.github.io/migrate/commands) / [命令行参考](https://exc-works.github.io/migrate/zh/commands)
- [JavaScript and TypeScript](https://exc-works.github.io/migrate/scripting) / [JavaScript 与 TypeScript](https://exc-works.github.io/migrate/zh/scripting)
- [Go integration](https://exc-works.github.io/migrate/go-library) / [Go 集成](https://exc-works.github.io/migrate/zh/go-library)
- [SSH tunnels and proxies](https://exc-works.github.io/migrate/ssh-tunnels) / [SSH 隧道与代理](https://exc-works.github.io/migrate/zh/ssh-tunnels)
- [Other language guides](https://exc-works.github.io/migrate/languages)
- [Runnable script import examples](examples/script-imports/README.en.md) / [可运行示例](examples/script-imports/README.md)
- [Migration API declarations](docs/migration-api.d.ts)
- [Changelog](CHANGELOG.md)

## Install

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Inspect supported formats and database capabilities:

```bash
migrate capabilities --output json
```

The script engine is embedded; execution never downloads an interpreter.

## Quick Start

```bash
migrate new config
# Set DATABASE_DSN or edit migration_config.yml before database commands.
migrate new version init_users
migrate create
migrate up
migrate status
migrate status --output json
```

Generate commented YAML (or strict JSON with `--format json`), with optional
multi-source, SSH and proxy settings:

```bash
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
migrate --config shards.yml up --parallel 8
```

The multi-source configuration shares one migration directory and keeps separate
history in each database. All sources run with bounded concurrency; failures are
reported per source and do not stop the remaining sources. See the
[configuration guide](docs/site/pages/configuration.md) for inheritance and output.

## Test

```bash
go test ./...
go test -tags=integration ./...
```

Run a specific integration dialect:

```bash
INTEGRATION_DB=sqlite go test -tags=integration ./integrationtest/...
INTEGRATION_DB=postgres,mysql,mariadb go test -tags=integration ./integrationtest/...
```

Run ClickHouse/TiDB/MSSQL SQL and script integration in disposable containers (or set
`INTEGRATION_CLICKHOUSE_DSN` / `INTEGRATION_TIDB_DSN` / `INTEGRATION_MSSQL_DSN` to use external test instances):

```bash
INTEGRATION_DB=clickhouse,tidb,mssql go test -tags=integration ./integrationtest/...
```

TiDB containers use Unistore; distributed TiKV behavior remains unverified.
Oracle and Redshift tests require dedicated external instances and the corresponding
`INTEGRATION_ORACLE_DSN` / `INTEGRATION_REDSHIFT_DSN`. Without these values, optional
checks are skipped, which does not count as verification.

Run Oracle integration (requires reachable Oracle DSN):

```bash
INTEGRATION_DB=oracle INTEGRATION_ORACLE_DSN='<oracle dsn>' go test -tags=integration ./integrationtest/...
```

## Enable Pre-commit Security Scan

```bash
./scripts/install-git-hooks.sh
```

This enables repository hooks via `core.hooksPath=.githooks` and blocks commits that add:

- private key material
- absolute filesystem paths (Unix and Windows)
- hardcoded credentials (`password` / `token` / `secret`)

For remote enforcement, configure protected branch push protection or server-side pre-receive scanning.
