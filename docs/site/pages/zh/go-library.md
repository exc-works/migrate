# 在 Go 中使用 migrate

导入根包即可在测试、应用启动或管理操作中执行迁移。
公开 API 遵循 SemVer；`internal/*` 不属于公开 API。

## 安装并注册驱动

```sh
go get github.com/exc-works/migrate
```

调用方自行创建 `*sql.DB` 时，需要注册对应驱动。下面的完整示例使用 SQLite，
从 `./migrations` 读取 SQL 文件：

```go
package main

import (
    "context"
    "database/sql"
    "log"

    "github.com/exc-works/migrate"
    _ "modernc.org/sqlite"
)

func main() {
    if err := run(context.Background()); err != nil {
        log.Fatal(err)
    }
}

func run(ctx context.Context) error {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        return err
    }
    defer db.Close()

    service, err := migrate.NewService(ctx, migrate.Config{
        Dialect: migrate.NewSQLiteDialect(),
        DB:      db,
        MigrationSource: migrate.DirectorySource{
            Directory: "./migrations",
        },
    })
    if err != nil {
        return err
    }
    if err := service.Create(); err != nil {
        return err
    }
    return service.Up()
}
```

数据库连接池及生命周期由调用方管理。Context 应采用符合部署流程的超时，
同一数据库上的迁移应串行执行。

## 选择来源与方言

| 来源 | 用途 |
| --- | --- |
| `DirectorySource` | 磁盘版本文件 |
| `StringSource` | 内存中的迁移文本，尤其适合测试 |
| `FSSource` | 任意 `fs.FS`，包括 `embed.FS` 与测试文件系统 |
| `CombinedSource` | 合并多个来源 |

使用 `NewPostgresDialect()`、`NewSQLiteDialect()` 等构造函数，
也可用 `DialectFromName("postgres")` 按名称查找。

通过 `//go:embed migrations/*.sql` 嵌入 SQL，再将
`migrate.FSSource{FS: files, Root: "migrations"}` 传给来源字段。
脚本 helper 与 npm 包的接入见[共享模块](./modules)。

## 服务方法

| 方法 | 行为 |
| --- | --- |
| `Create()` | 幂等创建历史表 |
| `Up()` | 执行待应用版本 |
| `Down(toVersion, all)` | 回退高于目标的版本，或回退全部 |
| `Status()` | 返回 `[]migrate.MigrationStatus` |
| `Baseline()` | 登记待执行版本，不执行内容 |

设置 `Config.DryRun` 和 `Config.DryRunOutput` 可预览用户迁移。
`Create()` 不受 dry-run 影响，仍会创建历史表。默认日志器为 `NoopLogger`，
需要时使用 `NewStdLogger("info", os.Stdout)` 或自行实现 `Logger`。

## 脚本与工具管理的连接

脚本执行默认包含在构建中，通过 `Config.ScriptOptions` 设置脚本限制。
[嵌入示例](./examples)演示将脚本与 helper 一起编入 Go 二进制。

需要 migrate 管理连接（包括可选 SSH 隧道）时，使用 `OpenConnection`。
先通过 `Connection.Context()` 对 `Connection.DB` 执行 Ping，再把 DB 和 context
传给 `NewService`，最后关闭 connection。该 helper 注册网络驱动，SQLite 仍需调用方注册。
库凭据输入通过 `ConnectionOptions.Prompt` 显式启用，主机密钥批准使用独立的 `HostKeyPrompt`。
空回调不提示、不信任未知密钥，两者都必须响应取消。参见 [SSH 隧道](./ssh-tunnels)。

更多可运行 API 示例见
[example_test.go](https://github.com/exc-works/migrate/blob/main/example_test.go)。
