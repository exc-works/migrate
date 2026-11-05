# Настройка

Создайте и отредактируйте `migration_config.yml` в каталоге проекта:

```sh
migrate new config
```

## Подключение

PostgreSQL со строкой подключения из `DATABASE_DSN`:

```yaml
# Тип базы данных
dialect: postgres
# Задайте DATABASE_DSN перед запуском
data_source_name: "${DATABASE_DSN}"
# Каталог файлов миграций
migration_source: migrations
# Таблица применённых версий
schema_name: migration_schema
```

| Поле | Назначение | По умолчанию |
| --- | --- | --- |
| `dialect` | Тип базы | `postgres` |
| `data_source_name` | Одна база, несовместимо с `data_sources` | Обязательно для одной базы |
| `data_sources` | Именованные подключения | Выключено |
| `migration_source` | Каталог файлов | `migrations` |
| `schema_name` | Таблица истории | `migration_schema` |
| `working_directory` | Основа поиска файлов | Текущий каталог |
| `migrate_out_of_order` | Выполнять старые ожидающие версии | `false` |
| `logger_level` | `debug`, `info`, `warn`, `error` | `info` |

### Примеры подключений

Замените хост, порт, базу или сервис Oracle и задайте `DB_USER`, `DB_PASSWORD`. SQLite использует путь к файлу.

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

В URL кодируйте специальные символы имени и пароля (`@`, `:`, `/`): подстановка переменных этого не делает. Формат `user:password@tcp(...)` MySQL/MariaDB/TiDB не использует URL-кодирование. TLS/SSL задаётся по требованиям сервера.

### Уровни журнала {#logging}

| Значение | Сообщения |
| --- | --- |
| `debug` | Отладка, информация, предупреждения, ошибки |
| `info` | Информация, предупреждения, ошибки |
| `warn` | Предупреждения и ошибки; также `warning` |
| `error` | Только ошибки |

```yaml
logger_level: debug # Показывать отладочные сообщения
```

## Переменные окружения {#environment-variables}

`${DATABASE_DSN}` требует заданную переменную; `${DB_NAME:app}` использует `app`, если переменной нет. YAML можно разделять между окружениями, передавая секреты через переменные или хранилище секретов. Хост и пользователь SSH, адрес прокси тоже поддерживают шаблоны.

## Порядок поиска {#configuration-file-search-order}

В рабочем каталоге используется первый найденный файл:

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

Файлы не объединяются. Ошибка или недоступность найденного файла останавливает команду. `-c` выбирает только указанный файл, даже если его нет.

```sh
migrate -c ./configs/dev.yml status
```

## Генерация настроек

| Сценарий | Команда |
| --- | --- |
| Одна база | `migrate new config` |
| Другое имя | `migrate new config dev.yml` |
| Несколько баз | `migrate new config --multi` |
| Четыре базы | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| Закрытый ключ | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| Пароль SSH | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

Параметры можно сочетать. Перезапись требует `--force`; все параметры — `migrate new config --help`.

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## Несколько баз {#multiple-databases}

Вместо `data_source_name` задайте `data_sources`. Тип, файлы и имя таблицы истории общие. Создаваемые имена начинаются с `shard_0`, переменные — с `SHARD_0_DSN`.

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # Подключение первой базы
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # Подключение второй базы
```

![Файлы миграций → shard_0 / shard_1 → Результаты](../assets/diagrams/multiple-databases-ru.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

Каждая база по своей истории выбирает ожидающие версии и выполняет их по порядку. Успешные не повторяются. `new version` создаёт один общий файл. Одновременно обрабатывается до 8 баз; `--parallel 1` выполняет последовательно.

```sh
migrate up --parallel 4
migrate status --output json
```

Ошибка не останавливает остальные базы. Результаты группируются по имени; любая ошибка даёт ненулевой код завершения. Общей транзакции нет: изменения других баз не откатываются. Ctrl+C прекращает запуск новых баз и отменяет активные операции.

Общий `ssh_tunnel` задаётся сверху. Локальный объект заменяет его целиком; отсутствие или `null` наследует общий. См. [SSH](./ssh-tunnels). Указывайте каждую реальную базу один раз, не запускайте перекрывающиеся команды и не меняйте файлы и зависимости при работе. Перед повтором после ошибки проверяйте данные и историю.

## Рабочий каталог

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` определяет поиск настроек и заменяет `working_directory`. Относительный `migration_source` считается от него; один `-c` не меняет основу. Относительные SSH-пути считаются от файла настроек.
