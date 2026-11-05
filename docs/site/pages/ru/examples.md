# Запуск примеров

Из копии репозитория. Все используют SQLite и проверяют свои данные.

## Временная CLI

Продолжайте в том же shell.

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## Локальный модуль {#local}

[Исходники](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local): две миграции используют общий TypeScript.

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

После up три версии и имена `alice`, `bob`. `down 2` оставляет 1, 2; `down --all` удаляет таблицу.

## npm

Node/npm нужен лишь для установки. Lockfile ссылается на локальный пакет, без реестра.

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

Вторая миграция импортирует `@migrate-example/normalizer/names` и проверяет `alice`; применено две версии.

## Встроить в Go

```sh
go run ./examples/script-imports/embedded
```

Использует `embed.FS` и отдельную память SQLite, без npm. Вывод:

```text
after up: alice
after down: all migrations pending
```

## Очистка

```sh
rm -r "$example_workspace"
```

npm оставляет игнорируемый `node_modules`. Для автоматической проверки временных копий: `./scripts/test-script-examples.sh` из корня. См. [воспроизводимость](./modules#checksums).
