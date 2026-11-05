# 마이그레이션 실행

대상 연결과 파일을 확인합니다.

```sh
migrate create
```

필요하면 이력 테이블만 만들고 파일은 실행하지 않습니다. 변경이 이미 존재하면 [기준선](./history#baseline)을 확인합니다.

```sh
migrate up --dry-run
migrate up
migrate status
```

미적용 버전을 순서대로 실행하고 성공을 기록합니다. 재실행해도 적용된 버전은 반복하지 않습니다. 출력이 없더라도 status로 확인합니다. Dry-run은 이력을 읽거나 SSH를 열 수 있지만 변경을 적용하지 않습니다. 스크립트를 실행하지 않아 동적 SQL은 예측하지 못합니다.

적용한 버전보다 낮은 미적용 버전은 기본적으로 건너뜁니다. 현재 스키마와 호환되는지 확인 후 `migrate_out_of_order`를 설정합니다.

```sh
migrate --non-interactive up
migrate --non-interactive status --output json
```

환경 변수와 자격 증명을 미리 주입하고 같은 DB의 명령은 직렬화합니다. 이력만으로 동시 실행, 충돌, 암묵적 커밋을 넘는 정확히 한 번 실행을 보장하지 않습니다. 실패 후 상태를 확인하고 재시도하세요. [문제 해결](./troubleshooting)을 참고하세요.
