---
title: Parse 명령어
description: 소스 코드에서 구조적 정보를 추출하여 .skt 파일 또는 JSON으로 저장합니다.
---

`codeknit parse`는 소스 파일을 읽고 해당 심볼(및 선택적으로 심볼 간의 관계)을 `.skt` 파일 또는 JSON으로 작성합니다. 이 가이드는 가장 자주 접하게 될 상황을 다룹니다. 모든 명령어에서 공유되는 플래그 목록은 [CLI 플래그 참조](../../reference/cli-flags/)를, 출력 문법은 [출력 형식 참조](../../reference/output-format/)를 확인하세요.

## 프로젝트를 처음 파싱하기

```bash
codeknit parse ./src
```

codeknit은 청크된 `.skt` 파일을 `./skeleton/`에 작성합니다: `map_001.skt`, `map_002.skt` 등. 모든 것을 한 번에 로드하지 않고 필요한 청크만 선택적으로 열 수 있습니다.

다른 위치에 작성하려면 출력 디렉토리를 두 번째 인수로 전달하세요:

```bash
codeknit parse ./src ./output
```

출력 디렉토리로 소스 디렉토리 또는 그 상위 디렉토리를 사용하지 마세요.

## 출력 모드 선택하기

`--output-mode`로 모드를 선택합니다.

| 모드               | 사용 시기                                    | 결과                                                                         |
| ------------------ | -------------------------------------------- | ---------------------------------------------------------------------------- |
| `directory-flat`   | 대부분의 프로젝트 파싱(기본값)              | `map_001.skt`, `map_002.skt`, ...                                            |
| `directory-tree`   | 하나의 파일 구조를 조회하고자 할 때         | 소스 파일당 하나의 `.skt` 파일, 소스 트리를 반영                            |
| `inline`           | 단일 파일 파싱 또는 다른 도구로 파이핑할 때 | stdout으로 출력; 출력 디렉토리를 사용하지 않음                              |

트리 모드에서는 파일 이름이 소스 확장자를 유지합니다: `src/foo.c`는 `src/foo.c.skt`가 되고, `src/foo.h`는 `src/foo.h.skt`가 됩니다. 큰 파일은 `foo.c_part1.skt`, `foo.c_part2.skt` 등으로 분할됩니다.

```bash
codeknit parse ./src --output-mode directory-tree
```

전체 프로젝트에 `inline` 모드를 사용하지 마세요. 모든 출력이 stdout으로 작성되어 터미널이나 LLM 컨텍스트 창을 넘치게 할 수 있습니다.

모드 선택에 대한 자세한 내용은 [출력 모드](../output-modes/)를 참조하세요.

## 동일한 프로젝트에서 다시 실행하기

codeknit은 선택한 형식에 대해 이미 생성된 출력이 있는 디렉토리에 작성하는 것을 거부합니다. 이는 오래된 청크와 새로운 청크가 섞이는 것을 방지합니다. 이전 출력을 대체하려면 `--clean`을 추가하세요:

```bash
codeknit parse ./src --clean
```

`--clean`은 미리 아무것도 삭제하지 않습니다. codeknit은 새로운 출력을 스테이징 디렉토리에 생성하고, 생성이 성공한 후에만 이전 생성 파일을 대체합니다. 생성이 실패하거나 프로세스가 중단되면 이전 출력이 유지되거나 다음 실행 시 복원됩니다. codeknit이 생성하지 않은 출력 디렉토리의 파일은 보존됩니다. 생성된 경로와 충돌하는 파일이 있으면 실행이 실패합니다.

JSON 모드에서는 `codeknit.json`에도 동일한 규칙이 적용됩니다.

**함정: 혼합된 청크.** 심볼 ID(`S1`, `S2`, ...)와 관계 결과는 실행 및 버전 간에 변경될 수 있습니다. 0.5.0으로 업그레이드한 후에는 `--clean`을 사용하여 전체 출력 세트를 다시 생성하세요. 서로 다른 실행의 청크를 절대 결합하지 마세요.

## 단일 파일 파싱하기

`parse`를 파일에 지정하고 `inline` 모드를 사용하세요:

```bash
codeknit parse ./src/main.go --output-mode inline
```

