# डेटाबेस API

सिंक्रोनस `db` में `dialect`, `query(sql, params?)`, `exec(sql, params?)` हैं। [प्रकार परिभाषा](../assets/migration-api.d.ts)।

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected दशमलव स्ट्रिंग है, जैसे "1"।
  }
}
```

query पहला परिणाम-समूह पूरा मेमोरी में लाता है; समान नाम के कॉलम को alias दें। exec `{ rowsAffected: string }` में ड्राइवर का मान रखता है, `"0"` भी।

पैरामीटर array हैं: PostgreSQL/Redshift `$1`; MySQL/MariaDB/SQLite/TiDB/ClickHouse `?`; Oracle `:1`; MSSQL `@p1`। `db.dialect` मानक नाम देता है।

सटीक पूर्णांक/दशमलव स्ट्रिंग, NULL `null`, बाइनरी कॉपी किए byte array, समय सटीक UTC स्ट्रिंग हैं। JavaScript number और JSON.parse सटीकता घटा सकते हैं। SQLite के float दशमलव अस्वीकार हैं। [विशेष प्रकार](./databases#value-conversion) देखें।

SQL बिना विभाजन या `GO` व्याख्या भेजता है; MERGE का सेमीकोलन बनाए रखें। MySQL batch को DSN विकल्प चाहिए हो सकता है। साफ़ किए ड्राइवर errors पकड़े जा सकते हैं, लेकिन catch ट्रांज़ैक्शन नहीं सुधारता। savepoint/rollback करें या दोबारा फेंकें। सामान्य वापसी परिणाम स्वीकारती है। रद्दीकरण, कुल समय सीमा, गलत मान और संसाधन सीमा catch के बाद भी घातक हैं और अगली DB कॉल रोकते हैं। केवल ड्राइवर timeout पकड़ा जा सकता है। [ट्रांज़ैक्शन](./script-transactions) देखें।
