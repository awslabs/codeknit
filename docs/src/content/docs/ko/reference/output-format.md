---
title: 출력 형식 참조
description: codeknit에서 사용하는 .skt 출력 형식에 대한 완전한 참조입니다.
---

`.skt`(skeleton) 형식은 `codeknit`에서 추출된 코드 구조를 표현하기 위해 사용되는 간결하고 사람이 읽을 수 있는 텍스트 형식입니다. 이는 심볼, 관계 및 메타데이터를 LLM 소비와 구조 분석에 적합한 최소한의 형태로 포함합니다.

`.skt` 파일은 섹션으로 나뉩니다. 각 섹션은 대괄호로 묶인 헤더로 시작합니다. 섹션은 어떤 순서로든 나타날 수 있지만, 일반적으로 `[symbols]`가 먼저 옵니다.

## [symbols]

`[symbols]` 섹션은 소스 파일별로 그룹화된 모든 추출된 심볼을 나열합니다. 각 파일은 `##` 헤더와 파일 경로로 시작합니다.

### 행 형식

각 심볼은 다음 구조를 가진 단일 행으로 표현됩니다:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### 필드

- **ShortID**: 각 심볼에 할당된 순차 식별자(예: `S1`, `S2`, `S3`). 엣지 및 기타 섹션에서 참조로 사용됩니다.
- **category/kind**: 심볼의 카테고리와 특정 종류를 나타내는 슬래시로 구분된 쌍입니다.
- **Lstart-Lend**: 심볼이 정의된 소스 파일의 행 범위(예: `L10-L15`)입니다.
- **signature**: 심볼의 이름과 유형 정보입니다. 심볼에 따라 형식이 다릅니다:
  - `name` — 타입, 값, 모듈용
  - `name(params)` — 반환 유형이 없는 호출 가능 객체용
  - `name(params) -> returnType` — 반환 유형이 있는 호출 가능 객체용
- **{properties}**: 중괄호로 묶인 선택적 메타데이터입니다. 여러 속성은 쉼표로 구분됩니다.

### 매개변수

- 타입이 지정되지 않은 언어: `paramName`
- 타입이 지정된 언어: `paramName: type`
- 알려진 심볼과 일치하는 타입 참조는 해당 ShortID로 대체됩니다(예: `config: S5` 대신 `config: Config`).

### 속성

일반적인 속성에는 다음이 포함됩니다:

- `async`: `true` 또는 `false`
- `exported`: `true` 또는 `false`
- `static`: 심볼이 정적일 경우 표시됨
- `visibility=public|private|protected`
- `receiver=*TypeName`: 메서드의 경우, 수신자 유형을 나타냄

### 심볼 카테고리 및 종류

| Category   | Kinds                          | Examples                               |
| ---------- | ------------------------------ | -------------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

### 예시

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
```

## [edges]

`[edges]` 섹션은 ShortID를 사용하여 심볼 간의 관계를 정의합니다.

### 행 형식

```
FromID --kind--> ToID1, ToID2
```

여러 대상 ID는 쉼표로 구분됩니다. 각 행은 하나의 방향성 관계를 나타냅니다.

### 엣지 종류

| Kind         | Meaning                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | 함수/메서드 호출                                |
| `contains`   | 클래스가 메서드를 포함, 모듈이 함수를 포함      |
| `inherits`   | 클래스가 다른 클래스를 확장                     |
| `implements` | 클래스가 인터페이스를 구현                      |
| `overrides`  | 메서드가 부모 메서드를 재정의                   |
| `references` | 심볼이 다른 심볼을 참조                        |
| `imports`    | 모듈이 다른 모듈을 임포트                       |
| `decorates`  | 데코레이터가 심볼에 적용됨                      |

### 예시

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

`[errors]` 섹션은 완전히 파싱할 수 없었던 파일을 나열합니다.

### 형식

각 행은 `-`로 시작하며 파일 경로와 오류 메시지가 뒤따릅니다:

```
- path/to/file.go: syntax error at line 42
```

### 예시

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

`[dict]` 섹션은 `--minify` 플래그가 사용된 경우에만 나타납니다. 이 섹션은 짧은 사전 코드를 반복되는 문자열 토큰에 매핑하여 출력 크기를 줄입니다.

### 형식

각 행은 사전 코드(`d0`, `d1` 등)를 확장된 값에 매핑합니다:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

파일의 나머지 부분에서 이러한 코드는 전체 값을 대체합니다.

### 예시

```skt
[dict]
- d0: async=false
- d1: callable/method
- d2: exported

[symbols]
## src/handler.py
S1 type/class L1-L6 Handler {}
S2 d1 L2-L3 __init__(name) {d0}
S3 d1 L5-L6 handle(request) {d0}

[edges]
S1 --contains--> S2, S3
```

## 전체 예시

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {d0}
S3 d1 L10-L12 NewServer(addr: string) -> *S2 {d0}
S4 callable/method L14-L20 Serve() {d0, receiver=*Server}
S5 callable/function L22-L25 handleError(err: error) -> bool {}

[edges]
S2 --contains--> S4
S4 --calls--> S5
S3 --returns--> S2

[errors]
- utils/broken.go: syntax error at line 5
```