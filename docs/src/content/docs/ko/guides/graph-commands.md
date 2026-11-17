---
title: 그래프 명령어
description: 그래프 알고리즘을 사용하여 코드베이스 구조를 시각화하고 분석하세요.
---

`graph` 하위 명령어를 사용하여 코드가 어떻게 연결되어 있는지 확인하고, 구조적 문제를 찾으며, 어떤 파일을 먼저 리팩터링할지 결정하세요. 이 가이드는 [codeknit을 설치](../../installation/)했으며 [`codeknit parse`](../parse-command/)를 실행할 수 있다고 가정합니다.

| 작업 | 명령어 |
| --- | --- |
| 그래프 시각적으로 탐색 | `codeknit graph show` |
| 구조적 문제 찾기 | `codeknit graph analyze` |
| 변동성이 크고 중요한 파일 찾기 | `codeknit graph hotspots` |

## 코드 그래프 시각화

`graph show`는 입력을 파싱하고 독립 실행형 HTML 파일을 작성합니다. 심볼(함수, 클래스, 타입)은 노드입니다. 관계(호출, 포함, 구현)는 엣지입니다.

```bash
codeknit graph show ./myproject
```

파일은 `./skeleton/codeknit-graph.html`로 작성되며 기본 브라우저에서 열립니다.

출력 경로를 선택하거나 테스트 파일을 포함하려면:

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| 플래그 | 기본값 | 목적 |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | 출력 HTML 파일 경로 |
| `--collect-test` | `false` | 테스트 파일 포함 |
| `--workers` | `0` (`NumCPU`) | 최대 동시 파싱 고루틴 수 |
| `--verbose` | `false` | 진행 정보 출력 |

### 미해결 관계 검사

뷰어는 해결된 관계만 그립니다. 불확실한(`unresolved` 또는 `ambiguous`) 관계의 개수는 표시하지만, 이를 그리거나 후보 링크를 표시하지 않습니다.

개별 불확실한 엣지를 보려면 `--edges`와 함께 `parse`를 실행하세요:

```bash
codeknit parse ./src --output-mode inline --edges
```

JSON 출력을 원하면 `--format json`을 추가하세요. 엣지와 후보가 어떻게 작성되는지는 [출력 형식 참조](../../reference/output-format/)를 확인하세요.

## 구조 분석 및 문제 찾기

`graph analyze`는 22개의 그래프 알고리즘을 실행하고 LLM이 읽을 수 있는 `.skt` 보고서를 작성합니다.

```bash
codeknit graph analyze ./myproject
```

보고서는 `./skeleton/graph_analysis.skt`로 작성됩니다. 직접 읽거나 관련 소스와 함께 LLM에 제공하세요.

### 보고서 조정

