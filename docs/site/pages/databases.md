# Builds and database capabilities

## SQL and full builds

```sh
# SQL build
go install github.com/exc-works/migrate/cmd/migrate@latest

# Full build: SQL, JS, TS and MJS
go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest
migrate capabilities
migrate capabilities --output json
```

Release downloads without a suffix are SQL-only; `_full` assets include scripts.
Both builds run without downloading an interpreter. npm imports require packages
to be installed beforehand; script execution does not provide Node.js APIs.

## Inspect your binary

SQL dialects include PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse,
TiDB and Redshift. SQL availability and script verification are separate capabilities.
Inspect the output of your binary before choosing a build or database:

- `formats` lists the migration formats present in the build.
- `script_dialect_capabilities` reports script verification, transaction modes and limitations.
- `aliases` lists accepted names; configurations normalize to the canonical dialect.

The [script database matrix](./scripting#limits-and-guarantees) is maintained in the
shared script guide. Experimental support has limited coverage; a skipped external
check does not establish real-instance verification.

## Transactions and rollback

Database engines differ: some DDL implicitly commits, and ClickHouse does not offer
the framework's automatic rollback. An explicitly nontransactional script can leave
partial changes without successful history if it fails. Inspect data and history before retrying.

Review [transaction modes](./scripting#values-and-sql) and the database matrix,
then use `migrate capabilities` to check your installed build's limits.
