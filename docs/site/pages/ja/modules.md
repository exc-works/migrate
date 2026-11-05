# 共有モジュールの読み込み

`.js`、`.ts`、`.mjs` の静的 import と再エクスポートを使用できます。共有ファイルは独立した移行として実行しません。

## 相対 import

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

パスは import 元のファイルから解決します。`shared/names.ts` に `normalize` を定義し、`users` を先に作成します。[完全な例](./examples#local)も参照してください。

## npm パッケージ

事前に lockfile に基づいてインストールします。`exports`、サブパス、入れ子の依存、静的 CommonJS に対応し、`neutral` 条件と `module`、次に `main` を使います。Node 固有の API、ネイティブ拡張、動的 import/require は使用できません。migrate はパッケージをインストールしません。

## 埋め込みとメモリ

`FSSource` は `embed.FS` などの `fs.FS` 内で相対モジュールを解決し、拡張子探索とディレクトリの `index` に対応します。仮想 `package.json` や `node_modules` の探索はしません。npm の依存には実ディスクのディレクトリを指定します。

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

相対ファイルは仮想ファイルシステムに残り、`DependencyDir` がなければ作業ディレクトリを探索しません。`StringSource` は `WithModules` の `FS` と任意の `Root`、または `Directory` を使用します。独自ソースは `MigrationEntriesSource` を実装できます。[Go](./go-library)も参照してください。

## 履歴の再現性 {#checksums}

ハッシュが対象にするのは移行ファイル本体のみです。共有ファイル、npm パッケージ、lockfile を変更しても不一致にはならず、適用済みの移行も再実行しません。依存を固定し、旧動作を維持してください。選択範囲全体の依存を実行前に検査するため、後続ファイルのエラーも先行ファイルの実行を止めます。
