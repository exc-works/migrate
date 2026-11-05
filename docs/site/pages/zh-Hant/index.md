---
layout: home
hero:
  name: migrate
  text: 管理資料庫變更。
  tagline: 從單庫到多庫，讓資料庫變更有序可控。
  actions:
    - theme: brand
      text: 開始使用
      link: ./quick-start
    - theme: alt
      text: 瀏覽文檔
      link: ./guide
features:
  - title: 多種資料庫，一套工具
    details: 支持 PostgreSQL、MySQL、MariaDB、SQLite、Oracle、SQL Server、ClickHouse、TiDB 和 Redshift。
    link: ./databases
    linkText: 查看資料庫支持
  - title: SQL 與指令碼遷移
    details: 支持編寫 SQL、TS、JS 或 MJS 檔案執行資料庫遷移。
    link: ./scripting
    linkText: 編寫遷移指令碼
  - title: 一次執行，更新多個資料庫
    details: 多個資料庫共享一套遷移檔案，併發執行、獨立記錄歷史，按資料庫匯總結果。
    link: ./configuration#multiple-databases
    linkText: 設定多個資料庫
  - title: 環境變數設定
    details: 通過環境變數切換資料庫連接與部署環境，在 YAML 中維護項目設置。
    link: ./configuration#environment-variables
    linkText: 設定連接與環境變數
  - title: SSH 隧道與Proxy
    details: 通過 SSH 堡壘機訪問資料庫，支持密碼、私鑰、SSH Agent，以及 SOCKS5 和 HTTP CONNECT Proxy。
    link: ./ssh-tunnels
    linkText: 連接遠程資料庫
  - title: 預覽、校驗與回退
    details: 執行前預覽變更，檢查已應用檔案是否被修改，並用 Down 操作回退版本。
    link: ./applying
    linkText: 執行資料庫變更
---
