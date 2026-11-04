# 02 Functional & CLI Spec

## 命令兼容（最小集合）

- `create`：创建迁移记录表。
- `baseline`：将当前 migration 文件标记为 baseline。
- `status`：输出 migration 文件与 DB 状态对比。
- `version`：输出 CLI 工具版本，不读取配置、不连接数据库。
- `up`：迁移到最新版本。
- `down [to-version] --all`：回滚到目标版本（不含）或回滚全部。
- `new version <desc> [-v version]`：创建迁移文件。
- `new config [filename]`：创建配置文件。

## 版本来源

- `version` 和 `capabilities` 使用相同的版本来源。
- 优先使用发布构建显式注入的版本号。
- 未注入版本号时，读取 Go 构建信息中的主模块版本，使 `go install ...@版本` 显示实际安装版本。
- 本地开发构建或缺少模块版本信息时显示 `dev`。

## 迁移文件语法

需兼容：

- `-- +migrate Up`
- `-- +migrate Down`
- `-- +migrate StatementBegin`
- `-- +migrate StatementEnd`

## 行为约束

- 解析错误必须返回 `error`，禁止 `panic`/`os.Exit`。
- `MigrateOutOfOrder=false` 时，out-of-order 文件必须被跳过。
- `hash mismatch` 与 `filename mismatch` 必须阻断执行。
- `up` 必须幂等。

## 配置约束

- 默认 `schema_name`：`migration_schema`。
- 默认 `migration_source`：`migrations`。
- 默认 `dialect`：`postgres`。
- 环境变量模板：`${KEY}` 与 `${KEY:default}`。
