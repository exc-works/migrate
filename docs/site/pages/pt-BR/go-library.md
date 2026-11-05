# Usar em Go

A API pública do pacote raiz segue SemVer; `internal/*` não é público. O chamador fecha o pool, usa context apropriado e serializa migrações no mesmo banco.

`Config.DryRun` e `DryRunOutput` mostram a prévia, mas `Create()` sempre cria o histórico. Logger padrão: `NoopLogger`. Para [scripts embutidos](./modules), use `FSSource` com dependências.

Com `OpenConnection`, use `Connection.Context()` para Ping e serviço e depois `Close()`. Drivers de rede são registrados; SQLite exige registro pelo chamador. `ConnectionOptions.Prompt` e `HostKeyPrompt` são opcionais e devem aceitar cancelamento. Sem callbacks não há entrada nem confiança em hosts desconhecidos.

Além do CLI, `github.com/exc-works/migrate` pode ser importado diretamente do código do seu serviço para executar migrações — útil para testes unitários, hooks de inicialização ou painéis administrativos.

## Instalação

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

## Exemplo mínimo

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

## API principal

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

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
