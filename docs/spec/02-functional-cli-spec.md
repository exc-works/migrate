# 02 Functional & CLI Spec

## 命令兼容（最小集合）

- `create`：创建迁移记录表。
- `baseline`：将当前 migration 文件标记为 baseline。
- `status`：输出 migration 文件与 DB 状态对比。
- `up`：迁移到最新版本。
- `down [to-version] --all`：回滚到目标版本（不含）或回滚全部。
- `new version <desc> [-v version]`：创建迁移文件。
- `new config [filename]`：创建配置文件。

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
