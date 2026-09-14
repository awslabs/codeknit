---
title: 출력 형식 참조
description: codeknit에서 사용되는 .skt 출력 형식에 대한 완전한 참조입니다.
---

`.skt`(skeleton) 형식은 `codeknit`에서 추출된 **코드 구조**를 표현하기 위해 사용되는 간결하고 사람이 읽을 수 있는 텍스트 형식입니다. 이 형식은 심볼, 관계 및 메타데이터를 LLM 소비와 구조 분석에 적합한 최소한의 형태로 포함합니다.

`.skt` 파일은 섹션으로 나뉩니다. 각 섹션은 대괄호로 묶인 헤더로 시작합니다. 섹션은 어떤 순서로든 나타날 수 있지만 `[symbols]`가 일반적으로 먼저 나옵니다.

## [symbols]

`[symbols]` 섹션은 추출된 모든 심볼을 소스 파일별로 그룹화하여 나열합니다. 각 파일은 `##` 헤더와 파일 경로로 시작합니다.

### 행 형식

각 심볼은 다음과 같은 구조로 한 줄에 표현됩니다:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### 필드

- **ShortID**: 각 심볼에 할당된 순차 식별자(예: `S1`, `S2`, `S3`). 엣지와 기타 섹션에서 참조로 사용됩니다.
- **category/kind**: 심볼의 카테고리와 특정 종류를 슬래시로 구분한 쌍입니다.
- **Lstart-Lend**: 심볼이 정의된 소스 파일의 **행 범위**(예: `L10-L15`).
- **signature**: 심볼의 이름과 타입 정보입니다. 심볼 유형에 따라 형식이 다릅니다:
  - `name` — 타입, 값, 모듈의 경우
  - `name(params)` — 반환 타입이 없는 호출 가능한 경우
  - `name(params) -> returnType` — 반환 타입이 있는 호출 가능한 경우
- **{properties}**: 중괄호로 묶인 선택적 메타데이터입니다. 여러 속성은 쉼표로 구분됩니다.

### 매개변수

- 타입이 없는 언어: `paramName`
- 타입이 있는 언어: `paramName: type`
- 알려진 심볼과 일치하는 타입 참조는 단축 ID로 대체됩니다(예: `config: S5` 대신 `config: Config`).

### 속성

일반적인 속성은 다음과 같습니다:

- `async`: `true` 또는 `false`
- `exported`: `true` 또는 `false`
- `static`: 심볼이 정적일 경우 표시됨
- `visibility=public|private|protected`
- `receiver=*TypeName`: 메서드의 경우, 수신자 타입을 나타냄

### 심볼 카테고리와 종류

| Category   | Kinds                                      | Examples                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

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

`[edges]` 섹션은 `--edges` 플래그가 활성화된 경우 관계를 기록합니다. 알려진 엔드포인트는 ShortID를 사용하며, 심볼로 확인할 수 없는 대상은 이름으로 나타날 수 있습니다.

### 행 형식

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

상태가 생략된 경우 분석기가 관계를 확인한 것입니다. `unresolved`는 의존성을 확립할 수 없음을 의미하며, `ambiguous`는 경쟁하는 해석이 가능함을 의미합니다. 후보 ID는 가능한 대상을 식별하며, 확인된 연결은 아닙니다. 확인은 구문과 범위를 기반으로 하며, 컴파일러 보장은 아닙니다.

대상은 소스, 종류, 확인 상태 및 후보 집합이 일치할 때만 그룹화됩니다. 심볼이 없는 이름은 소스 파일에 상대적입니다. 복잡한 식별자는 따옴표로 묶인 전체 ID를 사용하며, 따옴표, 백슬래시 및 줄 바꿈은 이스케이프됩니다. ShortID와 유사한 리터럴 이름은 전체 ID로 따옴표로 묶어야 합니다(예: `"app.go::S2"`). 심볼 정의는 이후 파일이나 출력 청크에서 발생할 수 있습니다.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### 파일 및 포함 엔드포인트

C/C++ 포함은 명시적 메타데이터 심볼을 가집니다. 헤더 철자는 심볼 목록에 한 번 나타나며, 엣지는 헤더 확인 상태와 관계없이 ShortID를 사용합니다:

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

포함 심볼은 작성된 지시문을 기록하며, 확인된 헤더 선언은 아닙니다. 파일 내 반복된 포함은 하나의 엔드포인트를 공유합니다. 파일/포함 메타데이터는 의존성 순위 및 **dead code** 발견에서 제외됩니다.

불확실한 관계는 파싱 출력에 보존되지만 의존성 메트릭에서 제외됩니다. 그래프 보고서는 제외된 관계 수를 표시하며, HTML 뷰어는 개별 불확실한 엣지를 그리지 않고 수만 표시합니다.

`0.5.0`으로 업그레이드할 때는 모든 출력 청크를 함께 다시 생성하십시오. 심볼 ID와 관계 결과는 변경될 수 있으므로, 서로 다른 실행의 청크를 혼합해서는 안 됩니다.

### 엣지 종류

| Kind         | Meaning                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | 함수/메서드 호출                                |
| `contains`   | 클래스가 메서드를 포함하거나 모듈이 함수를 포함 |
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

`[dict]` 섹션은 `--minify` 플래그가 사용된 경우에만 나타납니다. 이 섹션은 출력 크기를 줄이기 위해 반복되는 문자열 토큰을 짧은 사전 코드에 매핑합니다.

### 형식

각 행은 사전 코드(`d0`, `d1` 등)를 확장된 값에 매핑합니다:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

파일의 나머지 부분에서 이러한 코드는 전체 값을 대체합니다.

관계 종류도 사전 코드를 사용할 수 있지만 상태 레이블은 읽을 수 있게 유지됩니다: `S1 --d2[unresolved]--> S3`. `d2`를 사전을 사용하여 확장하고 `unresolved` 상태를 유지합니다.

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
S3 --references--> S2

[errors]
- utils/broken.go: syntax error at line 5
```

## JSON 관계

스크립트나 통합에서 구조화된 출력이 필요한 경우 JSON을 사용하십시오:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

각 엣지는 `from`, `to`, `kind`를 포함합니다. 알려진 소스 ID는 `from_short`에 나타나며, 확인된 대상 ID는 `to_short`에 나타납니다. 불확실한 엣지는 `resolution`을 추가하며, 사용 가능한 경우 `candidates`에 전체 심볼 ID를 포함합니다. 이들은 `to_short`를 생략하며, 대상에 포함 메타데이터 심볼이 있더라도 마찬가지입니다.

예를 들어, 모호한 호출은 다음과 같이 표현될 수 있습니다:

```json
{
  "from": "app.py::run",
  "from_short": "S1",
  "to": "app.py::Helper",
  "kind": "calls",
  "resolution": "ambiguous",
  "candidates": ["helpers/a.py::Helper", "helpers/b.py::Helper"]
}
```

JSON과 SKT는 동일한 불확실성을 다른 인코딩으로 보존합니다. 확인 규칙과 알려진 제한 사항은 [언어 지원 및 API 충실도](../language-support/)를 참조하십시오.