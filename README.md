# sql-migrate

SQL migration tool with PostgreSQL/MySQL/MariaDB support.

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

## Enable Pre-commit Secret Scan

```bash
./scripts/install-git-hooks.sh
```

This enables repository hooks via `core.hooksPath=.githooks` and runs secret checks before each commit.
