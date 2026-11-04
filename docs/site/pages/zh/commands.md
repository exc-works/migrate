# 命令参考

通过 `migrate --help` 和 `migrate <command> --help` 查看已安装版本的准确参数，
通过 `migrate version` 查看工具版本。

| 任务 | 命令 | 说明 |
| --- | --- | --- |
| 生成配置 | `migrate new config` | [初始化](./guide#_2-初始化) |
| 创建历史表 | `migrate create` | [初始化](./guide#_2-初始化) |
| 接管已有数据库 | `migrate baseline` | [基线](./guide#_2-初始化) |
| 创建版本文件 | `migrate new version <description>` | [版本文件](./guide#_3-创建版本文件) |
| 执行待应用版本 | `migrate up` | [升级](./guide#_4-升级-数据库迁移) |
| 回退版本 | `migrate down <version>` 或 `migrate down --all` | [回退](./guide#_5-回退-数据库迁移) |
| 查看迁移历史 | `migrate status` | [状态](./guide#_6-查看状态) |
| 查看构建与数据库支持 | `migrate capabilities --output json` | [能力查询](./databases) |

`down <version>` 回退高于目标的版本，目标版本保持已应用。
`baseline` 只登记历史，不执行迁移；使用前应确认现有数据库结构与登记版本一致。

执行 up/down 前先使用 `--dry-run`。[全局参数](./guide#_10-全局参数)介绍配置选择与日志设置；
[常见错误排查](./guide#_11-常见错误排查)涵盖配置缺失、环境变量和历史不一致。

在 Go 应用中嵌入工具，见[作为 Go 库使用](./guide#_12-作为-go-库嵌入使用)。
