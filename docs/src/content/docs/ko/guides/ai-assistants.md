---
title: AI 어시스턴트와 함께 사용하기
description: Kiro, Claude Code 및 기타 AI 코딩 어시스턴트에서 codeknit을 스킬로 설정하기
---

codeknit은 AI 코딩 어시스턴트가 효과적으로 사용할 수 있도록 준비된 스킬을 제공합니다. 이러한 스킬을 통해 어시스턴트는 코드 구조를 추출하고, 중복을 감지하며, 수동 프롬프트 없이 구조적 분석을 수행할 수 있습니다.

## 스킬 개요

codeknit은 두 가지 스킬을 제공합니다:

- **`codeknit-parse`**: 어시스턴트가 코드 구조(함수, 클래스, 메서드, 변수)와 관계(호출, 상속, 포함)를 `.skt` 파일로 추출하는 방법을 가르칩니다.
- **`codeknit-fingerprint`**: 어시스턴트가 퍼지 해싱을 사용하여 중복 및 근사 중복 코드를 감지하는 방법을 가르칩니다.

각 스킬에는 어시스턴트가 사용법, 플래그, 출력 형식 및 워크플로를 이해하는 데 도움이 되는 문서가 포함되어 있습니다.

## 설치

설치 도우미를 사용하여 스킬 디렉터리를 어시스턴트의 스킬 폴더로 복사하세요. 설치 프로그램은 번들된 스킬 파일만 다운로드하므로 저장소를 클론할 필요가 없습니다.

**Codex**, **Kiro**, **Claude Code**용 설치:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash
```

하나의 어시스턴트용 설치:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant codex
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant kiro
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant claude
```

로컬 체크아웃에서 Makefile 도우미를 사용할 수 있습니다:

```bash
make skills-install-dry-run
make skills-install
```

설치 프로그램은 기본적으로 기존 스킬 디렉터리를 건너뜁니다. 이를 대체하려면 `--force`를 추가하세요:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant all --force
```

설치 후, 어시스턴트는 자동으로 codeknit 명령어를 호출하고 적절한 플래그를 선택하며 `.skt` 출력을 해석하는 방법을 알게 됩니다.

## 각 스킬이 가르치는 내용

### codeknit-parse

`codeknit-parse` 스킬은 어시스턴트에게 다음을 가르칩니다:

- 다양한 시나리오에 적합한 플래그로 `codeknit parse` 실행하기
- 적절한 출력 모드 선택:
  - 대부분의 프로젝트에 `directory-flat`(기본값)
  - 단일 파일 또는 작은 입력에 `inline`
  - 소스 구조를 미러링하기 위해 `directory-tree`
- `.skt` 출력 파일 읽기 및 해석, `[symbols]`, `[edges]`, 선택적 `[dict]` 섹션 포함
- 구조적 데이터를 사용하여 리팩토링, 의존성 매핑 및 코드 리뷰 수행
- 더 깊은 코드 품질 통찰력을 위해 `codeknit graph analyze` 실행(순환 의존성, 허브 심볼, god class 등)

### codeknit-fingerprint

`codeknit-fingerprint` 스킬은 어시스턴트에게 다음을 가르칩니다:

- 중복 감지, DRY 감사 및 리팩터 식별을 위해 `codeknit fingerprint` 사용하기
- 적절한 유사도 범위 선택(`--min-similarity`, `--max-similarity`)
- `[duplicates]` 섹션을 읽어 근사 중복 코드 식별
- 지문이 의미적 의도가 아닌 구조적 형태를 측정한다는 점 이해
- 필요한 경우 Ollama 임베딩으로 `--rerank`을 사용하여 거짓 양성 감소

## 워크플로 예시

### 구조 분석

1. 어시스턴트에게 코드베이스 구조 분석 요청
2. 어시스턴트가 `codeknit parse ./src`를 실행하고 결과 `.skt` 파일 읽기
3. 구조적 질문(의존성, 호출 체인, 데드 코드)에 답변
4. 더 깊은 통찰력을 위해 `codeknit graph analyze ./src` 실행 및 보고서 해석

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### 중복 감지

1. 어시스턴트에게 중복 코드 찾기 요청
2. 어시스턴트가 `codeknit fingerprint ./src` 실행
3. 출력에서 `[duplicates]` 섹션 읽기
4. 플래그가 지정된 쌍 조사 및 통합 제안

```skt
[duplicates]
S1, S2: 87% 유사도
S3, S4: 76% 유사도
```

## 팁

- **구조적 질문에 대해서는 원본 소스가 아닌 `.skt` 파일을 읽으세요** — 이들은 압축되고 신뢰할 수 있는 형식으로 추출된 구조를 포함합니다
- 순환 의존성, 허브 심볼 및 깊은 상속 체인과 같은 코드 품질 문제를 발견하기 위해 `codeknit graph analyze` 사용
- 대규모 리팩터 전에 `codeknit fingerprint`를 실행하여 통합해야 할 복사-붙여넣기 코드 식별
- `.skt` 형식은 토큰 효율성을 위해 설계되어 LLM 컨텍스트 윈도우에 이상적입니다
- 대규모 코드베이스를 처리할 때 토큰 사용량을 더욱 줄이기 위해 `--minify` 사용