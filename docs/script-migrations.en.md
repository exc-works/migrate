# JavaScript and TypeScript migrations

Use scripts to transform data and execute SQL through the `db` interface. The examples below use SQLite.

## Install and inspect

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate capabilities
```

The default installation and release binaries embed the script engine and run SQL, JS, TS and MJS without extra build tags.

## Write a migration

Create a file in your configured migration project:

```bash
migrate new version create_projects --format js
```

Replace its contents with:

```js
export function up(db) {
  db.exec("CREATE TABLE projects (id INTEGER PRIMARY KEY, name TEXT NOT NULL)");
}

export function down(db) {
  db.exec("DROP TABLE projects");
}
```

Preview, apply and check the result:

```bash
migrate up --dry-run
migrate up
migrate status
```

Write changes in a synchronous `up(db)` function and reversal in `down(db)`.
Rollback checks every selected script for a `down` function.
Choose `--format ts` for TypeScript or `--format mjs` for MJS.
Assign each migration its own version; SQL and scripts share one version sequence.

Templates provide editor type hints. Download [migration-api.d.ts](migration-api.d.ts)
for shared modules and configure TypeScript in your project when you want type checking with `pnpm exec tsc --noEmit`.

## Shared modules

Save a helper in `shared/normalize.ts`:

```ts
export const normalize = (value: string) => value.trim();
```

Import it by relative path. This example uses an existing `users(id, name)` table:

```ts
import { normalize } from "../shared/normalize.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

To use npm packages, run `pnpm add <package>`, then use static package-name imports. Commit the lockfile and run `pnpm install --frozen-lockfile` during deployment.
Choose packages that use synchronous JavaScript and the `db` API for database access.
See the [runnable examples](../examples/script-imports/README.en.md) for a complete project.

The tool bundles dependencies for the selected range before executing migrations.
Use `--dry-run` to check bundling and the migration plan.
Checksums cover the entry file. Preserve published entries, helpers and lockfiles for replay;
add a new version when changing behavior.

## Embedded and in-memory modules

Use `FSSource` to load an `embed.FS` or in-memory filesystem in a Go application.
Point `DependencyDir` at the project containing installed packages:

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

Relative imports use the supplied filesystem; package-name imports use `DependencyDir`.
For `StringSource`, provide helpers through `WithModules.FS`, or use `Directory` for local files.
See the [embedded example](../examples/script-imports/embedded/main.go) for the full configuration.

## Values and SQL

| Task | Call |
| --- | --- |
| Query rows | `db.query(sql, [parameters])` returns an array of rows |
| Execute SQL | `db.exec(sql, [parameters])` returns an object with a `rowsAffected` string |
| Choose dialect syntax | `db.dialect` returns the canonical database name |

Read the count with `db.exec(sql, parameters).rowsAffected`.
Use native database placeholders and unique query-column names.
SQL follows the driver's native syntax; `query` reads the first result set.

Integers and exact decimals are returned as strings; preserve their exact representation when calculating or storing them.
Times use UTC strings, NULL uses `null`, and binary results use byte arrays.
Choose a precision-preserving SQLite storage type when storing exact decimals.

On transactional databases, the tool manages the transaction and writes data and successful history together.
For statements outside transactions, such as SQLite `VACUUM`, or for manual transaction control, export:

```js
export const config = { transaction: false };
```

This applies to both up and down. Finish manual transactions and restore changed session settings before returning.
When handling SQL exceptions, recover the transaction with a savepoint or rollback; rethrow when recovery fails.

## Limits and guarantees

Set execution time and per-query result limits in the configuration:

```json
{
  "script_timeout": "5m",
  "script_max_rows": 1000,
  "script_max_bytes": 16777216
}
```

Reaching a timeout or result limit terminates the migration. Adjust the query or limits before retrying.
Use ordered paging for large tables. Paged reads follow the migration's transaction scope.
Go applications set these values through `migrate.Config.ScriptOptions`. The script engine is included by default.

Run `migrate capabilities --output json` to inspect database status, aliases and transaction modes for your build.

| Database | Script status | Placeholders | Transactions |
| --- | --- | --- | --- |
| PostgreSQL | Verified | `$1` | Framework transaction |
| MySQL / MariaDB | Verified | `?` | Use transactional tables; DDL may implicitly commit |
| SQLite | Verified | `?` | Framework transaction |
| TiDB | Experimental; CI uses Unistore | `?` | Validate on your TiKV deployment; DDL may implicitly commit |
| Redshift | Experimental; real-instance verification pending | `$1` | Check transaction support for each statement |
| Oracle | Experimental; Oracle Free verified | `:1` | Transactional DML; DDL implicitly commits |
| ClickHouse | Experimental | `?` | Nontransactional |
| MSSQL / SQL Server | Experimental | `@p1` | Framework transaction; validate on your Azure SQL instance |

Convert database-specific values into supported scalars in SQL:

| Value | Approach |
| --- | --- |
| MSSQL GUID | Read text with `CAST(id AS VARCHAR(36))` |
| Oracle BLOB | Use bounded `DBMS_LOB.SUBSTR` to read RAW, or `RAWTOHEX` for hex text |
| Redshift VARBYTE | Write with `FROM_HEX($1)` and read with `FROM_VARBYTE(payload, 'hex')` |
| ClickHouse Array / Map / Tuple | Convert to strings; use `SETTINGS mutations_sync=2` when waiting for mutations |

After a nontransactional failure, inspect data and `migrate status`, repair partial changes, then retry.
Use a single migration job per database in your deployment pipeline.

See the [type declarations](migration-api.d.ts) for interfaces and the [script examples](../examples/script-imports/README.en.md) for complete projects.
