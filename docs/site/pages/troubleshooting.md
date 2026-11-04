# Troubleshooting

Start with `migrate version`, `migrate capabilities --output json` and
`migrate status --output json`. Status needs a working database connection;
version and capabilities do not.

## Command or configuration not found

If `migrate` is not on PATH, check the binary location in [Installation](./installation).
If `migration_config.json` cannot be found, run from the project directory or use
`-c` / `-w`. See [Configuration](./configuration) for path resolution.

The config must be JSON only. A syntax error includes line/column information;
remove comments, trailing commas or text after the closing object.

## Missing environment variable

`can't find env value for XXX` means a required `${XXX}` template could not be
resolved. Provide it in the command's environment, or use `${XXX:default}` for a
suitable nonsecret fallback. SSH secret fields instead contain the variable name
itself; see [SSH authentication](./ssh-tunnels).

## Unsupported format, dialect or module

- Use `migrate capabilities` to confirm SQL, JS, TS and MJS support. Upgrade an
  older SQL-only binary using the [installation instructions](./installation).
- For `unsupported dialect`, check [canonical names and aliases](./databases#aliases).
- If a module cannot resolve, verify its path from the importer and install npm
  dependencies beforehand. Virtual sources need explicit disk dependency settings.
- A package that requires Node APIs or asynchronous execution cannot run merely
  because it was bundled. See [Shared modules](./modules).

## Hash or filename mismatch

Restore the exact applied file and filename from version control. Put subsequent
changes in a new higher version. Do not bypass the error by baselining an unknown
database state. See [History and baseline](./history).

## Nothing happened after up

Successful commands may print nothing. Inspect status: versions may already be
applied or baseline, or an older pending version may be out of order and skipped.
Review [ordering behavior](./applying) before changing configuration.

## Rollback rejected

`to-version must be set, or use --all` requires a target or `--all`, but not both.
The target stays applied. A missing script down function makes its selected range
irreversible. See [Rollback](./rollback).

## A migration, timeout or connection failed

Inspect actual data and history before retrying. Nontransactional operations,
implicit DDL commits and connection loss around COMMIT can leave changes despite
a failed command. Baseline does not repair partial execution.

For scripts, review [transaction and resource limits](./script-transactions).
For SSH, verify trusted known-hosts, credentials, the proxy and the bastion's access
to the original database host. Transport loss does not reconnect or replay SQL.
Diagnostics are sanitized; use controlled database/server logs when investigating
vendor errors and avoid sharing credentials or row data.
