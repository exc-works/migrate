# 配置

在迁移项目目录运行命令。migrate 默认读取 `migration_config.json`，可先生成配置：

```sh
migrate new config
```

指定其他文件名可以使用 `migrate new config dev.json`。
已有文件只有在显式传入 `--force` 时才会被覆盖。

## 连接数据库

下面的 PostgreSQL 配置通过环境变量提供连接字符串：

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "${DATABASE_DSN}",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

执行数据库命令前，通过 shell 或部署环境的 secret 存储设置 `DATABASE_DSN`。
本地 SQLite 项目可改为 `"dialect": "sqlite"` 和
`"data_source_name": "./demo.sqlite"`。

| 字段 | 用途 | 省略时的默认值 |
| --- | --- | --- |
| `schema_name` | 迁移历史表名称 | `migration_schema` |
| `dialect` | 数据库方言，见[可接受名称](./databases#aliases) | `postgres` |
| `data_source_name` | 驱动连接字符串 | 必填 |
| `working_directory` | 相对迁移目录的解析基准 | 当前目录 |
| `migration_source` | 存放版本文件的目录 | `migrations` |
| `migrate_out_of_order` | 是否执行低于已应用版本的待执行版本 | `false` |
| `logger_level` | 日志级别 | `info` |

## 环境变量模板

`data_source_name` 支持 `${KEY}` 和 `${KEY:default}`。前者要求变量存在，
后者在变量不存在时使用默认值。例如非敏感的数据库名可写为 `${DB_NAME:app}`。
凭据应通过环境变量提供。

## 选择配置与工作目录

```sh
migrate -c ./configs/dev.json status
migrate -w ./deploy up
```

`-c` 选择配置文件。`-w` 指定工作目录，从该目录解析配置路径，并覆盖配置中的
`working_directory` 用于查找迁移文件。相对 `migration_source` 从工作目录解析，
不会因为使用自定义 `-c` 文件就自动改为相对配置文件所在目录。

脚本限制见[事务与执行限制](./script-transactions)；经过堡垒机连接时，
添加 [SSH 隧道配置](./ssh-tunnels)。配置完成后，可以[编写第一条 SQL 迁移](./sql-migrations)。
