# 설정

프로젝트 디렉터리에서 `migration_config.yml`을 생성하고 편집합니다.

```sh
migrate new config
```

## 데이터베이스 연결

환경 변수 `DATABASE_DSN`을 사용하는 PostgreSQL 예제입니다.

```yaml
# 데이터베이스 종류
dialect: postgres
# 실행 전에 DATABASE_DSN 설정
data_source_name: "${DATABASE_DSN}"
# 마이그레이션 파일 디렉터리
migration_source: migrations
# 적용한 버전의 이력 테이블
schema_name: migration_schema
```

| 필드 | 용도 | 기본값 |
| --- | --- | --- |
| `dialect` | 데이터베이스 종류 | `postgres` |
| `data_source_name` | 단일 연결, `data_sources`와 동시 사용 불가 | 단일 연결에서 필수 |
| `data_sources` | 이름이 있는 여러 연결 | 비활성화 |
| `migration_source` | 파일 디렉터리 | `migrations` |
| `schema_name` | 이력 테이블 | `migration_schema` |
| `working_directory` | 파일 검색 기준 | 현재 디렉터리 |
| `migrate_out_of_order` | 적용한 버전보다 오래된 미적용 버전 실행 | `false` |
| `logger_level` | `debug`, `info`, `warn`, `error` | `info` |

### 연결 예제

호스트, 포트, 데이터베이스 또는 Oracle 서비스 이름을 바꾸고 `DB_USER`, `DB_PASSWORD`를 설정합니다. SQLite는 파일 경로입니다.

::: code-group

```yaml [PostgreSQL]
dialect: postgres
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@db.example:5432/app?sslmode=require"
```

```yaml [MySQL]
dialect: mysql
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [MariaDB]
dialect: mariadb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:3306)/app?parseTime=true"
```

```yaml [SQLite]
dialect: sqlite
data_source_name: "./demo.sqlite"
```

```yaml [Oracle]
dialect: oracle
data_source_name: "oracle://${DB_USER}:${DB_PASSWORD}@db.example:1521/FREEPDB1"
```

```yaml [SQL Server]
dialect: mssql
data_source_name: "sqlserver://${DB_USER}:${DB_PASSWORD}@db.example:1433?database=app"
```

```yaml [ClickHouse]
dialect: clickhouse
data_source_name: "clickhouse://${DB_USER}:${DB_PASSWORD}@db.example:9000/app"
```

```yaml [TiDB]
dialect: tidb
data_source_name: "${DB_USER}:${DB_PASSWORD}@tcp(db.example:4000)/app?parseTime=true"
```

```yaml [Redshift]
dialect: redshift
data_source_name: "postgres://${DB_USER}:${DB_PASSWORD}@cluster.example:5439/app?sslmode=require"
```

:::

URL의 사용자 이름과 암호에 있는 `@`, `:`, `/` 같은 문자는 URL 인코딩해야 합니다. 변수 치환은 자동 인코딩하지 않습니다. MySQL/MariaDB/TiDB의 `user:password@tcp(...)`는 URL 인코딩을 사용하지 않습니다. TLS/SSL은 서버 요구에 맞춥니다.

### 로그 수준 {#logging}

| 값 | 표시 내용 |
| --- | --- |
| `debug` | 디버그, 정보, 경고, 오류 |
| `info` | 정보, 경고, 오류 |
| `warn` | 경고, 오류. `warning`도 가능 |
| `error` | 오류만 |

```yaml
logger_level: debug # 문제 조사 시 디버그 메시지 표시
```

## 환경 변수 {#environment-variables}

`${DATABASE_DSN}`은 변수가 없으면 오류입니다. `${DB_NAME:app}`은 없을 때 `app`을 사용합니다. 같은 YAML을 환경별로 재사용하고 자격 증명은 환경 변수나 비밀 저장소로 전달합니다. SSH 호스트, 사용자, 프록시 주소도 변수를 지원합니다.

