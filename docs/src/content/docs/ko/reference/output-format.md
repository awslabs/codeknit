---
title: 출력 형식 참조
description: codeknit에서 사용하는 .skt 출력 형식의 전체 참조입니다.
---

이 페이지는 `.skt` 출력의 문법과 동등한 JSON 출력에 대해 설명합니다. 관계 해결 규칙에 대해서는 [언어 지원 및 API 충실도](../language-support/)를 참조하세요.

## 개요

`.skt` 파일은 섹션으로 구성된 간결한 텍스트 파일입니다. 각 섹션은 대괄호로 묶인 헤더 줄로 시작합니다.

| 섹션        | 존재 조건                                       | 내용                                          |
|-------------|------------------------------------------------|-----------------------------------------------|
| `[symbols]` | 항상                                           | 소스 파일별로 그룹화된 심볼 정의              |
| `[edges]`   | `--edges`가 설정된 경우                        | 심볼 간의 관계                               |
| `[errors]`  | 스캔 또는 파싱 진단이 발생한 경우              | 완전히 처리되지 못한 파일                    |
| `[dict]`    | `--minify`가 설정된 경우                       | 반복되는 토큰을 짧은 코드로 매핑하는 사전    |

디렉터리 출력 모드에서는 `[errors]`와 `[dict]`가 보조 파일인 `warnings.skt`와 `dict.skt`에 기록될 수 있습니다. `inline` 모드에서는 다른 섹션과 동일한 출력에 나타납니다.