명시적으로 일반 소스 파일을 선택하면 codeknit은 해당 파일에 대해 무시 파일 및 테스트 파일 필터링을 건너뜁니다. 포함/제외 패턴과 크기 제한은 여전히 적용됩니다.

## 스크립트를 위한 JSON 얻기

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

JSON 문서는 최상위 필드 `files`, `symbols`, `edges`(없을 경우 생략), `errors`(없을 경우 생략)를 가집니다. 작은 예시:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

대신 JSON을 디스크에 작성하려면 `--output-mode inline`을 생략하세요. 디렉토리 모드에서는 JSON이 `codeknit.json`에 작성됩니다. JSON은 단일 문서이며 행 수로 분할되지 않습니다.

## 관계 포함하기

기본적으로 출력은 토큰을 절약하기 위해 `[symbols]` 섹션만 포함합니다. 무엇이 호출하고, 포함하고, 상속하고, 참조하는지 알 필요가 있을 때는 `--edges`를 추가하세요:

```bash
codeknit parse ./src --edges
```

### 불확실한 관계 읽기

Codeknit은 컴파일러 없이 구문을 분석하므로 일부 관계는 확립할 수 없습니다. 각 엣지는 얼마나 확실한지 나타냅니다:

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->`는 확정된 관계입니다.
- `--kind[unresolved]-->`는 codeknit이 대상을 확립할 수 없음을 의미합니다.
- `--kind[ambiguous]-->`는 둘 이상의 대상이 가능함을 의미합니다. 후보 ID는 사용 가능한 경우 나열됩니다.

후보는 가능한 대상이지만 확인된 의존성은 아닙니다. 알려진 엔드포인트는 짧은 ID를 사용합니다. 일치하는 심볼이 없는 미해결 대상은 소스 파일에 상대적인 이름으로 나타날 수 있습니다.

JSON에서 불확실한 엣지는 `resolution`(`"unresolved"` 또는 `"ambiguous"`)을 가지며, 사용 가능한 경우 `candidates`로 전체 심볼 ID를 가집니다. 이들은 `to_short`를 생략합니다.

C 및 C++의 경우 `#include` 지시문은 `meta/file` 및 `meta/include` 심볼이 되므로 해당 엣지는 짧은 ID를 가집니다. 이는 지시문을 기록한 것입니다. codeknit이 헤더를 찾았다는 의미는 아닙니다.

그래프 메트릭은 불확실한 엣지를 제외하고 제외된 수를 보고합니다. HTML 그래프 뷰어는 불확실한 관계의 수를 보여주고 이를 그리지 않습니다. 개별 불확실한 엣지를 검사하려면 `parse`에 `--edges`를 사용하세요. 해당 보고서는 [그래프 명령어](../graph-commands/)를 참조하세요.

