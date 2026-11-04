# Write script migrations

Use JavaScript or TypeScript for data transformations that are easier to express
in code. Scripts run synchronously in an embedded engine without
a Node.js runtime. Follow the [installation instructions](./installation) before continuing.

## Create a script

```sh
migrate new version normalize_names --format ts
```

Supported formats are `js`, `ts` and `mjs`; `.mjs` has the same runtime restrictions
as `.js`. Files use `V<version>__<description>.<format>` and share their version
sequence and history with SQL files.

## Implement the entrypoints

For a SQLite `users` table containing a text `name` column:

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

Generated up functions throw until implemented. Entrypoints must be synchronous:
async functions, generators and Promise results are unsupported. Ordinary synchronous
return values are ignored. SQL placeholders are dialect specific; see the
[database API](./script-api).

Export `down(db)` only when you have implemented reversal. Trimming whitespace loses
information, so this example deliberately has no down function. A rollback range
containing a script without down is blocked. Preserve original values if you need
that transformation to be reversible.

## Preview and run

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run compiles dependencies and prints a plan without evaluating scripts.
Actual execution prepares the entire selected range before migration writes;
module initialization happens without database access. Each migration has its own
VM and module state.

## Editor support and runtime boundaries

Generated JS templates annotate `db` with JSDoc; TS templates include equivalent
type declarations. Transpilation does not type-check. The shared
[migration-api.d.ts](./assets/migration-api.d.ts) can replace local declarations
and be referenced with a triple-slash path comment.

Scripts are trusted repository code. There are no Node built-ins, runtime module
loading, network, filesystem, environment or timer APIs. Cancellation interrupts
JavaScript and database operations but does not guarantee hard real-time termination
of native calls or per-VM heap isolation.

Continue with [shared modules](./modules), [transactions and limits](./script-transactions),
or the [runnable examples](./examples).
