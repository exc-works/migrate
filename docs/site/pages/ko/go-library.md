# Go에서 사용

루트 패키지의 공개 API는 SemVer를 따릅니다. `internal/*`는 공개 API가 아닙니다. 호출자가 풀을 닫고 적절한 context를 사용하며 같은 DB의 마이그레이션을 직렬화합니다.

`Config.DryRun`, `DryRunOutput`은 미리 보기지만 `Create()`는 이력을 생성합니다. 기본 로거는 `NoopLogger`입니다. [스크립트 내장](./modules)은 `FSSource`와 의존성 설정을 사용합니다.

`OpenConnection`이면 `Connection.Context()`로 Ping과 서비스를 호출하고 마지막에 `Close()`합니다. 네트워크 드라이버는 등록되지만 SQLite는 호출자가 등록합니다. `ConnectionOptions.Prompt`, `HostKeyPrompt`는 명시적이며 취소를 지원해야 합니다. 콜백이 없으면 입력이나 알 수 없는 호스트 신뢰를 허용하지 않습니다.

CLI 외에도 `github.com/exc-works/migrate`를 서비스 코드에서 직접 import하여 마이그레이션을 실행할 수 있습니다. 유닛 테스트, 시작 훅, 관리자 패널 등에 유용합니다.

## 설치

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

## 최소 예제

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

## 핵심 API

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

[Go API](https://github.com/exc-works/migrate/blob/main/migrate.go) · [Go example](https://github.com/exc-works/migrate/blob/main/example_test.go)