엣지 문법은 [출력 형식 참조](../../reference/output-format/)에 있습니다. 해결 규칙 및 제한은 [언어 지원 참조](../../reference/language-support/#relationship-resolution)에 있습니다.

## 출력 크기 줄이기

반복되는 토큰을 압축하려면 `--minify`를 추가하세요:

```bash
codeknit parse ./src --minify
```

이렇게 하면 `dict.skt`(또는 인라인 `[dict]` 섹션)에 사전이 작성됩니다. `d0`, `d1`과 같은 코드를 읽을 때 사전 값으로 대체하세요. 관계 종류도 인코딩되지만 `[unresolved]`와 같은 해결 레이블은 읽을 수 있게 유지됩니다.

청크 크기를 제어하려면 `--max-lines`(기본값 `500`)을 사용하세요:

- `directory-flat` 모드에서는 목표값입니다. 소스 블록은 분할되지 않으므로 청크가 한 블록만큼 초과할 수 있습니다.
- `directory-tree` 모드에서는 각 소스 청크에 대한 하드 제한입니다. 최소값은 `3`입니다. 각 청크는 `[symbols]` 및 `## source-path` 헤더를 반복합니다.

## `.h`로 명명된 C++ 헤더 파싱하기

Codeknit은 기본적으로 `.h` 파일을 C로 파싱합니다. 해당 확장자를 가진 C++ 헤더의 경우 `--header-language cpp`를 설정하세요:

```bash
codeknit parse ./src --header-language cpp
```

이 플래그는 `.h` 파일에만 영향을 미칩니다. `.hpp` 및 `.hxx`로 명명된 파일은 이미 C++로 파싱되며, 다른 확장자는 일반적인 파서가 유지됩니다.

## 어떤 파일을 스캔할지 제어하기

모든 다섯 가지 소스 처리 명령어는 다음 플래그를 허용합니다. 전체 목록은 [CLI 플래그 참조](../../reference/cli-flags/)에 있습니다.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` 및 `--exclude GLOB`은 입력 디렉토리에 상대적인 gitignore 스타일 패턴으로 반복 가능합니다. 셸이 확장하지 않도록 따옴표로 묶으세요. 제외가 우선합니다. 포함은 `.gitignore`, 테스트 파일 필터링 또는 기본 제외를 덮어쓰지 않습니다.
- `--max-file-size BYTES`는 더 큰 파일을 건너뜁니다. 기본값은 `10485760`(10 MiB)입니다. `0`은 제한을 비활성화합니다.
- `--strict`는 스캔 또는 파싱이 불완전한 경우 출력을 작성하기 전에 실패합니다. 기본값은 `false`입니다.

Codeknit은 심볼릭 링크와 비정규 파일을 건너뛰고 보고합니다. 조상 `.gitignore` 규칙은 작업 디렉토리와 독립적으로 적용됩니다.

**함정: 부분 결과.** 읽을 수 없는 경로, 과도하게 큰 파일, 건너뛴 심볼릭 링크 및 구문 경고는 분석을 불완전하게 만듭니다. `--strict` 없이 codeknit은 사용 가능한 결과를 유지하고 문제를 보고합니다:

- stderr에,
- JSON `errors` 필드에, 또는
- SKT `[errors]` 섹션(인라인 또는 디렉토리 모드의 `warnings.skt`)에.

선택된 지원 소스가 없거나 읽은 모든 소스가 실패한 입력은 항상 실패로 종료됩니다. 정확한 보장은 [분석 신뢰성](../../reference/analysis-reliability/)을 참조하세요.

## 파싱 플래그

다음 플래그는 `parse`에만 해당됩니다. 공유 플래그는 [CLI 플래그 참조](../../reference/cli-flags/)에 있습니다.

| 플래그               | 기본값          | 설명                                                                                       |
| -------------------- | ---------------- | ------------------------------------------------------------------------------------------ |
| `--output-mode`      | `directory-flat` | 출력 모드: `inline`, `directory-flat`, 또는 `directory-tree`.                              |
| `--format`           | `skt`            | 출력 형식: `skt` 또는 `json`.                                                              |
| `--header-language`  | `c`              | `.h` 파일에 대한 파서: `c` 또는 `cpp`. 다른 확장자는 변경되지 않습니다.                     |
| `--max-lines`        | `500`            | 플랫 출력의 행 목표; 각 트리 소스 청크에 대한 하드 제한(최소 3).                           |
| `--collect-test`     | `false`          | 테스트 파일 포함.                                                                          |
| `--minify`           | `false`          | 사전 기반 압축 사용.                                                                       |
| `--edges`            | `false`          | `[edges]` 섹션 포함(토큰 절약을 위해 기본적으로 비활성화).                                 |
| `--clean`            | `false`          | 이전 생성 출력을 대체하며, 생성이 성공한 후에만 대체합니다.                               |
| `--workers`          | `NumCPU`         | 최대 동시 파싱 고루틴 수.                                                                  |
| `--verbose`          | `false`          | 진행 정보 출력.                                                                            |

## 팁

- 대부분의 프로젝트에는 `directory-flat`를 사용하고 필요한 청크만 읽으세요.
- 토큰을 절약하려면 큰 코드베이스에 `--minify`를 사용하세요.
- 관계 데이터가 필요할 때만 `--edges`를 전달하세요.
- 특정 언어에 대한 추출을 신뢰하기 전에 [테스트된 기능 매트릭스 및 알려진 제한](../../reference/language-support/)을 확인하세요.

## 함께 보기

- [출력 모드](../output-modes/)
- [출력 형식 참조](../../reference/output-format/)
- [분석 신뢰성](../../reference/analysis-reliability/)
- [그래프 명령어](../graph-commands/)