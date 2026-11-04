# Quick start with SQLite

Create, apply and roll back one SQL migration. SQLite needs no database server.
Use a new directory: the rollback below removes the example table.

## Install

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

This installs the SQL build. See the [user guide](./guide#_1-installation) for
source installation and PATH setup, or [database capabilities](./databases) for the full build.

## Configure

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

Replace the generated `migration_config.json` with:

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

## Create a migration

```sh
migrate create
migrate new version init_users -v 202611030001
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

Version `202611030001` should be `applied`. Successful migration commands may print
nothing; `status` is the verification step.

## Roll back

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

The table is removed and the version returns to `pending`. A rollback only reverses
what its Down SQL implements. Write and review a real reverse operation for each migration.

Continue with the [command reference](./commands), [configuration guide](./guide#_2-initialization)
or [JavaScript and TypeScript migrations](./scripting).
