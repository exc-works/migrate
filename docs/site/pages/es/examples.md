# Ejecutar ejemplos

Desde una copia del repositorio, usa SQLite sin servidor. Las migraciones verifican sus resultados.

## Compilar una CLI temporal

Mantén el mismo shell abierto.

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## Archivo auxiliar local {#local}

[Fuentes](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local): dos migraciones importan el mismo TypeScript auxiliar.

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

Tras up hay tres versiones aplicadas y nombres `alice` y `bob`. `down 2` conserva 1 y 2; `down --all` elimina la tabla.

## Paquete npm

Node/npm solo se necesita para instalar. El lockfile referencia un paquete local, sin registro externo.

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

La segunda migración importa `@migrate-example/normalizer/names` y verifica `alice`. Se aplican dos versiones.

## Embebido en Go

```sh
go run ./examples/script-imports/embedded
```

Usa `embed.FS` y una base privada en memoria, sin npm. Salida:

```text
after up: alice
after down: all migrations pending
```

## Limpiar

```sh
rm -r "$example_workspace"
```

npm deja el directorio ignorado `node_modules`. Desde la raíz, `./scripts/test-script-examples.sh` valida copias temporales automáticamente. Consulta la [reproducibilidad](./modules#checksums).
