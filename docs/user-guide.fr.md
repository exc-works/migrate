# Guide utilisateur de migrate

Ce guide est destiné aux utilisateurs qui démarrent. Les commandes et flags sont basés sur l’implémentation actuelle (`cmd/migrate`).

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
migrate new config dev.json
migrate new config --force
```

Modèle de configuration par défaut :

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
migrate new version init_users
```

### 3.2 Version explicite

```bash
migrate new version add_email -v 202604140002
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

Remarque : `migrate down <to-version>` et `migrate down --all` sont mutuellement exclusifs.

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

## 7. Mettre à niveau ou rétrograder l’outil lui-même

### 7.1 Mettre à niveau la version de l’outil

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 Rétrograder la version de l’outil

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

Exemple :

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

Si le dépôt est privé et que `go install github.com/...@...` n’est pas disponible, récupérez la version cible dans le code source
puis exécutez :

```bash
go install ./cmd/migrate
```

### 7.3 Vérifier la version actuelle de l’outil

```bash
migrate version
```

Remarque : les artefacts de release affichent la version de release ; les builds locaux depuis `go install ./cmd/migrate` affichent généralement `dev`.

## 8. Modèles de variables d’environnement

`data_source_name` prend en charge :

- `${KEY}` : requis, doit exister
- `${KEY:default}` : utilise `default` si `KEY` est absent

Exemple :

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

Assurez-vous que `DB_PASSWORD` est déjà défini dans votre environnement, puis exécutez :

```bash
migrate status
```

## 9. Démo premier lancement en 10 minutes (SQLite)

### 9.1 Préparer le répertoire et la configuration

Vérifiez d’abord la disponibilité de la commande :

```bash
migrate --help
```

Créer le répertoire de démo (macOS/Linux) :

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Équivalent Windows PowerShell :

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

Mettez à jour `migration_config.json` en :

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

### 9.2 Initialiser et créer les fichiers de migration

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

Modifiez `migrations/V202604140001__init_users.sql` :

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

Modifiez `migrations/V202604140002__add_email.sql` :

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

### 9.3 Appliquer, vérifier le statut et rollback

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

Attendu :

- après `up` : les deux versions sont `applied`
- après `down 202604140001` : `202604140001=applied`, `202604140002=pending`
- après `down --all` : les deux versions sont `pending`

## 10. Flags globaux

Utiliser un fichier de configuration spécifique :

```bash
migrate -c ./configs/dev.json status
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

- assurez-vous que `migration_config.json` existe dans le répertoire courant
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

- ne modifiez pas les fichiers de migration déjà appliqués
- créez une nouvelle migration avec un numéro de version supérieur pour vos changements
