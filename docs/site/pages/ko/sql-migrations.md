# SQL 마이그레이션 작성

파일 하나에 논리적 변경 하나를 둡니다. SQL과 스크립트는 버전을 공유합니다.

```sh
migrate new version "init users"
migrate new version "add email" -v 202611040002
```

공백이 있는 설명은 따옴표로 묶습니다. 파일 이름에서는 `_`가 됩니다. `-v`가 없으면 버전을 생성합니다. 이름은 `V<version>__<description>.sql`입니다. 적용된 파일을 수정하거나 이름을 바꾸지 말고 새 버전을 추가합니다.

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

Up은 적용, Down은 반대 작업입니다. 이 Down은 데이터도 삭제하므로 복구 방법을 계획하세요.

문장 내부의 세미콜론으로 분리되면 안 될 경우 `StatementBegin` / `StatementEnd`를 사용합니다.

```sql
-- +migrate Up
-- +migrate StatementBegin
CREATE FUNCTION migration_answer() RETURNS integer AS $$
BEGIN
    RETURN 42;
END;
$$ LANGUAGE plpgsql;
-- +migrate StatementEnd

-- +migrate Down
DROP FUNCTION migration_answer();
```

SQL 문법과 DDL 트랜잭션은 DB에 따라 다릅니다. 다음은 [실행](./applying)입니다.
