# Démarrage rapide avec SQLite

Créez, appliquez et annulez une migration sans serveur de base de données.

## Installer

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

## Configurer

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

Remplacez `migration_config.yml` par :

```yaml
# Table de l’historique
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # Fichier SQLite local
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # Dossier des migrations
```

## Créer une migration

```sh
migrate create
migrate new version "init users" -v 202611030001
```

Remplacez `migrations/V202611030001__init_users.sql` par :

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## Exécuter et vérifier

```sh
migrate up --dry-run
migrate up
migrate status
```

La version `202611030001` doit être `applied`.

## Annuler

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

La table disparaît et la version redevient `pending`. Continuez avec [SQL](./sql-migrations), la [configuration](./configuration) ou les [scripts](./scripting).
