# sql-migrate

SQL migration tool with PostgreSQL/MySQL/MariaDB/Oracle/SQLite support.

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

## Enable Pre-commit Security Scan

```bash
./scripts/install-git-hooks.sh
```

This enables repository hooks via `core.hooksPath=.githooks` and blocks commits that add:

- private key material
- absolute filesystem paths (Unix and Windows)
- hardcoded credentials (`password` / `token` / `secret`)

For remote enforcement, configure protected branch push protection or server-side pre-receive scanning.
