# migrate

SQL migration tool with PostgreSQL/MySQL/MariaDB/Oracle/SQLite/MSSQL/ClickHouse/TiDB/Redshift support.
An optional full build adds JavaScript and TypeScript migrations with local and npm imports for PostgreSQL/MySQL/MariaDB/SQLite, plus experimental TiDB/Redshift/Oracle/ClickHouse/MSSQL support.
Use `migrate capabilities` to inspect verification status and transaction limits.

## Documentation

- [User Guide (中文)](docs/user-guide.md)
- [User Guide (English)](docs/user-guide.en.md)
- [User Guide (繁體中文)](docs/user-guide.zh-Hant.md)
- [User Guide (日本語)](docs/user-guide.ja.md)
- [User Guide (한국어)](docs/user-guide.ko.md)
- [User Guide (Español)](docs/user-guide.es.md)
- [User Guide (Français)](docs/user-guide.fr.md)
- [User Guide (Deutsch)](docs/user-guide.de.md)
- [User Guide (Português do Brasil)](docs/user-guide.pt-BR.md)
- [User Guide (Русский)](docs/user-guide.ru.md)
- [User Guide (العربية)](docs/user-guide.ar.md)
- [User Guide (हिन्दी)](docs/user-guide.hi.md)
- [SQL + JS/TS guide (中文)](docs/script-migrations.md)
- [SQL + JS/TS guide (English)](docs/script-migrations.en.md)
- [Runnable script import examples](examples/script-imports/README.en.md) / [可运行示例](examples/script-imports/README.md)
- [Migration API declarations](docs/migration-api.d.ts)
- [Changelog](CHANGELOG.md)

## Install

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

The default install supports SQL. Install the full build when scripts are needed:

```bash
go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest
migrate capabilities --output json
```

Release assets without a suffix remain SQL-only; `_full` assets include JS/TS.
Neither build downloads an interpreter during execution.

## Quick Start

```bash
migrate new config
migrate new version init_users
migrate create
migrate up
migrate status
migrate status --output json
```

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
INTEGRATION_DB=clickhouse,tidb,mssql go test -tags=integration,full ./integrationtest/...
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
