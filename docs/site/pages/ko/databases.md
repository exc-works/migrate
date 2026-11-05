# 지원 데이터베이스

SQL: PostgreSQL, MySQL, MariaDB, SQLite, Oracle, MSSQL, ClickHouse, TiDB, Redshift. `dialect`로 선택합니다.

## 스크립트와 트랜잭션 {#script-support}

| DB | 상태 | 매개변수 | 트랜잭션 |
| --- | --- | --- | --- |
| PostgreSQL | 검증됨 | `$1` | migrate 관리 |
| MySQL / MariaDB | 검증됨 | `?` | 트랜잭션 테이블 필요, DDL 암묵 커밋 가능 |
| SQLite | 검증됨 | `?` | migrate 관리 |
| TiDB | 실험적 | `?` | 분산 TiKV 미검증, DDL 암묵 커밋 가능 |
| Redshift | 실험적, 실제 인스턴스 미검증 | `$1` | 문장별 동작, 전체 PostgreSQL 호환 아님 |
| Oracle | 실험적, 제한적 Oracle Free 검증 | `:1` | DML 관리, DDL 암묵 커밋 |
| ClickHouse | 실험적 | `?` | 트랜잭션과 자동 롤백 없음 |
| MSSQL | 실험적, Azure SQL 미검증 | `@p1` | migrate 관리 |

실험적 스크립트는 미리 보기에도 경고합니다. 대상 환경에서 검증하고 [트랜잭션](./script-transactions)을 확인하세요.

## 이름과 별칭 {#aliases}

대소문자와 앞뒤 공백을 무시하고 `db.dialect`는 정규 이름입니다. MySQL, MariaDB, TiDB는 별개입니다.

| 이름 | 별칭 |
| --- | --- |
| `postgres` | `pg`, `pgsql`, `postgresql`, `pgx` |
| `mssql` | `sqlserver`, `azuresql` |
| `oracle` | `orcl`, `ora`, `go-ora`, `goora`, `godror` |
| `sqlite` | `sqlite3` |
| `mysql`, `mariadb`, `tidb`, `redshift`, `clickhouse` | 추가 별칭 없음 |

## 특수 값 {#value-conversion}

정확한 정수와 소수는 문자열을 유지합니다.

- MSSQL: MONEY는 정확한 문자열, UNIQUEIDENTIFIER는 바이너리. 텍스트는 `CAST(id AS VARCHAR(36))`, rowversion/timestamp는 바이너리입니다.
- Oracle: NUMBER는 텍스트, RAW는 바이트, CLOB는 텍스트. 메타데이터 없는 BLOB는 길이를 제한한 `DBMS_LOB.SUBSTR` 또는 `RAWTOHEX`를 사용합니다. 직접 연결은 `FAST+LOGIN=false`로 재연결을 안정화합니다. CLI는 명시 설정이 없으면 이를 사용합니다.
- Redshift: pgx의 알 수 없는 타입은 변환합니다. VARBYTE는 `FROM_HEX($1)`로 쓰고 `FROM_VARBYTE(payload, 'hex')`로 읽습니다.
- ClickHouse: 큰 수와 소수는 문자열. Array/Tuple/Map/UUID/IP는 스칼라로 변환합니다. String은 UTF-8, 임의 바이너리는 hex입니다. mutation 완료 대기는 `SETTINGS mutations_sync=2`, `rowsAffected`는 `"0"`일 수 있습니다.

```sh
migrate capabilities
```

설치 버전의 지원을 연결 없이 확인하며 자동화는 `--output json`을 사용합니다.
