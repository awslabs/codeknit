---
title: 시작하기
description: 5분 이내에 codeknit를 시작하고 실행해 보세요.
---

# 시작하기

5분 이내에 codeknit를 시작하고 실행해 보세요.

## 1. 사전 요구 사항

다음과 같은 항목이 필요합니다:

- Go 1.26+
- C 컴파일러 (tree-sitter를 위해 CGo가 필요합니다)

## 2. 소스에서 설치

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# 바이너리는 ./bin/codeknit에 있습니다
```

## 3. PATH에 추가

바이너리를 쉘의 PATH에 추가하세요:

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

변경 사항을 적용하려면 쉘을 다시 로드하거나 `source ~/.bashrc` (또는 `~/.zshrc`)를 실행하세요.

## 4. 설치 확인

codeknit가 작동하는지 확인하세요:

```bash
codeknit --version
```

## 5. 첫 번째 파싱

코드베이스에서 첫 번째 파싱을 실행하세요:

```bash
codeknit parse ./myproject
```

이 명령은:

- `./myproject` 내의 모든 소스 파일을 파싱합니다
- 구조적 정보(함수, 클래스, 관계)를 추출합니다
- 청크 단위의 `.skt` 파일을 `./skeleton/`(기본 출력 디렉터리)에 작성합니다

이 명령을 다시 실행할 경우, 이전 출력을 제거하려면 `--clean`을 사용하세요:

```bash
codeknit parse ./myproject --clean
```

## 6. 출력 결과 읽기

`.skt` 파일은 구조화된 코드 정보를 포함합니다. 다음은 작은 예시입니다:

```skt
[symbols]
## src/main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {exported}
S3 callable/function L10-L12 NewServer(addr: string) -> *S2 {exported}
S4 callable/method L14-L19 Start() {receiver=*Server}

[edges]
S2 --contains--> S4
S3 --returns--> S2
```

주요 섹션:

- `[symbols]`: 파일별로 그룹화된 정의로, 이름, **행 범위**, 메타데이터를 보여줍니다
- `[edges]`: `contains`, `calls`, `inherits`, `returns`와 같은 관계

## 7. 다음 단계

이제 첫 번째 파싱을 실행했으니:

- 파싱 명령에 대해 자세히 알아보기: [파싱 명령 가이드](/codeknit/ko/guides/parse-command/)
- 구조 분석 탐색하기: [그래프 명령 가이드](/codeknit/ko/guides/graph-commands/)
- 중복 감지 이해하기: [핑거프린트 명령 가이드](/codeknit/ko/guides/fingerprint-command/)
- 전체 출력 형식 읽기: [출력 형식 참조](/codeknit/ko/reference/output-format/)
- 사용 가능한 모든 플래그 확인하기: [CLI 플래그 참조](/codeknit/ko/reference/cli-flags/)
