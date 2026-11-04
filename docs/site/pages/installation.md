# Installation

The default build includes SQL, JavaScript, TypeScript and MJS migrations.
The embedded script engine does not need Node.js or download an interpreter.

## Install with Go

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
```

To pin or downgrade the tool, replace `@latest` with a published release tag.
When working from a source checkout, use `go install ./cmd/migrate` from the
repository root. No extra build tags are needed.

## Use a release binary

Download the asset matching your operating system and architecture from
[GitHub Releases](https://github.com/exc-works/migrate/releases). Each platform has
one binary containing all migration formats, without a build suffix.
Verify the published SHA-256, rename the executable to `migrate` (`migrate.exe`
on Windows), and place it in a directory on PATH.

## Verify the installed build

```sh
migrate --help
migrate version
migrate capabilities --output json
```

`version` identifies the binary; source installations commonly report `dev`.
`capabilities` lists formats, database verification and transaction limits without
connecting to a database. See [Database capabilities](./databases) for interpretation.

If a Go installation is not found on macOS/Linux, add its binary directory to PATH:

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
```

Continue with the [quick start](./quick-start) or [project configuration](./configuration).
