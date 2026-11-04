# Script transactions and limits

## Managed transactions

Where supported, each script runs in a framework transaction and successful history
writes share that transaction. Do not issue your own BEGIN, COMMIT or ROLLBACK in
managed mode; SQL is forwarded and can break the framework's rollback guarantees.

Handled database errors may recover, but uncaught exceptions, fatal bridge errors,
commit failures and history-write failures fail the migration. Implicit DDL commits,
procedure side effects and engine limits prevent universal rollback guarantees.
Check your [database's execution boundary](./databases#script-support).

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
a defined transaction value must be boolean. Config is evaluated and snapshotted
during actual execution preflight, not by status, baseline or dry-run.

All calls in one explicitly nontransactional step use a dedicated connection. If
you manage a transaction yourself, finish it before returning successfully. Restore
session settings you change; driver reset behavior varies. Do not rely on session
state carrying across migrations. On failure the dedicated connection is discarded.

Nontransactional failures can leave committed changes without successful history.
Inspect data and history before retrying; a Down function also has no automatic
rollback guarantee on a nontransactional engine.

## Bound execution and result sizes

Add these optional fields to your project configuration:

```json
{
  "script_timeout": "5m",
  "script_max_rows": 1000,
  "script_max_bytes": 16777216
}
```

Limit violations fail instead of truncating results. Page large reads explicitly;
paging does not create batch commits or resumable migrations. Cancellation interrupts
JS and cancels database operations, but native calls are not guaranteed to terminate
in hard real time.

Go callers configure `Config.ScriptOptions` with `migrate.ScriptOptions` and build
their application with `-tags=full`. See [Use from Go](./go-library).
