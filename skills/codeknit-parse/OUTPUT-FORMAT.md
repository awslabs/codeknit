# .skt output format reference

## Contents

- [symbols] — symbol definitions grouped by source file
- [edges] — relationships between symbols
- [errors] — files with parse errors
- [dict] — minification dictionary (optional)

## [symbols]

Symbols are grouped by source file. Each line:

```
ShortID category/kind Lstart-Lend signature {properties}
```

Where signature is one of:

- `name` — for types, values, modules
- `name(params)` — for callables without return type
- `name(params) -> returnType` — for callables with return type

Parameters use `name` for untyped languages and `name: type` for typed languages. Type references that match known symbols are replaced with their short IDs (e.g. `S3` instead of `Environment`).

- `ShortID` — sequential identifier (S1, S2, S3, ...)
- `category/kind` — classification (see categories below)
- `Lstart-Lend` — line span in source file
- `name(params)` — symbol name with parameters as `name` or `name: type` if callable
- `-> returnType` — return type (appended after params for typed languages)
- `{properties}` — metadata like `async`, `exported`, `static`, `visibility=public`, `receiver=*Type`

### Symbol categories

| Category   | Kinds                          | Examples                               |
| ---------- | ------------------------------ | -------------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

### Example

```
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
```

## [edges]

Relationships between symbols:

```
FromID --kind--> ToID1, ToID2
```

### Edge kinds

| Kind         | Meaning                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | function/method invocation                      |
| `contains`   | class contains method, module contains function |
| `inherits`   | class extends another class                     |
| `implements` | class implements interface                      |
| `overrides`  | method overrides parent method                  |
| `references` | symbol references another symbol                |
| `imports`    | module imports another module                   |
| `decorates`  | decorator applied to a symbol                   |

### Example

```
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

Files that had parse errors:

```
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict] (minified output only)

Maps short codes to repeated tokens. When present, substitute `d0`, `d1`, etc. with their dictionary values.

```
[dict]
- d0: async=false
- d1: callable/function
- d2: callable/method
- d3: exported
```

### Minified example

```
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
