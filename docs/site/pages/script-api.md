# Script database API

The synchronous `db` argument exposes the canonical `dialect`, `query(sql, params?)`
and `exec(sql, params?)`. Download [migration-api.d.ts](./assets/migration-api.d.ts)
for editor declarations.

## Query rows and execute writes

This example uses SQLite placeholders:

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected is a decimal string, for example "1".
  }
}
```

`query` materializes rows and returns only the first result set. Alias duplicate
column names. `exec` returns `{ rowsAffected: string }`, preserving the driver's count,
including a count of `"0"`; it does not infer the number of intended changes.

## Bind values

Pass parameters as an array using the dialect's native placeholders: `$1` for
PostgreSQL/Redshift, `?` for MySQL/MariaDB/SQLite/TiDB/ClickHouse, `:1` for Oracle,
and `@p1` for MSSQL. `db.dialect` is canonical even when configuration uses an alias.

Integer and exact decimal results are strings. NULL becomes `null`, binary values
are copied byte arrays, and Go timestamps become precise UTC strings. Do not convert
large integers or exact decimals through JavaScript numbers. `JSON.parse` can also
round large numeric literals; choose a lossless representation where needed.

SQLite NUMERIC affinity may store decimals as floats, which the runtime rejects.
Choose storage types that preserve the required precision. Vendor-specific values
may need explicit SQL conversion; see [Database capabilities](./databases#value-conversion).

## Native SQL behavior

The bridge forwards SQL unchanged. It does not split or validate batches, stored
procedures or administrative statements, and does not interpret client separators
such as `GO`. Preserve native terminators, including the semicolon required by MERGE.
MySQL batches may require driver DSN options. Permissions and driver support still apply.

## Handle database errors

Driver errors are sanitized exceptions that scripts may catch. They do not expose
raw vendor details or cause chains. Further database calls are allowed after a
catchable error, but catching does not repair a transaction. Use dialect-appropriate
savepoints or rollback; rethrow if the managed transaction cannot be repaired.
Returning normally after catching an error accepts that result.

Execution-context cancellation, the overall script timeout, invalid bridge values
and resource-limit failures remain fatal even when caught, and prevent further DB
calls. A driver-reported timeout alone is catchable. Review
[Transactions and limits](./script-transactions) before implementing recovery logic.
