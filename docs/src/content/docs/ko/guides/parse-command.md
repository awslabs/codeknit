---
title: parse 명령어
description: 소스 코드에서 구조적 정보를 추출하여 .skt 파일 또는 JSON으로 저장합니다.
---

`codeknit parse` 명령어는 코드베이스에서 함수, 클래스, 메서드, 변수 및 이들의 관계와 같은 구조적 정보를 추출하여 기본적으로 간결한 `.skt` 형식으로 출력합니다. 스크립트, 통합 또는 하위 도구에서 기계가 읽을 수 있는 출력이 필요할 때는 JSON을 사용하세요.

## 기본 사용법

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: 파싱할 디렉터리 또는 파일의 경로입니다.
- **`[output-dir]`**: 선택적 출력 디렉터리입니다. 제공하지 않으면 기본값은 `./skeleton`입니다.

### 예제

```bash
# 프로젝트를 파싱하고 기본 디렉터리 ./skeleton에 출력
codeknit parse ./src

# 파싱하고 사용자 정의 출력 디렉터리에 저장
codeknit parse ./src ./output

# 단일 파일을 파싱하고 stdout에 출력
codeknit parse ./src/main.go --output-mode inline

# 기계가 읽을 수 있는 JSON을 stdout에 출력
codeknit parse ./src --output-mode inline --format json
```

## 출력 모드

`--output-mode`를 사용하여 출력이 구조화되는 방식을 제어합니다. 세 가지 모드가 제공됩니다:

| 모드               | 설명                                                                                     | 사용 추천 사례                                      |
| ------------------ | ---------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `directory-flat`   | 청크 단위의 `.skt` 파일(예: `map_001.skt`, `map_002.skt`)을 출력 디렉터리에 저장합니다. | ✅ **대부분의 프로젝트** — 기본값이자 권장 모드     |
| `directory-tree`   | 소스 디렉터리 구조를 미러링하여 소스 파일당 하나의 `.skt` 파일을 생성합니다.             | 소스 코드와 함께 출력 탐색                          |
| `inline`           | 모든 출력을 stdout에 덤프합니다.                                                         | 단일 파일 또는 다른 도구로 파이핑할 때              |

> **팁**: 단일 파일로 작업하지 않는 한 기본값으로 `directory-flat`를 사용하세요. 대용량 입력 시 `inline`은 컨텍스트 윈도우를 압도할 수 있으므로 피하세요.

## 플래그

| 플래그             | 기본값           | 설명                                                                                     |
| ------------------ | ---------------- | ---------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | 출력 모드: `inline`, `directory-flat`, 또는 `directory-tree`                             |
| `--format`         | `skt`            | 출력 형식: `skt` 또는 `json`                                                             |
| `--max-lines`      | `500`            | 플랫/트리 모드에서 출력 파일당 최대 행 수                                                |
| `--collect-test`   | `false`          | 분석에 테스트 파일 포함                                                                  |
| `--minify`         | `false`          | 토큰 사용량을 줄이기 위한 사전 기반 압축 활성화                                          |
| `--edges`          | `false`          | 관계 데이터(호출, 포함 등)가 포함된 `[edges]` 섹션 포함                                  |
| `--clean`          | `false`          | 쓰기 전에 출력 디렉터리의 기존 `.skt` 파일 제거                                          |
| `--workers`        | `NumCPU`         | 최대 동시 파싱 고루틴 수(0 = 모든 CPU 코어 사용)                                         |
| `--verbose`        | `false`          | 처리 중 진행 상황 및 타이밍 정보 출력                                                    |

## 일반적인 패턴

```bash
# 프로젝트에서 첫 실행
codeknit parse ./src
```

```bash
# 이전 출력 정리 후 재실행
codeknit parse ./src --clean
```

```bash
# 단일 파일을 파싱하여 stdout에 출력
codeknit parse ./src/main.go --output-mode inline
```

```bash
# 대용량 코드베이스에 대해 출력 최소화
codeknit parse ./src --minify
```

```bash
# 의존성 분석을 위한 관계 엣지 포함
codeknit parse ./src --edges
```

```bash
# 다른 도구를 위한 JSON 출력
codeknit parse ./src --output-mode inline --format json --edges
```

JSON 출력 예시:

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
# 출력에서 소스 트리 구조 미러링
codeknit parse ./src --output-mode directory-tree
```

## 오래된 출력 보호

출력 디렉터리에 이전 실행에서 생성된 `.skt` 파일이 이미 존재하는 경우, `codeknit`은 오래된 데이터와 새로운 데이터가 혼합되는 것을 방지하기 위해 새 출력을 쓰는 것을 거부합니다.

이 동작을 무시하고 출력 디렉터리를 정리한 후 쓰려면 `--clean` 플래그를 사용하세요:

```bash
codeknit parse ./src --clean
```

이렇게 하면 새롭고 일관된 출력 세트가 보장됩니다.

## 팁

- ✅ 대부분의 프로젝트에서는 **기본값으로 `directory-flat`를 사용**하세요. 가독성과 관리 가능성의 균형을 맞출 수 있습니다.
- 🔍 대용량 코드베이스에서는 `--minify`를 사용하여 공유 사전(`dict.skt`)을 통해 토큰 사용량을 줄이세요.
- 🔗 `[edges]` 섹션은 토큰을 절약하기 위해 **기본적으로 제외**됩니다. `calls`, `contains`, `inherits`와 같은 관계 데이터가 필요할 때 `--edges`를 사용하세요.
- 🧾 스크립트나 통합에서 `.skt` 대신 구조화된 데이터가 필요할 때는 `--format json`을 사용하세요.
- 🧹 동일한 출력 디렉터리에서 재실행할 때는 항상 `--clean`을 사용하세요.
- 📁 `.skt` 파일을 편집기에서 소스 파일과 직접 연관시키려면 `directory-tree`를 사용하세요.