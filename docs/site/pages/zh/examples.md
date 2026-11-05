# 运行脚本示例

以下示例在仓库检出目录中运行，均使用 SQLite，无须数据库服务。
它们演示模块解析，并在迁移内部校验数据变更。

## 构建临时 CLI

在仓库根目录执行；本地导入和 npm 示例继续使用同一个 shell：

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## 共享本地 helper {#local}

[浏览本地示例](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local)。
其中两条迁移导入迁移目录之外的同一个 TypeScript helper。

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

up 后三个版本均已应用，用户名为 `alice` 和 `bob`。
`down 2` 保留版本 1、2；`down --all` 删除示例表。

## 解析已安装的包

[浏览 npm 示例](https://github.com/exc-works/migrate/tree/main/examples/script-imports/npm)。
Node/npm 仅用于安装。Lockfile 引用仓库内的本地包，没有 registry 依赖，
因此下面的安装可以离线完成：

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

第二条迁移通过 exports 子路径导入 `@migrate-example/normalizer/names`，
并校验归一化后的名称 `alice`。up 后两个版本均已应用。

## 将 helper 嵌入 Go 二进制

[嵌入式应用](https://github.com/exc-works/migrate/tree/main/examples/script-imports/embedded)
将版本文件和共享模块一起放入 `embed.FS`。它使用私有内存数据库，无须安装 npm 依赖：

```sh
go run ./examples/script-imports/embedded
```

预期输出：

```text
after up: alice
after down: all migrations pending
```

## 清理与验证

手动完成 CLI 示例后，删除临时数据库与可执行文件：

```sh
rm -r "$example_workspace"
```

npm 安装还会留下已忽略的 `examples/script-imports/npm/node_modules` 目录。
需要自动检查并自行管理临时副本时，在仓库根目录执行 `./scripts/test-script-examples.sh`。

共享代码的导入方式，以及修改依赖对旧迁移的影响，见[共享模块](./modules)。
