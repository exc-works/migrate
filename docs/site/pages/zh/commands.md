# 命令行参考

通过 `migrate --help` 或 `migrate <command> --help` 确认已安装版本的参数，
通过 `migrate version` 查看二进制版本。

## 全局选项

| 选项 | 默认值 | 含义 |
| --- | --- | --- |
| `--config`、`-c` | 自动查找 | 指定配置文件；省略时见[查找顺序](./configuration#配置文件的查找顺序) |
| `--working-dir`、`-w` | 当前目录 | 从该目录解析配置，并覆盖迁移工作目录 |
| `--non-interactive` | `false` | 缺少凭据或遇到未知 SSH 主机密钥时失败，不请求交互输入或批准 |

## new config

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

生成 `migration_config.yml`。用 `--multi` 生成多数据库配置，`--shards N` 指定数据库数量；
`--ssh-tunnel` 添加 SSH 设置，`--proxy socks5|http` 添加 SSH 上游代理。
`--ssh-auth agent|private_key|password` 选择 SSH 认证方式。
`--shards N` 需要同时使用 `--multi`，数量必须是正整数，编号从 0 开始。
`--proxy` 需要同时使用 `--ssh-tunnel`。
`--force` 允许覆盖已有文件。可通过 `.json` 文件名或 `--format json` 选择 JSON 输出。
全部选项见 `migrate new config --help`，常见用法见[配置](./configuration)。

## new version

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

在迁移目录创建版本文件。`--version` / `-v` 覆盖自动生成的版本，
`--format` 默认为 `sql`。描述包含空格时使用引号，例如 `migrate new version "init users"`；
文件名中的空格会替换为下划线。
参见 [SQL 迁移](./sql-migrations)或[脚本迁移](./scripting)。

## 多数据库

配置 `data_sources` 后，create/baseline/status/up/down 自动遍历全部数据源，
均支持 `--parallel N`（默认 8）。单库失败不取消其他目标，结果按数据源汇总；
任一失败返回非零退出码。共享目录与 SSH 设置见[多数据库配置](./configuration#multiple-databases)。
`status --output json` 按数据源名称排序返回结果，每项包含 `data_source`、`status`，
成功时附带 `migrations`，失败时附带 `error`。

## create

```sh
migrate create
```

幂等创建历史表，不执行用户迁移文件。

## baseline

```sh
migrate baseline
```

将待执行迁移登记为基线，不执行内容。先确认现有数据库已与这些文件一致。
参见[历史与基线](./history)。

## up

```sh
migrate up [--dry-run]
```

按顺序应用待执行版本。Dry-run 输出 SQL 或脚本计划，不执行用户迁移，
但可能连接数据库读取历史。参见[执行迁移](./applying)。

## down

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

回退高于目标的版本并保留目标，或回退全部已应用版本。
目标和 `--all` 互斥，必须选择其一。参见[回退迁移](./rollback)。

## status

```sh
migrate status [--output table|json]
```

对比文件和数据库历史，`--output` / `-o` 默认为 `table`。
状态含义见[历史与基线](./history)。

## capabilities

```sh
migrate capabilities [--output text|json]
```

列出构建格式、别名、脚本验证状态及事务限制，无须读取配置或连接数据库。
`--output` / `-o` 默认为 `text`。参见[支持的数据库](./databases)。

## version

```sh
migrate version
```

输出 CLI 版本，无须读取配置或连接数据库。这里显示的是工具版本，不是数据库最新迁移版本。