## 파일 검색 순서 {#configuration-file-search-order}

작업 디렉터리에서 처음 발견한 파일을 사용합니다.

1. `migration_config.yml`
2. `migration_config.yaml`
3. `migration_config.json`

파일을 합치지 않습니다. 발견한 파일이 잘못되었거나 읽을 수 없으면 오류입니다. `-c`를 지정하면 그 파일만 사용하며 없어도 다른 파일을 찾지 않습니다.

```sh
migrate -c ./configs/dev.yml status
```

## 상황별 설정 생성

| 상황 | 명령 |
| --- | --- |
| 단일 데이터베이스 | `migrate new config` |
| 사용자 지정 이름 | `migrate new config dev.yml` |
| 여러 데이터베이스 | `migrate new config --multi` |
| 네 데이터베이스 | `migrate new config --multi --shards 4` |
| SSH Agent | `migrate new config --ssh-tunnel` |
| 개인 키 | `migrate new config --ssh-tunnel --ssh-auth private_key` |
| SSH 암호 | `migrate new config --ssh-tunnel --ssh-auth password` |
| SOCKS5 | `migrate new config --ssh-tunnel --proxy socks5` |
| HTTP CONNECT | `migrate new config --ssh-tunnel --proxy http` |

옵션을 조합할 수 있습니다. 덮어쓰기는 `--force`가 필요합니다. 전체 옵션은 `migrate new config --help`를 참고하세요.

```sh
migrate new config shards.yml --multi --shards 4 --ssh-tunnel --proxy socks5
```

## 여러 데이터베이스 {#multiple-databases}

`data_source_name` 대신 `data_sources`를 사용합니다. 종류, 파일 디렉터리, 이력 테이블 이름은 공유합니다. 생성 이름은 `shard_0`, 변수는 `SHARD_0_DSN`부터 시작합니다.

```yaml
dialect: postgres
migration_source: migrations
schema_name: migration_schema
data_sources:
  shard_0:
    data_source_name: "${SHARD_0_DSN}" # 첫 번째 데이터베이스 연결
  shard_1:
    data_source_name: "${SHARD_1_DSN}" # 두 번째 데이터베이스 연결
```

![마이그레이션 파일 → shard_0 / shard_1 → 결과 수집](../assets/diagrams/multiple-databases-ko.svg)

```sh
migrate create
migrate up --dry-run
migrate up
migrate status
```

각 데이터베이스가 자체 이력으로 미적용 버전을 골라 순서대로 실행하고 성공한 버전은 반복하지 않습니다. `new version`은 공유 파일 하나만 만듭니다. 기본 동시 실행 수는 8이며 `--parallel 1`은 순차 실행입니다.

```sh
migrate up --parallel 4
migrate status --output json
```

하나가 실패해도 다른 데이터베이스는 계속됩니다. 결과는 이름별로 모이고 실패가 있으면 종료 코드는 0이 아닙니다. 데이터베이스 간 트랜잭션이 아니므로 다른 성공을 되돌리지 않습니다. Ctrl+C는 새 작업을 중단하고 실행 중 작업을 취소합니다.

최상위 `ssh_tunnel`을 공유합니다. 개별 SSH 객체는 전체를 교체하며, 생략 또는 `null`은 상속합니다. [SSH](./ssh-tunnels)를 참고하세요. 실제 DB는 한 번만 등록하고 같은 DB에 명령을 중복 실행하지 마세요. 실행 중 파일과 의존성을 변경하지 말고 실패 후 데이터와 이력을 확인한 뒤 재시도합니다.

## 작업 디렉터리

```sh
migrate -w ./deploy up
migrate -w ./deploy -c configs/dev.yml status
```

`-w`에서 설정을 찾고 `working_directory`를 덮어씁니다. 상대 `migration_source`는 작업 디렉터리 기준이며 `-c`만으로 바뀌지 않습니다. 상대 SSH 파일 경로는 설정 파일 디렉터리 기준입니다.
