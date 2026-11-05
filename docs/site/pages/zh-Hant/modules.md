# 導入共享模塊

指令碼遷移支持靜態導入和重新導出共享的 `.js`、`.ts` 與 `.mjs` 模塊。
共享檔案僅供遷移導入，不會作為獨立遷移執行。

## 磁盤上的相對導入

路徑從每個導入檔案所在位置解析，helper 可以放在遷移目錄之外：

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

在 `shared/names.ts` 中定義 `normalize`，並確保示例使用的 `users` 表已存在。
[本地導入示例](./examples#local)包含完整的建表步驟。

## 已安裝的 npm 包

`@company/migration-utils` 這樣的裸導入可以解析已安裝的包。
執行 migrate 前，通過項目的 lockfile 流程安裝依賴。支持 `exports`、子路徑、嵌套依賴和靜態 CommonJS 導入。
包同時提供多個入口時，使用 `neutral` 條件，並優先選擇 `module`，其次為 `main`。

包需要能夠同步運行，且不依賴 Node.js 專有功能。migrate 不會安裝包，也不提供 Node 內置模塊、
原生擴展、動態 import 或動態 require。

## 嵌入與內存來源

`FSSource` 在提供的 `fs.FS` 內解析相對 helper，包括 `embed.FS` 和 `fstest.MapFS`。
支持擴展名查找及目錄 `index` 檔案，不支持虛擬 package.json 入口或虛擬 node_modules
包解析。需要 npm 包時，顯式指定磁盤依賴目錄：

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

`embeddedFiles` 是提供的檔案系統，`projectDirectory` 包含已安裝的依賴。
相對 helper 仍從虛擬檔案系統讀取。沒有 `DependencyDir` 時，虛擬來源不會搜索宿主工作目錄。

`StringSource` 可包裝為 `WithModules`，通過 `FS` 提供 helper、可選 `Root` 定位入口，
或通過 `Directory` 使用磁盤 helper 和包。自定義來源可以實現 `MigrationEntriesSource`
提供入口上下文。參見[在 Go 中使用](./go-library)。

## 保持舊遷移可重復執行 {#checksums}

migrate 只檢查遷移檔案本身的內容是否改變。修改 helper、已安裝的包或 lockfile，不會觸發哈希不一致，
也不會重放已應用遷移。需要重現舊遷移時，應保留歷史 helper 行為並固定依賴；
改變行為時引入新的 helper 版本。

寫入前會編譯整個選中範圍的依賴。較後遷移中的依賴錯誤，也會阻止較早的選中步驟開始執行。
