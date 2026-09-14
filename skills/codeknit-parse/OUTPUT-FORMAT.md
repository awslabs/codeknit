# .skt output format reference

The default `codeknit parse` format is `.skt`, a compact text format optimized for LLM context. Use `--format json` when a script, editor integration, or CI check needs structured data.

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

| Category   | Kinds                                      | Examples                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

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
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

An omitted status means the analyzer resolved the relationship. `unresolved` means it could not establish the dependency; `ambiguous` means more than one interpretation remains possible. These statuses are not compiler guarantees. Candidate IDs identify possible targets, not confirmed connections.

Known endpoints use symbol IDs, including endpoints of unresolved edges. Names without symbols can appear directly, relative to the source file. Complex identities are quoted full IDs with escaped quotes, backslashes, and line breaks. A literal name such as `S2` must be written as a quoted full ID (`"app.go::S2"`) so it cannot be mistaken for a symbol ID. IDs may refer to symbols in later files or output chunks; load their definitions before interpreting the edges.

Targets are grouped only when their source, kind, resolution, and candidates agree:

```text
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### File and include endpoints

C/C++ includes have metadata symbols. Header spellings appear once, while edges use short IDs:

```text
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

An include symbol describes the written directive, not a resolved header declaration. Repeated includes within a file share an endpoint. File/include metadata is excluded from dependency rankings and dead-code findings. Uncertain edges are excluded from dependency metrics; graph reports and HTML show their counts, and the HTML viewer currently does not draw them.

Symbol IDs can change when extraction changes. Regenerate all output together rather than mixing chunks from different runs.

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

Minification encodes relationship kinds while preserving status labels, for example `S1 --d2[unresolved]--> S3`. Expand `d2` using the dictionary; retain the unresolved status.

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

## JSON output

Run JSON output with:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

In directory output modes, JSON is written to `codeknit.json`. In `inline` mode, JSON is written to stdout.

Top-level fields:

| Field     | Meaning                                                                  |
| --------- | ------------------------------------------------------------------------ |
| `files`   | Source files included in the parse result                                |
| `symbols` | Extracted symbols with IDs, short IDs, source file, category, kind, span |
| `edges`   | Relationships between symbols; omitted when no edges are present         |
| `errors`  | Parse warnings/errors; omitted when no parse errors are present          |

Edges include `from`, `to`, and `kind`. Resolved endpoints have short IDs in `from_short` and `to_short`. Uncertain edges additionally carry `resolution` and, when available, `candidates` as full symbol IDs. Their `to_short` is omitted, even when the target has an include metadata symbol. JSON and SKT preserve the same uncertainty despite their different encodings.

### Example

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```
