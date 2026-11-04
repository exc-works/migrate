# CLI reference

Use `migrate --help` or `migrate <command> --help` to confirm the flags in your
installed version. Use `migrate version` to identify the binary.

## Global flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `--config`, `-c` | `migration_config.json` | Config path |
| `--working-dir`, `-w` | Current directory | Resolve config from this directory and override migration working directory |
| `--non-interactive` | `false` | Fail instead of requesting secrets or approval for unknown SSH host keys |

## new config

```sh
migrate new config [filename] [--force]
```

Generate a JSON config, defaulting to `migration_config.json`. `--force` permits
overwriting an existing file. See [Configuration](./configuration).

## new version

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

Create a version file in the migration directory. `--version` / `-v` overrides the
generated version; `--format` defaults to `sql`. All formats are included by default. See [SQL migrations](./sql-migrations) or [script migrations](./scripting).

## create

```sh
migrate create
```

Idempotently create the history table. User migration files are not applied.

## baseline

```sh
migrate baseline
```

Record pending migrations as baseline without executing them. First confirm the
existing database already matches those files. See [History and baseline](./history).

## up

```sh
migrate up [--dry-run]
```

Apply pending versions in order. Dry-run prints SQL or script plans without executing
user migrations; it may connect to read history. See [Apply migrations](./applying).

## down

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

Reverse versions above the target, keeping the target applied, or reverse all
applied versions. The target and `--all` are mutually exclusive; one is required.
See [Rollback](./rollback).

## status

```sh
migrate status [--output table|json]
```

Compare files and database history. `--output` / `-o` defaults to `table`.
See [status values](./history).

## capabilities

```sh
migrate capabilities [--output text|json]
```

Report build formats, aliases, script verification and transaction limits without
reading config or connecting to a database. `--output` / `-o` defaults to `text`.
See [Database capabilities](./databases).

## version

```sh
migrate version
```

Print the CLI version without reading config or opening a database connection.
This reports the tool version, not the database's latest migration version.
