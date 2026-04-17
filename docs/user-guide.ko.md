# migrate 사용자 가이드

이 가이드는 처음 사용하는 사용자를 위한 것입니다. 명령어와 플래그는 현재 구현(`cmd/migrate`)을 기준으로 합니다.

## 1. 설치

### 1.1 모듈에서 설치

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

특정 버전 설치:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

`vX.Y.Z`를 실제 버전으로 바꾸세요. 예: `v0.2.3`.

### 1.2 로컬 소스에서 설치 (프라이빗 저장소 또는 내부 네트워크)

저장소 루트에서 실행:

```bash
go install ./cmd/migrate
```

### 1.3 설치 확인

```bash
migrate --help
```

명령어를 찾을 수 없는 경우:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

`Repository not found`가 보이면 위의 로컬 소스 설치 경로를 사용하세요.

## 2. 초기화

### 2.1 설정 파일 생성

```bash
migrate new config
```

선택 사항:

```bash
migrate new config dev.json
migrate new config --force
```

기본 설정 템플릿:

```json
{
  "schema_name": "migration_schema",
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD:change_me} dbname=${DB_NAME:postgres} sslmode=disable",
  "working_directory": "",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 2.2 핵심 설정 필드 업데이트

- `dialect`: `postgres`, `mysql`, `mariadb`, `oracle`, `sqlite`, `mssql`, `clickhouse`, `tidb`, `redshift`
- `data_source_name`: DB 연결 문자열
- `migration_source`: 마이그레이션 디렉터리 (기본값: `migrations`)

### 2.3 마이그레이션 이력 테이블 초기화

```bash
migrate create
```

`create`는 출력 없이 성공할 수 있습니다. 다음으로 확인하세요:

```bash
migrate status
```

기존 스키마가 있고 오래된 SQL을 다시 실행하고 싶지 않다면 다음을 사용하세요:

```bash
migrate baseline
```

## 3. 마이그레이션 버전 파일 생성

### 3.1 자동 생성 버전

```bash
migrate new version init_users
```

### 3.2 명시적 버전

```bash
migrate new version add_email -v 202604140002
```

생성되는 파일명 형식:

```text
V<version>__<description>.sql
```

기본 파일 템플릿:

```sql
-- +migrate Up

