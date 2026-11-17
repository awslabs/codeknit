---
title: AI 어시스턴트와 함께 사용하기
description: Claude Code, Codex 및 Kiro용 codeknit 플러그인을 설치합니다.
---

Claude Code, Codex 또는 Kiro에 codeknit 플러그인을 설치하고 최신 상태로 유지한 다음, 어시스턴트에게 구조 및 중복 코드 관련 질문을 합니다. 플러그인에는 `codeknit`이라는 하나의 스킬이 포함되어 있습니다. 이 스킬은 어시스턴트에게 `codeknit parse`, `codeknit graph analyze`, `codeknit fingerprint`를 언제 실행하고 `.skt` 출력을 어떻게 읽어야 하는지 알려줍니다.

`codeknit` 바이너리를 직접 설치할 필요는 없습니다. 스킬이 처음 실행될 때 번들된 `scripts/install-codeknit.sh`를 실행하며, 이 스크립트는 다음과 같은 작업을 수행합니다:

1. `codeknit`이 이미 `PATH` 또는 `~/.local/bin`에 있는 경우 즉시 종료됩니다.
2. 그렇지 않으면 사용자의 OS 및 아키텍처에 맞는 최신 GitHub 릴리스를 다운로드합니다.
3. 릴리스의 `checksums.txt`와 다운로드를 검증합니다.
4. 바이너리를 `~/.local/bin`에 설치하고 출력 마지막 줄에 절대 경로를 표시합니다.

`~/.local/bin`이 `PATH`에 없는 경우, 어시스턴트는 세션의 나머지 부분 동안 해당 전체 경로로 codeknit을 호출합니다.

설치를 제어하려면 다음 환경 변수를 설정하세요:

| 변수 | 효과 |
| --- | --- |
| `CODEKNIT_VERSION` | 릴리스를 고정합니다(예: `0.6.0`). |
| `CODEKNIT_INSTALL_DIR` | 설치 위치를 변경합니다. |

대신 바이너리를 직접 설치하려면 [설치](../../installation/)를 참조하세요.

## 플러그인 설치

사용 중인 어시스턴트에 맞는 섹션을 선택하세요.

### Claude Code

셸에서 다음 명령을 실행하세요. Claude Code 세션 내에서 동일한 명령을 `/plugin ...`으로 실행합니다.

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

1. Powers 패널을 엽니다.
2. **Add Custom Power**를 선택한 다음 **Import power from GitHub**를 선택합니다.
3. `https://github.com/awslabs/codeknit`을 입력합니다.

### Kiro CLI

Kiro CLI는 로컬 경로에서 설치하므로 먼저 저장소를 클론합니다:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### `--sparse`에 대하여

`--sparse` 옵션은 선택 사항입니다. 이 옵션은 전체 소스 트리 대신 플러그인 매니페스트와 `skills/`로 클론을 제한합니다.

### 스킬을 수동으로 설치하기

어시스턴트가 Agent Skills를 읽지만 플러그인 지원이 없는 경우, `skills/codeknit/`을 어시스턴트의 스킬 디렉토리로 복사합니다:

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## 플러그인 업데이트

| 어시스턴트 | 명령 또는 작업 |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Powers 패널에서 **Check for updates**를 선택합니다. |
| Kiro CLI | `git -C ~/.codeknit-power pull`을 실행한 다음 `kiro-cli powers install ~/.codeknit-power`를 다시 실행합니다. |

플러그인을 업데이트해도 이미 설치된 바이너리는 업데이트되지 않습니다. 바이너리 버전을 변경하려면 위에서 설명한 대로 `CODEKNIT_VERSION`을 설정하세요.

codeknit을 업그레이드하고 추출 방식이 변경되면 심볼 ID도 변경될 수 있습니다. 서로 다른 실행에서 나온 청크를 혼합하지 말고 어시스턴트에게 모든 출력을 다시 생성하도록 요청하세요.

## 스킬이 어시스턴트에게 가르치는 것

스킬은 어시스턴트에게 다음과 같은 규칙을 제공하므로, 모든 요청에서 반복할 필요가 없습니다:

- **질문에 따라 명령을 선택합니다.**
  - `parse`는 무엇이 존재하는지, 어디에 있는지, 무엇이 무엇을 호출하거나 상속하거나 포함하는지를 답변합니다.
  - `graph analyze`는 순환 의존성, 허브, 데드 코드, 갓 클래스 및 병목을 찾습니다.
  - `fingerprint`는 복사-붙여넣기 및 리팩터링 후보를 찾습니다.
