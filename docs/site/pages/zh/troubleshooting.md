# 问题排查

先查看 `migrate version`、`migrate capabilities --output json` 和
`migrate status --output json`。Status 需要数据库连接，version 和 capabilities 不需要。

## 找不到命令或配置

PATH 中找不到 `migrate` 时，按[安装](./installation)检查二进制位置。
找不到配置文件时，在项目目录执行 `migrate new config`，或使用 `-c` / `-w` 指定已有配置的位置。
路径解析见[配置](./configuration)。

默认配置文件是 `migration_config.yml`。配置报错时，根据提示的位置检查缩进和字段值；
YAML 缩进使用空格，不使用 Tab。配置文件的[查找顺序](./configuration#配置文件的查找顺序)决定实际使用哪一份配置。

## 缺少环境变量

`can't find env value for XXX` 表示必需的 `${XXX}` 模板无法解析。
在命令环境中提供变量，或对合适的非敏感字段使用 `${XXX:default}`。
SSH 的 `password_env` 等字段应直接填写变量名，参见 [SSH 认证](./ssh-tunnels)。

## 不支持的格式、方言或模块

- 通过 `migrate capabilities` 确认 SQL、JS、TS 和 MJS 支持。旧版 SQL 二进制
  应按[安装说明](./installation)升级。
- `unsupported dialect` 应检查[规范名称与别名](./databases#aliases)。
- 模块无法解析时，检查相对导入文件的路径，并提前安装 npm 依赖。
  在 Go 中嵌入脚本时，npm 包需通过 `DependencyDir` 指定安装目录。
- 依赖 Node API 或异步执行的包，即使成功打包也不能直接运行。参见[共享模块](./modules)。

## 哈希或文件名不一致

从版本控制恢复已应用文件的原始内容和文件名，后续变更新增更高版本。
不要通过对未知数据库状态建立基线来绕过错误。参见[历史与基线](./history)。

## up 后没有变化

成功命令可能没有输出。查看 status：版本可能已经应用或登记为基线，
也可能因低于已应用版本而被视为乱序并跳过。修改配置前先检查[版本顺序行为](./applying)。

## 回退被拒绝

`to-version must be set, or use --all` 表示必须指定目标或 `--all`，两者不能同时使用。
目标版本会保留。脚本缺少 down 函数时，对应选中范围不可回退。参见[回退迁移](./rollback)。

## 迁移、超时或连接失败

重试前检查实际数据和历史。非事务操作、DDL 隐式提交、COMMIT 前后断线，
均可能导致命令失败但变更已经生效。Baseline 不能修复部分执行。

脚本请检查[事务与资源限制](./script-transactions)。SSH 请检查可信主机记录、凭据、
代理以及堡垒机对原始数据库地址的访问。传输丢失不会自动重连或重放 SQL。
诊断信息经过脱敏；调查厂商错误时可检查受控的数据库或服务端日志，避免分享凭据或行数据。
