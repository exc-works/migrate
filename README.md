# migrate

SQL migration tool with PostgreSQL/MySQL/MariaDB/Oracle/SQLite/MSSQL/ClickHouse/TiDB/Redshift support.

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

## Install

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

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
