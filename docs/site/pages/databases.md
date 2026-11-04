# Install and choose a database

Choose a build for the migration files you use:

| Files | Install |
| --- | --- |
| SQL | `go install github.com/exc-works/migrate/cmd/migrate@latest` |
| SQL, JavaScript and TypeScript | `go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest` |

For [binary downloads](https://github.com/exc-works/migrate/releases), choose a `_full` asset to run scripts.

Configure `dialect` for PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB or Redshift.
Check your build's formats, database aliases and transaction modes:

```sh
migrate capabilities
migrate capabilities --output json
```

See the [database table](./scripting#limits-and-guarantees) for script support and transaction behavior.
