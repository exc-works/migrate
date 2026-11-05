# Installation

Install migrate with Go or download a release binary.

## Install with Go

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
```

To pin or downgrade the tool, replace `@latest` with a published release tag.
When working from a source checkout, use `go install ./cmd/migrate` from the
repository root.

## Use a release binary

Download the asset matching your operating system and architecture from
[GitHub Releases](https://github.com/exc-works/migrate/releases).
Verify the published SHA-256, rename the executable to `migrate` (`migrate.exe`
on Windows), and place it in a directory on PATH.

## Verify the installation

```sh
migrate --help
migrate version
migrate capabilities --output json
```

`version` reports the release version for downloaded binaries and the installed
module version for `go install ...@version`. Local development builds report `dev`.
For older versions that report `dev` after `go install`, use `go version -m <binary>`
to inspect the installed module version.
`capabilities` lists formats, database verification and transaction limits without
connecting to a database. See [Supported databases](./databases) for interpretation.

If a Go installation is not found on macOS/Linux, add its binary directory to PATH:

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
```

Continue with the [quick start](./quick-start) or [project configuration](./configuration).
