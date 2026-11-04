# 命令行参考

通过 `migrate --help` 或 `migrate <command> --help` 确认已安装版本的参数，
通过 `migrate version` 查看二进制版本。

## 全局选项

| 选项 | 默认值 | 含义 |
| --- | --- | --- |
| `--config`、`-c` | `migration_config.json` | 配置文件路径 |
| `--working-dir`、`-w` | 当前目录 | 从该目录解析配置，并覆盖迁移工作目录 |
| `--non-interactive` | `false` | 缺少凭据或遇到未知 SSH 主机密钥时失败，不请求交互输入或批准 |

## new config

```sh
migrate new config [filename] [--force]
```

生成 JSON 配置，默认文件名为 `migration_config.json`。
`--force` 允许覆盖已有文件。参见[配置](./configuration)。

## new version

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

在迁移目录创建版本文件。`--version` / `-v` 覆盖自动生成的版本，
`--format` 默认为 `sql`。
参见 [SQL 迁移](./sql-migrations)或[脚本迁移](./scripting)。

## 多数据库

配置 `data_sources` 后，create/baseline/status/up/down 自动遍历全部数据源，
均支持 `--parallel N`（默认 8）。单库失败不取消其他目标，结果按数据源汇总；
任一失败返回非零退出码。共享 SSH 与 JSON 输出结构见[配置](./configuration)。

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
`--output` / `-o` 默认为 `text`。参见[数据库能力](./databases)。

## version

```sh
migrate version
```

输出 CLI 版本，无须读取配置或连接数据库。这里显示的是工具版本，不是数据库最新迁移版本。
