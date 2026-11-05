# Référence CLI

Consultez `migrate --help` ou `migrate <command> --help` pour votre version, et `migrate version` pour l’identifier.

## Options globales

| Option | Effet |
| --- | --- |
| `-c`, `--config` | Fichier précis ; sinon [recherche automatique](./configuration#configuration-file-search-order) |
| `-w`, `--working-dir` | Remplace le dossier de travail |
| `--non-interactive` | Échoue au lieu de demander un secret ou l’approbation d’une clé SSH inconnue |

## Générer

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

Défaut : `migration_config.yml`. `--shards N` exige `--multi` et un nombre positif, numéroté à partir de 0 ; `--proxy` exige `--ssh-tunnel`. Authentification : `--ssh-auth agent|private_key|password`. `--force` écrase un fichier ; `.json` ou `--format json` choisit JSON.

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

Format par défaut SQL, version automatique ou `-v`. Écrivez `"init users"` pour les descriptions avec espaces, convertis en `_` dans le nom.

## Gérer la base

```sh
migrate create
```

Crée l’historique de façon idempotente, sans appliquer les fichiers.

```sh
migrate baseline
```

Enregistre sans exécuter, uniquement si la base correspond déjà aux fichiers. Voir [baseline](./history#baseline).

```sh
migrate up [--dry-run]
```

Applique les versions en attente. Dry-run peut lire l’historique mais ne fait pas les changements.

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

Annule les versions au-dessus de la cible en la conservant, ou toutes les versions. Choisissez la cible ou `--all`, pas les deux.

```sh
migrate status [--output table|json]
```

Table par défaut. En mode multibase, le JSON trié par nom contient `data_source`, `status`, puis `migrations` en succès ou `error` en échec.

create/baseline/status/up/down traitent toutes les connexions et acceptent `--parallel N` (8 par défaut). Une erreur ne bloque pas les autres bases et produit une sortie non nulle.

## Informations sans connexion

```sh
migrate capabilities [--output text|json]
```

Affiche formats, alias, prise en charge des scripts et transactions ; défaut `text`.

```sh
migrate version
```

Affiche la version de l’outil, pas celle des migrations.
