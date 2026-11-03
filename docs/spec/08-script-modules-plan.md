# Script Module Imports: PR Stack Plan

Status: disk-relative and installed-package imports are implemented on this branch;
virtual imports and workflow integration follow in the layers below. The current runtime
contract remains in [07-script-migrations.md](07-script-migrations.md). PR0 reconciles
that contract with the scope below and marks the rollout explicitly. This document
describes implementation order and acceptance, not an announcement of shipped support.

## Agreed scope

- Full builds support local static ES module imports and re-exports across JS, TS and
  MJS files, plus installed npm packages whose dependencies can be bundled for Goja.
- Statically resolvable CommonJS dependencies are bundled; no general runtime module
  loader is added. Native addons, Node built-ins, dynamic imports, dynamic require,
  top-level await and asynchronous migration entrypoints remain unsupported.
- Package installation remains a user/build step. Migrate does not invoke package
  managers, download dependencies or introduce a Node runtime.
- Keep entry-source SHA-256, SQL hashes and history schema unchanged. Dependency
  files, package metadata, lockfiles and bundle output are excluded from checksums.
- Changing a dependency does not invalidate entry checksums or replay applied
  migrations. Document this explicitly; retaining historical dependency behavior and
  pinning installed versions are author responsibilities, not enforced version rules.
- Migration discovery and dependency loading remain separate. Preserve versioned
  entry naming and direct-child discovery; helpers are loaded only through imports.
- Preserve synchronous execution, per-migration VM isolation, transaction policies,
  cancellation, database error handling and SQL-only build behavior.

## Dependency chain

```text
main
  -> PR0 docs/script-module-contract
  -> PR1 feat/script-module-context
  -> PR2 feat/script-relative-imports
  -> PR3 feat/script-package-imports
  -> PR4 feat/script-virtual-imports
  -> PR5 test/script-module-workflows
```

Names are proposed branch suffixes, not existing branches or pull requests. Each PR
targets its direct predecessor. Every implementation layer must remain buildable,
include its own tests and update the documentation for behavior it enables. PR5 adds
cross-layer verification rather than deferring foundational tests or user guidance.

## PR0: contract and acceptance

Changes:

- Update the module sections of spec 07 to distinguish local/package import support
  from Node runtime compatibility and unsupported asynchronous loading.
- Define the source matrix: disk sources resolve installed packages natively;
  virtual sources resolve relative modules in their supplied filesystem and may use
  an explicitly supplied disk directory for installed package resolution.
- Define package resolution: use esbuild's resolver, a neutral platform, explicit
  main-field precedence (`module`, then `main`), and no implicit browser/node export
  condition. Let esbuild apply import/require/default conditions according to the
  import kind. Keep its documented exports precedence rather than a custom parser.
- Keep esbuild's relative extension lookup rules; MJS may be imported with an explicit
  extension. Do not add TypeScript path aliases or automatic project tsconfig loading
  in this stack. Update the spec wording that currently prohibits all imports.
- Specify preflight, diagnostic sanitization and unchanged checksum/metadata behavior.

Acceptance: the contract is reviewable without production changes and records the
staged implementation status. Remaining unsupported features and source differences
are explicit; imports are not advertised as already available.

## PR1: preserve module context through sources and planning

Owning files: `internal/source/source.go`, `source.go`,
`internal/migrate/service.go`, and internal planning types.

Changes:

- Introduce an optional source extension returning migration entries with module
  context. Preserve the existing `LoadMigrations()` interface and public `SourceFile`
  field layout; legacy custom sources remain usable without implementing the extension.
- Carry the context alongside each entry through sorting and plan selection. Keep
  execution-only context out of public status values and history records.
- Disk context retains the resolved entry location. Virtual context retains the
  filesystem and entry path, with an optional explicit disk dependency directory.
- Keep contexts attached to entries when combining sources. Do not use a global map
  keyed solely by filename or an implicit process working-directory fallback.
- A load operation uses one source snapshot, not separate metadata and source loads.

Acceptance/tests:

- Legacy custom sources and existing public source examples still compile and run.
- Context survives source combination, ordering and selected rollback ranges.
- Identical helper names from different source roots remain distinct; duplicate
  migration-version detection keeps its current behavior.
- Entry bytes and all existing hashes/status/baseline behavior remain unchanged.
- Context itself does not enable module evaluation or imports in this layer.

## PR2: local relative imports and bundled execution

Owning files: `internal/script/runtime_full.go`, new focused module compiler helpers,
`internal/script/runtime_stub.go`, `internal/migrate/plan.go`, runtime/planning tests.

Changes:

- Pass optional context to compilation and replace the blanket import rejection for
  supported disk-relative imports and re-exports. Preserve no-context standalone use.
- Resolve relative paths from the importing module, including shared files outside
  the migration directory; use esbuild's filesystem resolver and bundler.
- Compile the entry from the source bytes already loaded and hashed. Preserve its
  module identity so cyclic references cannot reload a different copy of the entry.
- Keep Node built-ins and asynchronous module loading unsupported. Do not enable a
  generic runtime require function or accidentally rely on esbuild's browser defaults.
- Compile and initialize selected migrations before migration writes; execution uses
  the retained program/VM and does not reload dependency files. Do not cache programs
  globally across service invocations.
- Preserve source maps for imported files. Diagnostics identify logical module names,
  import sites and locations without leaking source lines, absolute paths or raw
  filesystem/driver errors.

Acceptance/tests:

