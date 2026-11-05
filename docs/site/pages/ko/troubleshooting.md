# 문제 해결

`migrate version`, `migrate capabilities --output json`, `migrate status --output json`으로 시작합니다. status만 DB 연결이 필요합니다.

[설치](./installation)의 PATH를 확인하고 프로젝트에서 설정을 생성하거나 `-c` / `-w`를 사용하세요. YAML 오류 위치, 값, 공백 들여쓰기를 확인하고 Tab은 쓰지 않습니다. [검색 순서](./configuration#configuration-file-search-order)를 참고하세요.

`can't find env value for XXX`는 `${XXX}`가 없습니다. 변수를 지정하거나 적절한 비밀 아닌 값에는 `${XXX:default}`를 사용합니다. SSH password_env는 변수 이름입니다.

capabilities로 형식을 확인하고 오래된 버전을 업데이트합니다. unsupported dialect는 [별칭](./databases#aliases)을 확인합니다. 모듈은 경로, npm 설치, 내장 소스의 `DependencyDir`를 확인하며 Node/비동기 의존 패키지는 호환되지 않습니다.

적용된 파일 내용과 이름을 Git에서 복원하고 새 버전을 추가하세요. 상태가 불명확할 때 baseline으로 우회하지 마세요. 출력 없는 명령도 성공할 수 있으니 applied, baseline, outOfOrder를 확인합니다.

`to-version must be set, or use --all`은 대상 또는 `--all`이 필요합니다. 대상은 유지되고 down 없는 스크립트는 되돌릴 수 없습니다.

실패 후 재시도 전 데이터와 이력을 조사하세요. DDL 커밋, 비트랜잭션 작업, COMMIT 근처 단절은 변경을 남길 수 있고 baseline으로 복구하지 못합니다. SSH 자격 증명, known_hosts, 프록시, 점프 서버의 DB 접근을 확인하세요. 자동 재접속이나 재실행은 없습니다. 비밀과 행 데이터를 공유하지 말고 통제된 서버 로그를 확인합니다.
