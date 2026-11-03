# 07 Script Migrations

## Distribution and compatibility

- The default build remains SQL-only. The `full` build tag enables SQL, JS and TS.
- Both builds use the same commands, configuration, drivers and history table.
- Existing release filenames continue to identify the SQL build; full assets add `_full`.
- Execution never downloads a runtime. Full builds embed Goja and esbuild without cgo.
- Pin dependencies compatible with the module's minimum Go version.
- `capabilities [--output json]` reports formats without configuration or a connection.
- Preserve the existing `version` output and existing SQL source hashes.

## Files and planning

- Discover direct children named `V<version>__<description>.sql`, `.js`, `.ts` or `.mjs`.
- Directory, string, combined and `fs.FS` sources share the same format rules.
- All formats share version ordering, duplicate detection and the existing history table.
- Scripts use the lowercase hex SHA-256 of original source bytes; SQL retains MD5.
- Discover all known formats even in SQL-only builds. Never silently skip a script.
- Check format support before any migration or baseline history writes.
- Before execution, validate history and prepare every selected migration in the plan.
- Compilation does not execute script code. For actual up/down, initialize all selected
  modules in isolated VMs without a DB capability and validate exports before the first
  migration transaction. Keep those instances for execution; do not evaluate modules twice.
- For down, validate the whole selected rollback range before rolling back anything.
- Applied/baseline migrations and disabled out-of-order migrations are not executed.
- Unsupported script dialects fail explicitly; existing SQL dialects remain supported.

## Script contract

The `.mjs` extension is a JS alias with identical compilation, execution and restrictions;
full builds advertise it in capabilities. SQL-only builds discover it and reject execution
before writes. It does not enable Node.js or a separate runtime module loader.

Scripts are UTF-8 files exporting synchronous `up(db)` and optional `down(db)`.
Missing down is irreversible, not a successful no-op. The runtime validates the entry
before database writes; invalid exports anywhere in a rollback plan must not cause
partial rollback. Ordinary synchronous return values are ignored without exporting or
inspecting object properties. Promise results and async/generator entrypoints remain
unsupported; the runtime does not await or consume returned work.

ES module export syntax and TS syntax are transformed internally into a VM-loadable
module. TS is transpiled without type checking; provide declarations for editor
assistance. Do not advertise full Node.js compatibility.

### Module import rollout

The implementation stack is recorded in
[08-script-modules-plan.md](08-script-modules-plan.md). Full builds implement
static disk-relative, installed-package and virtual-source imports under the
following contract.

- Bundle static local imports and re-exports from JS/TS/MJS, installed npm packages
  and statically resolvable CommonJS dependencies for execution in Goja. Installation
  is the author's build/deployment step; migrate does not install or download packages.
- Disk-relative paths resolve from their importing file, including paths outside the
  migration directory. Bare package names use esbuild's installed-package resolver.
  Use the neutral platform, main-field precedence `module`, then `main`, and explicit
  empty custom conditions; import/require/default conditions remain applicable.
  Preserve exports precedence and esbuild's extension rules. No automatic project
  tsconfig loading or TypeScript path aliases are introduced.
- Virtual sources resolve relative modules within their supplied filesystem and may
  explicitly provide a disk dependency directory for installed packages. They never
  implicitly use the process working directory. Fully virtual node_modules package
  resolution is outside this rollout. Combined sources preserve each entry's context.
- Keep the existing source interface and SourceFile field layout compatible. Legacy
  sources without module context continue to support standalone scripts.
- Compile the entry bytes loaded for hashing and retain each selected compiled
  program for execution. Never reload modules during execution or share VM state
  between migrations. Resolve and validate the selected plan before migration writes;
  dry-run compiles without evaluating modules, while status/baseline inspect metadata.
- Entry-source SHA-256 remains unchanged. Dependencies, package metadata, lockfiles
  and bundle output are excluded; changing them does not invalidate checksums or replay
  applied migrations. Authors maintain historical dependency behavior and versions.
- Node built-ins, native addons, dynamic import, runtime dynamic require, top-level
  await and asynchronous migration entrypoints remain unsupported. Resolving a package
  does not certify that its runtime requirements are available in Goja.
- Diagnostics identify logical module names and import locations without exposing
  source excerpts, absolute local paths or raw filesystem errors.