- JS/TS/MJS imports, nested imports, re-exports, shared files outside the migration
  directory and normal cyclic module graphs work with esbuild's bundling semantics.
- Resolution does not depend on a later working-directory change. Imported modules
  initialize once per migration VM; different migrations do not share module state.
- Missing modules, syntax errors and invalid imported entrypoints fail before any
  selected up/down step writes data; dry-run compiles but does not evaluate modules.
- Changing dependency contents affects a new invocation without changing the entry
  checksum. Changing files after preparation does not replace the prepared program.
- Existing Promise rejection, cancellation, recovery and source-map tests still pass.

## PR3: installed npm packages

Owning files: module compiler helpers, runtime tests and local package fixtures.

Changes:

- Enable bare package specifiers through esbuild's native disk resolver, rooted at
  each importing file's directory, including scoped names and exported subpaths.
- Apply PR0's platform, main-field and condition settings explicitly. Bundle eligible
  ESM and static CommonJS dependencies rather than leaving packages external.
- Follow installed layouts through esbuild's normal filesystem handling; do not parse
  lockfiles or implement package-manager-specific installers/resolvers.
- Report unresolvable modules and unsupported native/built-in dependencies clearly.
  State that successful package resolution does not certify Goja runtime compatibility.

Acceptance/tests:

- Local fixture packages exercise plain/scoped names, nested dependencies, subpath
  exports, import/require conditions, entry precedence and CommonJS interoperation.
- Two source roots can resolve different installed versions of the same package.
- Package tests need no network or package manager; fixtures are created locally.
- Test a linked-package layout representative of pnpm installations without claiming
  compatibility with every workspace or Yarn PnP arrangement.
- Packages requiring Node-only conditions or native runtime facilities do not acquire
  those capabilities implicitly; existing no-Node behavior remains covered.
- Modifying installed package files leaves entry checksums unchanged.

## PR4: virtual and in-memory module sources

Owning files: source context adapters, module compiler plugin, source/runtime tests,
public source examples.

Changes:

- Enable relative imports within the supplied `fs.FS`; allow in-memory callers to
  provide a module filesystem such as `fstest.MapFS`. Preserve entry source bytes as
  authoritative and keep the entry path separate from its history filename.
- Resolve each nested relative import against its own virtual importer. Normalize
  virtual paths, and respect the supplied filesystem boundary.
- Allow an explicit disk dependency directory for bare package imports. After a
  package is resolved to disk, its internal dependencies use disk resolution.
- Without an explicit disk context, virtual sources do not inspect the host working
  directory. Purely virtual `node_modules` package resolution is outside this stack;
  report the missing capability rather than implementing a second package resolver.
- Ensure combined disk/virtual sources retain their separate resolution contexts.

Acceptance/tests:

- `embed.FS` and `fstest.MapFS` cover nested relative modules and re-exports.
- A virtual migration can import a local helper that imports an installed package
  through an explicit dependency directory.
- Virtual path traversal, missing dependencies and accidental host filesystem
  fallback have deterministic failures before migration writes.
- Separate virtual roots with identical filenames cannot cross-load dependencies.
- Legacy string sources continue to execute standalone migrations unchanged.

## PR5: workflow integration and delivery

Owning files: CLI/facade tests, migration integration tests, script/user guides,
API declarations as needed, changelog and capabilities help where applicable.

Changes:

- Exercise real directory-based CLI and library workflows with shared modules and
  fixture packages, including up/down, managed/manual transactions and mixed SQL.
- Verify unchanged status/baseline semantics, hash mismatch handling, dry-run,
  out-of-order exclusion, selected rollback preflight and SQL-only rejection.
- Prove that a missing dependency in a later selected migration prevents earlier
  selected migrations from writing data; similarly verify the full rollback range.
- Add focused native-driver coverage using existing disposable integration fixtures;
  do not duplicate the entire dialect matrix for compiler-only behavior.
- Consolidate examples covering dependency installation, project layout, source
  configuration, checksums, historical dependency maintenance and runtime boundaries.
- Confirm packaging remains self-contained, without Node/npm executables or runtime
  package downloads. Preserve default-build exclusion of Goja/esbuild.

Acceptance: public examples match the shipped source capabilities, documented commands
work, and all final checks/reviews below are complete. Conditional external database
tests that skip are reported as skips, not successful real-instance verification.

## Checks and review for every layer

- Follow [01-quality-gates.md](01-quality-gates.md) and
  [06-reviewer-security-rules.md](06-reviewer-security-rules.md).
- Run the five repository merge gates in default and full configurations:
  `go test ./...`, `go test -tags=integration ./...`, `go vet ./...`,
  `go build ./...`, `go build ./cmd/migrate`; full counterparts add `-tags=full`,
  and integration uses `-tags=integration,full`.
- Keep each layer focused and include its meaningful tests. Aim for 200-400 added
  production lines; reconsider a layer above 800 handwritten diff lines. Split a
  layer when its actual review size warrants it, without deferring required tests.
- Use an ignored worktree per implementation branch and target the direct predecessor.
- Request `@codex review` for each PR, record the reviewed head/base and trigger, read
  findings and fix valid findings in the owning layer. If no fix is needed, explain
  the evidence explicitly. Update descendants and obtain fresh checks/reviews when
  relevant head/base changes invalidate the previous result.
- Completion requires green CI and a completed current head/base review with no
  unresolved valid findings. Pending/unavailable reviews remain pending.
- Do not merge or publish a release automatically. After the user merges a parent,
  update the next PR's base and revalidate its diff, checks and review.
