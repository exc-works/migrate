# sql-migrate 用户指南

本文档面向首次使用者，命令与参数基于当前实现（`cmd/sql-migrate`）。

## 1. 安装

### 1.1 从模块安装（仓库可公开访问时）

```bash
go install github.com/exc-works/sql-migrate/cmd/sql-migrate@latest
```

安装指定版本：

```bash
go install github.com/exc-works/sql-migrate/cmd/sql-migrate@vX.Y.Z
```

把 `vX.Y.Z` 替换成真实版本号（例如 `v0.2.3`）。

### 1.2 从本地源码安装（仓库私有或内网场景）

在仓库根目录执行：

```bash
go install ./cmd/sql-migrate
```

### 1.3 验证安装

```bash
sql-migrate --help
```

如果提示命令不存在：

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

如果提示 `Repository not found`，优先改用“从本地源码安装”方式。

## 2. 初始化

### 2.1 生成配置文件

```bash
sql-migrate new config
```

可选：

```bash
sql-migrate new config dev.json
sql-migrate new config --force
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
sql-migrate create
```

`create` 成功时可能无输出，执行以下命令确认：

```bash
sql-migrate status
```

已有历史数据库且不希望重放历史 SQL 时，可用：

```bash
sql-migrate baseline
```

## 3. 创建版本文件

### 3.1 自动版本号

```bash
sql-migrate new version init_users
```

### 3.2 指定版本号

```bash
sql-migrate new version add_email -v 202604140002
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
sql-migrate up --dry-run
```

正式执行：

```bash
sql-migrate up
```

执行后查看状态：

```bash
sql-migrate status
```

`up` 成功时可能无输出，请以 `status` 结果为准。

## 5. 回退（数据库迁移）

### 5.1 回退到目标版本（不含目标版本）

```bash
sql-migrate down 202604140001
```

语义：仅回退版本号大于 `202604140001` 的已应用迁移。

### 5.2 回退全部

```bash
sql-migrate down --all
```

### 5.3 回退演练（不落库）

```bash
sql-migrate down 202604140001 --dry-run
sql-migrate down --all --dry-run
```

`down` 成功时可能无输出，请执行 `sql-migrate status` 验证。

## 6. 查看状态

```bash
sql-migrate status
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
go install github.com/exc-works/sql-migrate/cmd/sql-migrate@latest
```

### 7.2 回退工具版本

```bash
go install github.com/exc-works/sql-migrate/cmd/sql-migrate@vX.Y.Z
```

例如：

```bash
go install github.com/exc-works/sql-migrate/cmd/sql-migrate@v0.2.3
```

建议把当前使用版本写入团队文档或 CI，避免本地行为不一致。

如果仓库私有且无法直接 `go install github.com/...@...`，请在对应版本源码目录执行：

```bash
go install ./cmd/sql-migrate
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
sql-migrate status
```

## 9. 新用户 10 分钟演练（SQLite）

### 9.1 准备目录与配置

先确认命令可用：

```bash
sql-migrate --help
```

创建演练目录（macOS/Linux）：

```bash
mkdir -p ./sql-migrate-demo
cd ./sql-migrate-demo
sql-migrate new config
```

Windows PowerShell 可使用：

```powershell
mkdir .\sql-migrate-demo
cd .\sql-migrate-demo
sql-migrate new config
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
sql-migrate create
sql-migrate new version init_users -v 202604140001
sql-migrate new version add_email -v 202604140002
```

编辑 `migrations/V202604140001__init_users.sql`：

```sql
-- +migrate Up
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

编辑 `migrations/V202604140002__add_email.sql`：

```sql
-- +migrate Up
ALTER TABLE users ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name FROM users;
DROP TABLE users;
ALTER TABLE users_tmp RENAME TO users;
```

### 9.3 升级、查看状态、回退

```bash
sql-migrate up --dry-run
sql-migrate up
sql-migrate status
sql-migrate down 202604140001 --dry-run
sql-migrate down 202604140001
sql-migrate status
sql-migrate down --all
sql-migrate status
```

预期：

- 执行 `up` 后两个版本都是 `applied`
- 执行 `down 202604140001` 后：`202604140001=applied`，`202604140002=pending`
- 执行 `down --all` 后两个版本都为 `pending`

## 10. 全局参数

指定配置文件：

```bash
sql-migrate -c ./configs/dev.json status
```

指定工作目录：

```bash
sql-migrate -w ./deploy create
sql-migrate -w ./deploy up
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

- 使用 `sql-migrate down <version>`
- 或使用 `sql-migrate down --all`

### 11.4 方言不支持

报错：`unsupported dialect: xxx`

处理：改为支持值：

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 迁移文件不一致

报错：`hash mismatch` 或 `filename mismatch`

处理：

- 不要修改已应用迁移文件
- 需要变更时新增更高版本迁移文件
