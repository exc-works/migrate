# Exécuter les exemples

Depuis une copie du dépôt, utilisez SQLite sans serveur. Les migrations vérifient leurs propres résultats.

## Construire la CLI temporaire

Conservez le même shell pour la suite.

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## Utilitaire local {#local}

[Sources](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local) : deux migrations importent un utilitaire TypeScript partagé.

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

Après up, trois versions sont appliquées et les noms sont `alice` et `bob`. `down 2` garde 1 et 2 ; `down --all` supprime la table.

## Paquet npm

Node/npm sert uniquement à l’installation. Le lockfile utilise un paquet local, sans accès au registre.

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

La deuxième migration importe `@migrate-example/normalizer/names` et vérifie `alice`. Deux versions sont appliquées.

## Fichiers embarqués dans Go

```sh
go run ./examples/script-imports/embedded
```

L’exemple utilise `embed.FS` et une base mémoire privée, sans npm. Résultat :

```text
after up: alice
after down: all migrations pending
```

## Nettoyer

```sh
rm -r "$example_workspace"
```

Le dossier ignoré `node_modules` reste après npm. Depuis la racine, `./scripts/test-script-examples.sh` vérifie automatiquement des copies temporaires. Voir la [reproductibilité](./modules#checksums).
