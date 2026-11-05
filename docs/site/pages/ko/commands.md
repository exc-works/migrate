# CLI 참조

`migrate version`, `migrate --help`, `migrate <command> --help`로 확인합니다.

| 전역 옵션 | 동작 |
| --- | --- |
| `-c`, `--config` | 지정 파일, 생략하면 [자동 검색](./configuration#configuration-file-search-order) |
| `-w`, `--working-dir` | 작업 디렉터리 덮어쓰기 |
| `--non-interactive` | 입력이나 SSH 신뢰 요청 대신 실패 |

## 생성

```sh
migrate new config [filename] [--format yaml|json] [--multi] [--shards N] [--ssh-tunnel] [--force]
migrate new config shards.yml --multi --shards 4
migrate new config shards-ssh.yml --multi --ssh-tunnel --proxy socks5
```

기본 `migration_config.yml`. `--shards N`은 `--multi`와 양수가 필요하고 0부터 번호를 붙입니다. `--proxy`는 `--ssh-tunnel`이 필요합니다. 인증은 `--ssh-auth agent|private_key|password`, 덮어쓰기는 `--force`, JSON은 `.json` 또는 `--format json`입니다.

```sh
migrate new version <description> [-v <version>] [--format sql|js|ts|mjs]
```

기본 SQL과 자동 버전 또는 `-v`입니다. `"init users"`처럼 공백은 따옴표로 묶으며 파일에서 `_`로 바뀝니다.

## 데이터베이스 작업

```sh
migrate create
```

이력 테이블만 멱등적으로 생성합니다.

```sh
migrate baseline
```

이미 파일과 일치할 때 실행 없이 기록합니다. [기준선](./history#baseline)을 참고하세요.

```sh
migrate up [--dry-run]
```

미적용 버전을 순서대로 실행합니다. Dry-run은 변경 없이 이력을 읽을 수 있습니다.

```sh
migrate down <to-version> [--dry-run]
migrate down --all [--dry-run]
```

대상보다 높은 버전 또는 전부를 되돌립니다. 대상은 유지되며 대상과 `--all`은 배타적입니다.

```sh
migrate status [--output table|json]
```

기본 표 형식입니다. 여러 DB의 JSON은 이름 순으로 `data_source`, `status`, 성공 시 `migrations`, 실패 시 `error`입니다. create/baseline/status/up/down은 모든 연결을 방문하고 `--parallel N`(기본 8)을 지원합니다. 하나의 실패가 다른 연결을 멈추지 않고 실패 시 0이 아닌 종료 코드입니다.

## 연결 없는 정보

```sh
migrate capabilities [--output text|json]
```

형식, 별칭, 스크립트, 트랜잭션 지원이며 기본 `text`입니다.

```sh
migrate version
```

마이그레이션 버전이 아닌 도구 버전입니다.