임계값을 높이면 발견 항목이 줄어들고, 낮추면 더 많은 항목이 표시됩니다:

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0`은 순위가 매겨진 섹션의 상한을 제거합니다.

| 플래그 | 기본값 | 목적 |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | 출력 `.skt` 파일 경로 |
| `--collect-test` | `false` | 테스트 파일 포함 |
| `--workers` | `0` (`NumCPU`) | 최대 동시 파싱 고루틴 수 |
| `--verbose` | `false` | 진행 정보 출력 |
| `--fan-threshold` | `10` | 허브 심볼로 플래그 지정할 최소 팬인 또는 팬아웃 |
| `--god-threshold` | `15` | god class 또는 함수로 플래그 지정할 최소 포함 엣지 수 |
| `--max-inheritance-depth` | `5` | 이보다 깊은 상속 체인을 플래그 지정 |
| `--top-n` | `30` | 순위가 매겨진 섹션의 상한; `0`은 무제한을 의미 |
| `--betweenness-threshold` | `0.001` | 보고할 최소 매개 중심성 |
| `--propagation-cutoff` | `0.05` | 변경 전파를 계속하기 위한 최소 휴리스틱 점수 |

### 어떤 질문에 어떤 발견이 답하는지 알기

| 알고리즘 | 이를 사용하여 찾기 |
| --- | --- |
| 순환 의존성(Tarjan의 SCC) | 서로 의존하는 심볼 그룹 |
| 허브 탐지 | 높은 팬인 또는 팬아웃 결합도를 가진 심볼 |
| 고아 탐지 | 데드 코드 후보 |
| god class/함수 탐지 | 과도한 수의 자식을 포함하는 심볼 |
| 불안정성 지표 | 취약하거나 안정적인 심볼 및 패키지, Ce/(Ca+Ce) 사용 |
| 깊은 상속 체인 | `--max-inheritance-depth`보다 깊은 계층 구조 |
| 매개 중심성 | 많은 의존 경로가 통과하는 병목 심볼 |
| 절단점 | 제거 시 그래프를 분할하는 단일 실패 지점 |
| PageRank | 재귀적으로 중요한 심볼 |
| 추이적 팬인 | 영향 범위: 심볼에 직접 또는 간접적으로 의존하는 모든 것 |
| 변경 전파 시뮬레이션 | 심볼 변경이 얼마나 퍼질 가능성이 있는지 |
| 순환 패키지 의존성 | 패키지 수준의 의존성 루프 |
| 계층 위반 탐지 | 계층을 잘못된 방향으로 가로지르는 의존성 |
| 진입점 도달 가능성 | 진입점에서 도달하지 못하는 코드 |
| 약하게 연결된 구성 요소 | 격리된 코드 클러스터 |
| 의존성 가중치 | 패키지가 얼마나 강하게 결합되어 있는지 |
| 메인 시퀀스로부터의 거리 | 추상성과 불안정성이 균형을 이루지 못한 패키지 |
| 샷건 수술 탐지 | 패키지 내에서 많은 곳을 건드릴 변경 |
| 기능 부러움 탐지 | 자신의 패키지보다 다른 패키지를 더 많이 사용하는 심볼 |
| 안정적 의존성 위반 | 안정적인 패키지가 덜 안정적인 패키지에 의존 |
| 인터페이스 분리 원칙 위반 | 사용자가 대부분 사용하지 않는 인터페이스 멤버 |
| 포함 깊이 | 깊게 중첩된 심볼 |

### 결과를 올바르게 해석하기

이러한 규칙은 발견의 신뢰도를 결정합니다:

- **불확실한 관계는 제외됩니다.** 의존성 지표는 해결된 엣지만 사용합니다. 보고서는 제외된 `unresolved` 및 `ambiguous` 관계의 개수를 셉니다.
- **파일 및 포함 메타데이터는 무시됩니다.** `meta/file` 및 `meta/include` 심볼은 의존성 순위나 데드 코드 발견에 영향을 미치지 않습니다.
- **팬인 및 팬아웃은 고유한 이웃을 셉니다.** 포함은 의존성으로 간주되지 않습니다.
- **상속 깊이는 모든 베이스 및 인터페이스를 포함합니다.** 순환 계층 구조에는 깊이가 할당되지 않습니다.
- **도달 가능성은 `main`, `Main`, 또는 `init` callable에서 시작**되며 의존성 엣지를 따릅니다. 도달 가능한 멤버는 해당 컨테이너를 관련 있게 만들지만, 형제 멤버는 그렇지 않습니다. 스캔에서 인식된 진입점을 찾지 못하면 도달 가능성 발견은 생략됩니다.
- **콜백은 참조이지 호출이 아닙니다.** 콜백 인수 또는 반환된 callable은 콜백이 실행된다는 것을 증명하지 않습니다.
- **점수는 휴리스틱입니다.** 변경 전파는 가장 강한 의존 경로를 사용합니다. 점수와 계층 규칙은 측정된 확률이나 컴파일러 보장이 아닙니다.

### 발견 사항을 안전하게 처리하기

모든 발견 사항을 검토 후보로 취급하세요. 코드를 변경하기 전에 보고서에 명시된 심볼을 열고 다음을 확인하세요:

- 스캔된 디렉터리 외부의 외부 호출자.
- 동적 디스패치.
- 스캔하지 않은 의존성.

분석은 구문 기반이므로 이러한 요소 중 어느 하나라도 결론을 바꿀 수 있습니다. 해결 규칙 및 한계에 대해서는 [언어 지원 참조의 관계 해결](../../reference/language-support/#relationship-resolution)을 확인하세요.

## 기록을 고려한 핫스팟 찾기

`graph hotspots`는 자주 변경되고 구조적으로 중요한 위치에 있는 파일을 순위화합니다. 또한 동일한 커밋에서 반복적으로 변경되는 파일인 시간적 결합도 보고합니다. Git 저장소 내의 디렉터리에서 실행하세요.

```bash
codeknit graph hotspots ./myproject
```

보고서는 `./skeleton/hotspots.skt`로 작성됩니다.

다음 두 그룹의 지표를 결합합니다:

- **기록:** 커밋 빈도, 추가 및 삭제된 라인 변경량, 최근성.
- **구조:** 파일 수준 PageRank, 추이적 팬인, 매개 중심성.

보고서에는 분석된 관련 커밋 수를 기반으로 한 신뢰 수준이 포함됩니다. 신뢰도가 낮은 보고서의 순위는 잠정적으로 취급하세요.

### 기록 창 및 필터 변경

```bash
# 2년을 분석하고 JSON 작성
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# 더 큰 커밋 허용 및 더 강한 시간적 결합 요구
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

