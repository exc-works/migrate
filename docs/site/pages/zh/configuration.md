# 配置

在迁移项目目录运行命令。migrate 默认依次查找 YAML 和 JSON 配置，可先生成带注释的 YAML 配置：

```sh
migrate new config
```

指定其他文件名可以使用 `migrate new config dev.json`。
已有文件只有在显式传入 `--force` 时才会被覆盖。

## 初始化 YAML、多数据源、SSH 与代理配置

`new config` 默认生成带字段注释的 `migration_config.yml`。使用 `--format json`
或 `.json` 文件名可生成严格 JSON，显式格式与文件名扩展名必须一致。初始化不会
读取环境变量值、密钥或现有配置，也不会连接网络。

```sh
migrate new config
migrate new config shards.yml --multi
migrate new config shards-four.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel
migrate new config shards-key.yml --multi --ssh-tunnel --ssh-auth private_key
migrate new config shards-proxy.yml --multi --ssh-tunnel --proxy socks5
migrate new config shards-http.yml --multi --ssh-tunnel --proxy http
migrate new config --format json
```

`--multi` 默认生成 `shard_0`、`shard_1`；`--shards N` 指定正整数数量，要求同时启用
`--multi`。编号从 0 开始，例如 `--shards 4` 生成 `shard_0` 到 `shard_3`，各项引用对应的
`SHARD_0_DSN`、`SHARD_1_DSN` 等环境变量。

SSH 地址、用户和代理地址默认使用环境变量占位符；密码与密钥口令使用环境变量
名称，没有明文密码参数。`--proxy-username` 与 `--proxy-password-env` 必须成对设置，
SSH/代理从属参数要求启用对应开关。执行 `migrate new config --help` 查看地址、端口、
认证方式、密钥路径和 known_hosts 参数。共享 SSH/代理只在顶层生成一份，各库填写 DSN。
代理地址也可使用 `${CORP_PROXY_ADDRESS}`、`proxy.example:${PROXY_PORT}` 等自定义引用。
初始化保留原始引用且不读取其值，运行迁移时再验证展开后的地址。

## YAML 与 JSON 查找

读取器支持 YAML 和 JSON。未显式传入 `-c` 时，依次查找
`migration_config.yml`、`migration_config.yaml`、`migration_config.json`。
只有文件不存在才回退；YAML 无效或不可读时直接报错。显式 `-c` 只读取指定文件。
YAML 使用原生 `#` 注释，JSON 保持严格语法。两种格式共享字段名、默认值、环境变量
和相对路径规则，嵌套 SSH/proxy 的字段名也一致。

读取器同时识别 `data_sources` 及共享 SSH 默认配置。批量命令在后续编排层接入；
当前中间层会在连接数据库之前明确拒绝多源数据库命令。

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
