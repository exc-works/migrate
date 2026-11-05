# Script transactions and limits

## Default transaction behavior

Where supported, each script runs in a framework transaction and successful history
writes share that transaction. Do not issue your own BEGIN, COMMIT or ROLLBACK in
managed mode; SQL is forwarded and can break the framework's rollback guarantees.

A failed script causes the transaction to roll back where supported. Some databases
implicitly commit DDL, and others, such as ClickHouse, do not support transactions,
so changes may remain after a failure. Check [transaction behavior for your database](./databases#script-support).

## Explicitly disable the framework transaction

For SQL that requires execution outside a framework transaction, export:

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: this statement cannot run inside a transaction block.
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

This setting applies to up and down. Omitted config/transaction defaults to true;
a defined transaction value must be boolean. `status`, `baseline` and dry-run do not run scripts,
so they do not validate the configuration exported at runtime.

All calls in one explicitly nontransactional step use a dedicated connection. If
you manage a transaction yourself, finish it before returning successfully. Restore
session settings you change; driver reset behavior varies. Do not rely on session
state carrying across migrations. On failure the dedicated connection is discarded.

Nontransactional failures can leave committed changes without successful history.
Inspect data and history before retrying; a Down function also has no automatic
rollback guarantee on a nontransactional engine.

## Bound execution and result sizes

Add these optional fields to your project configuration:

```yaml
script_timeout: 5m # Execution timeout for each script
script_max_rows: 1000 # Maximum rows returned by one query
script_max_bytes: 16777216 # Maximum bytes returned by one query (16 MiB)
```

Limit violations fail instead of truncating results. Page large reads explicitly;
paging does not create batch commits or resumable migrations. Canceling interrupts the script and requests cancellation of database operations;
the database may take some time to finish stopping a running statement.

Go callers configure `Config.ScriptOptions` with `migrate.ScriptOptions`. See [Use from Go](./go-library).
