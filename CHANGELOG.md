# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Experimental JS/TS/MJS migration execution for TiDB, Redshift, Oracle and ClickHouse.
- Per-dialect verification status, transaction modes and limitations in capabilities,
  plus one warning per selected experimental script plan.
- ClickHouse scalar value conversion and explicit partial-state errors for
  nontransactional execution/history failures; unknown affected-row counts are null.
- ClickHouse/TiDB container CI and optional SQL/full Oracle/Redshift integration.

### Fixed
- Upgrade the Oracle driver to go-ora v3 for modern authentication; disable its fast
  login cache by default while preserving explicit options. Correct Oracle history
  DDL, native status binding and missing result metadata handling.
- Serialize external SQL/full CI against each dedicated database to avoid fixture races.

### Compatibility
- Default builds remain SQL-only. Existing script_dialects JSON output is preserved;
  script_dialect_capabilities adds detailed status. New dialects remain experimental.
- Oracle DDL implicitly commits. ClickHouse scripts cannot automatically roll back;
  inspect data and history before retrying failures. Generated declarations include
  all executable dialects and nullable affected-row counts.

## [1.4.0] - 2026-11-02

### Added
- Accept `.mjs` as a JavaScript migration alias in full builds; SQL-only builds
  detect and reject it before migration writes. Support `new version --format mjs`.
- Optional `full` builds support standalone synchronous JS/TS migrations alongside SQL,
  with framework-owned transactions, exact value conversion, source checksums,
  execution deadlines and query result limits.
- `capabilities` reports compiled formats without configuration or a connection.
- `new version --format sql|js|ts|mjs` creates migration templates; script timeout and
  result limits can be configured through JSON or the Go facade.
- Generated JS/MJS templates include JSDoc database types; TS templates provide
  equivalent method, parameter and result declarations for editor assistance.
- SQL/full CI matrices and real-driver script integration for PostgreSQL, MySQL,
  MariaDB and SQLite; full release binaries and editor API declarations.

### Fixed
- Migration callbacks that panic now roll back instead of entering deferred commit.
- Execution validates complete selected plans before changing data, including
  irreversible rollback ranges and unsupported script capabilities.

### Compatibility
- Default binaries and existing release filenames remain SQL-only. Existing SQL
  hashes, migration history and version command output remain unchanged.
- Known JS/TS/MJS migration files are discovered by both builds; SQL-only execution
  refuses them instead of silently skipping required versions.

## [1.3.1] - 2026-04-22

### Fixed
- Configuration JSON parse errors now include the config file path, line, and
  column. Extra non-JSON content after the config object is reported with a
  clear remediation hint instead of only the raw Go parser message.

### Added
- Regression coverage for `migrate status -w <dir>` when
  `<dir>/migration_config.json` contains trailing non-JSON content.
- Configuration parsing coverage for DSNs that use environment defaults and
  passwords containing special characters such as `$`, `,`, and `>`.

## [1.3.0] - 2026-04-20

### Added
- Public root-package facade (`github.com/exc-works/migrate`) so other Go
  services can embed migration execution in-process alongside the existing
  `cmd/migrate` CLI. Exports `Service`, `Config`, `Status` constants, dialect
  constructors (`migrate.NewPostgresDialect()`, `NewMySQLDialect()`, etc.),
  migration sources, and logger types. `internal/*` remains private.
- `FSSource` loads migrations from any `fs.FS` — works with `//go:embed`,
  `os.DirFS`, or `fstest.MapFS` (test-friendly).
- User guides translated into 10 additional languages: ar, de, es, fr, hi,
  ja, ko, pt-BR, ru, zh-Hant. Each guide gains a new section 12 documenting
  library usage.

## [1.2.1] - 2026-04-17

### Changed
- CLI error messages and help output refined for both human users and
  agent-driven workflows.

## [1.2.0] - 2026-04-14

### Added
- Top-level `migrate --version` command.

## [1.1.1] - 2026-04-14

### Fixed
- Track `cmd/migrate` sources in git; narrowed `.gitignore` so the binary
  build artifact no longer excludes source files.

## [1.1.0] - 2026-04-14

### Changed
- Renamed binary and module from `sql-migrate` to `migrate`.
- Release workflow now targets Linux arm64 / x64 / x86.
- GitHub Actions bumped to Node 24–compatible versions.

### Fixed
- Oracle integration secret gating in CI.

## [1.0.0] - 2026-04-14

Initial tagged release.

### Added
- Dialect support: PostgreSQL, MySQL, MariaDB, MSSQL, Oracle, SQLite,
  ClickHouse, TiDB, Redshift.
- Bilingual (zh-CN, en) user guides.
- Tag-based multi-platform release workflow.
