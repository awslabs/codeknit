---
title: 시작하기
description: 5분 이내에 codeknit을 시작하고 실행하세요.
---

이 튜토리얼에서는 codeknit을 소스로부터 빌드하고, 첫 번째 파싱을 실행하며, 작성된 `.skt` 출력을 읽게 됩니다. codeknit 저장소의 작은 패키지를 파싱하므로 자신의 프로젝트가 필요하지 않습니다.

## 필수 조건

- **Go 1.26 이상.** `go version`을 실행하여 확인하세요.
- **C 컴파일러.** codeknit은 tree-sitter를 사용하며, 이는 CGo가 필요합니다.
- **Git** 및 **Make**.

## 1. codeknit 빌드

저장소를 클론하고 빌드합니다:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

빌드가 오류 없이 완료되고, `./bin/codeknit`에 바이너리가 생성되는 것을 확인해야 합니다.

## 2. codeknit을 PATH에 추가

`codeknit` 디렉토리에서 `bin`을 `PATH`에 추가합니다. 사용 중인 셸에 맞는 명령을 사용하세요.

bash 또는 zsh의 경우:

```bash
export PATH="$PATH:$(pwd)/bin"
```

fish의 경우:

```fish
fish_add_path (pwd)/bin
```

이 명령은 현재 bash 또는 zsh 세션에만 적용됩니다. 이후 세션에서도 유지하려면 동일한 명령을 `~/.bashrc` 또는 `~/.zshrc`에 추가하되, `$(pwd)/bin` 대신 `bin`의 절대 경로를 사용하세요.

## 3. 설치 확인

```bash
codeknit --version
```

codeknit 버전이 출력되는 것을 확인해야 합니다.

## 4. 첫 번째 파싱 실행

`codeknit` 디렉토리에 머무르며 `internal/console` 패키지를 파싱합니다:

```bash
codeknit parse ./internal/console --edges
```

codeknit은 디렉토리를 스캔하고, 지원되는 각 소스 파일을 파싱한 후 `.skt` 파일을 작성합니다. *`.skt` 파일*은 코드의 심볼과 심볼 간의 관계를 나열하는 간결한 텍스트 파일입니다. `--edges` 플래그는 관계를 추가합니다. 기본적으로는 공간을 절약하기 위해 생략됩니다.

명령이 완료되면 간단한 요약이 표시되고, 현재 디렉토리에 새로운 `./skeleton` 디렉토리가 생성됩니다.

출력 목록을 확인합니다:

```bash
ls ./skeleton
```

`map_001.skt`와 같은 파일이 표시됩니다.

## 5. 출력 읽기

파일을 엽니다:

```bash
cat ./skeleton/map_001.skt
```

먼저 `[symbols]` 섹션이 표시됩니다. *심볼*은 함수, 타입 또는 필드와 같은 코드의 명명된 항목입니다. 심볼은 `## <소스 파일>` 헤딩 아래 그룹화됩니다. 다음과 같은 출력이 포함됩니다:

```skt
[symbols]
## <path to>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exported}
...
S10 callable/function L28-L41 New() -> *S2 {exported}
...
S12 callable/method L49-L49 Error(msg: string) {exported, receiver=*Console}
```

각 줄은 동일한 부분을 포함합니다:

| 부분 | 예시 | 의미 |
| --- | --- | --- |
| Short ID | `S10` | 출력에서 다른 곳에서 사용되는 심볼의 짧은 레이블 |
| 카테고리 및 종류 | `callable/function` | 심볼의 종류 |
| 행 범위 | `L28-L41` | 소스 파일에서 심볼이 포함된 행 |
| 시그니처 | `New() -> *S2` | 이름, 매개변수 및 반환 타입 |
| 속성 | `{exported}` | 중괄호 안의 추가 세부 정보 |

`New() -> *S2`에서 `S2`는 `Console` 타입의 Short ID입니다. codeknit은 심볼을 알면 타입 이름을 Short ID로 대체합니다. 체크아웃 버전에 따라 행 번호가 약간 다를 수 있습니다.

`[edges]` 섹션까지 스크롤합니다. *엣지*는 두 심볼 간의 관계입니다. 각 엣지 줄은 소스 심볼, 관계의 종류, 대상 심볼을 명시합니다:

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

이 예시는 엣지 줄의 형태를 보여줍니다. "`S2`가 `S4`를 포함한다"와 "`S4`가 `S5`를 호출한다"를 의미합니다. `[symbols]`에서 각 Short ID를 찾아 어떤 심볼을 가리키는지 확인할 수 있습니다.

## 달성한 내용

codeknit을 빌드하고, `PATH`에 추가했으며, `codeknit parse`를 실행하고 `.skt` 출력의 `[symbols]` 및 `[edges]` 섹션을 읽었습니다.

## 다음 단계

- 자신의 프로젝트를 파싱하고, 다른 출력 레이아웃을 선택하며, 안전하게 다시 실행하세요: [parse 명령 가이드](../guides/parse-command/)를 참조하세요.
- 코드 구조를 시각화하고 분석하세요: [그래프 명령 가이드](../guides/graph-commands/)를 참조하세요.
- 중복 및 근사 중복 코드를 찾으세요: [fingerprint 명령 가이드](../guides/fingerprint-command/)를 참조하세요.
- `.skt` 형식의 모든 부분을 확인하세요: [출력 형식 참조](../reference/output-format/)를 참조하세요.
- 모든 명령줄 플래그를 확인하세요: [CLI 플래그 참조](../reference/cli-flags/)를 참조하세요.