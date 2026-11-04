# 多数据库编排 PR Stack

状态：计划阶段，尚未创建实现分支或远程 PR。技术方案以
[10-multi-database-migrations.md](10-multi-database-migrations.md) 为准。

## 依赖顺序

```text
当前基础分支
  -> PR0 docs/multi-database-orchestration
  -> PR1 feat/multi-database-config
  -> PR2 feat/config-initialization
  -> PR3 feat/batch-iteration
  -> PR4 feat/multi-database-cli
```

第一版增加“多源配置与初始化 + 有界并发遍历 + 现有命令接入”。不把迁移核心重构、
迁移锁、全局预检、共享编译或公共批量 API 放入此 stack。

## PR0 收敛技术方案

位置：spec 10、spec 11 和规范索引。

明确编排层只负责遍历和结果汇总，复用现有单库 Service。固定每库失败后继续
其余库、父取消终止任务的行为，记录同库互斥仍沿用现有操作约定。

验收：方案不要求改造 Up/Down、事务或脚本运行时，没有新增方言支持门槛，
示例和索引可用，计划状态与已实现状态清楚。

## PR1 多数据源配置与双格式读取

位置：internal/config、嵌套配置类型的 yaml tag、CLI 默认配置路径、go.mod 和配置测试。

- 在原 FileConfig 中增加可选 data_sources，每项为名称、DSN 和可选 SSH 设置。
- 复用既有默认值、环境展开和路径处理，将共享设置展开为独立目标配置。
- 仅顶层 DSN 与 data_sources 互斥；顶层 SSH 作为共享默认值，局部非 null SSH
  对象整体覆盖，未提供或 null 时继承，避免为每个数据源重复填写。
- 校验空集合、缺失 DSN 和 SSH 配置错误，复用既有相对路径规则。
- 复用仓库已有的 go.yaml.in/yaml/v3，补齐所有嵌套配置的 yaml tag；YAML/JSON
  解码到同一配置结构，后续处理共用，不增加 JSON 注释预处理。
- 未显式 -c 时按 yml/yaml/json 查找，仅文件不存在时回退；显式 -c 不回退。
  调整 CLI 固定 JSON 默认值，保留 --working-dir 与 SSH 相对路径规则。
- YAML 拒绝重复 key、多文档和无效结构，安全报告解析位置；旧 JSON 语法及
  显式其他扩展名的 JSON 读取保留，不增加配置版本和目标筛选。

验收：旧配置测试通过；顶层 SSH 与多源共存、继承/覆盖/直连测试通过；每项 SSH
可变状态隔离；两种格式的嵌套字段、特殊字符和环境展开等价；找到的 YAML 无效时
不回退，显式 -c 不回退，旧 JSON 自动回退正确；配置无效时尚未建立连接。
CLI 在批量执行接入前遇到多源数据库命令明确报未支持，不错误使用空顶层 DSN。

## PR2 配置初始化

位置：internal/config 模板渲染、cmd/migrate 的 new config 和对应测试。

- 增加 --multi、--ssh-tunnel、--ssh-auth、SSH 地址/认证引用相关参数，以及
  --proxy socks5|http、代理地址/认证引用参数，具体参数契约见 spec 10。
- 默认生成带原生注释的 migration_config.yml；--format json 生成严格 JSON。
  显式文件名推断格式，冲突报错；保留可选文件名和 --force 行为。
- 使用 YAML node 注释和标准 JSON 编码器生成语义相同的两种模板。
- 参数组合在写文件前校验，值经对应编码器处理，DSN/秘密使用环境变量引用；
  不读取环境变量值、密钥或现有配置，不联网或提示秘密。
- 生成顶层共享 SSH/proxy，多源各项无需重复配置。更新初始化帮助及使用示例。

验收：单/多源、agent/private_key/password、SOCKS5/HTTP、代理认证模板组合可读；
YAML 和 JSON 版解码结果一致；缺少环境变量仍能初始化；格式/扩展名、非法参数、
默认拒绝覆盖和 --force 均正确；特殊字符编码正确。多源执行仍在 PR4 完成接入。

## PR3 有界并发遍历器

位置：新增 internal/batch 及调度单测。

- 提供目标列表、parallel 和执行回调的简单 Run 入口，返回每目标结果。
- 固定数量 worker；每库执行失败继续其余目标，父 context 取消停止新任务。
- 汇总成功、失败、取消和 callback 错误，结果稳定排序。
- 使用 fake callback 验证调度，不依赖数据库、Service、方言或迁移计划。

验收：通道屏障证明并发且不超上限，parallel=1 正确串行；1000 个假目标无需
1000 个 worker；部分失败仍遍历完成；取消时每目标有终态，无 race 或任务泄漏。

## PR4 接入现有命令并交付

位置：cmd/migrate、CLI/集成测试、中英文用户指南和 CHANGELOG。

- 多源配置自动遍历全部数据库，新增 --parallel，单源继续走原有兼容路径。
- 将现有 Service 构造和 cleanup 封装为每目标回调，复用 Create/Baseline/Status/
  Up/Down 与 dry-run，不修改迁移执行核心。
- 每目标使用继承后的 SSH 配置复用现有 Connection 生命周期，不新增共享 transport。
- 注入每目标 logger/output，串行化终端交互，避免共享 Cobra 可变状态。
- 汇总多源结果和退出码，保留原单源 JSON；new 命令仍只生成一次文件。
- 同层加入临时 SQLite 多库 CLI 验证、现有网络数据库/SSH 代表性集成回归、
  并发 race 检查、配置示例和失败重跑说明。

验收：同一目录在多个不同进度数据库上并发 up/status/down；其中一库失败不影响
其他库；取消与初始化失败正确清理；dry-run 无写入，JSON 和日志不交错；
旧 CLI 和 Go API 回归通过。文档中的命令与本层实际提供的行为一致。

## 每层检查

每个 PR 以直接前序为 base，保持可构建，包含本层测试和必要文档。
遵守 AGENTS.md 及 spec 01、04、06；合并前完成：

```bash
go test ./...
go test -tags=integration ./...
go vet ./...
go build ./...
go build ./cmd/migrate
```

PR3/PR4 另外对编排器和 CLI 并发路径运行 go test -race。每层记录实际测试与
跳过项，不将集成测试跳过视为真实数据库验证。实际开始实施后再创建分支和 PR，
补充对应链接、head/base、CI 和审查结果。
