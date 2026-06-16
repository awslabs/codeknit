---
title: 출력 모드
description: 프로젝트 크기와 워크플로우에 맞는 적절한 출력 모드를 선택하세요.
---

codeknit는 `--output-mode` 플래그로 제어되는 세 가지 출력 모드를 지원합니다. 각 모드는 추출된 코드 구조를 디스크(또는 stdout)에 기록하는 방식을 결정합니다.

### directory-flat (기본값, 권장)

- **동작**: `map_001.skt`, `map_002.skt` 등과 같은 청크 단위의 `.skt` 파일을 작성합니다.
- **출력 디렉터리**: 기본값 `./skeleton/`
- **분할**: 파일이 `--max-lines` 제한(기본값: 500줄)을 초과하면 분할됩니다.
- **사용 사례**: 대부분의 프로젝트에 가장 적합합니다. 파일 크기를 제한하여 출력을 체계적이고 읽기 쉽게 유지합니다. 작업과 관련된 청크만 읽을 수 있습니다.
- **축소(minification)**: `--minify`가 활성화되면 출력 디렉터리에 토큰 매핑이 포함된 `dict.skt` 파일도 생성됩니다.

예시:

```bash
codeknit parse ./src
# 출력: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **동작**: 소스 디렉터리 구조를 정확히 미러링합니다.
- **출력 디렉터리**: 기본값 `./skeleton/`
- **매핑**: 소스 파일당 하나의 `.skt` 파일이 해당 경로에 생성됩니다.
- **사용 사례**: 특정 파일의 구조를 빠르게 조회하고자 할 때 이상적입니다. 원본 코드베이스와 함께 탐색하는 데 유용합니다.

예시:

```bash
codeknit parse ./src --output-mode directory-tree
# 출력: ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt 등
```

### inline

- **동작**: 모든 출력을 stdout으로 덤프합니다.
- **출력 디렉터리**: 생성되지 않음
- **사용 사례**: 단일 파일 또는 매우 작은 프로젝트(5개 미만 파일)에만 권장됩니다. 다른 도구로 출력을 파이핑하거나 단일 파일을 대화식으로 검사할 때 유용합니다.

예시:

```bash
codeknit parse ./src/main.go --output-mode inline
# 출력: 터미널에 직접 출력됨
```

### 결정 표

| 모드             | 가장 적합한 경우                | 출력 위치                                           |
| ---------------- | ------------------------------- | --------------------------------------------------- |
| `directory-flat` | 대부분의 프로젝트(기본값, 권장) | `./skeleton/map_001.skt`, `map_002.skt`, ...        |
| `directory-tree` | 소스 코드와 함께 출력 탐색      | `./skeleton/<미러링된 경로>.skt`                    |
| `inline`         | 단일 파일, 다른 도구로 파이핑   | stdout — 단일 파일 또는 매우 작은 프로젝트에만 사용 |

### 경험칙

- **확신이 서지 않을 때** → `directory-flat`(기본값) 사용
- **단일 파일 검사** → `inline` 허용
- **몇 개 이상의 파일** → `directory-flat` 또는 `directory-tree` 선호
- **대규모 코드베이스** → 토큰 사용량을 줄이기 위해 `--minify` 추가
- **동일한 출력에 재실행** → 오래된 `.skt` 파일을 제거하기 위해 `--clean` 사용

### 축소(minification)

`--minify` 플래그는 반복되는 토큰(예: `exported`, `async`와 같은 속성 키 또는 일반적인 타입 이름)의 사전 기반 압축을 활성화합니다. 활성화되면:

- 반복되는 값은 짧은 코드(`d0`, `d1`, `d2`, ...)로 대체됩니다.
- 출력 디렉터리에 코드를 원래 값에 매핑하는 `dict.skt` 파일이 작성됩니다.
- 대규모 코드베이스의 출력 크기를 크게 줄입니다.
- `directory-flat` 및 `directory-tree` 모드에서 모두 작동합니다.

예시 축소 출력:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```

이 형식은 전체 정보를 보존하면서 토큰 사용량을 최소화하여 LLM 기반 분석에 이상적입니다.
