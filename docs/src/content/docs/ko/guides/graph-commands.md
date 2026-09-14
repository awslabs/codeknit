---
title: 그래프 명령어
description: 그래프 알고리즘을 사용하여 코드베이스 구조를 시각화하고 분석합니다.
---

codeknit은 그래프 명령어를 제공하여 구조를 시각화하고, 자동화된 분석을 실행하며, 현재 의존성 그래프를 Git 변경 기록과 결합합니다.

## graph show

코드베이스의 대화형 HTML 그래프 시각화를 생성합니다.

```bash
codeknit graph show <input-path>
```

이 명령어는 코드베이스를 파싱하고 대화형 그래프 시각화가 포함된 독립 실행형 HTML 파일을 생성합니다. 심볼(함수, 클래스, 타입)은 노드로 나타나고, 이들의 관계(호출, 포함, 구현)는 엣지로 나타납니다. 시각화는 기본 브라우저에서 자동으로 열립니다.

뷰어는 해결된 관계를 그립니다. 해결되지 않거나 모호한 관계 및 후보 링크는 개별적으로 그리지 않지만 불확실성 카운트는 표시합니다. 이러한 관계를 검사하려면 파싱 출력을 사용하세요:

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### 플래그

| 플래그             | 기본값                          | 설명                                  |
| ---------------- | -------------------------------- | ------------------------------------ |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | 출력 HTML 파일 경로                   |
| `--collect-test` | `false`                          | 분석에 테스트 파일 포함               |
| `--workers`      | `NumCPU`                         | 최대 동시 파싱 고루틴 수              |
| `--verbose`      | `false`                          | 처리 중 진행 정보 출력                |

### 예제

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

코드베이스에 구조적 그래프 알고리즘을 실행하고 LLM이 읽을 수 있는 `.skt` 보고서를 생성하여 코드 품질 인사이트를 제공합니다.

```bash
codeknit graph analyze <input-path>
```

이 명령어는 순환 의존성, 허브 심볼, 데드 코드, 갓 클래스, 아키텍처 병목 등 검토 후보를 보고합니다.

### 알고리즘

분석에는 22개의 구조적 그래프 알고리즘이 포함됩니다:

- 순환 의존성 (Tarjan의 SCC)
- 허브 탐지 (높은 팬인/팬아웃 결합)
- 고아 탐지 (데드 코드 후보)
- 갓 클래스/함수 탐지 (과도한 자식)
- 불안정성 지표 (Robert C. Martin의 Ce/(Ca+Ce))
- 깊은 상속 체인
- 매개 중심성 (병목 탐지)
- 절단점 (단일 실패 지점)
- PageRank (재귀적 중요도)
- 전이 팬인 (영향 범위)
- 변경 전파 시뮬레이션
- 순환 패키지 의존성
- 계층 위반 탐지
- 진입점에서의 도달 가능성
- 약하게 연결된 구성 요소
- 의존성 가중치 (패키지 결합 강도)
- 메인 시퀀스로부터의 거리 (A+I 균형)
- 샷건 수술 탐지
- 기능 탐욕 탐지
- 안정적 의존성 위반
- 인터페이스 분리 위반
- 포함 깊이

### 보고서 해석

의존성 지표는 해결된 관계를 사용합니다. 보고서는 제외된 해결되지 않거나 모호한 관계를 카운트합니다. 파일/포함 메타데이터는 의존성 순위 및 데드 코드 결과에서 제외됩니다.

- 팬인과 팬아웃은 고유한 이웃을 카운트합니다. 반복된 관계는 이러한 카운트를 증가시키지 않으며, 포함은 의존성이 아닙니다.
- 상속 깊이는 모든 기본 클래스와 인터페이스를 고려합니다. 순환 계층 구조에는 깊이가 할당되지 않습니다.
- 도달 가능성은 `main`, `Main`, 또는 `init`으로 명명된 호출 가능 요소에서 시작하여 의존성 엣지를 따릅니다. 구성원에 도달하면 해당 컨테이너는 관련성이 있지만 모든 형제 요소가 관련성이 있는 것은 아닙니다. 인식된 진입점이 없으면 도달 가능성 결과는 생략됩니다.
- 콜백 인수와 반환된 호출 가능 요소는 참조일 뿐 콜백이 실행된다는 증거는 아닙니다.
- 변경 전파는 가장 강한 의존성 경로를 사용합니다. 점수와 디렉터리 기반 계층 규칙은 휴리스틱이며 측정된 확률이 아닙니다.

외부 호출자, 동적 디스패치 및 스캔 외부 의존성은 결론을 변경할 수 있습니다. 해결 규칙 및 한계에 대해서는 [언어 지원 및 API 충실도](../../reference/language-support/)를 참조하세요.

### 플래그

| 플래그                      | 기본값                         | 설명                                              |
| ------------------------- | ------------------------------- | ------------------------------------------------ |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | 출력 `.skt` 파일 경로                            |
| `--collect-test`          | `false`                          | 분석에 테스트 파일 포함                          |
| `--workers`               | `NumCPU`                        | 최대 동시 파싱 고루틴 수                         |
| `--verbose`               | `false`                          | 처리 중 진행 정보 출력                           |
| `--fan-threshold`         | `10`                            | 허브 심볼로 플래그 지정할 최소 팬인 또는 팬아웃   |
| `--god-threshold`         | `15`                            | 갓 클래스/함수로 플래그 지정할 최소 포함 엣지 수  |
| `--max-inheritance-depth` | `5`                             | 이보다 깊은 상속 체인 플래그 지정                |
| `--top-n`                 | `30`                            | 순위 출력 섹션 제한; 0 = 무제한                  |
| `--betweenness-threshold` | `0.001`                         | 보고할 최소 매개 중심성 값                       |
| `--propagation-cutoff`    | `0.05`                          | 변경 전파를 계속할 최소 휴리스틱 점수            |

### 예제

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

자주 변경되고 구조적으로 중요한 파일을 순위화합니다:

```bash
codeknit graph hotspots <input-path>
```

점수는 커밋 빈도, 라인 변경량 및 최신성과 파일 수준 PageRank, 전이 팬인, 매개 중심성을 결합합니다. 또한 동일한 커밋에서 반복적으로 변경되는 파일의 시간적 결합도 식별합니다.

기본적으로 병합 커밋은 제외됩니다. 50개 이상의 파일을 변경하는 커밋도 제외되어 생성, 벤더링 또는 기계적 대량 변경이 결과에 왜곡을 주지 않습니다.

### 플래그

| 플래그                     | 기본값                   | 설명                                     |
| ------------------------ | ------------------------- | --------------------------------------- |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | 출력 파일 경로                          |
| `--format`               | `skt`                     | 출력 형식: `skt` 또는 `json`            |
| `--since`                | `12mo`                    | 기록 기간, 예: `180d`, `12mo`, 또는 `2y` |
| `--max-commits`          | `2000`                    | 검사할 최대 커밋 수                     |
| `--max-files-per-commit` | `50`                      | 더 많은 파일을 변경하는 커밋 제외       |
| `--min-cochanges`        | `3`                       | 시간적 결합을 위한 최소 공유 커밋 수    |
| `--top-n`                | `30`                      | 보고서 섹션별 최대 결과 수              |
| `--include-merges`       | `false`                   | 병합 커밋 포함                           |
| `--collect-test`         | `false`                   | 테스트 파일 포함                        |
| `--workers`              | `NumCPU`                  | 최대 동시 파싱 고루틴 수                |
| `--verbose`              | `false`                   | 진행 정보 출력                          |

### 예제

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```