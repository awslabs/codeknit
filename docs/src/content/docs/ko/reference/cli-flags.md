---
title: CLI 참조
description: 모든 codeknit 명령어와 플래그의 완전한 참조입니다.
---

이 페이지는 모든 codeknit 명령어, 인수 및 플래그를 나열합니다. 작업 지향적인 지침은 [parse 명령 가이드](../../guides/parse-command/), [graph 명령 가이드](../../guides/graph-commands/), [fingerprint 명령 가이드](../../guides/fingerprint-command/)를 참조하세요.

## 명령 요약

| 명령                                       | 목적                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | 대화형 터미널 UI를 실행합니다.                            |
| `codeknit parse <input-path> [output-dir]`    | 심볼과 관계를 `.skt` 또는 JSON 출력으로 추출합니다.  |
| `codeknit graph show <input-path>`            | 대화형 HTML 그래프를 생성합니다.                            |
| `codeknit graph analyze <input-path>`         | 구조적 그래프 알고리즘을 실행하고 `.skt` 보고서를 작성합니다.     |
| `codeknit graph hotspots <input-path>`        | Git 기록과 그래프 구조를 사용하여 변경 핫스팟을 순위화합니다.    |
| `codeknit fingerprint <input-path>`           | 중복 및 근사 중복 코드를 감지합니다.                      |
| `codeknit completion <shell>`                 | 셸 완성 스크립트를 생성합니다.                            |

