# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