`--include-merges`를 전달하지 않으면 병합 커밋은 제외됩니다. `--max-files-per-commit`보다 많은 파일을 변경한 커밋은 대량 커밋으로 간주되어 건너뛰므로, 생성된 또는 기계적인 변경이 순위에 영향을 미치지 않습니다.

| 플래그 | 기본값 | 목적 |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | 출력 파일 경로 |
| `--format` | `skt` | 출력 형식: `skt` 또는 `json` |
| `--since` | `12mo` | 기록 창, 예: `180d`, `12mo`, 또는 `2y` |
| `--max-commits` | `2000` | 검사할 최대 커밋 수 |
| `--max-files-per-commit` | `50` | 이보다 많은 파일을 변경한 커밋 제외 |
| `--min-cochanges` | `3` | 시간적 결합을 보고할 최소 공유 커밋 수 |
| `--top-n` | `30` | 섹션별 최대 결과 수 |
| `--include-merges` | `false` | 기록 지표에 병합 커밋 포함 |
| `--collect-test` | `false` | 테스트 파일 포함 |
| `--workers` | `0` (`NumCPU`) | 최대 동시 파싱 고루틴 수 |
| `--verbose` | `false` | 진행 정보 출력 |

### JSON 출력 읽기

`--format json`을 사용하면 각 핫스팟에는 `score`, `history_score`, `structure_score`, `recency_score`, `pagerank`, `betweenness`, `commits`, `churn`, `transitive_fan_in`이 포함됩니다. 각 시간적 결합 항목에는 `left`, `right`, `strength`, `co_changes`가 포함됩니다. 최상위 수준에는 `commits_visited`, `commits_analyzed`, `skipped_merges`, `skipped_bulk_commits`도 기록됩니다. 이러한 개수를 사용하여 순위가 얼마나 많은 기록에 기반하는지 확인하세요.

## 스캔할 소스 제어

모든 다섯 가지 소스 처리 명령어(`parse`, `fingerprint`, `graph show`, `graph analyze`, `graph hotspots`)는 동일한 선택 플래그를 허용합니다:

| 플래그 | 기본값 | 목적 |
| --- | --- | --- |
| `--include GLOB` | 모든 지원 소스 | 일치하는 경로 포함; 반복 가능 |
| `--exclude GLOB` | 없음 | 일치하는 경로 또는 디렉터리 제외; 반복 가능. 제외가 포함보다 우선 |
| `--max-file-size BYTES` | `10485760` (10 MiB) | 더 큰 파일 건너뛰기; `0`은 제한 비활성화 |
| `--strict` | `false` | 스캔 또는 파싱이 불완전한 경우 출력 작성 전에 실패 |

패턴은 입력 디렉터리를 기준으로 하는 gitignore 스타일 글로브입니다. 셸이 확장하지 않도록 따옴표로 묶으세요. 포함은 `.gitignore` 규칙이나 테스트 필터링을 무시하지 않습니다.

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

심볼릭 링크 및 비정규 파일은 건너뛰고 보고됩니다. 스캔 및 파싱 문제는 stderr에 나타납니다. `--strict` 없이도 codeknit은 사용 가능한 부분 결과를 유지하므로, 보고서를 신뢰하기 전에 stderr을 확인하세요. `--strict`을 사용하면 불완전한 분석은 출력이 작성되기 전에 중단됩니다. 모든 소스 읽기가 실패한 실행은 항상 실패합니다.

정확한 보장은 [분석 신뢰성 참조](../../reference/analysis-reliability/)를 확인하세요.

## 관련 페이지

- [CLI 플래그 참조](../../reference/cli-flags/): 모든 명령어의 모든 플래그.
- [출력 형식 참조](../../reference/output-format/): `.skt` 및 JSON 구문.
- [언어 지원 참조](../../reference/language-support/): 추출 및 해결 한계.
- [파싱 명령어 가이드](../parse-command/): 개별 파일에 대한 심볼 및 엣지 추출.