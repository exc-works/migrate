# Guia do Usuário do migrate

Este guia é para usuários de primeira viagem. Os comandos e flags são baseados na implementação atual (`cmd/migrate`).

## 1. Instalação

### 1.1 Instalar a partir do módulo

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Instale uma versão específica:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Substitua `vX.Y.Z` por uma versão real, por exemplo `v0.2.3`.

### 1.2 Instalar a partir do código-fonte local (repositório privado ou rede interna)

Execute na raiz do repositório:

```bash
go install ./cmd/migrate
```

### 1.3 Verificar instalação

```bash
migrate --help
```

Se o comando não for encontrado:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Se você vir `Repository not found`, use o caminho de instalação pelo código-fonte local acima.

## 2. Inicialização

### 2.1 Gerar arquivo de configuração

```bash
migrate new config
```

Opcional:

```bash
migrate new config dev.json
migrate new config --force
```

Template padrão de configuração:

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 2.2 Atualizar campos principais da configuração

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: string de conexão do banco
- `migration_source`: diretório de migração (padrão: `migrations`)

### 2.3 Inicializar tabela de histórico de migrações

```bash
migrate create
```

`create` pode ser concluído com sucesso sem saída. Confirme com:

```bash
migrate status
```

Se você já tem um schema existente e não quer reaplicar SQL antigo, use:

```bash
migrate baseline
```

## 3. Criar arquivos de versão de migração

### 3.1 Versão gerada automaticamente

```bash
migrate new version init_users
```

### 3.2 Versão explícita

```bash
migrate new version add_email -v 202604140002
```

Formato de nome de arquivo gerado:

```text
V<version>__<description>.sql
```

Template padrão de arquivo:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. Upgrade (aplicar migrações)

Faça primeiro um dry run:

```bash
migrate up --dry-run
```

Aplique de fato:

```bash
migrate up
```

Depois verifique o status:

```bash
migrate status
```

`up` pode ser concluído com sucesso sem saída. Use `status` como fonte de verdade.

## 5. Rollback

### 5.1 Fazer rollback para uma versão alvo (a versão alvo é mantida)

```bash
migrate down 202604140001
```

Semântica: apenas versões aplicadas maiores que `202604140001` são revertidas.

### 5.2 Reverter todas as versões aplicadas

```bash
migrate down --all
```

Nota: `migrate down <to-version>` e `migrate down --all` são mutuamente exclusivos.

### 5.3 Rollback em dry-run

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` pode ser concluído com sucesso sem saída. Execute `migrate status` para verificar.

## 6. Verificar status

```bash
migrate status
```

Saída legível por máquina (recomendada para scripts e agentes de IA):

```bash
migrate status --output json
```

Colunas de saída: `Version`, `Filename`, `Hash`, `Status`.

Status comuns:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. Fazer upgrade ou downgrade da própria ferramenta

### 7.1 Fazer upgrade da versão da ferramenta

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 Fazer downgrade da versão da ferramenta

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Exemplo:

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

Se o repositório for privado e `go install github.com/...@...` não estiver disponível, faça checkout da versão alvo no código-fonte
e execute:

```bash
go install ./cmd/migrate
```

### 7.3 Verificar versão atual da ferramenta

```bash
migrate version
```

Nota: artefatos de release imprimem a versão de release; builds locais a partir de `go install ./cmd/migrate` normalmente imprimem `dev`.

## 8. Templates de variáveis de ambiente

`data_source_name` suporta:

- `${KEY}`: obrigatório, deve existir
- `${KEY:default}`: usa `default` se `KEY` estiver ausente

Exemplo:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

Garanta que `DB_PASSWORD` já esteja definido no seu ambiente, depois execute:

```bash
migrate status
```

## 9. Demo de primeira execução em 10 minutos (SQLite)

### 9.1 Preparar diretório e configuração

Primeiro verifique a disponibilidade do comando:

```bash
migrate --help
```

Crie o diretório de demonstração (macOS/Linux):

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Equivalente no Windows PowerShell:

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

Atualize `migration_config.json` para:

```json
{
  "schema_name": "migration_schema",
  "dialect": "sqlite",
  "data_source_name": "./demo.sqlite",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 9.2 Inicializar e criar arquivos de migração

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

Edite `migrations/V202604140001__init_users.sql`:

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

Edite `migrations/V202604140002__add_email.sql`:

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 Aplicar, verificar status e rollback

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

Esperado:

- após `up`: ambas versões estão `applied`
- após `down 202604140001`: `202604140001=applied`, `202604140002=pending`
- após `down --all`: ambas versões estão `pending`

## 10. Flags globais

Use arquivo de configuração específico:

```bash
migrate -c ./configs/dev.json status
```

Use diretório de trabalho específico:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. Erros comuns e troubleshooting

### 11.1 Arquivo de configuração não encontrado

Erro: `config file ... no such file or directory`

Correção:

- certifique-se de que `migration_config.json` existe no diretório atual
- ou passe o caminho da configuração com `-c`

### 11.2 Variável de ambiente ausente

Erro: `can't find env value for XXX`

Correção:

- `export XXX=...`
- ou use `${XXX:default}`

### 11.3 Argumentos incompletos de `down`

Erro: `to-version must be set, or use --all`

Correção:

- use `migrate down <version>`
- ou use `migrate down --all`

### 11.4 Dialeto não suportado

Erro: `unsupported dialect: xxx`

Correção: use um dos seguintes:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 Incompatibilidade de metadados de migração

Erro: `hash mismatch` ou `filename mismatch`

Correção:

- não edite arquivos de migração já aplicados
- crie uma nova migração com versão maior para mudanças

## 12. Usar migrate como biblioteca Go

Além do CLI, `github.com/exc-works/migrate` pode ser importado diretamente do código do seu serviço para executar migrações — útil para testes unitários, hooks de inicialização ou painéis administrativos.

### 12.1 Instalação

```bash
go get github.com/exc-works/migrate
```

Importe o driver de banco de dados que precisar (a biblioteca não fixa nenhum):

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 Exemplo mínimo

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/exc-works/migrate"
)

