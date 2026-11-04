# Install and choose a database

The default installation includes SQL, JavaScript, TypeScript and MJS migrations:

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
```

[Release binaries](https://github.com/exc-works/migrate/releases) include the same formats without extra build tags or a build suffix.

Configure `dialect` for PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB or Redshift.
Check your build's formats, database aliases and transaction modes:

```sh
migrate capabilities
migrate capabilities --output json
```

See the [database table](./scripting#limits-and-guarantees) for script support and transaction behavior.