- **기본적으로 `directory-flat` 출력 모드를 사용합니다.** 어시스턴트는 청크 파일을 `./skeleton/`에 작성하고 필요한 청크만 읽습니다. `inline` 출력은 단일 파일 또는 소수의 파일에만 사용합니다. `inline`은 모든 것을 컨텍스트 창에 덤프하기 때문입니다.
- **구조적 질문에 대해서는 원시 소스가 아닌 `.skt` 출력을 읽습니다.** 어시스턴트는 출력에서 행 범위를 사용하여 구현 세부 사항을 검사하거나 변경하기 위해 소스 파일을 엽니다. 스킬의 `references/output-format.md`에 전체 형식이 있습니다. 또한 [출력 형식 참조](../../reference/output-format/)를 참조하세요.
- **유사도 범위를 선택하고 일치 항목을 신호로 처리합니다.** 지문은 의도가 아닌 구조적 형태를 비교합니다. 어시스턴트는 리팩터링을 제안하기 전에 쌍의 두 심볼을 모두 읽습니다.

## 구조 분석 요청하기

1. 어시스턴트에게 코드베이스 구조를 분석하도록 요청합니다. 예를 들어: "`./src`의 구조를 매핑하고 auth 패키지에 의존하는 모듈을 알려주세요."
2. 어시스턴트는 바이너리가 설치되어 있는지 확인한 다음 엣지가 활성화된 상태로 `codeknit parse`를 실행하고 `.skt` 출력을 읽습니다.
3. 의존성, 호출 체인 또는 데드 코드에 대해 후속 질문을 합니다. 어시스턴트는 출력에서 답변합니다.
4. 더 심층적인 분석을 위해 구조적 검토를 요청합니다. 예를 들어: "`./src`에서 순환 의존성과 갓 클래스를 찾으세요." 어시스턴트는 `codeknit graph analyze`를 실행하고 보고서를 읽습니다.

출력에서 `[unresolved]` 또는 `[ambiguous]`로 표시된 관계는 불확실하며, 확인된 의존성이 아닙니다. 이에 기반한 답변은 검증해야 할 후보로 간주하세요.

이러한 요청 뒤에 있는 명령에 대해서는 [parse 명령 가이드](../parse-command/) 및 [그래프 명령 가이드](../graph-commands/)를 참조하세요.

## 중복 감지 요청하기

1. 어시스턴트에게 중복 코드를 찾도록 요청합니다. 예를 들어: "`./src`에서 복사-붙여넣기된 함수를 찾으세요."
2. 어시스턴트는 `codeknit fingerprint`를 실행하고 `./skeleton/fingerprints.skt`의 `[duplicates]` 섹션을 읽습니다.
3. 어시스턴트는 각 플래그된 쌍의 두 심볼을 열고 리팩터링을 제안합니다.
4. 결과가 너무 많으면 유사도 범위를 조정하거나 의미적 재순위를 추가하도록 요청합니다. 재순위에는 로컬 Ollama 서버와 가져온 임베딩 모델이 필요합니다.

이러한 요청 뒤에 있는 옵션에 대해서는 [fingerprint 명령 가이드](../fingerprint-command/)를 참조하세요.

## 팁

- 구조적 질문에 대해서는 원시 소스가 아닌 `.skt` 파일을 읽도록 어시스턴트에게 상기시키세요. 이 형식은 심볼, 관계 및 행 범위를 포함하며 간결하므로 컨텍스트를 작게 유지할 수 있습니다.
- 코드 품질 통찰력을 얻으려면 `graph analyze`를 사용하세요. 외부 호출자와 동적 디스패치로 인해 결과가 달라질 수 있으므로, 발견 사항은 검토 후보입니다.
- 대규모 리팩터링 전에 지문 분석을 실행하여 통합해야 할 복사-붙여넣기를 발견하세요.
- 어시스턴트가 `.skt` 출력이 이미 있는 디렉토리에 `parse`를 다시 실행하면 codeknit은 쓰기를 거부합니다. 이전 출력을 대체하려면 `--clean`을 전달하도록 요청하세요.
- 바이너리 다운로드가 실패한 경우(오프라인 또는 지원되지 않는 플랫폼), [릴리스 페이지](https://github.com/awslabs/codeknit/releases)에서 수동으로 릴리스를 받거나 [설치](../../installation/)를 참조하세요.
- 실행 결과가 부분적으로 표시되면 stderr 및 `[errors]` 섹션을 확인하세요. 자세한 내용은 [분석 신뢰성](../../reference/analysis-reliability/)을 참조하세요.