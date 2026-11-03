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

Files follow `V<version>__<description>.sql|js|ts|mjs` and share one version sequence.
The `.mjs` extension is a `.js` alias with the same execution restrictions; generate it
with `--format mjs`. The default remains SQL. It does not enable the Node.js runtime.
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
Generated up functions throw until implemented. Ordinary synchronous return values
are ignored, so `export const up = db => db.exec(...)` is supported. Promise results
and async/generator entrypoints remain unsupported; returned work is not awaited.

JS templates define `MigrationDB` in JSDoc and annotate `db`; TS templates provide
equivalent declarations. Both describe database methods, parameters and result types
for editor assistance. TS transpilation does not perform type checking. Optionally
replace the local declarations with the shared
[migration-api.d.ts](migration-api.d.ts), referenced by a triple-slash path comment.

## Shared modules

Disk-based migrations can use static relative imports and re-exports from `.js`,
`.ts` and `.mjs` files. Paths resolve from each importing file, including shared
files outside the migration directory. Helpers are dependencies, not migration entries.

```ts
import { normalize } from "../shared/normalize.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

The compiler bundles dependencies before migration writes. Each migration keeps its
own VM and module state. Dependency failures in a later selected migration prevent
execution of earlier steps; the whole selected rollback range is also prepared first.
Dry-run compiles dependencies without evaluating them; status/baseline inspect metadata.

Checksums cover only original entry bytes. Editing a helper does not cause checksum
mismatch or replay an applied migration. Keep historical helper behavior when old
migrations must remain reproducible; add a new helper version when needed.

Installed npm packages and virtual-source imports are enabled by later rollout layers.
Node built-ins, native addons, dynamic import, dynamic require and asynchronous
entrypoints remain unsupported. No runtime module loader is exposed.

## Values and SQL

`query(sql, params?)` returns materialized rows. `exec(sql, params?)` returns
`{ rowsAffected: string }`. Use native dialect placeholders and array parameters.
Integer and exact decimal results are strings. NULL is null, binary results are
copied byte arrays, and Go timestamps become precise UTC strings.

Do not round large integers or decimals through JS numbers. JSON.parse can also
round large JSON numeric literals; use a lossless representation. SQLite NUMERIC
affinity may store decimals as floats, which the runtime rejects; choose suitable
storage types when exact decimal preservation is required.

SQL is passed unchanged to the driver. Native comments, semicolons, batches,
procedure definitions/calls and administrative SQL follow database permissions and
driver support. The bridge does not split or validate SQL and does not interpret
client separators such as GO. Preserve native terminators (including MERGE's `;`).
MySQL batches may require driver DSN options. Alias duplicate column names; `query`
returns only the first result set. Exec counts preserve the driver's batch semantics.

Each script uses a framework transaction by default where supported. For SQL that
must run outside it, manual transactions or session control, export:

```js
export const config = { transaction: false }; // applies to up and down
export function up(db) {
  db.exec("BEGIN TRANSACTION"); // use your dialect's native syntax
  db.exec("UPDATE users SET profile=@p1", ["changed"]);
  db.exec("COMMIT TRANSACTION");
}
```

Omitted config/transaction defaults to true. A defined transaction value must be a
boolean. Config is evaluated and snapshotted during actual execution preflight;
status, baseline and dry-run do not evaluate it. All calls in an explicitly
nontransactional step share one dedicated connection. On success, it is returned to
the pool for reuse; database/sql and driver validity/reset hooks handle unhealthy
connections. Driver reset behavior varies and may preserve session settings, so restore
any settings you change and do not depend on session state across migrations.
Finish every manual transaction before returning successfully. Unfinished manual transactions on successful return are invalid
usage. An uncaught or fatal failure discards the dedicated connection, closing any
still-uncommitted transaction.

Driver errors from execution, affected-row retrieval, query metadata, scan, iteration
and result close are sanitized exceptions that scripts can catch and handle. Further
DB calls are allowed; the engine does not classify vendor errors or declare a
connection healthy. Catching an error and returning normally means the author accepts
it. Recovery calls such as ROLLBACK can run afterwards, but catching an error does not
automatically repair a transaction. Use a savepoint or manual rollback as required by
the dialect. If the entire managed transaction ended and cannot be repaired, rethrow;
use explicit manual mode for full transaction retries. Uncaught exceptions, commit
failures and history-write failures still fail the migration.
Cancellation of the actual execution context, the overall script timeout,
parameter/result validation and resource limits remain fatal even when caught, and
prevent further DB calls. A driver-reported timeout alone is catchable. Driver errors
never expose raw details or cause chains to JS.
Nontransactional failures may leave committed changes without successful history;
errors warn about partial state. Inspect data/history before retrying. In managed
mode, do not issue your own BEGIN/COMMIT/ROLLBACK: these are forwarded, and can break
the framework's rollback and history guarantees. Migration authors own their SQL.

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

Scripts support the following databases. Experimental dialects are executable, with
limited coverage of versions, types and deployment configurations. Text and JSON
`capabilities` output report verification, transaction mode and limitations. JSON
preserves `script_dialects` and adds `script_dialect_capabilities`. Selected plans
containing experimental scripts warn once without an extra confirmation step;
status, baseline and SQL-only plans do not issue this warning.

| Database | Status | Placeholders | Execution boundary |
| --- | --- | --- | --- |
| PostgreSQL | Verified | `$1` | Framework transaction |
| MySQL / MariaDB | Verified | `?` | Requires transactional tables; DDL may implicitly commit |
| SQLite | Verified | `?` | Framework transaction |
| TiDB | Experimental | `?` | CI uses Unistore; distributed TiKV remains unverified; DDL may implicitly commit |
| Redshift | Experimental; no real-instance verification | `$1` | Transactions depend on the statement; not full PostgreSQL equivalence |
| Oracle | Experimental; limited Oracle Free coverage | `:1` | DML uses a framework transaction; DDL implicitly commits |
| ClickHouse | Experimental | `?` | Nontransactional; no automatic rollback |
| MSSQL / SQL Server | Experimental | `@p1` | Framework transaction; Azure SQL instances remain unverified |

Names are trimmed and case-insensitive. Configuration, CLI, Go facade and script
capability checks share a single alias catalog:

| Canonical name | Accepted aliases |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | No extra aliases |

MySQL, MariaDB and TiDB remain separate dialects despite sharing a driver.
`db.dialect` always contains the canonical name: an `azuresql` input yields `mssql`.
Capabilities text and JSON `aliases` fields list accepted aliases.

MSSQL MONEY/SMALLMONEY values remain exact strings. UNIQUEIDENTIFIER retains the
driver's binary representation as copied bytes; use `CAST(id AS VARCHAR(36))` for
formatted GUID text and bind GUID strings on writes. Rowversion/timestamp is binary;
DATETIMEOFFSET and other time values follow the common time conversion contract.
Native T-SQL batches and procedure calls are supported by the driver.

In managed mode, script operations and successful history writes share the framework
transaction. Handled SQL errors can recover; uncaught errors and fatal bridge failures
abort execution. User transaction control,
procedure side effects, implicit DDL commits and engine limitations prevent universal
rollback guarantees. Nontransactional mode provides no data/history atomicity.

Oracle uses pinned go-ora v3. NUMBER values remain exact strings; RAW becomes bytes
and CLOB becomes text. BLOB results without type metadata require an explicit
projection, such as bounded `DBMS_LOB.SUBSTR` to RAW or `RAWTOHEX` to hex text.
Do not infer ambiguous binary values. CLI connections disable v3 fast-login cookies
by default for reliable reconnection, preserving explicit `FAST LOGIN` options.
Go facade callers constructing their own DB should set `FAST+LOGIN=false` in the DSN.
Native PL/SQL blocks and calls follow driver support. Redshift
specific types may have unknown pgx metadata: explicitly convert them to supported
scalars. For VARBYTE, write hex text with `FROM_HEX($1)` and read hex text with
`FROM_VARBYTE(payload, 'hex')` rather than assuming BYTEA compatibility.

ClickHouse integers (including 128/256-bit values) and decimals remain exact strings.
Unadapted values such as Array, Tuple, Map, UUID and IP require explicit SQL conversion
to supported scalars. String must contain valid UTF-8; encode arbitrary binary data
as hex text. `rowsAffected` preserves the driver count as a decimal string, including a ClickHouse driver value of `"0"`; the framework does not infer a replacement. Use statement-local
`SETTINGS mutations_sync=2` when mutation completion is required; session settings
must follow driver support; explicitly disable framework transactions for session control.

A ClickHouse failure, timeout or history-write error can leave successful prior
writes without successful history. Errors report possible partial changes. Down
also has no automatic rollback guarantee. Inspect data and history and recover
according to their actual state before retrying; do not blindly rerun or baseline.

Run container verification with the following command. External TiDB/ClickHouse/MSSQL
DSNs take precedence over disposable containers:

```bash
INTEGRATION_DB=clickhouse,tidb,mssql go test -tags=integration,full ./integrationtest/...
```

Optional Oracle/Redshift checks use `INTEGRATION_ORACLE_DSN` and
`INTEGRATION_REDSHIFT_DSN`. Missing DSNs cause skipped checks, not verification.

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
