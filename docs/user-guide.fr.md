# Guide utilisateur de migrate

## 1. Installation

### 1.1 Installer depuis le module

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

Installer une version spécifique :

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Remplacez `vX.Y.Z` par une vraie version, par exemple `v0.2.3`.

### 1.2 Installer depuis le code source local (repo privé ou réseau interne)

Exécutez dans la racine du dépôt :

```bash
go install ./cmd/migrate
```

### 1.3 Vérifier l’installation

```bash
migrate --help
```

Si la commande est introuvable :

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Si vous voyez `Repository not found`, utilisez le chemin d’installation depuis le code source local ci-dessus.

## 2. Initialisation

### 2.1 Générer le fichier de configuration

```bash
migrate new config
```

Optionnel :

```bash
migrate new config dev.yml
migrate new config --force
```

Modèle de configuration par défaut :

```yaml
schema_name: migration_schema
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:postgres} sslmode=disable"
working_directory: ""
migrate_out_of_order: false
logger_level: info
migration_source: migrations
```

### 2.2 Mettre à jour les champs clés de la configuration

- `dialect` : `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name` : chaîne de connexion à la base de données
- `migration_source` : répertoire des migrations (par défaut : `migrations`)

### 2.3 Initialiser la table d’historique des migrations

```bash
migrate create
```

`create` peut réussir sans sortie. Confirmez avec :

```bash
migrate status
```

Si vous avez déjà un schéma existant et que vous ne souhaitez pas rejouer les anciens SQL, utilisez :

```bash
migrate baseline
```

## 3. Créer des fichiers de version de migration

### 3.1 Version générée automatiquement

```bash
migrate new version "init users"
```

### 3.2 Version explicite

```bash
migrate new version "add email" -v 202604140002
```

Format de nom de fichier généré :

```text
V<version>__<description>.sql
```

Modèle de fichier par défaut :

```sql
-- +migrate Up

-- +migrate Down
```

## 4. Upgrade (appliquer les migrations)

Faites d’abord un dry run :

```bash
migrate up --dry-run
```

Appliquez réellement :

```bash
migrate up
```

Puis vérifiez le statut :

```bash
migrate status
```

`up` peut réussir sans sortie. Utilisez `status` comme source de vérité.

## 5. Rollback

### 5.1 Revenir à une version cible (la version cible est conservée)

```bash
migrate down 202604140001
```

Sémantique : seules les versions appliquées supérieures à `202604140001` sont annulées.

### 5.2 Annuler toutes les versions appliquées

```bash
migrate down --all
```


### 5.3 Rollback en dry run

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` peut réussir sans sortie. Exécutez `migrate status` pour vérifier.

## 6. Vérifier le statut

```bash
migrate status
```

Sortie lisible par machine (recommandée pour les scripts et les agents IA) :

```bash
migrate status --output json
```

Colonnes de sortie : `Version`, `Filename`, `Hash`, `Status`.

Statuts courants :

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. Vérifier la version

```bash
migrate version
```

## 8. Modèles de variables d’environnement

`data_source_name` prend en charge :

- `${KEY}` : requis, doit exister
- `${KEY:default}` : utilise `default` si `KEY` est absent

Exemple :

```yaml
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
```

Assurez-vous que `DB_PASSWORD` est déjà défini dans votre environnement, puis exécutez :

```bash
migrate status
```

## 9. Démo premier lancement en 10 minutes (SQLite)

[Démo premier lancement en 10 minutes (SQLite) · English](site/pages/quick-start.md)

## 10. Flags globaux

Utiliser un fichier de configuration spécifique :

```bash
migrate -c ./configs/dev.yml status
```

Utiliser un répertoire de travail spécifique :

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. Erreurs courantes et dépannage

### 11.1 Fichier de configuration introuvable

Erreur : `config file ... no such file or directory`

Correction :

- assurez-vous que `migration_config.yml` existe dans le répertoire courant
- ou passez le chemin de configuration avec `-c`

### 11.2 Variable d’environnement manquante

Erreur : `can't find env value for XXX`

Correction :

- `export XXX=...`
- ou utilisez `${XXX:default}`

### 11.3 Arguments `down` incomplets

Erreur : `to-version must be set, or use --all`

Correction :

- utilisez `migrate down <version>`
- ou utilisez `migrate down --all`

### 11.4 Dialect non pris en charge

Erreur : `unsupported dialect: xxx`

Correction : utilisez l’un de :

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 Incohérence des métadonnées de migration

Erreur : `hash mismatch` ou `filename mismatch`

Correction :

- restaurez le fichier original depuis le contrôle de version
- créez une nouvelle migration avec un numéro de version supérieur pour vos changements

## 12. Utiliser migrate comme bibliothèque Go

En plus du CLI, `github.com/exc-works/migrate` peut être importé directement depuis le code de votre service pour exécuter les migrations — pratique pour les tests unitaires, les hooks de démarrage ou les panneaux d’administration.

### 12.1 Installation

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

### 12.2 Exemple minimal

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

### 12.3 API clé

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

[Go API](../migrate.go) · [Go example](../example_test.go)
