# sql-migrate

SQL migration tool with PostgreSQL/MySQL/MariaDB/Oracle/SQLite/MSSQL/ClickHouse/TiDB/Redshift support.

## Documentation

- [User Guide (中文)](docs/user-guide.md)
- [User Guide (English)](docs/user-guide.en.md)

## Install

```bash
go install github.com/exc-works/sql-migrate/cmd/sql-migrate@latest
```

## Quick Start

```bash
sql-migrate new config
sql-migrate create
sql-migrate up
sql-migrate status
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
