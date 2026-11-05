# 설치

Go를 사용하거나 [Releases](https://github.com/exc-works/migrate/releases)에서 운영체제와 아키텍처에 맞는 파일을 받습니다. SHA-256을 확인하고 `migrate`(Windows는 `migrate.exe`)를 PATH에 넣습니다.

## 모듈에서 설치

```bash
go install github.com/exc-works/migrate/cmd/migrate@latest
```

특정 버전 설치:

```bash
go install github.com/exc-works/migrate/cmd/migrate@vX.Y.Z
```

## 로컬 소스에서 설치 (프라이빗 저장소 또는 내부 네트워크)

저장소 루트에서 실행:

```bash
go install ./cmd/migrate
```

## 설치 확인

```sh
migrate --help
migrate version
migrate capabilities --output json
```

명령어를 찾을 수 없는 경우:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

로컬 빌드는 `dev`입니다. 오래된 Go 설치가 `dev`로 나오면 `go version -m <binary>`로 확인합니다. capabilities는 연결하지 않습니다. [빠른 시작](./quick-start)으로 진행하세요.
