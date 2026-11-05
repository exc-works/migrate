# Escribir migraciones SQL

Guarda un cambio lógico por archivo. SQL y scripts comparten la secuencia de versiones.

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

Usa comillas para descripciones con espacios, convertidos en `_` en el archivo. `-v` elige la versión; sin él se genera. Formato: `V<version>__<description>.sql`. No edites ni renombres archivos aplicados: añade otra versión.

## Up y Down

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up aplica y Down revierte. Aquí Down borra también los datos; planifica su recuperación.

## Puntos y coma internos

Rodea una sentencia con separadores internos con `StatementBegin` y `StatementEnd`, como esta función PostgreSQL:

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

La sintaxis y las transacciones DDL dependen de la base. Continúa con [aplicar](./applying).
