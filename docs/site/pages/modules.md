# Import shared modules

Script migrations can statically import and re-export shared `.js`, `.ts` and `.mjs`
modules. Helpers are dependencies, not migration entries.

## Relative imports on disk

Paths resolve from each importing file, including helpers outside the migration directory:

```text
project/
  migrations/V2__normalize.ts
  shared/names.ts
```

```ts
import { normalize } from "../shared/names.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

Define `normalize` in `shared/names.ts`, and ensure the example's `users` table
already exists. The [local example](./examples#local) includes that setup.

## Installed npm packages

Bare imports such as `@company/migration-utils` resolve installed packages.
Install dependencies with your project's lockfile workflow before running migrate.
The compiler resolves exports, subpaths, nested dependencies and static CommonJS
imports using neutral conditions and `module`, then `main` entry fields.

Packages must work in Goja's synchronous runtime. migrate does not install packages
or provide Node built-ins, native addons, dynamic import or dynamic require.

## Embedded and in-memory sources

`FSSource` resolves relative helpers inside its supplied `fs.FS`, including `embed.FS`
and `fstest.MapFS`. File-extension lookup and directory `index` files are supported;
virtual package.json entrypoints and virtual node_modules resolution are not.
Use an explicit on-disk dependency directory when packages are needed:

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

`embeddedFiles` is your supplied filesystem; `projectDirectory` contains installed
packages. Relative helpers stay in the virtual filesystem. Without `DependencyDir`,
virtual sources do not search the host working directory.

For `StringSource`, wrap the source in `WithModules` with an `FS` of helpers and
optional `Root`, or a `Directory` for disk helpers/packages. Custom sources can
implement `MigrationEntriesSource` to provide entry contexts. See [Use from Go](./go-library).

## Preserve historical dependencies {#checksums}

Script hashes cover original entry bytes only. Editing a helper, installed package
or lockfile neither creates a hash mismatch nor replays an applied migration.
Retain historical helper behavior and pin dependencies when old migrations must be
reproducible; introduce a new helper version for changed behavior.

Dependencies for the whole selected range are compiled before writes. A dependency
failure in a later migration prevents earlier selected steps from executing.
