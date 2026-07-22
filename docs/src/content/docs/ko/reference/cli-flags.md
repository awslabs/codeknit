---
title: CLI 참조
description: 모든 codeknit 명령어와 플래그에 대한 완전한 참조입니다.
---

## codeknit

대화형 터미널 UI(TUI)를 실행하여 사용 가능한 명령어와 옵션을 안내합니다.

```bash
codeknit
```

## codeknit parse

소스 코드에서 구조적 정보를 추출하여 `.skt` 파일 또는 JSON으로 저장합니다.

```bash
codeknit parse <input-path> [output-dir]
```

| 플래그             | 유형    | 기본값           | 설명                                                                                     |
| ---------------- | ------ | ---------------- | -------------------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat` | 출력 모드: `inline`, `directory-flat` 또는 `directory-tree`                              |
| `--format`       | string | `skt`            | 출력 형식: `skt` 또는 `json`                                                           |
| `--max-lines`    | int    | `500`            | 출력 파일당 최대 행 수(`directory-flat` 및 `directory-tree` 모드에 적용)                     |
| `--collect-test` | bool   | `false`          | 분석에 테스트 파일 포함                                                                   |
| `--minify`       | bool   | `false`          | 사전 기반 출력 최소화 활성화                                                              |
| `--edges`        | bool   | `false`          | 출력에 `[edges]` 섹션 포함(토큰 절약을 위해 기본값은 비활성화)                                |
| `--clean`        | bool   | `false`          | 쓰기 전에 출력 디렉터리에서 오래된 `.skt` 파일 제거                                          |
| `--workers`      | int    | `0` (NumCPU)     | 최대 동시 파싱 고루틴 수                                                                |
| `--verbose`      | bool   | `false`          | 처리 중 진행 정보 출력                                                                   |

출력 디렉터리는 지정하지 않으면 `./skeleton`이 기본값입니다. `inline` 모드에서는 출력이 stdout으로 기록되며 디렉터리는 사용되지 않습니다. `--format json`을 사용하면 디렉터리 출력이 `codeknit.json`으로 저장됩니다.

## codeknit graph show

코드베이스 구조의 대화형 HTML 그래프 시각화를 생성합니다.

```bash
codeknit graph show <input-path>
```

| 플래그             | 유형    | 기본값                          | 설명                                  |
| ---------------- | ------ | -------------------------------- | ------------------------------------ |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | 출력 HTML 파일 경로                   |
| `--collect-test` | bool   | `false`                          | 분석에 테스트 파일 포함               |
| `--workers`      | int    | `0` (NumCPU)                     | 최대 동시 파싱 고루틴 수             |
| `--verbose`      | bool   | `false`                          | 처리 중 진행 정보 출력               |

생성된 HTML 파일은 독립 실행형이며 기본 브라우저에서 자동으로 열립니다.

## codeknit graph analyze

구조 분석 알고리즘을 실행하고 LLM이 읽을 수 있는 `.skt` 보고서를 생성합니다.

```bash
codeknit graph analyze <input-path>
```

| 플래그                      | 유형     | 기본값                         | 설명                                                   |
| ------------------------- | ------- | ------------------------------- | ------------------------------------------------------ |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | 출력 `.skt` 파일 경로                                  |
| `--collect-test`          | bool    | `false`                         | 분석에 테스트 파일 포함                                |
| `--workers`               | int     | `0` (NumCPU)                    | 최대 동시 파싱 고루틴 수                               |
| `--verbose`               | bool    | `false`                         | 처리 중 진행 정보 출력                                 |
| `--fan-threshold`         | int     | `10`                            | 허브 심볼로 플래그 지정할 최소 팬인 또는 팬아웃 값        |
| `--god-threshold`         | int     | `15`                            | god class/function으로 플래그 지정할 최소 contains-엣지 수 |
| `--max-inheritance-depth` | int     | `5`                             | 이 값보다 깊은 상속 체인 플래그 지정                   |
| `--top-n`                 | int     | `30`                            | 순위 출력 섹션 제한; `0`은 제한 없음                   |
| `--betweenness-threshold` | float64 | `0.001`                         | 보고할 최소 betweenness 중심성 값                      |
| `--propagation-cutoff`    | float64 | `0.05`                          | 변경 전파 시뮬레이션을 계속할 최소 확률                |

## codeknit graph hotspots

Git 기록과 구조적 중요성을 사용하여 파일을 순위화하고, 반복적으로 함께 변경되는 파일의 시간적 결합을 보고합니다.

```bash
codeknit graph hotspots <input-path>
```

| 플래그                     | 유형    | 기본값                   | 설명                                      |
| ------------------------ | ------ | ------------------------- | ----------------------------------------- |
| `-o`, `--output`         | string | `./skeleton/hotspots.skt` | 출력 파일 경로                            |
| `--format`               | string | `skt`                     | 출력 형식: `skt` 또는 `json`              |
| `--since`                | string | `12mo`                    | 기록 기간(예: `180d`, `12mo`, `2y`)      |
| `--max-commits`          | int    | `2000`                    | 검사할 최대 커밋 수                       |
| `--max-files-per-commit` | int    | `50`                      | 더 많은 파일을 변경하는 커밋 제외         |
| `--min-cochanges`        | int    | `3`                       | 시간적 결합을 위한 최소 공유 커밋 수      |
| `--top-n`                | int    | `30`                      | 보고서 섹션별 최대 결과 수                |
| `--include-merges`       | bool   | `false`                   | 병합 커밋 포함                            |
| `--collect-test`         | bool   | `false`                   | 테스트 파일 포함                          |
| `--workers`              | int    | `0` (NumCPU)              | 최대 동시 파싱 고루틴 수                  |
| `--verbose`              | bool   | `false`                   | 진행 정보 출력                            |

## codeknit fingerprint

퍼지 해싱을 사용하여 중복 및 근사 중복 코드를 감지합니다.

```bash
codeknit fingerprint <input-path>
```

| 플래그               | 유형    | 기본값                       | 설명                                                                                                                  |
| ------------------ | ------ | ----------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | 출력 `.skt` 파일 경로                                                                                                 |
| `--min-similarity` | int    | `65`                          | 보고할 최소 유사도 비율(0–100)                                                                                       |
| `--max-similarity` | int    | `95`                          | 보고할 최대 유사도 비율(0–100)                                                                                       |
| `--show-all`       | bool   | `false`                       | 원시 토큰 데이터를 포함하는 `[fingerprints]` 섹션 포함                                                                |
| `--rerank`         | bool   | `false`                       | 의미적 이웃을 찾고 Ollama 임베딩을 사용하여 후보를 재순위화(`ollama serve` 및 `ollama pull qwen3-embedding:0.6b` 필요) |
| `--model`          | string | `qwen3-embedding:0.6b`        | `--rerank`와 함께 사용할 Ollama 임베딩 모델                                                                          |
| `--collect-test`   | bool   | `false`                       | 분석에 테스트 파일 포함                                                                                               |
| `--workers`        | int    | `0` (NumCPU)                  | 최대 동시 파싱 고루틴 수                                                                                              |
| `--verbose`        | bool   | `false`                       | 처리 중 진행 정보 출력                                                                                                |

## codeknit completion

지원되는 셸에 대한 셸 완성 스크립트를 생성합니다.

```bash
codeknit completion <shell>
```

지원되는 셸: `bash`, `zsh`, `fish`, `powershell`.

## 전역 플래그

| 플래그           | 설명                       |
| -------------- | ------------------------- |
| `--version`    | 버전 정보 출력            |
| `--help`, `-h` | 현재 명령어에 대한 도움말 표시 |