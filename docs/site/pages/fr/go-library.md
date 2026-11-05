# Utiliser depuis Go

L’API publique du paquet racine suit SemVer ; `internal/*` n’en fait pas partie. L’appelant gère la fermeture du pool et sérialise les migrations sur une même base avec un context adapté.

`Config.DryRun` et `DryRunOutput` prévisualisent les changements, mais `Create()` crée toujours l’historique. Le logger par défaut est `NoopLogger`. Pour [embarquer des scripts](./modules), utilisez `FSSource` et ses dépendances.

Avec `OpenConnection`, utilisez `Connection.Context()` pour Ping et le service, puis `Close()`. Les pilotes réseau sont enregistrés ; SQLite exige un enregistrement côté appelant. `ConnectionOptions.Prompt` et `HostKeyPrompt` sont facultatifs et doivent gérer l’annulation. Sans eux, aucune saisie ni confiance automatique envers les hôtes inconnus.

En plus du CLI, `github.com/exc-works/migrate` peut être importé directement depuis le code de votre service pour exécuter les migrations — pratique pour les tests unitaires, les hooks de démarrage ou les panneaux d’administration.

## Installation

```bash
go get github.com/exc-works/migrate
```

Importez le pilote de base de données dont vous avez besoin (la bibliothèque n’en fige aucun) :

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

## Exemple minimal

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

    if err := svc.Create(); err != nil { // idempotent : crée la table d’historique si absente
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

## API clé

- `migrate.NewService(ctx, migrate.Config)` construit un exécuteur de migrations
- `svc.Create()` crée la table d’historique `migration_schema` (idempotent)
- `svc.Up()` applique toutes les migrations en attente
- `svc.Down(toVersion, all)` revient à une version cible ou à tout
- `svc.Status()` retourne `[]migrate.MigrationStatus`
- `svc.Baseline()` marque les fichiers en attente existants comme `baseline`

Types courants :

- Dialectes (préférez les constructeurs — ils renvoient l’interface `Dialect`) : `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, ou `migrate.DialectFromName("postgres")` pour une résolution par nom
- Sources : `DirectorySource` (système de fichiers), `StringSource` (slice en mémoire), `FSSource` (n’importe quel `fs.FS`, par ex. `//go:embed` ou `os.DirFS`), `CombinedSource` (fusionne plusieurs sources)
- Loggers : `migrate.NoopLogger{}` (par défaut), `migrate.NewStdLogger("info", os.Stdout)`, ou tout type implémentant `migrate.Logger`

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