-- +migrate Down
```

## 4. 업그레이드 (마이그레이션 적용)

먼저 Dry run:

```bash
migrate up --dry-run
```

실제로 적용:

```bash
migrate up
```

그다음 상태 확인:

```bash
migrate status
```

`up`은 출력 없이 성공할 수 있습니다. `status`를 기준 정보로 사용하세요.

## 5. 롤백

### 5.1 대상 버전으로 롤백 (대상 버전은 유지)

```bash
migrate down 202604140001
```

의미: `202604140001`보다 큰 적용된 버전만 롤백됩니다.

### 5.2 적용된 모든 버전 롤백

```bash
migrate down --all
```

참고: `migrate down <to-version>`과 `migrate down --all`은 서로 동시에 사용할 수 없습니다.

### 5.3 Dry-run 롤백

```bash
migrate down 202604140001 --dry-run
migrate down --all --dry-run
```

`down`은 출력 없이 성공할 수 있습니다. 검증하려면 `migrate status`를 실행하세요.

## 6. 상태 확인

```bash
migrate status
```

기계 판독 가능한 출력(스크립트 및 AI 에이전트에 권장):

```bash
migrate status --output json
```

출력 컬럼: `Version`, `Filename`, `Hash`, `Status`.

일반적인 상태:

- `pending`
- `applied`
- `baseline`
- `outOfOrder`
- `hashMismatch`
- `filenameMismatch`

## 7. 도구 자체 업그레이드 또는 다운그레이드

### 7.1 도구 버전 업그레이드

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

### 7.2 도구 버전 다운그레이드

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

예시:

```bash
go install github.com/exc-works/migrate/cmd/migrate@v0.2.3
```

저장소가 프라이빗이라 `go install github.com/...@...`를 사용할 수 없다면, 소스 코드에서 대상 버전을 체크아웃한 뒤
다음을 실행하세요:

```bash
go install ./cmd/migrate
```

### 7.3 현재 도구 버전 확인

```bash
migrate version
```

참고: 릴리스 아티팩트는 릴리스 버전을 출력하며, `go install ./cmd/migrate`로 빌드한 로컬 소스 빌드는 보통 `dev`를 출력합니다.

## 8. 환경 변수 템플릿

`data_source_name`은 다음을 지원합니다:

- `${KEY}`: 필수, 반드시 존재해야 함
- `${KEY:default}`: `KEY`가 없으면 `default` 사용

예시:

```json
{
  "dialect": "postgres",
  "data_source_name": "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
}
```

먼저 환경에 `DB_PASSWORD`가 이미 설정되어 있는지 확인한 뒤 다음을 실행하세요:

```bash
migrate status
```

## 9. 10분 첫 실행 데모 (SQLite)

### 9.1 디렉터리 및 설정 준비

먼저 명령어 사용 가능 여부를 확인하세요:

```bash
migrate --help
```

데모 디렉터리 생성(macOS/Linux):

```bash
mkdir -p ./migrate-demo
cd ./migrate-demo
migrate new config
```

Windows PowerShell 대응:

```powershell
mkdir .\migrate-demo
cd .\migrate-demo
migrate new config
```

`migration_config.json`을 다음과 같이 업데이트:

```json
{
  "schema_name": "migration_schema",
  "dialect": "sqlite",
  "data_source_name": "./demo.sqlite",
  "working_directory": ".",
  "migrate_out_of_order": false,
  "logger_level": "info",
  "migration_source": "migrations"
}
```

### 9.2 초기화 및 마이그레이션 파일 생성

```bash
migrate create
migrate new version init_users -v 202604140001
migrate new version add_email -v 202604140002
```

`migrations/V202604140001__init_users.sql` 편집:

```sql
-- +migrate Up
CREATE TABLE users
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

`migrations/V202604140002__add_email.sql` 편집:

```sql
-- +migrate Up
ALTER TABLE users
    ADD COLUMN email TEXT;

-- +migrate Down
CREATE TABLE users_tmp
(
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users_tmp (id, name)
SELECT id, name
FROM users;
DROP TABLE users;
ALTER TABLE users_tmp
    RENAME TO users;
```

### 9.3 적용, 상태 확인, 롤백

```bash
migrate up --dry-run
migrate up
migrate status
migrate down 202604140001 --dry-run
migrate down 202604140001
migrate status
migrate down --all
migrate status
```

예상 결과:

- `up` 이후: 두 버전 모두 `applied`
- `down 202604140001` 이후: `202604140001=applied`, `202604140002=pending`
- `down --all` 이후: 두 버전 모두 `pending`

## 10. 전역 플래그

특정 설정 파일 사용:

```bash
migrate -c ./configs/dev.json status
```

특정 작업 디렉터리 사용:

```bash
migrate -w ./deploy create
migrate -w ./deploy up
```

## 11. 일반적인 오류와 문제 해결

### 11.1 설정 파일을 찾을 수 없음

오류: `config file ... no such file or directory`

해결:

- 현재 디렉터리에 `migration_config.json`이 있는지 확인
- 또는 `-c`로 설정 경로 전달

### 11.2 누락된 환경 변수

오류: `can't find env value for XXX`

해결:

- `export XXX=...`
- 또는 `${XXX:default}` 사용

### 11.3 불완전한 `down` 인자

오류: `to-version must be set, or use --all`

해결:

- `migrate down <version>` 사용
- 또는 `migrate down --all` 사용

### 11.4 지원되지 않는 dialect

오류: `unsupported dialect: xxx`

해결: 다음 중 하나 사용:

- `postgres` `mysql` `mariadb` `oracle` `sqlite` `mssql` `clickhouse` `tidb` `redshift`

### 11.5 마이그레이션 메타데이터 불일치

오류: `hash mismatch` 또는 `filename mismatch`

해결:

- 이미 적용된 마이그레이션 파일은 수정하지 않기
- 변경이 필요하면 더 높은 버전의 새 마이그레이션 생성
