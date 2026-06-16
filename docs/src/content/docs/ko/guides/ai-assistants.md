---
title: AI 어시스턴트와 함께 사용하기
description: Kiro, Claude Code 및 기타 AI 코딩 어시스턴트를 위한 codeknit 스킬 설정하기
---

codeknit는 AI 코딩 어시스턴트가 효과적으로 사용할 수 있도록 준비된 스킬을 제공합니다. 이러한 스킬을 통해 어시스턴트는 수동 프롬프트 없이 코드 구조 추출, 중복 감지 및 구조 분석을 수행할 수 있습니다.

## 스킬 개요

codeknit는 두 가지 스킬을 제공합니다:

- **`codeknit-parse`**: 어시스턴트가 코드 구조(함수, 클래스, 메서드, 변수)와 관계(호출, 상속, 포함)를 `.skt` 파일로 추출하는 방법을 가르칩니다.
- **`codeknit-fingerprint`**: 어시스턴트가 퍼지 해싱(fuzzy hashing)을 사용하여 중복 및 근사 중복 코드를 감지하는 방법을 가르칩니다.

각 스킬에는 어시스턴트가 필요에 따라 읽어 사용법, 플래그, 출력 형식 및 워크플로를 이해하는 데 도움이 되는 문서가 포함되어 있습니다.

## 설치

스킬 디렉터리를 어시스턴트의 스킬 폴더로 복사합니다.

**Kiro**의 경우:

```bash
cp -r skills/codeknit-parse ~/.kiro/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.kiro/skills/codeknit-fingerprint
```

**Claude Code**의 경우:

```bash
cp -r skills/codeknit-parse ~/.claude/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.claude/skills/codeknit-fingerprint
```

설치 후, 어시스턴트는 자동으로 codeknit 명령어를 호출하고, 적절한 플래그를 선택하며, `.skt` 출력을 해석하는 방법을 알게 됩니다.

## 각 스킬이 가르치는 내용

### codeknit-parse

`codeknit-parse` 스킬은 어시스턴트에게 다음을 가르칩니다:

- 다양한 시나리오에 적합한 플래그와 함께 `codeknit parse` 실행하기
- 적절한 출력 모드 선택:
  - 대부분의 프로젝트에 적합한 `directory-flat`(기본값)
  - 단일 파일 또는 작은 입력에 적합한 `inline`
  - 소스 구조를 미러링하는 `directory-tree`
- `.skt` 출력 파일 읽기 및 해석, `[symbols]`, `[edges]`, 선택적 `[dict]` 섹션 포함
- 구조적 데이터를 활용한 리팩터링, 의존성 매핑 및 코드 리뷰
- 더 깊은 코드 품질 인사이트를 위해 `codeknit graph analyze` 실행(순환 의존성, 허브 심볼, god class 등)

### codeknit-fingerprint

`codeknit-fingerprint` 스킬은 어시스턴트에게 다음을 가르칩니다:

- 중복 감지, DRY 감사 및 리팩터 식별을 위해 `codeknit fingerprint` 사용하기
- 적절한 유사도 범위 선택(`--min-similarity`, `--max-similarity`)
- 출력에서 `[duplicates]` 섹션을 읽어 근사 중복 코드 식별하기
- fingerprints가 의미적 의도가 아닌 구조적 형태를 측정한다는 점 이해하기
- 필요한 경우 Ollama 임베딩과 함께 `--rerank`을 사용하여 거짓 양성 줄이기

## 워크플로 예시

### 구조 분석

1. 어시스턴트에게 코드베이스 구조 분석 요청
2. 어시스턴트가 `codeknit parse ./src`를 실행하고 결과 `.skt` 파일 읽기
3. 구조적 질문(의존성, 호출 체인, dead code)에 답변
4. 더 깊은 인사이트를 위해 `codeknit graph analyze ./src` 실행 및 보고서 해석

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

- **구조적 질문에 대해서는 원시 소스가 아닌 `.skt` 파일을 읽으세요** — `.skt` 파일은 추출된 구조를 간결하고 신뢰할 수 있는 형식으로 포함하고 있습니다.
- `codeknit graph analyze`를 사용하여 순환 의존성, 허브 심볼, 깊은 상속 체인과 같은 코드 품질 문제 발견하기
- 대규모 리팩터링 전에 `codeknit fingerprint`를 실행하여 통합해야 할 복사-붙여넣기 코드 식별하기
- `.skt` 형식은 토큰 효율성을 위해 설계되어 LLM 컨텍스트 윈도우에 이상적입니다.
- 대규모 코드베이스를 처리할 때 토큰 사용량을 더욱 줄이려면 `--minify` 사용하기
