# Справочник CLI

Версия: `migrate version`. Параметры: `migrate --help`, `migrate <command> --help`.

| Общий параметр | Действие |
| --- | --- |
| `-c`, `--config` | Конкретный файл, иначе [поиск](./configuration#configuration-file-search-order) |
| `-w`, `--working-dir` | Заменить рабочий каталог |
| `--non-interactive` | Ошибка вместо запроса секрета или доверия SSH |

## Генерация

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

По умолчанию `migration_config.yml`. `--shards N` требует `--multi` и положительное число, нумерация с 0. `--proxy` требует `--ssh-tunnel`. Аутентификация: `--ssh-auth agent|private_key|password`. `--force` перезаписывает; `.json` или `--format json` выбирает JSON.

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

SQL по умолчанию, версия автоматическая или `-v`. Описание `"init users"` заключайте в кавычки; пробелы в имени становятся `_`.

## База данных

```sh
migrate create
```

Идемпотентно создаёт только историю.

```sh
migrate baseline
```

Записывает без выполнения, только если база уже соответствует файлам. См. [baseline](./history#baseline).

```sh
migrate up [--dry-run]
```

Применяет ожидающие. Dry-run может читать историю, не изменяя данные.

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

Откат выше цели с её сохранением либо всего. Цель и `--all` исключают друг друга.

```sh
migrate status [--output table|json]
```

По умолчанию таблица. Для нескольких баз JSON сортируется по именам: `data_source`, `status`, при успехе `migrations`, при сбое `error`. create/baseline/status/up/down обходят все базы, поддерживают `--parallel N` (8). Ошибка не останавливает остальные и даёт ненулевой код.

## Сведения без подключения

```sh
migrate capabilities [--output text|json]
```

Форматы, псевдонимы, скрипты и транзакции. По умолчанию `text`.

```sh
migrate version
```

Версия инструмента, не миграций.
