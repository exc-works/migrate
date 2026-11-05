---
layout: home
hero:
  name: migrate
  text: "データベースの変更を管理。"
  tagline: "ひとつのデータベースから複数のデータベースまで、変更を順序立てて管理。"
  actions:
    - theme: brand
      text: "使ってみる"
      link: ./quick-start
    - theme: alt
      text: "ドキュメントを読む"
      link: ./guide
features:
  - title: "さまざまなデータベースに対応"
    details: "PostgreSQL、MySQL、MariaDB、SQLite、Oracle、SQL Server、ClickHouse、TiDB、Redshift に対応。"
    link: "./databases"
    linkText: "対応状況を見る"
  - title: "SQL とスクリプトで移行"
    details: "SQL、TS、JS、MJS ファイルでマイグレーションを作成できます。"
    link: "./scripting"
    linkText: "スクリプトを作成"
  - title: "複数のデータベースをまとめて更新"
    details: "同じファイルを並列実行し、履歴と結果をデータベースごとに管理します。"
    link: "./configuration#multiple-databases"
    linkText: "複数の接続を設定"
  - title: "環境変数による設定"
    details: "YAML にプロジェクト設定を保存し、環境変数で接続先や実行環境を切り替えます。"
    link: "./configuration#environment-variables"
    linkText: "接続を設定"
  - title: "SSH トンネルとプロキシ"
    details: "パスワード、秘密鍵、SSH Agent に対応。SOCKS5 や HTTP CONNECT 経由で踏み台に接続できます。"
    link: "./ssh-tunnels"
    linkText: "リモート接続"
  - title: "プレビュー、検証、ロールバック"
    details: "実行前に変更を確認し、適用済みファイルの変更を検出し、Down でバージョンを戻します。"
    link: "./applying"
    linkText: "変更を実行"
---
