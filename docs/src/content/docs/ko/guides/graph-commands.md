---
title: 그래프 명령어
description: 그래프 알고리즘을 사용하여 코드베이스 구조를 시각화하고 분석합니다.
---

codeknit는 코드베이스 구조를 이해하고 개선하는 데 도움이 되는 두 가지 강력한 그래프 명령어를 제공합니다: 대화형 시각화를 위한 `graph show`와 자동화된 구조 분석을 위한 `graph analyze`입니다.

## graph show

코드베이스의 대화형 HTML 그래프 시각화를 생성합니다.

```bash
codeknit graph show <input-path>
```

이 명령어는 코드베이스를 파싱하고 대화형 그래프 시각화가 포함된 독립 실행형 HTML 파일을 생성합니다. 심볼(함수, 클래스, 타입)은 노드로, 그들의 관계(호출, 포함, 구현)는 엣지로 나타납니다. 시각화는 기본 브라우저에서 자동으로 열립니다.

### 플래그

| 플래그           | 기본값                           | 설명                     |
| ---------------- | -------------------------------- | ------------------------ |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | 출력 HTML 파일 경로      |
| `--collect-test` | `false`                          | 분석에 테스트 파일 포함  |
| `--workers`      | `NumCPU`                         | 최대 동시 파싱 고루틴 수 |
| `--verbose`      | `false`                          | 처리 중 진행 정보 출력   |

### 예제

```skt
# 기본 시각화 생성
codeknit graph show ./myproject

# 사용자 정의 출력 파일
codeknit graph show ./myproject -o graph.html

# 테스트 파일 포함
codeknit graph show ./src --collect-test
```

## graph analyze

코드베이스에 구조적 그래프 알고리즘을 실행하고 코드 품질 인사이트가 포함된 LLM이 읽을 수 있는 `.skt` 보고서를 생성합니다.

```bash
codeknit graph analyze <input-path>
```

이 명령어는 순환 의존성, 허브 심볼, 데드 코드, god 클래스, 아키텍처 병목 등 일반적인 코드 품질 문제를 감지합니다.

### 알고리즘

분석에는 17개의 구조적 그래프 알고리즘이 포함됩니다:

- 순환 의존성 (Tarjan의 SCC)
- 허브 감지 (높은 팬인/팬아웃 결합)
- 고아 감지 (데드 코드 후보)
- god 클래스/함수 감지 (과도한 자식)
- 불안정성 지표 (Robert C. Martin의 Ce/(Ca+Ce))
- 깊은 상속 체인
- 매개 중심성 (병목 감지)
- 단절점 (단일 실패 지점)
- PageRank (재귀적 중요도)
- 전이 팬인 (영향 범위)
- 변경 전파 시뮬레이션
- 순환 패키지 의존성
- 레이어 위반 감지
- 진입점에서의 도달 가능성
- 약하게 연결된 구성 요소
- 의존성 가중치 (패키지 결합 강도)
- 메인 시퀀스로부터의 거리 (A+I 균형)

### 플래그

| 플래그                    | 기본값                          | 설명                                                  |
| ------------------------- | ------------------------------- | ----------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | 출력 `.skt` 파일 경로                                 |
| `--collect-test`          | `false`                         | 분석에 테스트 파일 포함                               |
| `--workers`               | `NumCPU`                        | 최대 동시 파싱 고루틴 수                              |
| `--verbose`               | `false`                         | 처리 중 진행 정보 출력                                |
| `--fan-threshold`         | `10`                            | 허브 심볼로 플래그 지정할 최소 팬인 또는 팬아웃       |
| `--god-threshold`         | `15`                            | god 클래스/함수로 플래그 지정할 최소 contains-엣지 수 |
| `--max-inheritance-depth` | `5`                             | 이보다 깊은 상속 체인 플래그 지정                     |
| `--top-n`                 | `30`                            | 순위 출력 섹션 제한; 0 = 제한 없음                    |
| `--betweenness-threshold` | `0.001`                         | 보고할 최소 매개 중심성 값                            |
| `--propagation-cutoff`    | `0.05`                          | 변경 전파를 계속할 최소 확률                          |

### 예제

```skt
# 기본값으로 구조 분석 실행
codeknit graph analyze ./myproject

# 사용자 정의 출력 및 임계값
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# 섹션별 더 많은 결과 표시
codeknit graph analyze ./myproject --top-n 50

# 테스트 파일 포함
codeknit graph analyze ./src --collect-test
```
