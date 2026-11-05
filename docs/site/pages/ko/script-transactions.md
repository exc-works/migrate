# 트랜잭션과 제한

지원 DB에서는 스크립트와 성공 이력을 같은 트랜잭션에 넣습니다. 성공은 커밋, 오류는 롤백입니다. BEGIN/COMMIT/ROLLBACK을 직접 실행하지 마세요. DDL 암묵적 커밋과 비트랜잭션 DB에서는 변경이 남을 수 있습니다.

```js
export const config = { transaction: false };

export function up(db) {
  // PostgreSQL: 이 명령은 트랜잭션 밖에서 실행해야 합니다.
  db.exec("CREATE INDEX CONCURRENTLY users_name_idx ON users (name)");
}

export function down(db) {
  db.exec("DROP INDEX CONCURRENTLY users_name_idx");
}
```

`transaction`은 boolean이며 기본 `true`, up/down 공통입니다. status/baseline/dry-run은 실행 시 내보내는 설정을 평가하지 않습니다. 한 마이그레이션은 전용 연결을 공유합니다. 직접 시작한 트랜잭션을 종료하고 세션 설정을 복원한 후 반환하세요. 마이그레이션 간 세션 지속에 의존하지 마세요. 비트랜잭션 실패는 성공 이력 없이 변경을 남길 수 있습니다.

## 자원 한도

```yaml
script_timeout: 5m # 스크립트별 제한 시간
script_max_rows: 1000 # 쿼리 한 번의 최대 행 수
script_max_bytes: 16777216 # 쿼리 한 번의 최대 바이트 수(16 MiB)
```

초과 시 잘라내지 않고 실패합니다. 대량 조회는 페이지로 나누되 배치 커밋이나 자동 재개가 생기는 것은 아닙니다. 취소 후 DB 중단은 시간이 걸릴 수 있습니다. Go는 `Config.ScriptOptions`를 씁니다. [DB별 동작](./databases#script-support)을 확인하세요.
