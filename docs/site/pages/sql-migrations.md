# Write SQL migrations

Keep one logical change in each version file. migrate reads files from the configured
`migration_source` and uses their versions to build an ordered migration plan.

## Create a version file

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

Quote descriptions containing spaces; spaces become underscores in filenames.
The first command generates a version; `-v` lets you choose one explicitly. Files
use `V<version>__<description>.sql`. SQL and script files share one version sequence.
Do not edit or rename a file after it has been applied; add a new version instead.

## Define Up and Down

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up performs the change; Down implements its reverse. Here, Down deletes the table
and its data. Plan rollback behavior as part of writing the migration, including
whether any lost data can be restored.

## Group statements containing semicolons

Use `StatementBegin` and `StatementEnd` around a statement whose internal semicolons
must not be treated as statement boundaries. For example, a PostgreSQL function:

```sql
-- +migrate Up
-- +migrate StatementBegin
CREATE FUNCTION migration_answer() RETURNS integer AS $$
BEGIN
    RETURN 42;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
DROP FUNCTION migration_answer();
```

These markers control SQL parsing; database syntax and DDL transaction behavior
still depend on your engine. Continue with [Apply migrations](./applying).
