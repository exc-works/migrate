# トラブルシューティング

`migrate version`、`migrate capabilities --output json`、`migrate status --output json` で調べます。status だけは DB 接続が必要です。

## コマンドや設定が見つからない

[インストール](./installation)の PATH を確認し、プロジェクトで `migrate new config` を実行するか、`-c` / `-w` を指定します。YAML のエラーは行と列を確認し、Tab ではなく空白で字下げします。[検索順序](./configuration#configuration-file-search-order)も確認してください。

## 環境変数がない

`can't find env value for XXX` は `${XXX}` が未設定です。値を渡すか、非機密値には `${XXX:default}` を使います。SSH の `password_env` には変数名を指定します。

## 非対応の形式や import

`capabilities` で対応形式を調べ、古いバージョンなら更新します。`unsupported dialect` は[名前](./databases#aliases)を確認。相対パスを確認し、npm を事前インストールします。埋め込みでは `DependencyDir` を明示し、Node API や非同期処理を使うパッケージは避けてください。

## 履歴の不一致、または up で変化がない

適用済みファイルの内容と名前をバージョン管理から復元し、変更は新しいバージョンにします。実状態不明の DB に baseline を使って回避しないでください。up が無出力でも成功している場合があります。applied、baseline、outOfOrder の状態を確認します。

## ロールバックや接続の失敗

`to-version must be set, or use --all` なら対象か `--all` のどちらかを指定します。対象バージョンは残ります。`down` がないスクリプトは戻せません。

失敗後も DDL やコミット済みのデータが残ることがあります。実データと履歴を調べてから再試行します。baseline は部分失敗を修復しません。SSH は認証、信頼済みホスト、プロキシ、踏み台から DB への到達性を確認。切断後の自動再接続や再実行はしません。機密情報を共有せず、管理された DB ログで詳細を調べます。
