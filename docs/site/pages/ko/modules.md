# 공유 모듈

`.js`, `.ts`, `.mjs`의 정적 import와 재내보내기를 지원합니다. 보조 파일은 독립 마이그레이션이 아닙니다.

```text
project/
  migrations/V2__normalize.ts
  shared/names.ts
```

```ts
import { normalize } from "../shared/names.ts";

export function up(db) {
  db.exec("UPDATE users SET name=? WHERE id=?", [normalize(" Alice "), "1"]);
}
```

가져오는 파일 기준 경로입니다. `shared/names.ts`에 `normalize`를 정의하고 `users`를 먼저 만드세요. [완전한 예제](./examples#local)를 참고하세요.

npm은 lockfile로 미리 설치합니다. exports, 하위 경로, 중첩 의존성, 정적 CommonJS를 지원하며 `neutral` 조건과 `module`, 다음 `main`을 사용합니다. Node API, 네이티브 확장, 동적 import/require는 없습니다. migrate는 패키지를 설치하지 않습니다.

## 내장 파일

`FSSource`는 `embed.FS` 같은 자체 `fs.FS` 안에서 확장자와 `index`를 찾지만 가상 package.json/node_modules는 찾지 않습니다. npm에는 실제 디렉터리가 필요합니다.

```go
source := migrate.WithModules{
    Source: migrate.FSSource{
        FS:   embeddedFiles,
        Root: "migrations",
    },
    DependencyDir: projectDirectory,
}
```

상대 모듈은 가상 파일시스템에 남습니다. `DependencyDir`가 없으면 호스트를 검색하지 않습니다. `StringSource`는 `WithModules`와 `FS`/`Root` 또는 `Directory`를 사용합니다. 사용자 정의 소스는 `MigrationEntriesSource`를 구현할 수 있습니다. [Go](./go-library)도 참고하세요.

## 재현성 {#checksums}

해시는 진입 파일만 포함합니다. 보조 파일, 패키지, lockfile 변경은 불일치나 적용된 버전의 재실행을 유발하지 않습니다. 의존성을 고정하고 이전 동작을 유지하세요. 선택 범위 전체 의존성을 쓰기 전에 컴파일하므로 나중 파일의 오류도 먼저 실행할 파일을 차단합니다.