func main() {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    svc, err := migrate.NewService(context.Background(), migrate.Config{
        Dialect:         migrate.NewSQLiteDialect(),
        DB:              db,
        MigrationSource: migrate.DirectorySource{Directory: "./migrations"},
    })
    if err != nil {
        panic(err)
    }

    if err := svc.Create(); err != nil { // idempotente: cria a tabela de histórico se ausente
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 API principal

- `migrate.NewService(ctx, migrate.Config)` constrói um executor de migrações
- `svc.Create()` cria a tabela de histórico `migration_schema` (idempotente)
- `svc.Up()` aplica todas as migrações pendentes
- `svc.Down(toVersion, all)` reverte até uma versão alvo ou tudo
- `svc.Status()` retorna `[]migrate.MigrationStatus`
- `svc.Baseline()` marca arquivos pendentes existentes como `baseline`

Tipos comuns:

- Dialetos (prefira os construtores — retornam a interface `Dialect`): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, ou `migrate.DialectFromName("postgres")` para busca por nome
- Fontes: `DirectorySource` (sistema de arquivos), `StringSource` (slice em memória, prático em testes), `CombinedSource` (combina várias fontes)
- Loggers: `migrate.NoopLogger{}` (padrão), `migrate.NewStdLogger("info", os.Stdout)` ou qualquer tipo que implemente `migrate.Logger`

### 12.4 Amigável para testes: StringSource + SQLite em memória

```go
src := migrate.StringSource{Migrations: []migrate.SourceFile{{
    Filename: "V1__init.sql",
    Source:   "-- +migrate Up\nCREATE TABLE t(id INT);\n-- +migrate Down\nDROP TABLE t;\n",
}}}

db, _ := sql.Open("sqlite", ":memory:")
svc, _ := migrate.NewService(ctx, migrate.Config{
    Dialect:         migrate.NewSQLiteDialect(),
    DB:              db,
    MigrationSource: src,
})
```

Sem dependência do sistema de arquivos — roda direto de um teste unitário.

### 12.5 Pré-visualizar SQL (DryRun)

```go
var buf bytes.Buffer
svc, _ := migrate.NewService(ctx, migrate.Config{
    Dialect:         migrate.NewPostgresDialect(),
    DB:              db,
    MigrationSource: src,
    DryRun:          true,
    DryRunOutput:    &buf,
})
_ = svc.Create() // Create() não é afetado por DryRun; prepara a tabela de histórico
_ = svc.Up()     // o SQL das migrações vai para buf; nenhuma tabela de usuário é criada
```

### 12.6 Contrato de estabilidade

- `github.com/exc-works/migrate` (pacote raiz) é a API pública e segue SemVer
- `internal/*` não está coberto pelo contrato de estabilidade — não importe diretamente
- Um exemplo completo executável está em `example_test.go` na raiz do repositório
