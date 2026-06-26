---
title: 출력 모드
description: 프로젝트 크기와 워크플로에 맞는 적절한 출력 모드를 선택하세요.
---

codeknit은 `--output-mode` 플래그로 제어되는 세 가지 출력 모드를 지원합니다. 각 모드는 추출된 코드 구조를 디스크(또는 stdout)에 기록하는 방식을 결정합니다.

출력 모드는 출력 형식과 별개입니다. 기본 형식은 `.skt`이며, `--format json`을 전달하면 동일한 파싱 결과를 기계가 읽을 수 있는 JSON으로 내보냅니다. 디렉터리 모드에서는 JSON이 `codeknit.json`에 기록됩니다. `inline` 모드에서는 JSON이 stdout에 기록됩니다.

### directory-flat (기본값, 권장)

- **동작**: `map_001.skt`, `map_002.skt` 등과 같은 청크 단위의 `.skt` 파일을 작성합니다.
- **출력 디렉터리**: 기본값은 `./skeleton/`
- **분할**: 파일이 `--max-lines` 제한(기본값: 500줄)을 초과하면 분할됩니다.
- **사용 사례**: 대부분의 프로젝트에 가장 적합합니다. 파일 크기를 제한하여 출력을 정리하고 읽기 쉽게 유지합니다. 작업과 관련된 청크만 읽을 수 있습니다.
- **축소**: `--minify`가 활성화되면 출력 디렉터리에 압축된 값에 대한 토큰 매핑을 포함하는 `dict.skt` 파일도 생성됩니다.

예시:

```bash
codeknit parse ./src
# 출력: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **동작**: 소스 디렉터리 구조를 정확히 미러링합니다.
- **출력 디렉터리**: 기본값은 `./skeleton/`
- **매핑**: 소스 파일당 하나의 `.skt` 파일이 해당 경로에 생성됩니다.
- **사용 사례**: 특정 파일의 구조를 빠르게 조회하고자 할 때 이상적입니다. 원본 코드베이스와 함께 탐색하는 데 유용합니다.

예시:

```bash
codeknit parse ./src --output-mode directory-tree
# 출력: ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt 등
```

### inline

- **동작**: 모든 출력을 stdout에 덤프합니다.
- **출력 디렉터리**: 생성되지 않음
- **사용 사례**: 단일 파일 또는 매우 작은 프로젝트(5개 미만 파일)에만 권장됩니다. 다른 도구로 출력을 파이핑하거나 단일 파일을 대화형으로 검사할 때 유용합니다.

예시:

```bash
codeknit parse ./src/main.go --output-mode inline
# 출력: 터미널에 직접 출력됨
```

### JSON 형식

- **동작**: `files`, `symbols`, 선택적 `edges`, 선택적 `errors`를 포함하는 단일 JSON 문서를 내보냅니다.
- **출력 위치**: 디렉터리 모드에서는 `codeknit.json`, `inline` 모드에서는 stdout
- **사용 사례**: 스크립트, 편집기 통합, CI 검사 및 구조화된 데이터가 필요한 도구에 가장 적합합니다.

예시:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

샘플 출력:

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

### 결정 표

| 모드               | 가장 적합한 경우                          | 출력 위치                                      |
| ------------------ | ----------------------------------------- | ---------------------------------------------- |
| `directory-flat`   | 대부분의 프로젝트(기본값, 권장)          | `./skeleton/map_001.skt`, `map_002.skt`, ...   |
| `directory-tree`   | 소스 코드와 함께 출력 탐색                | `./skeleton/<미러링된 경로>.skt`              |
| `inline`           | 단일 파일, 다른 도구로 파이핑             | stdout — 단일 파일 또는 매우 작은 프로젝트에만 사용 |

| 형식  | 가장 적합한 경우               | 출력                                              |
| ----- | ------------------------------ | ------------------------------------------------- |
| `skt` | LLM 컨텍스트 및 인적 검사      | `.skt` 파일 또는 stdout                           |
| `json`| 스크립트 및 구조화된 통합      | 디렉터리 모드에서는 `codeknit.json`, `inline` 모드에서는 stdout |

### 경험칙

- **확신이 서지 않을 때** → `directory-flat`(기본값) 사용
- **단일 파일 검사** → `inline` 사용 가능
- **몇 개 이상의 파일** → `directory-flat` 또는 `directory-tree` 선호
- **대규모 코드베이스** → 토큰 사용량을 줄이기 위해 `--minify` 추가
- **동일한 출력에 재실행** → 오래된 `.skt` 파일을 제거하기 위해 `--clean` 사용

### 축소

`--minify` 플래그는 반복되는 토큰(예: `exported`, `async` 또는 일반적인 타입 이름과 같은 속성 키)의 사전 기반 압축을 활성화합니다. 활성화되면:

- 반복되는 값은 짧은 코드(`d0`, `d1`, `d2`, ...)로 대체됩니다.
- 출력 디렉터리에 코드를 원래 값에 매핑하는 `dict.skt` 파일이 작성됩니다.
- 대규모 코드베이스의 출력 크기를 크게 줄입니다.
- `directory-flat` 및 `directory-tree` 모드 모두에서 작동합니다.

예시 축소 출력:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```

이 형식은 토큰 사용량을 최소화하면서 전체 정보를 보존하므로 LLM 기반 분석에 이상적입니다.