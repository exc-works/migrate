# SQLite 빠른 시작

별도 서버 없이 SQL 마이그레이션을 작성하고 실행한 뒤 되돌립니다.

## 설치

```sh
go install github.com/exc-works/migrate/cmd/migrate@latest
migrate --help
migrate version
```

## 설정

```sh
mkdir migrate-demo
cd migrate-demo
migrate new config
```

`migration_config.yml`을 다음으로 교체합니다.

```yaml
# 마이그레이션 이력 테이블
schema_name: migration_schema
dialect: sqlite
data_source_name: ./demo.sqlite # 로컬 SQLite 파일
working_directory: .
migrate_out_of_order: false
logger_level: info
migration_source: migrations # 마이그레이션 디렉터리
```

## 마이그레이션 생성

```sh
migrate create
migrate new version "init users" -v 202611030001
```

`migrations/V202611030001__init_users.sql`을 다음으로 교체합니다.

```sql
-- +migrate Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE users;
```

## 실행 및 확인

```sh
migrate up --dry-run
migrate up
migrate status
```

버전 `202611030001`이 `applied`인지 확인합니다.

## 되돌리기

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

테이블이 삭제되고 `pending`으로 돌아옵니다. [SQL](./sql-migrations), [설정](./configuration), [스크립트](./scripting)를 참고하세요.
