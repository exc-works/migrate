# 脚本模块导入示例

这些示例随完整版本运行，使用 SQLite，不需要启动数据库服务。请在仓库根目录执行下面的命令。

| 目录 | 展示内容 | 运行前需要 |
| --- | --- | --- |
| [local](local) | TS 和 MJS 迁移共用迁移目录外的 TS helper | Go |
| [npm](npm) | `package.json`、lockfile、已安装包及 `exports` 子路径 | Go、Node.js/npm（安装包时） |
| [embedded](embedded) | Go `embed.FS` 打包迁移和 helper，通过公开 API 执行 up/down | Go |

[English instructions](README.en.md)。完整契约见[脚本迁移指南](../../docs/script-migrations.md)。

## 1. 相对路径与共享 helper

先构建完整 CLI，并为示例建立独立的临时数据库：

```bash
demo_dir=$(mktemp -d)
go build -tags=full -o "$demo_dir/migrate" ./cmd/migrate
export MIGRATE_EXAMPLE_DB="$demo_dir/local.db"

"$demo_dir/migrate" --working-dir examples/script-imports/local create
"$demo_dir/migrate" --working-dir examples/script-imports/local up --dry-run
"$demo_dir/migrate" --working-dir examples/script-imports/local up
"$demo_dir/migrate" --working-dir examples/script-imports/local status
```

目录结构：

```text
local/
├── migration_config.json
├── migrations/
│   ├── V1__create_users.sql
│   ├── V2__normalize_names.ts
│   └── V3__add_user.mjs
└── shared/
    └── names.ts
```

V2 和 V3 都导入 `../shared/names.ts`。执行后，用户 1 的名字变为 `alice`，新增用户 2
的名字为 `bob`；迁移内的查询会验证转换结果。status 显示三个版本已应用。

可以回滚到版本 2，再回滚全部版本：

```bash
"$demo_dir/migrate" --working-dir examples/script-imports/local down 2
"$demo_dir/migrate" --working-dir examples/script-imports/local status
"$demo_dir/migrate" --working-dir examples/script-imports/local down --all
```

`down 2` 保留 V1/V2，只撤销 V3；`down --all` 还原并删除示例用户表。辅助文件不会作为
独立迁移执行。`MIGRATE_EXAMPLE_DB` 仅指定这个示例的 SQLite 文件，不改变其他项目配置。

## 2. package.json 与 npm 包

继续使用上面构建的 CLI。安装示例中附带的本地包：

```bash
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$demo_dir/npm.db"

"$demo_dir/migrate" --working-dir examples/script-imports/npm create
"$demo_dir/migrate" --working-dir examples/script-imports/npm up
"$demo_dir/migrate" --working-dir examples/script-imports/npm status
"$demo_dir/migrate" --working-dir examples/script-imports/npm down --all
```

这个项目的 `package.json` 用 `file:./packages/normalizer` 声明依赖，lockfile 固定安装布局；
安装后产生 `node_modules/@migrate-example/normalizer`。包自己的 `package.json` 通过
`exports` 暴露根入口和 `/names` 子路径，V2 使用：

```ts
import { normalizeName } from "@migrate-example/normalizer/names";
```

执行后用户名字为 `alice`，status 显示两个版本已应用。包是纯 JavaScript、没有外部依赖，
因此这份示例可以离线安装。换成 registry 包时，使用项目实际依赖和对应的 lockfile 安装方式。
Node.js/npm 只用于安装依赖；迁移运行时使用内嵌的 esbuild 和 Goja。

## 3. Go 嵌入

```bash
go run -tags=full ./examples/script-imports/embedded
```

预期输出：

```text
after up: alice
after down: all migrations pending
```

程序使用独立的 SQLite 内存数据库，并通过 `//go:embed migrations shared` 同时嵌入入口
和 helper。`FSSource.Root` 只选择 `migrations` 下的版本文件；helper 从同一个嵌入文件系统
读取，构建后的程序不再依赖磁盘上的 TS 文件。这个示例没有 npm 依赖。

内存模块与 `WithModules` 的另一个可执行示例见
[modules_full_test.go](../../modules_full_test.go)，可运行：

```bash
go test -tags=full . -run '^ExampleWithModules$' -v
```

## 运行全部示例

```bash
./scripts/test-script-examples.sh
```

这个检查在临时目录中构建 CLI、复制并安装 npm 示例，验证两个 CLI 生命周期和嵌入程序输出，
完成后清除临时数据库与可执行文件。CI 也运行同一命令。

手动执行前两节时，可在完成后删除临时目录及 npm 安装结果：

```bash
rm -r "$demo_dir"
rm -r examples/script-imports/npm/node_modules
```

checksum 只覆盖迁移入口原始源码。修改 helper、包或 lockfile 不会导致历史校验失败，也不会
重跑已应用迁移；维护历史可重放性时应保留依赖的历史行为。包必须兼容 Goja；Node 内置 API、
原生扩展、动态导入及异步迁移入口仍不支持。虚拟来源中的包解析边界见使用指南。
