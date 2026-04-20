# migrate 用户指南

本文档面向首次使用者，命令与参数基于当前实现（`cmd/migrate`）。

## 1. 安装

### 1.1 从模块安装

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

安装指定版本：

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

把 `vX.Y.Z` 替换成真实版本号（例如 `v0.2.3`）。

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

`create` 成功时可能无输出，执行以下命令确认：

```bash
migrate status
```

已有历史数据库且不希望重放历史 SQL 时，可用：

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

`up` 成功时可能无输出，请以 `status` 结果为准。

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

说明：`migrate down <to-version>` 与 `migrate down --all` 互斥，不能同时使用。

### 5.3 回退演练（不落库）

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down` 成功时可能无输出，请执行 `migrate status` 验证。

## 6. 查看状态

```bash
migrate status
```

机器可读输出（适合脚本与 AI Agent）：

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

## 7. 工具升级与回退

### 7.1 升级工具本身

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 回退工具版本

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

例如：

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

建议把当前使用版本写入团队文档或 CI，避免本地行为不一致。

如果仓库私有且无法直接 `go install github.com/...@...`，请在对应版本源码目录执行：

```bash
go install ./cmd/migrate
```

### 7.3 查看当前工具版本

```bash
migrate version
```

说明：发布产物会输出发布版本号；本地源码直接 `go install ./cmd/migrate` 构建通常输出 `dev`。

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

### 9.1 准备目录与配置

先确认命令可用：

```bash
migrate --help
```

创建演练目录（macOS/Linux）：

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Windows PowerShell 可使用：

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

把 `migration_config.json` 改为：

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

### 9.2 初始化与创建迁移文件

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

编辑 `migrations/V202604140001__init_users.sql`：

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

编辑 `migrations/V202604140002__add_email.sql`：

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 升级、查看状态、回退

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

预期：

- 执行 `up` 后两个版本都是 `applied`
- 执行 `down 202604140001` 后：`202604140001=applied`，`202604140002=pending`
- 执行 `down --all` 后两个版本都为 `pending`

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

### 11.1 配置文件找不到

报错：`config file ... no such file or directory`

处理：

- 确认当前目录存在 `migration_config.json`
- 或使用 `-c` 指定配置文件路径

### 11.2 环境变量缺失

报错：`can't find env value for XXX`

处理：

- `export XXX=...`
- 或改为 `${XXX:default}`

### 11.3 `down` 参数不完整

报错：`to-version must be set, or use --all`

处理：

- 使用 `migrate down <version>`
- 或使用 `migrate down --all`

### 11.4 方言不支持

报错：`unsupported dialect: xxx`

处理：改为支持值：

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 迁移文件不一致

报错：`hash mismatch` 或 `filename mismatch`

处理：

- 不要修改已应用迁移文件
- 需要变更时新增更高版本迁移文件

## 12. 作为 Go 库嵌入使用

除了 CLI，`github.com/exc-works/migrate` 也可以作为库直接在你的服务代码里触发迁移，便于写单元测试、集成到启动流程或嵌入管理后台。

### 12.1 安装

```bash
go get github.com/exc-works/migrate
```

同时按需引入数据库驱动（库本身不强绑定驱动）：

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

### 12.4 测试友好：用 StringSource + 内存 SQLite

```go
src := migrate.StringSource{Migrations: []migrate.SourceFile{{
    Filename: "V1__init.sql",
    Source:   "-- +migrate Up\nCREATE TABLE t(id INT);\n-- +migrate Down\nDROP TABLE t;\n",
}}}

db, _ := sql.Open("sqlite", ":memory:")
svc, _ := migrate.NewService(ctx, migrate.Config{
    Dialect:         migrate.NewSQLiteDialect(),
    DB:              db,
    MigrationSource: src,
})
```

这样整个用例不依赖磁盘，可直接在单元测试里运行。

### 12.5 用 //go:embed 将迁移打包进二进制

借助 Go 的 embed 能力，把迁移 SQL 直接打进程序二进制：

```go
import "embed"

//go:embed migrations/*.sql
var migrations embed.FS

// 然后在服务里这样接入：
// MigrationSource: migrate.FSSource{FS: migrations, Root: "migrations"},
```

`FSSource` 接收任意 `fs.FS`，所以 `os.DirFS` 和 `fstest.MapFS` 也能用同样写法——测试里可以换成合成的文件系统。

### 12.6 预览 SQL（DryRun）

```go
var buf bytes.Buffer
svc, _ := migrate.NewService(ctx, migrate.Config{
    Dialect:         migrate.NewPostgresDialect(),
    DB:              db,
    MigrationSource: src,
    DryRun:          true,
    DryRunOutput:    &buf,
})
_ = svc.Create() // Create() 不受 DryRun 影响，用于建立历史表
_ = svc.Up()     // 用户迁移 SQL 只写入 buf，不会实际建表
```

### 12.7 稳定性约定

- `github.com/exc-works/migrate` 根包是对外公开 API，按 SemVer 维护
- `internal/*` 不在稳定性承诺内，请勿直接 import
- 完整可运行示例见仓库根目录 `example_test.go`
