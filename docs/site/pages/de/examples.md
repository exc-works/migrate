# Beispiele ausführen

Aus einem ausgecheckten Repository mit SQLite, ohne Server. Die Migrationen prüfen ihre Daten selbst.

## Temporäre CLI erstellen

Dieselbe Shell offen lassen.

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## Lokale Hilfsdatei {#local}

[Quellen](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local): Zwei Migrationen nutzen dieselbe TypeScript-Hilfsdatei.

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

Nach up sind drei Versionen angewandt und die Namen `alice`, `bob`. `down 2` behält 1 und 2, `down --all` entfernt die Tabelle.

## npm-Paket

Node/npm wird nur zum Installieren gebraucht. Das Lockfile verweist auf ein enthaltenes lokales Paket, ohne Registryzugriff.

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

Die zweite Migration importiert `@migrate-example/normalizer/names` und prüft `alice`. Zwei Versionen sind angewandt.

## In Go einbetten

```sh
go run ./examples/script-imports/embedded
```

`embed.FS`, private Speicherdatenbank, kein npm. Ausgabe:

```text
after up: alice
after down: all migrations pending
```

## Aufräumen

```sh
rm -r "$example_workspace"
```

npm hinterlässt das ignorierte `node_modules`. Automatisch prüfen mit `./scripts/test-script-examples.sh` im Repository. Siehe [Reproduzierbarkeit](./modules#checksums).
