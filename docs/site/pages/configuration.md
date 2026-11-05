# Configuration

Generate a configuration file in your project directory:

```sh
migrate new config
```

Open `migration_config.yml`, fill in your database type and connection details, then run migration commands.

## Connect to a database

This configuration connects to PostgreSQL using the `DATABASE_DSN` environment variable:

```yaml
# Database type
dialect: postgres
# Set DATABASE_DSN in your shell or deployment environment before running commands
data_source_name: "${DATABASE_DSN}"
# Directory containing migration files
migration_source: migrations
# Table that records applied versions
schema_name: migration_schema
```

| Field | Purpose | Default when omitted |
| --- | --- | --- |
| `dialect` | Database type; see [supported names](./databases#aliases) | `postgres` |
| `data_source_name` | Database connection string; choose this or `data_sources` | Required for one database |
| `data_sources` | Named databases and their connection settings; see [multiple databases](#multiple-databases) | Disabled |
| `migration_source` | Directory containing migration files | `migrations` |
| `schema_name` | Migration history table name | `migration_schema` |
| `working_directory` | Starting directory for finding migration files | Current directory |
| `migrate_out_of_order` | Apply pending versions below the applied version | `false` |
| `logger_level` | `debug`, `info`, `warn` or `error`; see [logging levels](#logging) | `info` |

### Connection examples for each database

Choose your database and copy its `dialect` and `data_source_name` into your configuration.
Replace the host, port, database name or Oracle service name with your own values,
and set `DB_USER` and `DB_PASSWORD` before running commands. For SQLite, use the database file path directly.

::: code-group

```yaml [PostgreSQL]
dialect: postgres
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@db.example:5432/app?sslmode=require"
```

```yaml [MySQL]
dialect: mysql
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [MariaDB]
dialect: mariadb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [SQLite]
dialect: sqlite
data_source_name: "./demo.sqlite"
```

```yaml [Oracle]
dialect: oracle
data_source_name: "oracle://${DB_USER}:${DB_PASSWORD}@db.example:1521/FREEPDB1"
```

```yaml [SQL Server]
dialect: mssql
data_source_name: "sqlserver://${DB_USER}:${DB_PASSWORD}@db.example:1433?database=app"
```

```yaml [ClickHouse]
dialect: clickhouse
data_source_name: "clickhouse://${DB_USER}:${DB_PASSWORD}@db.example:9000/app"
```

```yaml [TiDB]
dialect: tidb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:4000)/app?parseTime=true"
```

```yaml [Redshift]
dialect: redshift
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@cluster.example:5439/app?sslmode=require"
```

:::

For URL connection strings, URL-encode special characters such as `@`, `:` and `/` in usernames
or passwords first; environment substitution does not encode them. The `user:password@tcp(...)`
format used by MySQL, MariaDB and TiDB does not use URL encoding.
Set TLS/SSL options according to your database service's connection requirements.

### Logging levels {#logging}

`logger_level` controls which messages are shown and defaults to `info`:

| Value | Messages shown |
| --- | --- |
| `debug` | Debug, information, warnings and errors |
| `info` | Information, warnings and errors |
| `warn` | Warnings and errors; `warning` is also accepted |
| `error` | Errors only |

```yaml
logger_level: debug # Show debug messages when troubleshooting
```

## Use environment variables {#environment-variables}

Connection strings accept `${KEY}` and `${KEY:default}`:

- `${DATABASE_DSN}` reads `DATABASE_DSN`; the command fails if it is unset.
- `${DB_NAME:app}` reads `DB_NAME`, using `app` if it is unset.

Use the same configuration in development, testing and production by supplying different
variable values in each environment. Provide database credentials through environment variables
or your deployment platform's secret store.
SSH hosts, usernames and proxy addresses also accept these templates; see [SSH tunnels and proxies](./ssh-tunnels).

## Configuration file search order

Without `-c`, migrate searches the current working directory in this order and uses the first file it finds:

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

Files are not merged. If the selected file contains an error or cannot be read, fix that file before running again.

Use `-c` to select a file directly. If that file is missing, the command fails without searching for another one:

```sh
migrate -c ./configs/dev.yml status
```

## Generate configuration for your setup

Choose a command for your setup, then fill in the generated connection settings:

| Setup | Command |
| --- | --- |
| One database | `migrate new config` |
| Custom filename | `migrate new config dev.yml` |
| Multiple databases | `migrate new config --multi` |
| Four databases | `migrate new config --multi --shards 4` |
| SSH with SSH Agent | `migrate new config --ssh-tunnel` |
| SSH with a private key | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| SSH with a password | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 proxy to an SSH bastion | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT proxy to an SSH bastion | `migrate new config --ssh-tunnel --proxy http` |

Combine options as needed. For example, generate shared SSH and proxy settings for four databases:

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

Existing files are preserved unless you pass `--force`.
Run `migrate new config --help` for all options.

## Apply one set of migrations to multiple databases {#multiple-databases}

For shards or tenant databases that need the same changes, list each database under
`data_sources` instead of using the top-level `data_source_name`. All databases share the
database type, migration directory and history table name.

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # Connection string for the first database
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # Connection string for the second database
```

Set `SHARD_0_DSN` and `SHARD_1_DSN` before running commands. Configurations generated with
`--multi --shards N` start at `shard_0` and reference `SHARD_0_DSN`, `SHARD_1_DSN`, and so on.

![Migrations from one directory are applied to shard_0 and shard_1, each with its own history, then results are collected.](./assets/diagrams/multiple-databases-en.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

These commands automatically process every configured database. Each database selects pending
versions from its own history and applies them in version order. Successful recorded versions
are skipped on later runs. Run `new version` once to create a file shared by all databases.

By default, up to 8 databases are processed at once. Adjust `--parallel` as needed;
use 1 to process databases one at a time:

```sh
migrate up --parallel 4
migrate status --output json
```

If one database fails, the others continue. Results are reported per database, and the command
exits nonzero if any database fails. This is not a transaction across databases: a failure does
not undo changes already completed elsewhere. Pressing Ctrl+C stops new databases from starting
and cancels running operations.

For databases behind the same bastion, add `ssh_tunnel` at the top level. To use different SSH
settings for one database, provide a complete `ssh_tunnel` in that database's entry; it replaces
the shared settings. Omitting it or setting it to `null` uses the shared settings.
See [SSH tunnels and proxies](./ssh-tunnels) for fields.

Configure each physical database once and run only one migration command against it at a time.
Keep migration files and script dependencies unchanged during execution.
After a failure, inspect that database's data and history before retrying.

## Working directory and relative paths

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` selects the working directory. Configuration files are located from there, and the option
overrides `working_directory` in the file. A relative `migration_source` is resolved from the
working directory. Passing only `-c` does not change it to the configuration file's directory.

For example, running `migrate -c configs/dev.yml up` from the project root with
`migration_source: migrations` uses the project's root-level `migrations` directory.
Relative SSH key paths are resolved from the configuration file's directory instead.

Once configured, [write your first SQL migration](./sql-migrations) or set [script execution limits](./script-transactions).
