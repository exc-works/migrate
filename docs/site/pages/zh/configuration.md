# 配置

在项目目录中生成配置文件：

```sh
migrate new config
```

打开生成的 `migration_config.yml`，填写数据库类型和连接信息，再执行迁移命令。

## 连接数据库

下面的配置连接 PostgreSQL，并从环境变量 `DATABASE_DSN` 获取连接字符串：

```yaml
# 数据库类型
dialect: postgres
# 执行命令前，在 shell 或部署环境中设置 DATABASE_DSN
data_source_name: "${DATABASE_DSN}"
# 迁移文件目录
migration_source: migrations
# 用于记录已执行版本的历史表
schema_name: migration_schema
```

| 字段 | 用途 | 省略时的默认值 |
| --- | --- | --- |
| `dialect` | 数据库类型，见[支持的名称](./databases#aliases) | `postgres` |
| `data_source_name` | 数据库连接字符串；与 `data_sources` 选择其一 | 单库时必填 |
| `data_sources` | 多个数据库的名称和连接设置，见[多数据库配置](#multiple-databases) | 不启用 |
| `migration_source` | 迁移文件目录 | `migrations` |
| `schema_name` | 迁移历史表名称 | `migration_schema` |
| `working_directory` | 查找迁移文件的起始目录 | 当前目录 |
| `migrate_out_of_order` | 是否执行低于已应用版本的待执行版本 | `false` |
| `logger_level` | 日志级别：`debug`、`info`、`warn`、`error`，见[日志级别](#logging) | `info` |

### 各数据库的连接示例

选择对应数据库，将下面的 `dialect` 和 `data_source_name` 填入配置。
把示例中的主机、端口、数据库名或 Oracle 服务名替换为实际值，
并在运行命令前设置 `DB_USER` 和 `DB_PASSWORD` 环境变量。SQLite 直接填写数据库文件路径。

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

URL 形式的连接字符串中，用户名或密码包含 `@`、`:`、`/` 等特殊字符时，需要先做 URL 编码；
环境变量替换不会自动编码。MySQL、MariaDB 和 TiDB 使用的 `user:password@tcp(...)` 格式不按 URL 编码。
TLS/SSL 参数请按数据库服务的连接要求设置。

### 日志级别 {#logging}

`logger_level` 控制显示哪些级别的日志，默认为 `info`：

| 值 | 显示内容 |
| --- | --- |
| `debug` | 调试信息、普通信息、警告和错误 |
| `info` | 普通信息、警告和错误 |
| `warn` | 警告和错误，也可写为 `warning` |
| `error` | 仅错误 |

```yaml
logger_level: debug # 排查问题时显示调试信息
```

## 使用环境变量 {#environment-variables}

连接字符串支持 `${KEY}` 和 `${KEY:default}`：

- `${DATABASE_DSN}`：读取 `DATABASE_DSN`；未设置时命令会报错。
- `${DB_NAME:app}`：读取 `DB_NAME`；未设置时使用 `app`。

同一份配置可以用于开发、测试和生产环境，只需在运行命令的环境中提供不同的变量值。
数据库凭据通过环境变量或部署平台的密钥管理功能提供。
SSH 地址、用户名和代理地址也支持这些写法，详见 [SSH 隧道与代理](./ssh-tunnels)。

## 配置文件的查找顺序

未使用 `-c` 指定文件时，migrate 在当前工作目录按以下顺序查找，使用找到的第一份配置：

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

多份配置不会合并。找到的文件若有错误或无法读取，命令会报错，需要修正该文件后再运行。

使用 `-c` 可以直接选择配置文件；指定文件不存在时会报错，不再查找其他文件：

```sh
migrate -c ./configs/dev.yml status
```

## 为不同场景生成配置

根据项目需要选择一条命令，生成后填写其中的连接信息：

| 场景 | 命令 |
| --- | --- |
| 单个数据库 | `migrate new config` |
| 使用自定义文件名 | `migrate new config dev.yml` |
| 多个数据库 | `migrate new config --multi` |
| 四个数据库 | `migrate new config --multi --shards 4` |
| 通过 SSH 连接，使用 SSH Agent | `migrate new config --ssh-tunnel` |
| 通过 SSH 连接，使用私钥 | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| 通过 SSH 连接，使用密码 | `migrate new config --ssh-tunnel --ssh-auth password` |
| 经 SOCKS5 代理连接 SSH 堡垒机 | `migrate new config --ssh-tunnel --proxy socks5` |
| 经 HTTP CONNECT 代理连接 SSH 堡垒机 | `migrate new config --ssh-tunnel --proxy http` |

这些选项可以组合使用，例如为四个数据库生成共享 SSH 与代理设置：

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

已有文件不会被覆盖；需要重新生成时加上 `--force`。
使用 `migrate new config --help` 查看全部选项。

## 将同一套迁移应用到多个数据库 {#multiple-databases}

分库或多个租户数据库需要相同变更时，使用 `data_sources` 列出各个数据库，
替代顶层的 `data_source_name`。所有数据库使用相同的数据库类型、迁移目录和历史表名称。

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # 第一个数据库的连接字符串
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # 第二个数据库的连接字符串
```

运行前设置 `SHARD_0_DSN` 和 `SHARD_1_DSN`。如果通过 `--multi --shards N`
生成配置，数据库从 `shard_0` 开始编号，分别引用 `SHARD_0_DSN`、`SHARD_1_DSN` 等变量。

![同一目录中的迁移分别应用到 shard_0 和 shard_1，各库独立记录历史，最后汇总结果。](../assets/diagrams/multiple-databases-zh.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

这些命令会自动处理全部配置的数据库。每个数据库根据自己的历史选择待执行版本，
并按版本顺序执行；已经成功的版本不会重复执行。`new version` 只需运行一次，
生成的文件供所有数据库共用。

默认最多同时处理 8 个数据库。使用 `--parallel` 调整数量，设为 1 时逐库执行：

```sh
migrate up --parallel 4
migrate status --output json
```

某个数据库失败时，其他数据库会继续执行；结束后按数据库汇总结果，任一失败则命令返回非零退出码。
这不是跨库事务，某个数据库失败不会撤销其他数据库已经完成的变更。
按 Ctrl+C 会停止后续数据库的执行，并取消正在运行的操作。

如果多个数据库经过同一台堡垒机，可以在顶层配置 `ssh_tunnel`。
某个数据库需要不同的 SSH 设置时，在该项下填写完整的 `ssh_tunnel`，它会替换顶层设置。
不填写或设为 `null` 时使用顶层设置。具体字段见 [SSH 隧道与代理](./ssh-tunnels)。

每个实际数据库只配置一次；不要同时启动多个命令修改同一个数据库。
执行期间保持迁移文件与脚本依赖不变。失败后先检查该库的数据和历史，再决定是否重试。

## 工作目录与相对路径

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w` 指定工作目录。配置文件从该目录查找，配置中的 `working_directory` 也会被该选项覆盖。
相对的 `migration_source` 从工作目录查找；只传入 `-c` 不会把工作目录改成配置文件所在目录。

例如在项目根目录运行 `migrate -c configs/dev.yml up`，配置中写了
`migration_source: migrations`，使用的就是项目根目录下的 `migrations`。
SSH 私钥等文件的相对路径则从配置文件所在目录查找。

配置完成后，可以[编写第一条 SQL 迁移](./sql-migrations)，或设置[脚本执行限制](./script-transactions)。
