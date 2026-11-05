# Écrire des migrations SQL

Conservez une modification logique par fichier. SQL et scripts partagent la même suite de versions.

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

Entourez les descriptions avec espaces de guillemets ; le nom du fichier utilise des `_`. `-v` choisit la version ; sans lui, elle est générée. Format : `V<version>__<description>.sql`. N’éditez ni ne renommez un fichier appliqué : ajoutez une nouvelle version.

## Up et Down

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up applique le changement ; Down l’annule. Ici, Down supprime aussi les données. Prévoyez leur récupération si nécessaire.

## Points-virgules internes

Encadrez une instruction contenant des points-virgules internes par `StatementBegin` et `StatementEnd`, comme cette fonction PostgreSQL :

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

La syntaxe et les transactions DDL restent propres à la base. Passez à l’[exécution](./applying).
