# Apply migrations

Before applying changes, check that the configuration points to the intended database
and that its migration directory contains the versions you want to deploy.

## Initialize history

```sh
migrate create
```

This idempotently creates the migration history table; it does not apply your
version files. If the database already contains the changes in those files,
review [Baseline an existing database](./history#baseline) before continuing.

## Preview and execute

```sh
migrate up --dry-run
migrate up
migrate status
```

`up` applies pending migrations in version order and records successful versions.
Re-running it does not replay already applied versions. Successful commands may
produce no output; use `status` to confirm the result.

Dry-run prints SQL or script plans without executing user migrations. It may read
database history and open an SSH tunnel. For scripts it compiles dependencies
without evaluating them, so it cannot predict dynamically generated SQL.

## Handle older pending versions

By default, pending files below the applied version are out of order and skipped.
If you deliberately need to apply them, enable `migrate_out_of_order` in
[configuration](./configuration) after reviewing their compatibility with the current schema.

## Run in a deployment step

```sh
migrate --non-interactive up
migrate --non-interactive status --output json
```

Supply required environment variables and SSH/proxy secrets before the command.
Serialize migrators targeting the same database. History does not provide an
exactly-once guarantee across concurrent runs, crashes or implicit commits.

If execution fails, inspect database state and history before retrying, especially
with DDL or nontransactional engines. See [Troubleshooting](./troubleshooting) and
[database transaction boundaries](./databases).