`db` is a narrow object with `dialect`, `query(sql, params?)` and `exec(sql, params?)`.
Statements use native dialect placeholders and bound parameters; do not concatenate
values into SQL. Each call forwards nonempty SQL unchanged to the driver; it is not
annotated migration text and the bridge does not interpret client separators such as GO.
No database object is exposed during module initialization.

Scripts are trusted repository code with the configured account's database permissions.
The bridge does not implement a SQL keyword allowlist/denylist, statement splitter or
SQL syntax validator. Comments, semicolons, native batches, administrative statements,
functions/procedure definitions and calls are forwarded as supplied; support depends
on the driver and database. Preserve native terminators, including MERGE's semicolon.
Use bound values; this API is not a SQL sandbox. SQL files retain their existing parser.
`query` materializes only the first result set; driver-reported exec counts retain their
native meaning for batches. Use separate calls when multiple result sets are needed.

A module may export `config = { transaction: false }` to disable the framework
transaction for both up and down. Omitted config/transaction defaults to true; supplied
config must be an object and a defined transaction value must be a boolean. Validate
and snapshot it while initializing every selected module before execution begins.
Entrypoint mutations cannot change the prepared policy. Metadata-only and dry-run
commands do not evaluate config or validate dynamically computed export values.

## Database values

- Integer results and driver-reported affected-row counts are decimal strings, including
  small integers and zero. Do not override a driver count with a dialect-specific value.
- DECIMAL/NUMERIC results are exact strings. Never round them through float64.
- Floating point values are finite JS numbers; booleans and text retain their types.
- NULL is `null`. `undefined`, arbitrary objects and unsafe numeric inputs are rejected.
- Binary values are copied byte arrays with explicit validation on bind.
- Driver `time.Time` values become RFC3339Nano strings, preserving the represented instant.
  Strings returned by drivers stay strings. Timezone-free values retain driver semantics;
  do not claim to recover missing timezone information or use JS Date implicitly.
- Inspect column metadata when drivers return bytes for text/numeric values. Ambiguous
  or unsupported values fail instead of silently treating text as binary or losing precision.
- Reject duplicate result column names; callers must alias them.
- Document that JSON.parse can itself lose precision for large JSON numeric literals.

## Execution, transactions and failures

- Each invocation owns a fresh VM; runtime values never cross invocations.
- Verified script support covers PostgreSQL, MySQL, MariaDB and SQLite.
- TiDB, Redshift, Oracle, ClickHouse and MSSQL are executable experimental dialects using their existing
  drivers. Compatibility does not establish broad deployment verification. Oracle uses
  pinned go-ora v3 (required for Oracle 23ai authentication) and native `:1`
  placeholders; NUMBER stays exact text, RAW is bytes and CLOB is text. BLOB values
  lacking type metadata require explicit RAW/hex projection (for example a bounded
  DBMS_LOB.SUBSTR); do not guess based on bytes alone. CLI and test connections disable
  fast-login cookies by default for reliable v3 reconnection, preserving explicit
  FAST LOGIN options. Go facade callers must configure their supplied DB likewise.
  Native PL/SQL follows driver execution semantics. If this driver omits
  SQL type names, use its numeric precision and Oracle scan-type metadata for NUMBER,
  binary and text conversion; keep ambiguous non-Oracle byte values rejected.
  Redshift-specific values with unknown pgx type metadata require explicit SQL casts;
  use FROM_VARBYTE/FROM_HEX for binary values through hex text.
- MSSQL uses the existing sqlserver driver and native `@p1` parameters. Native batches,
  DECLARE/SET, procedure definitions/calls and manual transactions are permitted by
  the bridge. SQL Server parses the original SQL; do not reimplement T-SQL grammar.
  MONEY/SMALLMONEY remain exact strings, GUIDs retain the driver's copied binary
  representation (cast to VARCHAR(36) for formatted text), and timestamp/rowversion
  is binary only for this dialect. DATE/TIME values retain the existing time contract.
- Database name resolution has one owning alias catalog, shared by SQL, script
  capability checks and runtime execution. Accept trimmed, case-insensitive aliases;
  expose only canonical names in `db.dialect` and generated result types.
  Empty SQL configuration keeps its existing PostgreSQL default; empty explicit
  script capability names are rejected. MySQL/MariaDB/TiDB remain distinct dialects.
