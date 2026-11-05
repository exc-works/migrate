# 데이터베이스 API

동기 `db`는 `dialect`, `query(sql, params?)`, `exec(sql, params?)`를 제공합니다. [타입 선언](../assets/migration-api.d.ts)을 받으세요.

```js
export function up(db) {
  const rows = db.query("SELECT id, name FROM users WHERE name=?", ["Alice"]);
  for (const row of rows) {
    const result = db.exec("UPDATE users SET name=? WHERE id=?", ["alice", row.id]);
    // result.rowsAffected는 "1" 같은 십진 문자열입니다.
  }
}
```

query는 첫 결과 집합 전체를 메모리로 읽습니다. 중복 열에는 별칭을 사용하세요. exec는 `{ rowsAffected: string }`으로 드라이버 값을 유지하며 `"0"`도 추측하여 바꾸지 않습니다.

매개변수는 배열입니다. PostgreSQL/Redshift `$1`, MySQL/MariaDB/SQLite/TiDB/ClickHouse `?`, Oracle `:1`, MSSQL `@p1`입니다. `db.dialect`는 정규 이름입니다.

정수와 정확한 소수는 문자열, NULL은 `null`, 바이너리는 복사한 바이트 배열, 시간은 정확한 UTC 문자열입니다. JavaScript number와 `JSON.parse`의 정밀도 손실을 피하세요. SQLite 소수가 float로 저장되면 거부됩니다. [특수 타입](./databases#value-conversion)을 참고하세요.

SQL을 분할하거나 `GO`를 처리하지 않습니다. MERGE의 세미콜론도 유지하세요. MySQL 배치는 DSN 옵션이 필요할 수 있습니다. 정리된 드라이버 오류는 catch할 수 있지만 트랜잭션을 복구하지는 않습니다. savepoint/rollback을 사용하고 복구 불가 시 다시 throw하세요. 정상 반환은 결과를 수락합니다. 취소, 전체 시간 초과, 잘못된 값, 자원 한도는 catch해도 실패이며 이후 DB 호출을 막습니다. 드라이버 자체 timeout은 catch할 수 있습니다. [트랜잭션](./script-transactions)을 보세요.
