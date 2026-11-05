# スクリプトのデータベース API

同期関数の `db` は `dialect`、`query(sql, params?)`、`exec(sql, params?)` を提供します。[型宣言](../assets/migration-api.d.ts)を利用できます。

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected は "1" のような 10 進文字列です。
  }
}
```

`query` は最初の結果セット全体をメモリに読み込みます。同名列には別名を付けます。`exec` の `{ rowsAffected: string }` はドライバーの件数を保持し、`"0"` でも推測で変更しません。

## パラメーターと型

パラメーターは配列で渡します。PostgreSQL/Redshift は `$1`、MySQL/MariaDB/SQLite/TiDB/ClickHouse は `?`、Oracle は `:1`、MSSQL は `@p1` です。`db.dialect` は別名ではなく正規名を返します。

整数と精密な小数は文字列、NULL は `null`、バイナリはコピーしたバイト配列、時刻は正確な UTC 文字列です。JavaScript number や `JSON.parse` で大きい数の精度を失わないようにしてください。SQLite NUMERIC が浮動小数点として保存した値は拒否されます。[データベース固有の変換](./databases#value-conversion)も確認してください。

## SQL とエラー

SQL はそのまま転送され、バッチ分割や `GO` の処理はしません。MERGE のセミコロンも保持します。MySQL のバッチには DSN オプションが必要な場合があります。

機密情報を除いたデータベース例外は catch できますが、catch だけでトランザクションは修復されません。セーブポイントやロールバックを使用し、回復できない場合は再スローします。キャンセル、全体のタイムアウト、不正な値、結果の上限超過は catch しても移行が失敗し、その後の DB 操作を禁止します。ドライバー単独のタイムアウトは catch できます。[トランザクション](./script-transactions)を参照してください。
