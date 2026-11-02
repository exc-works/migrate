# SQL + JS/TS migrations

Default builds remain SQL-only. Full builds embed the JS engine and TS transpiler;
no Node.js installation or runtime download is required. The authoritative contract
is [Script Migrations](spec/07-script-migrations.md).

## Install and inspect

```bash
go install -tags=full github.com/exc-works/migrate/cmd/migrate@latest
migrate capabilities --output json
```

Existing release filenames identify SQL builds. Assets ending in `_full` (before
`.exe` on Windows) include JS/TS. Verify the release SHA-256 and rename the selected
executable to `migrate`. Both variants share configuration and history. SQL builds
refuse known script files instead of skipping required versions.

## Write a migration

```bash
migrate new version normalize_profile --format js
migrate new version backfill_settings --format ts
migrate up --dry-run
migrate up
```

Files follow `V<version>__<description>.sql|js|ts` and share one version sequence.
Do not modify or rename applied migrations. SQL keeps its existing checksums;
scripts use original-source SHA-256.

```js
export function up(db) {
  const update = db.dialect === "postgres"
    ? "UPDATE users SET profile=$1 WHERE id=$2"
    : "UPDATE users SET profile=? WHERE id=?";
  for (const row of db.query("SELECT id, profile FROM users")) {
    const profile = JSON.parse(row.profile);
    profile.displayName = profile.name;
    delete profile.name;
    db.exec(update, [JSON.stringify(profile), row.id]);
  }
}
```

Export synchronous `down(db)` only when reversal is implemented. Missing down
blocks the selected rollback range; an empty function does not reverse an up.
Generated up functions throw until implemented.

JS templates define `MigrationDB` in JSDoc and annotate `db`; TS templates provide
equivalent declarations. Both describe database methods, parameters and result types
for editor assistance. TS transpilation does not perform type checking. Optionally
replace the local declarations with the shared
[migration-api.d.ts](migration-api.d.ts), referenced by a triple-slash path comment.

## Values and SQL

`query(sql, params?)` returns materialized rows. `exec(sql, params?)` returns
`{ rowsAffected: string }`. Use native dialect placeholders and array parameters.
Integer and exact decimal results are strings. NULL is null, binary results are
copied byte arrays, and Go timestamps become precise UTC strings.

Do not round large integers or decimals through JS numbers. JSON.parse can also
round large JSON numeric literals; use a lossless representation. SQLite NUMERIC
affinity may store decimals as floats, which the runtime rejects; choose suitable
storage types when exact decimal preservation is required.

Each bridge call accepts one statement, with an optional trailing terminator and
no embedded semicolons. Bind values containing semicolons. Only direct data/schema
statements listed in the spec are accepted; dynamic SQL, procedure calls, administrative
commands, manual transaction/session control and executable comments are rejected. Use SQL migration
files for compound routine definitions. Alias duplicate result column names.
SQL must start directly with a supported keyword after whitespace. Leading SQL comments
are unsupported because dialects disagree about nesting and line endings. Place annotations
in JS/TS comments.

## Limits and guarantees

Add optional fields to an existing configuration:

```json
{
  "script_timeout": "5m",
  "script_max_rows": 1000,
  "script_max_bytes": 16777216
}
```

Library callers use `Config.ScriptOptions` with `migrate.ScriptOptions`, and build
their application with `-tags=full`. Limit violations fail instead of truncating.
Page large result sets explicitly; paging reads does not create batch commits or
resumable migrations.

Scripts initially support PostgreSQL, MySQL, MariaDB and SQLite. Data operations
and successful history writes share a transaction. Caught database errors remain
fatal. MySQL/MariaDB require transactional tables; implicit DDL commits and other
engine limitations still apply.

There is no Node/npm runtime, runtime module loading, network, filesystem,
environment or timer API. Entrypoints must be synchronous. Treat scripts as trusted
repository code. Cancellation interrupts JS and cancels database operations, but
does not promise hard real-time termination of native calls or per-VM heap isolation.

Status, baseline and dry-run do not evaluate scripts. Actual execution validates
the whole selected range and initializes each module once without DB access.
Dry-run prints a plan, not predicted dynamic SQL. Errors retain source locations
without exposing raw exception payloads, SQL values or credentials.

Serialize migrators targeting the same database. Successful version history is
not an exactly-once guarantee across crashes, implicit commits or concurrency.
