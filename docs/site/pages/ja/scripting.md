# スクリプトマイグレーション

JavaScript、TypeScript、MJS でデータを変換できます。実行時に Node.js は不要です。

```sh
migrate new version "normalize names" --format ts
```

SQL と同じ `V<version>__<description>.<format>` の順序を使います。`.mjs` と `.js` の制限は同じです。

## エントリ関数

SQLite の `users` に `name` 列がある例です。

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

生成された `up` は実装するまで例外を投げます。`async`、generator、Promise の返却には対応しません。通常の同期的な戻り値は無視されます。[SQL のプレースホルダー](./script-api)はデータベースごとに異なります。

前後の空白除去は元の値を失うため、この例には `down` がありません。逆操作を実装したときだけ `down(db)` を公開してください。戻すには元の値を保存する必要があります。

## プレビュー

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run はスクリプトを実行しません。DB 操作はモジュール直下ではなく `up` または `down` に書きます。移行間のグローバル変数共有には依存しないでください。

テンプレートの型宣言や [migration-api.d.ts](../assets/migration-api.d.ts)を利用できます。変換時に型検査は行いません。Node の組み込み API、ネットワーク、ファイル、環境変数、タイマー、動的モジュール読み込みは使用できません。npm パッケージも同じ条件を満たす必要があります。[共有モジュール](./modules)と[トランザクション](./script-transactions)へ進んでください。
