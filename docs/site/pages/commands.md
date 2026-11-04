# Command reference

Run `migrate <command> --help` to see its options.

| Task | Command | Details |
| --- | --- | --- |
| Generate configuration | `migrate new config` | [Initialization](./guide#_2-initialization) |
| Create the history table | `migrate create` | [Initialization](./guide#_2-initialization) |
| Adopt an existing database | `migrate baseline` | [Baseline](./guide#_2-initialization) |
| Create a version file | `migrate new version <description>` | [Version files](./guide#_3-create-migration-version-files) |
| Apply pending versions | `migrate up` | [Up](./guide#_4-upgrade-apply-migrations) |
| Roll back versions | `migrate down <version>` or `migrate down --all` | [Down](./guide#_5-rollback) |
| Inspect migration history | `migrate status` | [Status](./guide#_6-check-status) |
| Inspect build/database support | `migrate capabilities --output json` | [Capabilities](./databases) |

`down <version>` rolls back versions above the target and keeps the target applied.
To adopt an existing database, match its schema to the migration files, then run `baseline` to record them.

Use `--dry-run` before up/down. [Global flags](./guide#_10-global-flags) cover
configuration selection and logging. [Troubleshooting](./guide#_11-common-errors-and-troubleshooting)
covers missing configuration, environment variables and history mismatches.

For embedded Go applications, see [library integration](./guide#_12-embedding-migrate-as-a-go-library).
