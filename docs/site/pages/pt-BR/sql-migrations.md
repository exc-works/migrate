# Escrever migrações SQL

Uma mudança lógica por arquivo. SQL e scripts usam a mesma sequência.

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

Descrições com espaços precisam de aspas; no nome eles viram `_`. `-v` escolhe a versão, caso contrário ela é gerada. Nome: `V<version>__<description>.sql`. Não modifique arquivos aplicados: adicione outra versão.

## Up e Down

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up aplica, Down reverte. Este Down também apaga os dados; planeje sua recuperação.

## Ponto e vírgula interno

Use `StatementBegin` e `StatementEnd` para não dividir instruções com ponto e vírgula interno:

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

Sintaxe e transações DDL dependem do banco. Continue com a [aplicação](./applying).
