# CLI reference

Use `migrate --help` or `migrate <command> --help` to confirm the flags in your
installed version. Use `migrate version` to identify the binary.

## Global flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `--config`, `-c` | Search `.yml`, `.yaml`, then `.json` | Explicit config path; otherwise discover `migration_config` with those extensions |
| `--working-dir`, `-w` | Current directory | Resolve config from this directory and override migration working directory |
| `--non-interactive` | `false` | Fail instead of requesting secrets or approval for unknown SSH host keys |

## new config

```sh
migrate new config [filename] [--format yaml|json] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

Generate a commented YAML config, defaulting to `migration_config.yml`.
Use `--format json` or a `.json` filename for strict JSON. `--force` permits
overwriting an existing file.

`--multi` generates two sources named `shard_0` and `shard_1`; `--shards N`
requires `--multi` and sets a positive source count, numbered from 0.
`--ssh-tunnel` adds shared SSH defaults, and `--proxy socks5|http` adds an upstream
proxy. Authentication, address and credential-reference flags are listed by
`migrate new config --help`. See [Configuration](./configuration) for examples
and flag dependencies.

## new version

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

Create a version file in the migration directory. `--version` / `-v` overrides the
generated version; `--format` defaults to `sql`. See [SQL migrations](./sql-migrations) or [script migrations](./scripting).

## Multiple databases

With `data_sources` configured, create/baseline/status/up/down visit every source.
Each accepts `--parallel N` (default 8). One failure does not cancel other targets;
results are grouped by source and any failure yields a nonzero exit. See
[configuration](./configuration) for shared SSH settings and JSON result shapes.

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
