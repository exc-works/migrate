# migrate 用户指南

## 1. 安装

### 1.1 从模块安装

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

安装指定版本：

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

将 `vX.Y.Z` 替换为 [Releases](https://github.com/exc-works/migrate/releases) 中的版本号。

### 1.2 从本地源码安装（仓库私有或内网场景）

在仓库根目录执行：

```bash
go install ./cmd/migrate
```

### 1.3 验证安装

```bash
migrate --help
```

如果提示命令不存在：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

如果提示 `Repository not found`，优先改用“从本地源码安装”方式。

## 2. 初始化

### 2.1 生成配置文件

```bash
migrate new config
```

可选：

```bash
migrate new config dev.json
migrate new config --force
```

默认配置模板：

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 2.2 调整关键配置

- `dialect`：`postgres`、`mysql`、`mariadb`、`oracle`、`sqlite`、`mssql`、`clickhouse`、`tidb`、`redshift`
- `data_source_name`：数据库连接串
- `migration_source`：迁移文件目录（默认 `migrations`）

### 2.3 初始化迁移记录表

```bash
migrate create
```

查看迁移状态：

```bash
migrate status
```

接管已有数据库时，确认结构与迁移文件一致，再登记基线：

```bash
migrate baseline
```

## 3. 创建版本文件

### 3.1 自动版本号

```bash
migrate new version init_users
```

### 3.2 指定版本号

```bash
migrate new version add_email -v 202604140002
```

生成文件名格式：

```text
V<version>__<description>.sql
```

默认模板内容：

```sql
-- +migrate Up

-- +migrate Down
```

## 4. 升级（数据库迁移）

先演练：

```bash
migrate up --dry-run
```

正式执行：

```bash
migrate up
```

执行后查看状态：

```bash
migrate status
```

执行后，已应用的版本显示为 `applied`。

## 5. 回退（数据库迁移）

### 5.1 回退到目标版本（不含目标版本）

```bash
migrate down 202604140001
```

语义：仅回退版本号大于 `202604140001` 的已应用迁移。

### 5.2 回退全部

```bash
migrate down --all
```


### 5.3 回退演练（不落库）

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

执行 `migrate status`，确认回退的版本已恢复为 `pending`。

## 6. 查看状态

```bash
migrate status
```

在脚本中读取 JSON：

```bash
migrate status --output json
```

输出列：`Version`、`Filename`、`Hash`、`Status`。

常见状态：

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. 查看工具版本

```bash
migrate version
```

## 8. 环境变量使用

`data_source_name` 支持以下模板：

- `${KEY}`：必须存在
- `${KEY:default}`：不存在时使用默认值

示例：

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

确保运行环境已设置 `DB_PASSWORD` 后执行：

```bash
migrate status
```

## 9. 新用户 10 分钟演练（SQLite）

[新用户 10 分钟演练（SQLite）](site/pages/zh/quick-start.md)

## 10. 全局参数

指定配置文件：

```bash
migrate -c ./configs/dev.json status
```

指定工作目录：

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. 常见错误排查

| 提示 | 操作 |
| --- | --- |
| `config file ... no such file or directory` | 运行 `migrate new config`，或用 `-c` 指定已有配置。 |
| `can't find env value for XXX` | 设置环境变量 `XXX`，或在配置中使用 `${XXX:default}`。 |
| `to-version must be set, or use --all` | 选择 `migrate down <version>` 或 `migrate down --all`。 |
| `unsupported dialect: xxx` | 将 `dialect` 改为第 2 节列出的数据库名称。 |
| `hash mismatch` / `filename mismatch` | 从版本控制恢复原迁移文件，再新增版本实现变更。 |

## 12. 作为 Go 库嵌入使用

在 Go 应用中调用迁移：

### 12.1 安装

```bash
go get github.com/exc-works/migrate
```

引入所用数据库的驱动：

```go
import (
    _ "modernc.org/sqlite"           // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 最小示例

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/exc-works/migrate"
)

func main() {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    svc, err := migrate.NewService(context.Background(), migrate.Config{
        Dialect:         migrate.NewSQLiteDialect(),
        DB:              db,
        MigrationSource: migrate.DirectorySource{Directory: "./migrations"},
    })
    if err != nil {
        panic(err)
    }

    if err := svc.Create(); err != nil { // 幂等，历史表不存在时创建
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 关键 API

- `migrate.NewService(ctx, migrate.Config)` 构造迁移执行器
- `svc.Create()` 建立 `migration_schema` 历史表（幂等）
- `svc.Up()` 应用所有未执行的迁移
- `svc.Down(toVersion, all)` 回退到指定版本或全部回退
- `svc.Status()` 返回 `[]migrate.MigrationStatus`
- `svc.Baseline()` 把当前已存在的 pending 文件标记为 `baseline`

常用类型：

- 方言（推荐使用构造函数，返回 `Dialect` 接口）：`migrate.NewPostgresDialect()`、`NewMySQLDialect()`、`NewSQLiteDialect()`、`NewMSSQLDialect()`、`NewOracleDialect()`、`NewClickHouseDialect()`、`NewMariaDBDialect()`、`NewTiDBDialect()`、`NewRedshiftDialect()`，或 `migrate.DialectFromName("postgres")` 按名字动态解析
- 迁移源：`DirectorySource`（读文件系统）、`StringSource`（内存数组）、`FSSource`（任意 `fs.FS`，例如 `//go:embed` 或 `os.DirFS`）、`CombinedSource`（合并多个源）
- 日志：`migrate.NoopLogger{}`（默认）、`migrate.NewStdLogger("info", os.Stdout)`，或自己实现 `migrate.Logger` 接口

[Go API](../migrate.go) · [Go example](../example_test.go)

## SQL、JS 和 TS 混合迁移

支持 SQL 和 JS/TS/MJS 脚本迁移，以及本地模块和已安装包的导入。安装方式、脚本接口与限制见[SQL、JS 和 TS 混合迁移](script-migrations.md).
