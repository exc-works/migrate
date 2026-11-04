# History and baseline

The migration history table records versions alongside their filenames and hashes.
Use it together with the actual schema to understand what has been applied.

## Inspect status

```sh
migrate status
migrate status --output json
```

The table reports Version, Filename, Hash and Status. JSON is useful for automation.

| Status | Meaning | Next step |
| --- | --- | --- |
| `pending` | File has not been applied | Review and run `up` |
| `applied` | Version was successfully applied | Keep the file unchanged |
| `baseline` | Version was recorded without running it | Preserve the adopted history |
| `outOfOrder` | Pending version is below the applied version | Review ordering policy |
| `hashMismatch` | Recorded and current file hashes differ | Restore the historical file |
| `filenameMismatch` | Recorded and current filenames differ | Restore the historical filename |

Hash or filename mismatches block execution. Do not rewrite an applied migration
to make a new schema change; add a higher version instead. Script checksums cover
entry source bytes, not imported helpers; see [Module reproducibility](./modules#checksums).

## Baseline an existing database {#baseline}

Use baseline only when the database already matches the migration files being
adopted. It records history without executing their SQL or scripts.

1. Compare the existing schema and data with the migration files.
2. Initialize the history table with `migrate create` if needed.
3. Run `migrate baseline`, then inspect `migrate status`.
4. Add a new higher version for future changes and apply it normally.

Baseline is not a repair command for a partially failed migration. First establish
the real database state; blindly recording history can hide missing changes.