다섯 개의 명령어(`parse`, `graph show`, `graph analyze`, `graph hotspots`, `fingerprint`)은 소스 코드를 처리합니다. 이들은 [공유 소스 처리 플래그](#공유-소스-처리-플래그)에 설명된 플래그를 공유합니다.

## 전역 플래그

| 플래그              | 타입 | 기본값 | 설명                  |
| ----------------- | ---- | ------- | ---------------------------- |
| `--version`       | bool | `false` | 버전 정보를 출력합니다.   |
| `-h`, `--help`    | bool | `false` | 명령어에 대한 도움말을 표시합니다.   |

## 공유 소스 처리 플래그

이 플래그는 각 소스 처리 명령어 테이블에 나타납니다.

| 플래그                    | 타입                | 기본값               | 허용 값       | 설명                                                                          |
| ----------------------- | ------------------- | --------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                   | `c`, `cpp`           | `.h` 파일에 사용되는 파서입니다.                                                          |
| `--include GLOB`        | string, repeatable  | 모든 지원 소스       | Gitignore 스타일 glob | 일치하는 소스 경로를 포함합니다. 여러 패턴은 합집합을 형성합니다.                       |
| `--exclude GLOB`        | string, repeatable  | 없음                  | Gitignore 스타일 glob | 일치하는 경로 또는 디렉터리를 제외합니다.                                               |
| `--max-file-size BYTES` | int                 | `10485760` (10 MiB)   | `0` 이상             | 이 크기보다 큰 소스 파일을 건너뜁니다. `0`은 제한을 해제합니다.                     |
| `--strict`              | bool                | `false`               | —                    | 스캔 또는 파싱 진단이 발생하면 출력을 작성하기 전에 실패합니다.                    |

### `--header-language`

`.h` 파일에만 적용됩니다. 다른 확장자는 일반적인 파서(`.hpp` 및 `.hxx`는 이미 C++를 사용)를 유지합니다.

- `c` (기본값)는 C 매크로, typedef 및 union 추출을 보존합니다.
- `cpp`는 C++ 파서를 선택합니다. 헤더 이름이 `.h`인 C++ 프로젝트에 사용하세요.

TUI는 `.h` 파일에 대해 기본값인 `c`를 사용합니다.

```bash
codeknit parse ./src --header-language cpp
```

테스트된 추출 기능 및 언어별 알려진 제한 사항은 [언어 지원 및 API 충실도](../language-support/)를 참조하세요. 해당 페이지는 완전한 언어 또는 컴파일러 의미론적 범위가 아닌 테스트된 구성만을 설명합니다.

### 소스 선택

- 패턴은 입력 디렉터리를 기준으로 슬래시로 구분된 gitignore 스타일 glob을 사용합니다. 단일 파일 입력의 경우 패턴은 파일의 기본 이름과 일치합니다.
- 패턴을 따옴표로 묶어 셸이 확장하지 않도록 합니다.
- 제외는 포함보다 우선합니다.
- 포함은 `.gitignore` 규칙, 테스트 파일 필터링 또는 내장 제외 규칙을 무시하지 않습니다.
- 부정 패턴(`!pattern`)은 `--include` 또는 `--exclude`에서 허용되지 않습니다.
- 디렉터리 스캔은 현재 작업 디렉터리와 무관하게 상위 및 하위 `.gitignore` 파일을 준수합니다.
- 명시적으로 일반 소스 파일을 선택하면 무시 및 테스트 필터링을 우회합니다. 포함, 제외 및 크기 제한은 여전히 적용됩니다.
- 테스트 파일은 `--collect-test`가 설정되지 않는 한 제외됩니다.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### 진단 및 `--strict`

- 심볼릭 링크, 디렉터리 심볼릭 링크 및 비정규 파일은 건너뛰고 보고됩니다.
- 읽을 수 없는 경로, 무시 파일 읽기 실패, 과도하게 큰 소스 및 구문 경고는 분석을 불완전하게 만듭니다.
- 기본적으로 사용 가능한 부분 결과는 유지됩니다. 진단은 stderr에 표시됩니다. `parse`는 JSON `errors` 또는 SKT `[errors]`에 기록합니다.
- `--strict`를 사용하면 불완전한 분석은 출력이 작성되기 전에 거부됩니다.
- 선택된 지원 소스가 없거나 모든 소스 읽기가 실패한 실행은 항상 실패로 종료됩니다.
- TUI는 동일한 심볼릭 링크 정책과 10 MiB 기본값을 적용합니다. TUI의 양식은 선택 및 strict 플래그를 노출하지 않습니다.

정확한 보장은 [분석 신뢰성](../analysis-reliability/)을 참조하세요.

## `codeknit`

```bash
codeknit
```

하위 명령 없이 `codeknit`을 실행하면 대화형 터미널 UI가 실행됩니다. UI는 `parse`, `graph show`, `graph analyze`, `graph hotspots`, `fingerprint` 명령어와 해당 옵션을 제공합니다.

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| 인수           | 필수 | 기본값      | 설명                                                              |
| -------------- | -------- | ------------ | ------------------------------------------------------------------------ |
| `input-path`   | 예      | —            | 소스 디렉터리 또는 단일 소스 파일.                                  |
| `output-dir`   | 아니요      | `./skeleton` | 출력 디렉터리. `--output-mode`가 `inline`일 때는 사용되지 않습니다.             |

| 플래그                    | 타입                | 기본값            | 허용 값                                   | 설명                                                                                         |
| ----------------------- | ------------------- | ------------------ | ------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`   | `inline`, `directory-flat`, `directory-tree`     | 출력 위치 및 방식입니다.                                                                    |
| `--format`              | string              | `skt`              | `skt`, `json`                                    | 출력 형식입니다.                                                                                      |
| `--max-lines`           | int                 | `500`              | `directory-tree`의 경우 3 이상               | 플랫 출력의 행 목표입니다. 트리 출력의 경우 하드 소스 청크 제한입니다.                               |
| `--collect-test`        | bool                | `false`            | —                                                | 테스트 파일을 포함합니다.                                                                                 |
| `--minify`              | bool                | `false`            | —                                                | 사전 기반 출력 압축을 활성화합니다.                                                         |
| `--edges`               | bool                | `false`            | —                                                | `[edges]` 섹션을 포함합니다. 기본적으로 토큰을 절약하기 위해 꺼져 있습니다.                                       |
| `--clean`               | bool                | `false`            | —                                                | 이전 생성된 출력을 성공적으로 생성한 후에만 교체합니다.                                   |
| `--workers`             | int                 | `0` (`NumCPU`)     | `0` 이상                                   | 최대 동시 파싱 고루틴 수입니다.                                                              |
| `--verbose`             | bool                | `false`            | —                                                | 진행 정보를 출력합니다.                                                                         |
| `--header-language`     | string              | `c`                | `c`, `cpp`                                       | `.h` 파일에 사용되는 파서입니다.                                                                         |
| `--include GLOB`        | string, repeatable  | 모든 지원 소스 | —                                             | 일치하는 소스 경로를 포함합니다.                                                                      |
| `--exclude GLOB`        | string, repeatable  | 없음               | —                                                | 일치하는 경로 또는 디렉터리를 제외합니다.                                                              |
| `--max-file-size BYTES` | int                 | `10485760`         | `0`은 제한을 해제함                           | 더 큰 소스 파일을 건너뜁니다.                                                                           |
| `--strict`              | bool                | `false`            | —                                                | 스캔 또는 파싱 진단이 발생하면 출력을 작성하기 전에 실패합니다.                                              |

### 출력 모드

| 값             | 동작                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| `inline`          | stdout에 작성합니다. 출력 디렉터리는 사용되지 않습니다.                                                    |
| `directory-flat`  | 출력 디렉터리에 청크 파일로 작성합니다. 소스 블록은 그대로 유지되므로 `--max-lines`는 하드 제한이 아닌 분할 목표입니다. |
| `directory-tree`  | 소스 트리를 미러링합니다. 이름은 소스 확장자를 유지합니다(`foo.c.skt`, `foo.h.skt`). 분할된 파일은 `foo.c_part1.skt`를 사용합니다. |

### 출력 동작

- `--clean` 없이 기존 생성된 출력 형식이 출력 디렉터리에 있으면 오류가 발생합니다.
- `--clean`을 사용하면 SKT 형식의 오래된 `.skt` 파일 또는 JSON 형식의 `codeknit.json`은 성공적으로 생성된 후에만 교체됩니다. 관련 없는 항목은 보존됩니다.
- JSON 형식에서 디렉터리 모드는 `codeknit.json`을 작성합니다. JSON은 단일 문서이며 행 수로 분할되지 않습니다.
- `dict.skt` 및 `warnings.skt`는 예약된 보조 파일 이름입니다.
- 소스 디렉터리 또는 그 상위 디렉터리를 출력 디렉터리로 사용하지 마세요.

자세한 내용은 [출력 모드](../../guides/output-modes/) 및 [출력 형식 참조](../output-format/)를 참조하세요.

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| 인수           | 필수 | 설명                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | 예      | 소스 디렉터리 또는 단일 소스 파일.  |

| 플래그                    | 타입                | 기본값                            | 허용 값          | 설명                                                       |
| ----------------------- | ------------------- | ---------------------------------- | ----------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`   | 파일 경로               | 출력 HTML 파일입니다.                                                 |
| `--collect-test`        | bool                | `false`                            | —                       | 테스트 파일을 포함합니다.                                               |
| `--workers`             | int                 | `0` (`NumCPU`)                     | `0` 이상          | 최대 동시 파싱 고루틴 수입니다.                            |
| `--verbose`             | bool                | `false`                            | —                       | 진행 정보를 출력합니다.                                       |
| `--header-language`     | string              | `c`                                | `c`, `cpp`              | `.h` 파일에 사용되는 파서입니다.                                       |
| `--include GLOB`        | string, repeatable  | 모든 지원 소스              | —                       | 일치하는 소스 경로를 포함합니다.                                    |
| `--exclude GLOB`        | string, repeatable  | 없음                               | —                       | 일치하는 경로 또는 디렉터리를 제외합니다.                            |
| `--max-file-size BYTES` | int                 | `10485760`                         | `0`은 제한을 해제함  | 더 큰 소스 파일을 건너뜁니다.                                         |
| `--strict`              | bool                | `false`                            | —                       | 스캔 또는 파싱 진단이 발생하면 출력을 작성하기 전에 실패합니다.            |

이 명령은 독립 실행형 HTML 파일을 작성하고 기본 브라우저에서 엽니다. 뷰어는 해결된 관계를 그리고 불확실한 관계의 수를 표시합니다.

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| 인수           | 필수 | 설명                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | 예      | 소스 디렉터리 또는 단일 소스 파일.  |

| 플래그                       | 타입                | 기본값                            | 허용 값          | 설명                                                  |
| -------------------------- | ------------------- | ---------------------------------- | ----------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`    | 파일 경로               | 출력 `.skt` 파일입니다.                                          |
| `--collect-test`           | bool                | `false`                            | —                       | 테스트 파일을 포함합니다.                                          |
| `--workers`                | int                 | `0` (`NumCPU`)                     | `0` 이상          | 최대 동시 파싱 고루틴 수입니다.                       |
| `--verbose`                | bool                | `false`                            | —                       | 진행 정보를 출력합니다.                                 |
| `--fan-threshold`          | int                 | `10`                               | —                       | 허브 심볼로 플래그 지정할 최소 팬인 또는 팬아웃입니다.              |
| `--god-threshold`          | int                 | `15`                               | —                       | god class 또는 함수로 플래그 지정할 최소 `contains` 엣지 수입니다. |
| `--max-inheritance-depth`  | int                 | `5`                                | —                       | 이 값보다 깊은 상속 체인을 플래그 지정합니다.              |
| `--top-n`                  | int                 | `30`                               | `0` = 제한 없음          | 순위 출력 섹션을 제한합니다.                                  |
| `--betweenness-threshold`  | float               | `0.001`                            | —                       | 보고할 최소 매개 중심성 값입니다.              |
| `--propagation-cutoff`     | float               | `0.05`                             | —                       | 변경 전파를 계속할 최소 확률입니다.          |
| `--header-language`        | string              | `c`                                | `c`, `cpp`              | `.h` 파일에 사용되는 파서입니다.                                  |
| `--include GLOB`           | string, repeatable  | 모든 지원 소스              | —                       | 일치하는 소스 경로를 포함합니다.                               |
| `--exclude GLOB`           | string, repeatable  | 없음                               | —                       | 일치하는 경로 또는 디렉터리를 제외합니다.                       |
| `--max-file-size BYTES`    | int                 | `10485760`                         | `0`은 제한을 해제함  | 더 큰 소스 파일을 건너뜁니다.                                    |
| `--strict`                 | bool                | `false`                            | —                       | 스캔 또는 파싱 진단이 발생하면 출력을 작성하기 전에 실패합니다.       |

의존성 메트릭은 불확실한 관계 및 파일/포함 메타데이터를 제외합니다. 보고서는 제외된 관계 수를 기록합니다.

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| 인수           | 필수 | 설명                                                  |
| -------------- | -------- | ------------------------------------------------------------ |
| `input-path`   | 예      | Git 저장소 내부의 소스 디렉터리입니다.                    |

| 플래그                       | 타입                | 기본값                    | 허용 값                              | 설명                                                  |
| -------------------------- | ------------------- | -------------------------- | ------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`  | 파일 경로                                   | 출력 파일입니다.                                                 |
| `--format`                 | string              | `skt`                      | `skt`, `json`                               | 출력 형식입니다.                                               |
| `--since`                  | string              | `12mo`                     | `180d`, `12mo`, `2y`와 같은 기간      | 기록 기간입니다.                                              |
| `--max-commits`            | int                 | `2000`                     | —                                           | 검사할 최대 커밋 수입니다.                                  |
| `--max-files-per-commit`   | int                 | `50`                       | —                                           | 이 값보다 많은 파일을 건드리는 커밋을 제외합니다.       |
| `--min-cochanges`          | int                 | `3`                        | —                                           | 보고할 시간적 결합의 최소 공유 커밋 수입니다.          |
| `--top-n`                  | int                 | `30`                       | —                                           | 섹션별 최대 결과 수입니다.                                 |
| `--include-merges`         | bool                | `false`                    | —                                           | 기록 메트릭에 병합 커밋을 포함합니다.                    |
| `--collect-test`           | bool                | `false`                    | —                                           | 테스트 파일을 포함합니다.                                          |
| `--workers`                | int                 | `0` (`NumCPU`)             | `0` 이상                              | 최대 동시 파싱 고루틴 수입니다.                       |
| `--verbose`                | bool                | `false`                    | —                                           | 진행 정보를 출력합니다.                                 |
| `--header-language`        | string              | `c`                        | `c`, `cpp`                                  | `.h` 파일에 사용되는 파서입니다.                                  |
| `--include GLOB`           | string, repeatable  | 모든 지원 소스      | —                                           | 일치하는 소스 경로를 포함합니다.                               |
| `--exclude GLOB`           | string, repeatable  | 없음                       | —                                           | 일치하는 경로 또는 디렉터리를 제외합니다.                       |
| `--max-file-size BYTES`    | int                 | `10485760`                 | `0`은 제한을 해제함                      | 더 큰 소스 파일을 건너뜁니다.                                    |
| `--strict`                 | bool                | `false`                    | —                                           | 스캔 또는 파싱 진단이 발생하면 출력을 작성하기 전에 실패합니다.       |

이 명령은 커밋 빈도, 라인 변경량 및 최근성을 파일 수준 PageRank, 전이 팬인 및 매개 중심성과 결합합니다. 함께 변경되는 파일 간의 시간적 결합도 보고합니다. 병합 커밋은 기본적으로 제외됩니다.

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| 인수           | 필수 | 설명                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | 예      | 소스 디렉터리 또는 단일 소스 파일.  |

| 플래그                    | 타입                | 기본값                          | 허용 값          | 설명                                                                                                      |
| ----------------------- | ------------------- | -------------------------------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`    | 파일 경로               | 출력 `.skt` 파일입니다.                                                                                              |
| `--min-similarity`      | int                 | `65`                             | `0`–`100`               | 보고할 최소 유사도 비율입니다.                                                                         |
| `--max-similarity`      | int                 | `95`                             | `0`–`100`               | 보고할 최대 유사도 비율입니다.                                                                         |
| `--show-all`            | bool                | `false`                          | —                       | 원시 토큰 데이터를 포함하는 `[fingerprints]` 섹션도 출력합니다.                                                      |
| `--rerank`              | bool                | `false`                          | —                       | 의미론적 이웃을 찾고 Ollama 임베딩으로 후보를 재순위화합니다. `ollama serve` 및 `ollama pull qwen3-embedding:0.6b`가 필요합니다. |
| `--model`               | string              | `qwen3-embedding:0.6b`           | Ollama 모델 이름       | `--rerank`와 함께 사용할 Ollama 임베딩 모델입니다.                                                                   |
| `--collect-test`        | bool                | `false`                          | —                       | 테스트 파일을 포함합니다.                                                                                              |
| `--workers`             | int                 | `0` (`NumCPU`)                   | `0` 이상          | 최대 동시 파싱 고루틴 수입니다.                                                                           |
| `--verbose`             | bool                | `false`                          | —                       | 진행 정보를 출력합니다.                                                                                      |
| `--header-language`     | string              | `c`                              | `c`, `cpp`              | `.h` 파일에 사용되는 파서입니다.                                                                                      |
| `--include GLOB`        | string, repeatable  | 모든 지원 소스            | —                       | 일치하는 소스 경로를 포함합니다.                                                                                   |
| `--exclude GLOB`        | string, repeatable  | 없음                             | —                       | 일치하는 경로 또는 디렉터리를 제외합니다.                                                                           |
| `--max-file-size BYTES` | int                 | `10485760`                       | `0`은 제한을 해제함  | 더 큰 소스 파일을 건너뜁니다.                                                                                        |
| `--strict`              | bool                | `false`                          | —                       | 스캔 또는 파싱 진단이 발생하면 출력을 작성하기 전에 실패합니다.                                                           |

### 유사도 범위

기본 범위인 65–95%는 100% 일치를 제외합니다. 정확한 정규화 토큰 일치를 보고하려면 두 경계를 모두 `100`으로 설정하세요:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

정확한 정규화 토큰 후보는 완전합니다. 500개 이상의 적합한 심볼이 있는 경우, 다른 근사 중복 후보는 근사 상위 50개 이웃 검색을 사용합니다. `--rerank`은 구조적으로 정확한 후보를 제거할 수 있습니다. 자세한 내용은 [분석 신뢰성](../analysis-reliability/#duplicate-detection-guarantees)을 참조하세요.

## `codeknit completion`

```bash
codeknit completion <shell>
```

| 인수    | 필수 | 허용 값                         | 설명                                   |
| --------- | -------- | -------------------------------------- | --------------------------------------------- |
| `shell`   | 예      | `bash`, `zsh`, `fish`, `powershell`    | 완성 스크립트를 생성할 셸입니다. |

스크립트는 stdout에 작성됩니다.

```bash
codeknit completion zsh >> ~/.zshrc
```

## 관련 페이지

- [출력 형식 참조](../output-format/)
- [언어 지원 및 API 충실도](../language-support/)
- [분석 신뢰성](../analysis-reliability/)