# トランザクションと制限

対応するデータベースでは、スクリプトと成功履歴は同じトランザクションに入り、成功時にコミット、失敗時にロールバックします。このモードで `BEGIN`、`COMMIT`、`ROLLBACK` を直接実行しないでください。DDL の暗黙のコミットや非対応エンジンでは変更が残ることがあります。

## 自動トランザクションを無効にする

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: この文はトランザクション内で実行できません。
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

`transaction` は真偽値で、既定は `true`、設定は up/down の両方に適用されます。status、baseline、dry-run はスクリプトを実行せず、実行時の設定は検証しません。

同じ移行の操作は同じ接続を使用します。自分で開始したトランザクションは関数を返す前に終了し、変更したセッション設定も戻します。次の移行へのセッション引き継ぎには依存しないでください。失敗すると成功履歴なしで変更が残る場合があるため、再試行前に状態を調べます。

## 実行時間と結果サイズ

```yaml
script_timeout: 5m # 各スクリプトの制限時間
script_max_rows: 1000 # 1 回のクエリの最大行数
script_max_bytes: 16777216 # 1 回のクエリの最大バイト数（16 MiB）
```

上限超過は切り捨てではなくエラーです。大量読み込みはページ分割しますが、自動的に分割コミットや再開可能な移行にはなりません。キャンセル後に DB 処理が終了するまで時間がかかる場合があります。Go では `Config.ScriptOptions` を使います。[データベース別の動作](./databases#script-support)も確認してください。