`--format json`은 동일한 정보를 담은 JSON 문서를 생성합니다. [JSON 출력](#json-출력)을 참조하세요.

## [symbols]

심볼은 소스 파일별로 그룹화됩니다. 각 그룹은 `## <path>` 헤더 줄로 시작하며, 그 뒤에 심볼당 한 줄이 옵니다.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### 심볼 줄

```text
ShortID category/kind Lstart-Lend signature {properties}
```

| 필드          | 설명                                                                                               |
|---------------|---------------------------------------------------------------------------------------------------|
| `ShortID`     | 순차 식별자: `S1`, `S2`, `S3` 등. 엣지는 이 값을 사용하여 심볼을 참조합니다.                      |
| `category/kind` | 심볼의 분류. [카테고리 및 종류](#카테고리-및-종류)를 참조하세요.                                   |
| `Lstart-Lend` | 소스 파일에서 심볼의 첫 번째 및 마지막 행.                                                        |
| `signature`   | 심볼 이름, 호출 가능한 경우 매개변수 및 반환 유형 포함. [시그니처 형식](#시그니처-형식)을 참조하세요. |
| `{properties}` | 중괄호 안의 메타데이터. [속성](#속성)을 참조하세요.                                               |

심볼 줄은 하나의 물리적 줄에 유지됩니다. 시그니처 내부의 줄 바꿈은 `\n` 및 `\r`로 이스케이프됩니다.

추출이 변경되면 심볼 ID가 변경될 수 있습니다. 서로 다른 실행에서 나온 청크를 혼합하지 말고 모든 출력 청크를 함께 다시 생성하세요. 특히 0.5.0으로 업그레이드할 때 심볼 ID와 관계 해결 결과가 변경될 수 있습니다.

### 카테고리 및 종류

| 카테고리   | 종류                                      | 예시                                           |
|------------|------------------------------------------|-----------------------------------------------|
| `callable` | function, method, constructor            | `callable/function`, `callable/method`        |
| `type`     | class, interface, struct, enum           | `type/class`, `type/interface`                |
| `value`    | variable, constant, field                | `value/variable`, `value/constant`            |
| `module`   | package, namespace                       | `module/package`                              |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### 시그니처 형식

| 형식                         | 사용 대상                                  |
|-----------------------------|------------------------------------------|
| `name`                      | 유형, 값 및 모듈                          |
| `name(params)`              | 반환 유형이 없는 호출 가능 항목           |
| `name(params) -> returnType` | 반환 유형이 있는 호출 가능 항목           |

매개변수는 다음과 같이 작성됩니다:

| 형식          | 사용 대상                  |
|---------------|---------------------------|
| `name`        | 유형이 없는 언어           |
| `name: type`  | 유형이 있는 언어           |

알려진 심볼과 일치하는 유형 참조의 경우, 유형 이름은 해당 심볼의 짧은 ID로 대체됩니다. 예를 들어, `-> *S2`는 심볼 `S2`에 대한 포인터를 참조합니다.

### 속성

속성은 시그니처 뒤에 중괄호로 표시되며 쉼표로 구분됩니다. 빈 중괄호(`{}`)는 심볼에 속성이 없음을 의미합니다.

| 속성 형식        | 예시                 |
|------------------|----------------------|
| 플래그           | `exported`, `async`, `static` |
| `key=value`      | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

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

`[edges]` 섹션은 심볼 간의 관계를 나열합니다. `--edges`가 요청된 경우에만 작성됩니다.

### 엣지 줄

```text
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

| 부분                     | 설명                                                                           |
|--------------------------|-------------------------------------------------------------------------------|
| `FromID`                 | 소스 엔드포인트.                                                              |
| `kind`                   | 관계 종류. [엣지 종류](#엣지-종류)를 참조하세요.                             |
| `[unresolved]`, `[ambiguous]` | 선택적 해결 마커. [해결 마커](#해결-마커)를 참조하세요.                     |
| `ToID1, ToID2`           | 쉼표로 구분된 하나 이상의 대상.                                               |
| `[candidates=...]`       | 모호한 엣지의 가능한 대상 목록으로, 대괄호 안에 쉼표로 구분된 목록입니다.     |

대상들은 소스, 종류, 해결 마커 및 후보 집합이 동일한 경우에만 한 줄에 공유됩니다. 엣지에서 참조하는 정의는 이후 파일이나 출력 청크에 있을 수 있습니다. 엣지를 해석하기 전에 모든 청크를 로드하세요.

### 해결 마커

| 마커         | 의미                                                                              |
|--------------|----------------------------------------------------------------------------------|
| 없음         | 분석기가 관계를 해결했습니다.                                                    |
| `[unresolved]` | 분석기가 의존성을 확립할 수 없었습니다.                                          |
| `[ambiguous]`  | 하나 이상의 해석이 가능합니다. 후보 ID는 가능한 대상을 나열합니다.               |

누락된 마커는 컴파일러 보장이 아닙니다. 후보 ID는 가능한 대상을 식별하며, 확인된 연결은 아닙니다. 불확실성은 항상 이 SKT 문법으로 작성되며, 임베디드 JSON으로 작성되지 않습니다.

### 엔드포인트 형식

| 형식                      | 사용 시기                                                                                         | 예시                     |
|---------------------------|--------------------------------------------------------------------------------------------------|--------------------------|
| Short ID                  | 엔드포인트 또는 후보가 알려진 심볼인 경우. 여기에는 해결되지 않은 엣지의 엔드포인트와 포함 심볼이 포함됩니다. | `S4`, `S8`       |
| Bare name                 | 대상이 심볼이 없는 경우. 이름은 소스 파일에 상대적입니다.                                        | `Refresh`, `string`      |
| Quoted full ID            | ID가 복잡하거나 리터럴 이름이 Short ID와 유사한 경우.                                            | `"app.go::S2"`           |

Quoted full ID는 따옴표, 백슬래시 및 줄 바꿈을 이스케이프합니다. `S2`와 같은 리터럴 이름은 Short ID로 읽히지 않도록 quoted full ID로 작성해야 합니다.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### 엣지 종류

| 종류         | 의미                                         |
|--------------|---------------------------------------------|
| `calls`      | 함수/메서드 호출                            |
| `contains`   | 클래스가 메서드를 포함하거나 모듈이 함수를 포함 |
| `inherits`   | 클래스가 다른 클래스를 확장                 |
| `implements` | 클래스가 인터페이스를 구현                  |
| `overrides`  | 메서드가 부모를 재정의                      |
| `references` | 심볼이 다른 심볼을 참조                     |
| `imports`    | 모듈이 다른 모듈을 임포트                   |
| `decorates`  | 데코레이터가 심볼에 적용                    |

반환된 호출 가능 항목 및 콜백 인수는 별도의 종류가 아닌 `references`로 기록됩니다.

### C 및 C++ 파일 및 포함 심볼

C 및 C++ 소스는 `meta/file` 및 `meta/include` 심볼을 생성하므로 포함 관계에 Short ID가 있습니다. 각 헤더 철자는 심볼 목록에 한 번 나타나며, 엣지는 Short ID를 사용합니다. 파일 내 반복된 포함은 하나의 엔드포인트를 공유합니다.

```skt
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

포함 심볼은 작성된 지시문을 기록합니다. 이는 해결된 헤더 선언을 나타내지 않습니다.

### 분석 및 시각화에 미치는 영향

- 불확실한 엣지와 파일/포함 메타데이터는 의존성 메트릭에 기여하지 않습니다. 보고서에는 제외된 관계의 수가 표시됩니다.
- HTML 그래프는 해결된 링크를 그리고 불확실한 관계의 수를 표시합니다. 개별 불확실한 엣지는 그리지 않습니다.

### 예시

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

`[errors]` 섹션은 완전히 처리되지 못한 파일을 나열합니다. 진단이 있는 경우에만 존재합니다.

```text
- <path>: <message>
```

```skt
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict]

`[dict]` 섹션은 축소된 출력(`--minify`)에서만 나타납니다. 짧은 코드를 반복되는 토큰에 매핑합니다.

| 부분       | 설명                                                                 |
|------------|----------------------------------------------------------------------|
| `d0`, `d1`, … | 사전 코드.                                                        |
| 엔트리 줄  | `- <code>: <token>`                                                  |

축소된 출력을 읽으려면 각 코드를 해당 사전 값으로 대체하세요. 코드는 `[symbols]`의 카테고리/종류 쌍 및 속성을 대체할 수 있습니다. 또한 `[edges]`의 관계 종류도 대체할 수 있습니다. 불확실성 마커는 읽을 수 있게 유지됩니다. 예를 들어, `S1 --d2[unresolved]--> S3`는 종류가 `d2` 값인 해결되지 않은 엣지입니다.

### 축소된 예시

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

여기서 `S2`는 `callable/method`이며 속성은 `async=false`입니다.

## 전체 예시

다음 출력은 `[symbols]`, `[edges]`, `[errors]`를 포함합니다. 축소되지 않았으므로 `[dict]` 섹션이 없습니다. [축소된 예시](#축소된-예시)를 참조하세요.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
## pkg/util/hash.go
S6 callable/function L3-L5 Hash(data: string) -> string {exported}
## pkg/util/hash_alt.go
S7 callable/function L3-L5 Hash(data: string) -> string {exported}

[edges]
S1 --contains--> S2, S3, S4, S5
S2 --contains--> S4
S3 --references--> S2
S4 --calls--> S5
S4 --calls[unresolved]--> Refresh
S5 --calls[ambiguous]--> Hash [candidates=S6, S7]

[errors]
- pkg/legacy/broken.go: syntax error at line 42
```

## JSON 출력

`--format json`