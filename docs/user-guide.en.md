# migrate User Guide

This guide is for first-time users. Commands and flags are based on the current implementation (`cmd/migrate`).

## 1. Installation

### 1.1 Install from module

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Install a specific version:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Replace `vX.Y.Z` with a real version, for example `v0.2.3`.

### 1.2 Install from local source (private repo or internal network)

Run in the repository root:

```bash
go install ./cmd/migrate
```

### 1.3 Verify installation

```bash
migrate --help
```

If the command is not found:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

If you see `Repository not found`, use the local source install path above.

## 2. Initialization

### 2.1 Generate config file

```bash
migrate new config
```

Optional:

```bash
migrate new config dev.json
migrate new config --force
```

Default config template:

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 2.2 Update key config fields

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: DB connection string
- `migration_source`: migration directory (default: `migrations`)

### 2.3 Initialize migration history table

```bash
migrate create
```

`create` may succeed without output. Confirm with:

```bash
migrate status
```

If you have an existing schema and do not want to replay old SQL, use:

```bash
migrate baseline
```

## 3. Create migration version files

### 3.1 Auto-generated version

```bash
migrate new version init_users
```

### 3.2 Explicit version

```bash
migrate new version add_email -v 202604140002
```

Generated filename format:

```text
V<version>__<description>.sql
```

Default file template:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. Upgrade (apply migrations)

Dry run first:

```bash
migrate up --dry-run
```

Apply for real:

```bash
migrate up
```

Then check status:

```bash
migrate status
```

`up` may succeed without output. Use `status` as the source of truth.

## 5. Rollback

### 5.1 Roll back to a target version (target version is kept)

```bash
migrate down 202604140001
```

Semantics: only applied versions greater than `202604140001` are rolled back.

### 5.2 Roll back all applied versions

```bash
migrate down --all
```

Note: `migrate down <to-version>` and `migrate down --all` are mutually exclusive.

### 5.3 Dry-run rollback

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` may succeed without output. Run `migrate status` to verify.

## 6. Check status

```bash
migrate status
```

Machine-readable output (recommended for scripts and AI agents):

```bash
migrate status --output json
```

Output columns: `Version`, `Filename`, `Hash`, `Status`.

Common statuses:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. Upgrade or downgrade the tool itself

### 7.1 Upgrade tool version

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 Downgrade tool version

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Example:

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

If the repo is private and `go install github.com/...@...` is not available, check out the target version in source code
and run:

```bash
go install ./cmd/migrate
```

### 7.3 Check current tool version

```bash
migrate version
```

Note: release artifacts print the release version; local source builds from `go install ./cmd/migrate` usually print `dev`.

## 8. Environment variable templates

`data_source_name` supports:

- `${KEY}`: required, must exist
- `${KEY:default}`: use `default` if `KEY` is missing

Example:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

Make sure `DB_PASSWORD` is already set in your environment, then run:

```bash
migrate status
```

## 9. 10-minute first-run demo (SQLite)

### 9.1 Prepare directory and config

First verify command availability:

```bash
migrate --help
```

Create demo directory (macOS/Linux):

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Windows PowerShell equivalent:

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

Update `migration_config.json` to:

```json
{
  "schema_name": "migration_schema",
  "dialect": "sqlite",
  "data_source_name": "./demo.sqlite",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 9.2 Initialize and create migration files

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

Edit `migrations/V202604140001__init_users.sql`:

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

Edit `migrations/V202604140002__add_email.sql`:

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 Apply, check status, and rollback

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

Expected:

- after `up`: both versions are `applied`
- after `down 202604140001`: `202604140001=applied`, `202604140002=pending`
- after `down --all`: both versions are `pending`

## 10. Global flags

Use specific config file:

```bash
migrate -c ./configs/dev.json status
```

Use specific working directory:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. Common errors and troubleshooting

### 11.1 Config file not found

Error: `config file ... no such file or directory`

Fix:

- make sure `migration_config.json` exists in current directory
- or pass the config path with `-c`

### 11.2 Missing environment variable

Error: `can't find env value for XXX`

Fix:

- `export XXX=...`
- or use `${XXX:default}`

### 11.3 Incomplete `down` arguments

Error: `to-version must be set, or use --all`

Fix:

- use `migrate down <version>`
- or use `migrate down --all`

### 11.4 Unsupported dialect

Error: `unsupported dialect: xxx`

Fix: use one of:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 Migration metadata mismatch

Error: `hash mismatch` or `filename mismatch`

Fix:

- do not edit already applied migration files
- create a new higher version migration for changes