- Capability metadata and text output list accepted aliases alongside canonical
  names. Test coverage must catch any SQL dialect or alias missing script support.
- Capabilities preserve `script_dialects` and add per-dialect verification, transaction
  mode and notes. Text output shows the same distinctions. SQL-only builds report no
  script dialects. Generated declarations include every executable dialect.
- Warn once per selected script plan for experimental dialects, including dry-run;
  metadata-only operations and SQL-only plans do not emit this warning.
- By default, each script migration and successful history write share the framework
  transaction where supported. Capability transaction_mode describes this default.
  Authors must not manually begin/commit/rollback that transaction; such SQL is not
  blocked, but violates managed-mode usage and invalidates automatic rollback guarantees.
- With config.transaction=false, use one dedicated sql.Conn for every call and history
  operation in that step, with no framework Begin/Commit/Rollback. On success, return
  the connection to the pool via sql.Conn.Close; rely on database/sql and the driver's
  validity/reset hooks for reuse rather than marking every successful connection bad.
  On error or panic, discard it because an unfinished manual transaction may remain.
  Driver reset support does not uniformly clear session settings: authors must restore
  changed settings and finish manual transactions before successful return. Successful
  return with an unfinished manual transaction is invalid usage. Uncaught or fatal
  failures discard the connection and clean up still-uncommitted native transactions.
- Warn for explicitly nontransactional scripts and annotate their errors (including
  history and down failures) with possible partial changes and recovery guidance.
  Earlier auto-committed operations can remain without successful history. Do not
  claim data/history atomicity or automatic rollback in this mode.
- ClickHouse retains its existing nontransactional default.
- Preserve existing SQL transaction behavior, including the nontransactional ClickHouse
  SQL fallback. Transaction guarantees below apply only where a transaction is supported.
- Default cleanup rolls back. Commit happens explicitly after successful execution.
- On uncaught script/SQL failures or fatal bridge failures the framework does not call
  Commit and attempts rollback.
  This can undo only changes still governed by its transaction; user SQL, procedures,
  implicit commits and nontransactional operations can defeat that guarantee. Capture
  unexpected callback panics, return a sanitized error and preserve rollback errors.
- Driver errors from execution, affected-row retrieval, query metadata, scan,
  iteration and result close are sanitized JS exceptions. Scripts may catch them
  and make further DB calls, or rethrow to fail the migration. Do not classify
  driver types, SQLSTATEs, vendor codes or connection/timeout sentinels to override
  that choice. Catching an error and returning normally means the author accepts it;
  the engine does not certify that the connection or transaction remains usable.
- Authors must repair transaction state where required (for example ROLLBACK TO
  SAVEPOINT on PostgreSQL, or explicit ROLLBACK in manual mode). If an error ended
  the entire managed transaction and a savepoint cannot recover it, rethrow. Full
  transaction retries belong in explicit manual mode. Do not automatically replay
  SQL or replace a failed transaction. Commit and history errors still fail the step.
- Cancellation of the actual execution context, the overall script deadline,
  parameter/result validation and resource-limit failures remain fatal even when
  caught, and block subsequent DB calls. A driver-reported timeout/cancellation
  alone does not cancel the execution context and may be handled by the script.
- Never expose raw driver errors, cause chains, SQL text or bound/result values to
  JS. Each driver exception contains only a sanitized operation message.
- Manual SQL transaction/session control requires config.transaction=false. Database
  permissions, SQL validity, balancing manual transactions and resetting desired session
  state are the migration author's responsibility; no SQL keyword enforcement is provided.
- MySQL/MariaDB atomicity requires transactional tables and operations. Oracle DDL
  implicitly commits. MySQL/MariaDB/TiDB DDL may implicitly
  commit: a framework transaction cannot make those statements rollback-safe.
- ClickHouse scripts may leave partial changes without successful history on failure;
  warn before execution and report this in execution/history errors for up and down.
  Inspect data/history before retries. Do not promise automatic rollback, resumability or
  exactly-once execution. Users must wait for mutations using native statement settings.
- ClickHouse scalar integers (including 128/256 bit) and decimal values remain exact
  strings; unwrap Nullable/LowCardinality metadata. Complex values (Array/Tuple/Map,
  etc.) require explicit scalar casts; arbitrary driver objects remain rejected.
