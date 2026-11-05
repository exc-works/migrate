# 마이그레이션 되돌리기

Down 또는 `down(db)`가 구현한 작업만 되돌리며 삭제한 데이터를 자동 복구하지 않습니다.

```sh
migrate down 202611030001 --dry-run
migrate down 202611030001
migrate status
```

대상보다 높은 버전만 되돌립니다. 1, 2, 3이면 `down 2`는 3만 되돌리고 1, 2는 유지합니다.

```sh
migrate down --all --dry-run
migrate down --all
migrate status
```

대상과 `--all`은 함께 사용하지 않습니다. 성공하면 `pending`으로 돌아옵니다. 전체 범위와 [트랜잭션](./script-transactions)을 확인하세요. down이 없으면 범위가 차단되고 빈 down은 아무것도 되돌리지 않습니다. DDL과 비트랜잭션 변경은 실패 후 남을 수 있습니다.
