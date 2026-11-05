# 예제 실행

저장소 체크아웃에서 실행합니다. 모두 SQLite를 사용하고 내부에서 데이터를 검증합니다.

## 임시 CLI

같은 셸을 유지하세요.

```sh
example_workspace=$(mktemp -d)
go build -o "$example_workspace/migrate" ./cmd/migrate
```

## 로컬 보조 모듈 {#local}

[소스](https://github.com/exc-works/migrate/tree/main/examples/script-imports/local)의 두 마이그레이션이 TypeScript를 공유합니다.

```sh
export MIGRATE_EXAMPLE_DB="$example_workspace/local.sqlite"
"$example_workspace/migrate" -w examples/script-imports/local create
"$example_workspace/migrate" -w examples/script-imports/local up --dry-run
"$example_workspace/migrate" -w examples/script-imports/local up
"$example_workspace/migrate" -w examples/script-imports/local status
"$example_workspace/migrate" -w examples/script-imports/local down 2
"$example_workspace/migrate" -w examples/script-imports/local down --all
```

up 뒤 3개 버전과 이름 `alice`, `bob`입니다. `down 2`는 1, 2를 유지하고 `down --all`은 테이블을 삭제합니다.

## npm 패키지

Node/npm은 설치에만 필요합니다. lockfile은 동봉된 로컬 패키지를 사용해 레지스트리 접속이 없습니다.

```sh
npm --prefix examples/script-imports/npm ci --ignore-scripts --no-audit --no-fund --offline
export MIGRATE_EXAMPLE_DB="$example_workspace/npm.sqlite"
"$example_workspace/migrate" -w examples/script-imports/npm create
"$example_workspace/migrate" -w examples/script-imports/npm up
"$example_workspace/migrate" -w examples/script-imports/npm status
"$example_workspace/migrate" -w examples/script-imports/npm down --all
```

두 번째 마이그레이션이 `@migrate-example/normalizer/names`를 가져와 `alice`를 확인합니다. 두 버전이 적용됩니다.

## Go에 내장

```sh
go run ./examples/script-imports/embedded
```

`embed.FS`와 전용 메모리 DB를 쓰며 npm이 필요 없습니다. 출력:

```text
after up: alice
after down: all migrations pending
```

## 정리

```sh
rm -r "$example_workspace"
```

npm의 무시된 `node_modules`는 남습니다. 루트에서 `./scripts/test-script-examples.sh`로 임시 복사본을 자동 검증합니다. [재현성](./modules#checksums)을 참고하세요.
