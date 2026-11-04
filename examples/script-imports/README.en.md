# Script module import examples

Run these commands from the repository root. All examples use SQLite and the default
build; no database server is needed. [中文说明](README.md).

| Directory | Demonstrates | Prerequisites |
| --- | --- | --- |
| [local](local) | TS/MJS migrations sharing a TS helper outside the migration directory | Go |
| [npm](npm) | package.json, a lockfile, installed packages and an exports subpath | Go, Node/npm for installation |
| [embedded](embedded) | embed.FS, the public Go API, up and down | Go |

## Relative imports

```bash
demo_dir=$(mktemp -d)
go build -o "$demo_dir/migrate" ./cmd/migrate
export MIGRATE_EXAMPLE_DB="$demo_dir/local.db"

"$demo_dir/migrate" --working-dir examples/script-imports/local create
"$demo_dir/migrate" --working-dir examples/script-imports/local up --dry-run
"$demo_dir/migrate" --working-dir examples/script-imports/local up
"$demo_dir/migrate" --working-dir examples/script-imports/local status
"$demo_dir/migrate" --working-dir examples/script-imports/local down 2
"$demo_dir/migrate" --working-dir examples/script-imports/local down --all
```

V2 and V3 both import `../shared/names.ts`. Up normalizes user 1 to `alice` and adds
user 2 as `bob`; queries inside the migrations check these results. Status reports
three applied versions. `down 2` keeps V1/V2 and reverts V3; `down --all` drops the
example users table. Helpers are dependencies, not separate migration entries.

## Installed npm packages

Using the same built CLI:

```bash
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$demo_dir/npm.db"

"$demo_dir/migrate" --working-dir examples/script-imports/npm create
"$demo_dir/migrate" --working-dir examples/script-imports/npm up
"$demo_dir/migrate" --working-dir examples/script-imports/npm status
"$demo_dir/migrate" --working-dir examples/script-imports/npm down --all
```

The project package.json declares `file:./packages/normalizer`; the committed lockfile
reproduces its installation under `node_modules/@migrate-example/normalizer`. The
package's own package.json exposes the root and `/names` subpath using `exports`.
V2 imports `@migrate-example/normalizer/names`, normalizes the name to `alice`, and
checks it. Status reports two applied versions.

This bundled example package is pure JavaScript without registry dependencies, so
installation works offline. For registry packages, install your actual dependencies
with the project's lockfile workflow. Node/npm is used for installation; migration
execution uses embedded esbuild and Goja.

## Embedded Go application

```bash
go run ./examples/script-imports/embedded
```

Expected output:

```text
after up: alice
after down: all migrations pending
```

The program uses a private SQLite in-memory database and embeds both `migrations`
and `shared`. FSSource selects only versioned entries under Root; imports read helpers
from the same supplied filesystem. The built program needs no TS files on disk.
This example has no npm dependency.

The in-memory WithModules example in [modules_test.go](../../modules_test.go)
is also runnable:

```bash
go test . -run '^ExampleWithModules$' -v
```

## Verify and clean up

```bash
./scripts/test-script-examples.sh
```

CI runs this same check. It builds the CLI, copies and installs the npm example,
checks both CLI lifecycles and embedded output, then removes its temporary files.
After running the manual commands above, clean up with:

```bash
rm -r "$demo_dir"
rm -r examples/script-imports/npm/node_modules
```

Entry-source checksums exclude helpers, packages and lockfiles. Dependency edits do
not trigger history mismatches or replay applied migrations; retain historical helper
behavior when reproducibility matters. Packages must run in Goja; Node APIs, native
addons, dynamic loading and async migration entrypoints remain unsupported. See the
[script guide](../../docs/script-migrations.en.md) for virtual package resolution limits.
