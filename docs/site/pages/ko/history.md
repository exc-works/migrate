# 이력과 기준선

버전, 파일 이름, 해시를 실제 스키마와 비교합니다.

```sh
migrate status
migrate status --output json
```

| 상태 | 의미 | 조치 |
| --- | --- | --- |
| `pending` | 미적용 | 확인 후 `up` |
| `applied` | 성공 | 파일 유지 |
| `baseline` | 실행 없이 기록 | 인수한 이력 유지 |
| `outOfOrder` | 더 오래된 미적용 버전 | 순서 정책 확인 |
| `hashMismatch` | 내용 변경 | 원본 복원 |
| `filenameMismatch` | 이름 변경 | 원래 이름 복원 |

불일치는 실행을 차단합니다. 변경은 새 버전으로 추가합니다. 스크립트 해시는 [모듈](./modules#checksums)을 포함하지 않습니다.

## 기존 데이터베이스 인수 {#baseline}

스키마와 데이터가 파일에 이미 일치할 때만 사용합니다.

1. 실제 상태와 파일을 비교합니다.
2. 필요하면 `migrate create`를 실행합니다.
3. `migrate baseline`, 이어서 `migrate status`를 확인합니다.
4. 더 높은 버전으로 다음 변경을 추가합니다.

baseline은 실행 없이 기록하며 부분 실패를 복구하지 않습니다.
