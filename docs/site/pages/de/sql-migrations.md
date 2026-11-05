# SQL-Migrationen schreiben

Eine logische Änderung je Datei; SQL und Skripte teilen dieselbe Versionsfolge.

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

Beschreibungen mit Leerzeichen in Anführungszeichen setzen; im Dateinamen werden sie zu `_`. Ohne `-v` wird die Version erzeugt. Format: `V<version>__<description>.sql`. Angewandte Dateien weder ändern noch umbenennen, sondern neue Versionen ergänzen.

## Up und Down

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up ändert, Down nimmt zurück. Dieses Down löscht auch die Daten. Planen Sie deren Wiederherstellung.

## Interne Semikolons

Eine Anweisung mit eigenen Semikolons mit `StatementBegin` und `StatementEnd` umschließen:

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

Syntax und DDL-Transaktionen hängen von der Datenbank ab. Weiter zur [Ausführung](./applying).
