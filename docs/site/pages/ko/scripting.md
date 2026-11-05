# 스크립트 작성

JS, TS, MJS로 데이터를 처리하며 실행에 Node.js가 필요 없습니다.

```sh
migrate new version "normalize names" --format ts
```

`V<version>__<description>.<format>`은 SQL과 버전을 공유하며 MJS와 JS의 제한은 같습니다.

```ts
export function up(db) {
  for (const row of db.query("SELECT id, name FROM users")) {
    db.exec("UPDATE users SET name=? WHERE id=?", [row.name.trim(), row.id]);
  }
}
```

SQLite의 `users.name` 예제입니다. 생성한 up은 구현 전까지 오류를 발생시킵니다. async, generator, Promise는 지원하지 않고 일반 동기 반환값은 무시합니다. [SQL 매개변수](./script-api)는 DB마다 다릅니다.

공백 제거는 원본을 잃으므로 이 예제에는 down이 없습니다. 역연산을 구현한 후에만 `down(db)`를 내보내세요. 복원하려면 원본 값을 보관합니다.

```sh
migrate up --dry-run
migrate up
migrate status
```

Dry-run은 스크립트를 실행하지 않습니다. DB 작업은 모듈 최상위가 아닌 up/down에 둡니다. 마이그레이션 사이의 전역 변수에 의존하지 마세요. 템플릿과 [migration-api.d.ts](../assets/migration-api.d.ts)에 타입이 있지만 변환은 타입 검사하지 않습니다. Node API, 네트워크, 파일, 환경 변수, 타이머, 동적 로딩은 제공하지 않습니다. npm도 확인하세요. [모듈](./modules), [트랜잭션](./script-transactions)을 참고하세요.
