# SQLite 快速开始

在新的 SQLite 项目中完成一次迁移的创建、执行和回退。

## 安装

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

源码安装和 PATH 设置见[安装](./installation)，
支持的迁移格式见[数据库能力](./databases)。

## 配置

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

将生成的 `migration_config.json` 替换为：

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

## 创建迁移

```sh
migrate create
migrate new version init_users -v 202611030001
```

将 `migrations/V202611030001__init_users.sql` 替换为：

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## 执行并验证

```sh
migrate up --dry-run
migrate up
migrate status
```

确认版本 `202611030001` 显示为 `applied`。

## 回退

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

确认示例表已删除，版本恢复为 `pending`。

继续阅读[编写 SQL 迁移](./sql-migrations)、[配置说明](./configuration)
或 [JavaScript 和 TypeScript 迁移](./scripting)。
