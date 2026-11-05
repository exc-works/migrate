# Documentation map

migrate tracks database changes as ordered version files. Use the CLI in a deployment
step, or embed the same migration service in a Go application.

## Make your first migration

Start with the [SQLite quick start](./quick-start) to create, apply and reverse one
SQL migration without running a database server. For an existing project, choose
a build in [Installation](./installation), then set up [Configuration](./configuration).

## Work with a database

1. [Write SQL migrations](./sql-migrations): version names, Up/Down sections and statement blocks.
2. [Apply migrations](./applying): preview the plan, execute it and verify the result.
3. [Roll back migrations](./rollback): retain a target version or reverse all applied versions.
4. [History and baseline](./history): understand status, checksums and adopting an existing schema.

## Go beyond SQL

Use [JavaScript or TypeScript](./scripting) for transformations that need code.
[Shared modules](./modules), the [database API](./script-api), and
[transaction boundaries](./script-transactions) cover the runtime separately.
The [examples](./examples) let you try local imports, npm packages and embedded files.

For application integration, see [Use from Go](./go-library). For databases behind
a bastion, see [SSH tunnels and proxies](./ssh-tunnels).

## Look something up

- [CLI reference](./commands): commands, arguments and flags.
- [Supported databases](./databases): verification, placeholders and engine limits.
- [Troubleshooting](./troubleshooting): configuration, history and execution failures.
- [Release notes](./releases): behavior shipped in a specific release.

::: info Version availability
This website follows main. Use `migrate version` and the release notes to check
whether a feature is available in your installed binary.
:::
