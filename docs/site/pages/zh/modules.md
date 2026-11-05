# 导入共享模块

脚本迁移支持静态导入和重新导出共享的 `.js`、`.ts` 与 `.mjs` 模块。
共享文件仅供迁移导入，不会作为独立迁移执行。

## 磁盘上的相对导入

路径从每个导入文件所在位置解析，helper 可以放在迁移目录之外：

```text
project/
  migrations/V2__normalize.ts
  shared/names.ts
```

```ts
import { normalize } from "../shared/names.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

在 `shared/names.ts` 中定义 `normalize`，并确保示例使用的 `users` 表已存在。
[本地导入示例](./examples#local)包含完整的建表步骤。

## 已安装的 npm 包

`@company/migration-utils` 这样的裸导入可以解析已安装的包。
执行 migrate 前，通过项目的 lockfile 流程安装依赖。支持 `exports`、子路径、嵌套依赖和静态 CommonJS 导入。
包同时提供多个入口时，使用 `neutral` 条件，并优先选择 `module`，其次为 `main`。

包需要能够同步运行，且不依赖 Node.js 专有功能。migrate 不会安装包，也不提供 Node 内置模块、
原生扩展、动态 import 或动态 require。

## 嵌入与内存来源

`FSSource` 在提供的 `fs.FS` 内解析相对 helper，包括 `embed.FS` 和 `fstest.MapFS`。
支持扩展名查找及目录 `index` 文件，不支持虚拟 package.json 入口或虚拟 node_modules
包解析。需要 npm 包时，显式指定磁盘依赖目录：

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

`embeddedFiles` 是提供的文件系统，`projectDirectory` 包含已安装的依赖。
相对 helper 仍从虚拟文件系统读取。没有 `DependencyDir` 时，虚拟来源不会搜索宿主工作目录。

`StringSource` 可包装为 `WithModules`，通过 `FS` 提供 helper、可选 `Root` 定位入口，
或通过 `Directory` 使用磁盘 helper 和包。自定义来源可以实现 `MigrationEntriesSource`
提供入口上下文。参见[在 Go 中使用](./go-library)。

## 保持旧迁移可重复执行 {#checksums}

migrate 只检查迁移文件本身的内容是否改变。修改 helper、已安装的包或 lockfile，不会触发哈希不一致，
也不会重放已应用迁移。需要重现旧迁移时，应保留历史 helper 行为并固定依赖；
改变行为时引入新的 helper 版本。

写入前会编译整个选中范围的依赖。较后迁移中的依赖错误，也会阻止较早的选中步骤开始执行。
