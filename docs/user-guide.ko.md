# migrate 사용자 가이드

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
migrate new config dev.yml
migrate new config --force
```

기본 설정 템플릿:

```yaml
schema_name: migration_schema
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:postgres} sslmode=disable"
working_directory: ""
migrate_out_of_order: false
logger_level: info
migration_source: migrations
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
migrate new version "init users"
```

### 3.2 명시적 버전

```bash
migrate new version "add email" -v 202604140002
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

## 7. 도구 버전 확인

```bash
migrate version
```

## 8. 환경 변수 템플릿

`data_source_name`은 다음을 지원합니다:

- `${KEY}`: 필수, 반드시 존재해야 함
- `${KEY:default}`: `KEY`가 없으면 `default` 사용

예시:

```yaml
dialect: postgres
data_source_name: "host=127.0.0.1 port=5432 user=${DB_USER:postgres} password=${DB_PASSWORD} dbname=${DB_NAME:app} sslmode=disable"
```

먼저 환경에 `DB_PASSWORD`가 이미 설정되어 있는지 확인한 뒤 다음을 실행하세요:

```bash
migrate status
```

## 9. 10분 첫 실행 데모 (SQLite)

[10분 첫 실행 데모 (SQLite) · English](site/pages/quick-start.md)

## 10. 전역 플래그

특정 설정 파일 사용:

```bash
migrate -c ./configs/dev.yml status
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

- 현재 디렉터리에 `migration_config.yml`이 있는지 확인
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

- 버전 관리에서 원래 마이그레이션 파일 복원
- 변경이 필요하면 더 높은 버전의 새 마이그레이션 생성

## 12. migrate를 Go 라이브러리로 내장하기

CLI 외에도 `github.com/exc-works/migrate`를 서비스 코드에서 직접 import하여 마이그레이션을 실행할 수 있습니다. 유닛 테스트, 시작 훅, 관리자 패널 등에 유용합니다.

### 12.1 설치

```bash
go get github.com/exc-works/migrate
```

필요한 데이터베이스 드라이버를 import하세요(라이브러리는 특정 드라이버에 고정되지 않습니다):

```go
import (
    _ "modernc.org/sqlite"             // sqlite
    _ "github.com/jackc/pgx/v5/stdlib" // postgres
    _ "github.com/go-sql-driver/mysql" // mysql / mariadb / tidb
    // ...
)
```

### 12.2 최소 예제

```go
package main

import (
    "context"
    "database/sql"

    _ "modernc.org/sqlite"

    "github.com/exc-works/migrate"
)

func main() {
    db, err := sql.Open("sqlite", "./app.sqlite")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    svc, err := migrate.NewService(context.Background(), migrate.Config{
        Dialect:         migrate.NewSQLiteDialect(),
        DB:              db,
        MigrationSource: migrate.DirectorySource{Directory: "./migrations"},
    })
    if err != nil {
        panic(err)
    }

    if err := svc.Create(); err != nil { // idempotent: 이력 테이블이 없으면 생성
        panic(err)
    }
    if err := svc.Up(); err != nil {
        panic(err)
    }
}
```

### 12.3 핵심 API

- `migrate.NewService(ctx, migrate.Config)`는 마이그레이션 실행기를 만듭니다
- `svc.Create()`는 이력 테이블 `migration_schema`를 생성합니다(idempotent)
- `svc.Up()`은 아직 적용되지 않은 모든 마이그레이션을 적용합니다
- `svc.Down(toVersion, all)`은 대상 버전까지 또는 전체를 롤백합니다
- `svc.Status()`는 `[]migrate.MigrationStatus`를 반환합니다
- `svc.Baseline()`은 기존 pending 파일을 `baseline`으로 표시합니다

자주 쓰는 타입:

- 방언(생성자를 권장 — `Dialect` 인터페이스를 반환): `migrate.NewPostgresDialect()`, `NewMySQLDialect()`, `NewSQLiteDialect()`, `NewMSSQLDialect()`, `NewOracleDialect()`, `NewClickHouseDialect()`, `NewMariaDBDialect()`, `NewTiDBDialect()`, `NewRedshiftDialect()`, 또는 이름 기반 조회용 `migrate.DialectFromName("postgres")`
- 소스: `DirectorySource`(파일 시스템), `StringSource`(인메모리 슬라이스), `FSSource`(모든 `fs.FS`, 예: `//go:embed` 또는 `os.DirFS`), `CombinedSource`(여러 소스 결합)
- 로거: `migrate.NoopLogger{}`(기본), `migrate.NewStdLogger("info", os.Stdout)`, 또는 `migrate.Logger`를 구현한 임의의 타입

[Go API](../migrate.go) · [Go example](../example_test.go)
