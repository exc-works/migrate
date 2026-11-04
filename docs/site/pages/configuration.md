# Configuration

Run commands from your migration project directory. By default, migrate searches YAML first and falls back to JSON. The current
initializer produces JSON; generate a starting file with:

```sh
migrate new config
```

Use `migrate new config dev.json` for a different filename. Existing files are
preserved unless you explicitly pass `--force`.

## YAML and JSON discovery

The configuration reader accepts YAML and JSON. Without an explicit `-c`, it checks
`migration_config.yml`, `migration_config.yaml`, then `migration_config.json` in that
order. Only a missing file permits fallback; invalid or unreadable YAML is an error.
An explicit `-c` reads exactly that file. YAML supports native `#` comments; JSON
remains strict. Fields, defaults, environment templates and relative-path rules are
the same in both formats. Nested SSH and proxy fields use the same names.

The reader also recognizes `data_sources` with shared SSH defaults. Batch command
execution is delivered by the subsequent orchestration layer; this intermediate
layer rejects multi-source database commands explicitly before opening a connection.

## Connect to a database

This PostgreSQL example keeps the connection string in an environment variable:

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "${DATABASE_DSN}",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

Set `DATABASE_DSN` through your shell or deployment secret store before running
database commands. For a local SQLite project, use `"dialect": "sqlite"` and
`"data_source_name": "./demo.sqlite"` instead.

| Field | Purpose | Default when omitted |
| --- | --- | --- |
| `schema_name` | Migration history table name | `migration_schema` |
| `dialect` | Database dialect; see [accepted names](./databases#aliases) | `postgres` |
| `data_source_name` | Driver connection string | Required |
| `working_directory` | Base for a relative migration directory | Current directory |
| `migration_source` | Directory containing version files | `migrations` |
| `migrate_out_of_order` | Allow pending versions below the applied version | `false` |
| `logger_level` | Logging level | `info` |

## Environment templates

`data_source_name` accepts `${KEY}` for a required variable and `${KEY:default}`
for a fallback when the variable is absent. For example, use `${DB_NAME:app}` for
a nonsecret database name. Keep credentials in the environment.

## Select a project or configuration

```sh
migrate -c ./configs/dev.json status
migrate -w ./deploy up
```

`-c` selects a config file. `-w` selects the working directory, resolves the config
path from there, and overrides `working_directory` for migration discovery.
A relative `migration_source` is resolved from the working directory, not automatically
from the directory containing a custom `-c` file.

Script limits are documented in [Transactions and limits](./script-transactions).
For bastion connections, add [SSH tunnel settings](./ssh-tunnels).
Once configured, [write your first SQL migration](./sql-migrations).
