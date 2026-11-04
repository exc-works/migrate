# Roll back migrations

Rollback executes the Down section or `down(db)` supplied by each migration.
It can only undo what that code implements; it cannot recover deleted data by itself.

## Keep a target version

```sh
migrate down 202611030001 --dry-run
migrate down 202611030001
migrate status
```

Only applied versions above the target are reversed. The target version stays
applied. For example, with versions 1, 2 and 3 applied, `migrate down 2` reverses
version 3 and keeps versions 1 and 2.

## Reverse all applied versions

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

A target and `--all` are mutually exclusive. After a successful reversal, those
version files become pending and can be applied again.

## Review the whole rollback range

Check every selected migration's Down behavior and database transaction limits.
Script rollback prepares the whole selected range before writing; a missing
`down(db)` blocks that range. An empty down function does not undo an up function.

DDL implicit commits and nontransactional changes may survive a failed rollback.
Inspect data and history before recovery; see [Transactions and limits](./script-transactions)
and [Troubleshooting](./troubleshooting).