- Resumable backfills remain deferred.
- History tracks successful versions, not an exactly-once guarantee across crashes,
  implicit commits or concurrent migrators. Operators serialize migration invocations.

## Resource and observation contract

- Default script execution timeout: 5 minutes; configurable positive duration.
- Default query limits: 1,000 rows and 16 MiB of returned values; configurable positive limits.
- Close each query result before returning materialized rows to JS or allowing writes.
- Propagate query, scan, iteration, close, commit and rollback errors.
- Tie cancellation to both VM interruption and database Context cancellation.
- A VM interrupt cannot forcibly terminate a blocking native callback. Do not promise
  hard real-time termination or per-VM hard heap isolation.
- Limit call-stack depth; restrict source maps to internally generated inline maps.
- Errors identify file/direction and original source locations, without bound values,
  query text, credentials, result data or raw script exception messages.
- Scripts are trusted repository code. No filesystem, network, environment, process,
  timers or raw Go database objects are exposed.

Status and baseline only inspect source metadata. Dry-run may parse/compile and print
SQL or script plan entries, but must not evaluate script code, modify user data or
write successful migration history. It does not predict dynamically generated SQL.

## CLI and API

- `new version <description> --format sql|js|ts|mjs` defaults to SQL.
- SQL-only builds refuse JS/TS generation with an actionable full-build error.
- Generated JS templates annotate `db` with JSDoc; TS templates declare equivalent
  database, parameter and result types for editor assistance.
- Script limits are configurable through both file config and the Go facade.
- Preserve public migration/source types where possible; keep prepared execution state internal.

## Validation and delivery

- Keep runnable local, installed-package and embedded-source examples under
  `examples/script-imports`; CI runs `scripts/test-script-examples.sh` using temporary
  SQLite databases and the committed local npm package without registry access.

- Every implementation layer includes its own meaningful tests and remains buildable.
- Cover mixed ordering, duplicates, hashes, preflight failure, rollback-range validation,
  irreversible down, baseline, out-of-order, dry-run and default-build rejection.
- Cover ignored synchronous return values, Promise rejection, caught SQL recovery,
  uncaught failures, fatal bridge failures, conversion fidelity, timeout, cancellation,
  resource caps, module restrictions, isolated VM state and original TS error locations.
- Use testcontainers for PostgreSQL/MySQL/MariaDB and temporary SQLite files.
- Explicit ClickHouse/TiDB/MSSQL selection without an external DSN uses disposable pinned
  containers; TiDB uses Unistore, not a verified distributed TiKV deployment. Preserve
  external DSN overrides and default optional-database skips.
- CI covers ClickHouse/TiDB/MSSQL SQL and full builds. Oracle/Redshift and external MSSQL SQL/full checks are
  conditional on dedicated external DSNs; skipped jobs are not real-instance verification.
- New dialects remain experimental despite limited container coverage; keep capability
  notes accurate about what has been exercised.
- Run all five repository merge gates in both default and full configurations.
  Combine `integration,full` for full integration tests.
- Test minimum Go compatibility, release platform builds and native executable smoke tests.
- Verify default runtime dependencies exclude Goja/esbuild; distinguish cross-build from runtime tests.

## Stacked review workflow

PR0 defines this contract; PR1 execution/transactions; PR2 VM/transpilation;
PR3 value conversion; PR4 DB bridge; PR5 mixed planning; PR6 CLI;
PR7 database integration; PR8 distribution and user documentation.

Each branch has its own ignored worktree and targets its direct predecessor. Keep each
PR focused; aim for 200-400 added production lines and reconsider above 800 handwritten
diff lines. Split a layer when its review topics or size warrant it; do not omit tests.

Create/update PRs and post `@codex review`. Record head/base commit identities and the
trigger comment. Read findings, repair them in the owning layer, update descendants,
rerun checks and request a fresh review after relevant changes. Resolve findings only
with a fix or documented evidence that they do not apply.

Completion requires CI success and a completed review of the current head/base with
no unresolved valid findings. Silence, an in-progress reaction, rate limiting and stale
reviews do not count as approval. Report unavailable review services as pending.
Do not auto-merge or publish a release. Revalidate descendants after parents change.
