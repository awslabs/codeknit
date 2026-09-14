---
title: parse 명령어
description: 소스 코드에서 구조적 정보를 추출하여 .skt 파일 또는 JSON으로 출력합니다.
---

`codeknit parse` 명령어는 코드베이스에서 함수, 클래스, 메서드, 변수 및 이들의 관계와 같은 구조적 정보를 추출하여 기본적으로 간결한 `.skt` 형식으로 출력합니다. 스크립트, 통합 또는 다운스트림 도구에 기계가 읽을 수 있는 출력이 필요한 경우 JSON을 사용하세요.

## 기본 사용법

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: 파싱할 디렉터리 또는 파일의 경로입니다.
- **`[output-dir]`**: 선택적 출력 디렉터리입니다. 제공하지 않으면 기본값은 `./skeleton`입니다.

### 예제

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## 출력 모드

`--output-mode`를 사용하여 출력이 구조화되는 방식을 제어합니다. 세 가지 모드가 제공됩니다:

| 모드               | 설명                                                                                     | 최적 사용 사례                                      |
| ------------------ | ---------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `directory-flat`   | 청크된 `.skt` 파일(예: `map_001.skt`, `map_002.skt`)을 출력 디렉터리에 작성합니다.       | ✅ **대부분의 프로젝트** — 기본 및 권장 모드        |
| `directory-tree`   | 소스 디렉터리 구조를 미러링하여 소스 파일당 하나의 `.skt` 파일을 생성합니다.             | 소스 코드와 함께 출력 탐색                          |
| `inline`           | 모든 출력을 stdout에 덤프합니다.                                                         | 단일 파일 또는 다른 도구로 파이핑                   |

> **팁**: 단일 파일로 작업하지 않는 한 기본적으로 `directory-flat`를 사용하세요. 큰 입력의 경우 `inline`은 컨텍스트 창을 압도할 수 있습니다.

## 플래그

| 플래그             | 기본값           | 설명                                                                                     |
| ------------------ | ---------------- | ---------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | 출력 모드: `inline`, `directory-flat` 또는 `directory-tree`                              |
| `--format`         | `skt`            | 출력 형식: `skt` 또는 `json`                                                             |
| `--max-lines`      | `500`            | 플랫/트리 모드에서 출력 파일당 최대 행 수                                                |
| `--collect-test`   | `false`          | 분석에 테스트 파일 포함                                                                  |
| `--minify`         | `false`          | 토큰 사용량을 줄이기 위한 사전 기반 압축 활성화                                          |
| `--edges`          | `false`          | 관계 데이터(호출, 포함 등)가 포함된 `[edges]` 섹션 포함                                  |
| `--clean`          | `false`          | 작성 전에 출력 디렉터리의 기존 `.skt` 파일 제거                                          |
| `--workers`        | `NumCPU`         | 동시에 파싱하는 고루틴의 최대 수(0 = 모든 CPU 코어 사용)                                 |
| `--verbose`        | `false`          | 처리 중 진행 상황 및 타이밍 정보 출력                                                    |

## 일반적인 패턴

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

JSON 출력 예제:

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

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## 관계 불확실성

`--edges`를 사용하면 Codeknit은 구문과 범위가 고유한 대상을 확립하지 않더라도 관계를 보존합니다:

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

생략된 상태는 분석기에 의해 해결된 것을 의미합니다. `unresolved`는 의존성을 확립할 수 없음을 의미하며, `ambiguous`는 경쟁하는 해석이 가능함을 의미합니다. 후보는 가능한 대상이지만 확인된 연결은 아닙니다. 알려진 엔드포인트는 ShortID를 사용하며, 심볼이 없는 미해결 대상은 이름으로 나타날 수 있습니다.

JSON은 동일한 정보를 `resolution` 및 `candidates`에 담고 있으며, 불확실한 대상에 대해서는 `to_short`를 생략합니다. C/C++ include는 `meta/file` 및 `meta/include` 심볼을 포함하므로 엣지는 ShortID를 사용합니다. 이는 컴파일러 수준의 헤더 해석을 주장하지 않고 작성된 지시어를 기록합니다.

의존성 메트릭은 불확실한 관계를 제외합니다. HTML 그래프는 개별 불확실한 링크를 그리지 않고 그 수를 표시합니다. 문법은 [출력 형식 참조](../../reference/output-format/)를, 해상도 제한은 [언어 지원 참조](../../reference/language-support/)를 참조하세요.

## 오래된 출력 보호

출력 디렉터리에 이전 실행에서 생성된 `.skt` 파일이 이미 있는 경우, `codeknit`은 오래된 데이터와 새로운 데이터가 혼합되는 것을 방지하기 위해 새 출력을 작성하지 않습니다.

이 동작을 무시하고 출력 디렉터리를 정리한 후 작성하려면 `--clean` 플래그를 사용하세요:

```bash
codeknit parse ./src --clean
```

이렇게 하면 새롭고 일관된 출력 세트가 보장됩니다.

`0.5.0`으로 업그레이드할 때는 전체 출력 세트를 다시 생성하세요. 심볼 ID와 관계 결과는 변경될 수 있으므로 서로 다른 실행의 청크를 혼합하지 마세요.

## 팁

- ✅ 대부분의 프로젝트에는 **기본적으로 `directory-flat`를 사용**하세요. 가독성과 관리 가능성의 균형을 맞춥니다.
- 🔍 대규모 코드베이스에서 토큰 사용량을 줄이려면 `--minify`를 사용하여 공유 사전(`dict.skt`)을 통해 압축하세요.
- 🔗 `[edges]` 섹션은 토큰을 절약하기 위해 **기본적으로 제외**됩니다. `calls`, `contains` 또는 `inherits`와 같은 관계 데이터가 필요한 경우 `--edges`를 사용하세요.
- 🧾 스크립트나 통합에서 구조화된 데이터가 필요한 경우 `.skt` 대신 `--format json`을 사용하세요.
- 🧹 동일한 출력 디렉터리에서 다시 실행할 때는 항상 `--clean`을 사용하세요.
- 📁 `.skt` 파일을 에디터에서 소스 파일과 직접 연관시키려면 `directory-tree`를 사용하세요.