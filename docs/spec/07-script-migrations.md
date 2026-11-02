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

- Discover direct children named `V<version>__<description>.sql`, `.js` or `.ts`.
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

Scripts are independent UTF-8 files exporting synchronous `up(db)` and optional `down(db)`.
Missing down is irreversible, not a successful no-op. The runtime validates the entry
before database writes; invalid exports anywhere in a rollback plan must not cause
partial rollback. Up and down return `undefined`; promises and generators are rejected.

ES module export syntax and TS syntax are transformed internally into a VM-loadable
module. Imports, re-exports, dynamic imports, `require`, npm packages, asynchronous
entrypoints and top-level await are unsupported. TS is transpiled without type checking;
provide declarations for editor assistance. Do not advertise full Node.js compatibility.

`db` is a narrow object with `dialect`, `query(sql, params?)` and `exec(sql, params?)`.
Statements use native dialect placeholders and bound parameters; do not concatenate
values into SQL. One call executes one driver statement, not annotated migration text.
No database object is exposed during module initialization.

## Database values

- Integer results and affected-row counts are decimal strings, including small integers.
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
- Full script support initially covers PostgreSQL, MySQL, MariaDB and SQLite.
- Each script migration and its successful history write use the same framework transaction.
- Preserve existing SQL transaction behavior, including the nontransactional ClickHouse
  SQL fallback. Transaction guarantees below apply only where a transaction is supported.
- Default cleanup rolls back. Commit happens explicitly after successful execution.
- Runtime/bridge failures never commit. Capture unexpected callback panics at the
  execution boundary, return a sanitized error and preserve rollback errors.
- A database bridge failure remains fatal even if script code catches the JS exception.
- Scripts must not execute transaction-control SQL. No nested/manual transaction API
  is exposed. Statement restrictions are checked before executing user SQL.
- MySQL/MariaDB atomicity requires transactional tables and operations. DDL may implicitly
  commit: a framework transaction cannot make those statements rollback-safe.
- General nontransactional script execution and resumable backfills are deferred.
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

- `new version <description> --format sql|js|ts` defaults to SQL.
- SQL-only builds refuse JS/TS generation with an actionable full-build error.
- Script limits are configurable through both file config and the Go facade.
- Preserve public migration/source types where possible; keep prepared execution state internal.

## Validation and delivery

- Every implementation layer includes its own meaningful tests and remains buildable.
- Cover mixed ordering, duplicates, hashes, preflight failure, rollback-range validation,
  irreversible down, baseline, out-of-order, dry-run and default-build rejection.
- Cover conversion fidelity, driver errors, timeout, cancellation, sticky failures,
  resource caps, module restrictions, isolated VM state and original TS error locations.
- Use testcontainers for PostgreSQL/MySQL/MariaDB and temporary SQLite files.
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
