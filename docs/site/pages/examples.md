# Run the script examples

Try these from a checkout of the repository. All use SQLite, so no database server
is needed. The examples exercise the full build's module resolver and verify their
data changes from inside the migrations.

## Build a temporary CLI

Run from the repository root; keep this shell open for the local and npm examples:

```sh
example_workspace=$(mktemp -d)
go build -tags=full -o "$example_workspace/migrate" ./cmd/migrate
```

## Share a local helper {#local}

[Browse the local example](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local).
Two migrations import a shared TypeScript helper outside the migration directory.

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

After up, three versions are applied and the names are `alice` and `bob`.
`down 2` keeps versions 1 and 2; `down --all` removes the example table.

## Resolve an installed package

[Browse the npm example](https://github.com/exc-works/migrate/tree/main/examples/script-imports/npm).
Node/npm is required for installation only. Its lockfile refers to a bundled local
package, so the install below works offline without registry dependencies:

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

The second migration imports `@migrate-example/normalizer/names` through an exports
subpath and verifies the normalized name `alice`. After up, two versions are applied.

## Embed helpers in a Go binary

The [embedded application](https://github.com/exc-works/migrate/tree/main/examples/script-imports/embedded)
includes both version files and shared modules in `embed.FS`. It uses a private
in-memory database and needs no npm installation:

```sh
go run -tags=full ./examples/script-imports/embedded
```

Expected output:

```text
after up: alice
after down: all migrations pending
```

## Clean up and verify

After the manual CLI examples, remove their temporary database files and executable:

```sh
rm -r "$example_workspace"
```

The npm install also leaves an ignored `examples/script-imports/npm/node_modules`
directory. For an automated check that manages its own temporary copies, run
`./scripts/test-script-examples.sh` from the repository root.

See [shared module behavior](./modules) for resolution and checksum boundaries.
