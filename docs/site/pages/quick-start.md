# Quick start with SQLite

Use a new SQLite project to create, apply and roll back one SQL migration.

## Install

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

See [Installation](./installation) for
source installation and PATH setup, or [database capabilities](./databases) for supported migration formats.

## Configure

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

Replace the generated `migration_config.yml` with:

```yaml
# Table that records migration history
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # Local SQLite file
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # Migration directory
```

## Create a migration

```sh
migrate create
migrate new version "init users" -v 202611030001
```

Replace `migrations/V202611030001__init_users.sql` with:

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## Apply and verify

```sh
migrate up --dry-run
migrate up
migrate status
```

Confirm version `202611030001` is `applied`.

## Roll back

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

Confirm the table is removed and the version is `pending`.

Continue with [writing SQL migrations](./sql-migrations), [configuration](./configuration)
or [JavaScript and TypeScript migrations](./scripting).
