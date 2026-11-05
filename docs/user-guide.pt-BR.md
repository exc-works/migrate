# Guia do Usuário do migrate

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
migrate new config dev.yml
migrate new config --force
```

Template padrão de configuração:

```yaml
schema_name: migration_schema
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:postgres} sslmode=disable"
working_directory: ""
migrate_out_of_order: false
logger_level: info
migration_source: migrations
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
migrate new version "init users"
```

### 3.2 Versão explícita

```bash
migrate new version "add email" -v 202604140002
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

## 7. Verificar a versão

```bash
migrate version
```

## 8. Templates de variáveis de ambiente

`data_source_name` suporta:

- `${KEY}`: obrigatório, deve existir
- `${KEY:default}`: usa `default` se `KEY` estiver ausente

Exemplo:

```yaml
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
```

Garanta que `DB_PASSWORD` já esteja definido no seu ambiente, depois execute:

```bash
migrate status
```

## 9. Demo de primeira execução em 10 minutos (SQLite)

[Demo de primeira execução em 10 minutos (SQLite) · English](site/pages/quick-start.md)

## 10. Flags globais

Use arquivo de configuração específico:

```bash
migrate -c ./configs/dev.yml status
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

- certifique-se de que `migration_config.yml` existe no diretório atual
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

- restaure o arquivo original a partir do controle de versão
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
- Fontes: `DirectorySource` (sistema de arquivos), `StringSource` (slice em memória), `FSSource` (qualquer `fs.FS`, como `//go:embed` ou `os.DirFS`), `CombinedSource` (combina várias fontes)
- Loggers: `migrate.NoopLogger{}` (padrão), `migrate.NewStdLogger("info", os.Stdout)` ou qualquer tipo que implemente `migrate.Logger`

[Go API](../migrate.go) · [Go example](../example_test.go)
