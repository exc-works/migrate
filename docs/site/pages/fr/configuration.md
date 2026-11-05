# Configuration

Depuis le dossier du projet, générez puis modifiez `migration_config.yml` :

```sh
migrate new config
```

## Connexion à la base

Cette configuration PostgreSQL lit la connexion dans `DATABASE_DSN` :

```yaml
# Type de base
dialect: postgres
# Définir DATABASE_DSN avant l’exécution
data_source_name: "${DATABASE_DSN}"
# Dossier des fichiers de migration
migration_source: migrations
# Table des versions appliquées
schema_name: migration_schema
```

| Champ | Rôle | Valeur par défaut |
| --- | --- | --- |
| `dialect` | Type de base | `postgres` |
| `data_source_name` | Connexion unique, incompatible avec `data_sources` | Obligatoire en mode unique |
| `data_sources` | Connexions nommées | Désactivé |
| `migration_source` | Dossier des migrations | `migrations` |
| `schema_name` | Table d’historique | `migration_schema` |
| `working_directory` | Base de recherche des fichiers | Dossier courant |
| `migrate_out_of_order` | Appliquer les versions en attente plus anciennes | `false` |
| `logger_level` | `debug`, `info`, `warn`, `error` | `info` |

### Exemples de connexion

Adaptez hôte, port, base ou service Oracle et définissez `DB_USER` et `DB_PASSWORD`. SQLite utilise directement un chemin de fichier.

::: code-group

```yaml [PostgreSQL]
dialect: postgres
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@db.example:5432/app?sslmode=require"
```

```yaml [MySQL]
dialect: mysql
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [MariaDB]
dialect: mariadb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [SQLite]
dialect: sqlite
data_source_name: "./demo.sqlite"
```

```yaml [Oracle]
dialect: oracle
data_source_name: "oracle://${DB_USER}:${DB_PASSWORD}@db.example:1521/FREEPDB1"
```

```yaml [SQL Server]
dialect: mssql
data_source_name: "sqlserver://${DB_USER}:${DB_PASSWORD}@db.example:1433?database=app"
```

```yaml [ClickHouse]
dialect: clickhouse
data_source_name: "clickhouse://${DB_USER}:${DB_PASSWORD}@db.example:9000/app"
```

```yaml [TiDB]
dialect: tidb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:4000)/app?parseTime=true"
```

```yaml [Redshift]
dialect: redshift
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@cluster.example:5439/app?sslmode=require"
```

:::

Dans une URL, encodez les caractères spéciaux du nom et du mot de passe (`@`, `:`, `/`). Les variables ne sont pas encodées automatiquement. Le format `user:password@tcp(...)` de MySQL/MariaDB/TiDB n’utilise pas cet encodage. Réglez TLS/SSL selon les exigences du serveur.

### Niveaux de journalisation {#logging}

| Valeur | Messages affichés |
| --- | --- |
| `debug` | Débogage, informations, avertissements, erreurs |
| `info` | Informations, avertissements, erreurs |
| `warn` | Avertissements et erreurs ; alias `warning` |
| `error` | Erreurs uniquement |

```yaml
logger_level: debug # Afficher les messages de débogage
```

## Variables d’environnement {#environment-variables}

`${DATABASE_DSN}` exige une variable définie. `${DB_NAME:app}` utilise `app` si la variable manque. Réutilisez le YAML entre environnements, en fournissant les identifiants par variables ou coffre de secrets. Hôte SSH, utilisateur et adresse du proxy acceptent aussi ces modèles.

## Ordre de recherche {#configuration-file-search-order}

Dans le dossier de travail, le premier fichier existant est utilisé :

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

Les fichiers ne sont pas fusionnés. Un fichier sélectionné invalide ou illisible provoque une erreur. Avec `-c`, seul le fichier indiqué est consulté, même s’il manque.

```sh
migrate -c ./configs/dev.yml status
```

## Générer selon le besoin

| Besoin | Commande |
| --- | --- |
| Une base | `migrate new config` |
| Nom personnalisé | `migrate new config dev.yml` |
| Plusieurs bases | `migrate new config --multi` |
| Quatre bases | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| Clé privée | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| Mot de passe SSH | `migrate new config --ssh-tunnel --ssh-auth password` |
| Proxy SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| Proxy HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

Combinez les options. L’écrasement exige `--force`. Toutes les options sont dans `migrate new config --help`.

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## Plusieurs bases {#multiple-databases}

Remplacez `data_source_name` par `data_sources`. Type de base, fichiers et nom de table d’historique sont partagés. Les noms générés commencent à `shard_0`, avec les variables `SHARD_0_DSN`, etc.

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # Connexion à la première base
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # Connexion à la seconde base
```

![Fichiers → shard_0 / shard_1 → Résultats](../assets/diagrams/multiple-databases-fr.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

Chaque base choisit les versions en attente selon son propre historique et les applique dans l’ordre. Les versions réussies sont ignorées aux exécutions suivantes. `new version` crée un seul fichier partagé. Au plus 8 bases sont traitées simultanément par défaut ; `--parallel 1` impose un traitement séquentiel.

```sh
migrate up --parallel 4
migrate status --output json
```

Une erreur n’arrête pas les autres bases. Les résultats sont regroupés par nom et toute erreur produit un code de sortie non nul. Ce n’est pas une transaction entre bases : les changements ailleurs restent appliqués. Ctrl+C arrête les nouveaux traitements et annule ceux en cours.

Un `ssh_tunnel` global sert de configuration commune. Un objet SSH local le remplace entièrement ; absent ou `null`, il l’hérite. Voir [SSH](./ssh-tunnels). Ne configurez chaque base réelle qu’une fois, ne lancez pas de migrations concurrentes dessus et ne modifiez pas les dépendances pendant l’exécution. Vérifiez données et historique avant de relancer après un échec.

## Dossier de travail

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` détermine la recherche de configuration et remplace `working_directory`. `migration_source` relatif est résolu depuis ce dossier ; `-c` seul ne le déplace pas. Les chemins SSH relatifs partent du dossier du fichier de configuration.
