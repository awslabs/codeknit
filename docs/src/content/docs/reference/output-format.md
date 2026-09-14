---
title: Output Format Reference
description: Complete reference for the .skt output format used by codeknit.
---

The `.skt` (skeleton) format is a compact, human-readable text format used by `codeknit` to represent extracted code structure. It contains symbols, relationships, and metadata in a minimal form suitable for LLM consumption and structural analysis.

A `.skt` file is divided into sections. Each section begins with a header in square brackets. Sections may appear in any order, though `[symbols]` typically comes first.

## [symbols]

The `[symbols]` section lists all extracted symbols grouped by their source file. Each file is introduced with a `##` header followed by the file path.

### Line format

Each symbol is represented on a single line with the following structure:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Fields

- **ShortID**: A sequential identifier assigned to each symbol (e.g., `S1`, `S2`, `S3`). Used as a reference in edges and other sections.
- **category/kind**: A slash-separated pair indicating the symbol's category and specific kind.
- **Lstart-Lend**: The line span in the source file where the symbol is defined (e.g., `L10-L15`).
- **signature**: The symbol's name and type information. Format depends on the symbol:
  - `name` — for types, values, modules
  - `name(params)` — for callables without return type
  - `name(params) -> returnType` — for callables with return type
- **{properties}**: Optional metadata enclosed in braces. Multiple properties are comma-separated.

### Parameters

- In untyped languages: `paramName`
- In typed languages: `paramName: type`
- Type references that match known symbols are replaced with their short IDs (e.g., `config: S5` instead of `config: Config`).

### Properties

Common properties include:

- `async`: `true` or `false`
- `exported`: `true` or `false`
- `static`: present if symbol is static
- `visibility=public|private|protected`
- `receiver=*TypeName`: for methods, indicates receiver type

### Symbol categories and kinds

| Category   | Kinds                                      | Examples                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### Example

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

The `[edges]` section records relationships when `--edges` is enabled. Known endpoints use their ShortIDs; targets that cannot be resolved to a symbol may appear as names.

### Line format

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

An omitted status means the analyzer resolved the relationship. `unresolved` means it could not establish the dependency; `ambiguous` means competing interpretations remain possible. Candidate IDs identify possible targets, not confirmed connections. Resolution is based on syntax and scope, not a compiler guarantee.

Targets are grouped only when their source, kind, resolution, and candidate set match. Names without symbols are relative to the source file. Complex identities use quoted full IDs with escaped quotes, backslashes, and line breaks. A literal name that resembles a short ID must be quoted as a full ID, such as `"app.go::S2"`. Symbol definitions may occur in later files or output chunks.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### File and include endpoints

C/C++ includes have explicit metadata symbols. Header spellings appear once in the symbol list, and edges use short IDs even though header resolution remains unresolved:

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

An include symbol records the written directive, not a resolved header declaration. Repeated includes within a file share an endpoint. File/include metadata is excluded from dependency rankings and dead-code findings.

Uncertain relationships are preserved in parse output but excluded from dependency metrics. Graph reports show excluded relationship counts; the HTML viewer shows their count without drawing individual uncertain edges.

When upgrading to `0.5.0`, regenerate all output chunks together. Symbol IDs and relationship results can change, so chunks from different runs must not be mixed.

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

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

The `[errors]` section lists files that could not be parsed completely.

### Format

Each line starts with `-` followed by the file path and error message:

```
- path/to/file.go: syntax error at line 42
```

### Example

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

The `[dict]` section appears only when the `--minify` flag is used. It maps short dictionary codes to repeated string tokens to reduce output size.

### Format

Each line maps a dictionary code (`d0`, `d1`, etc.) to its expanded value:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

In the rest of the file, these codes replace their full values.

Relationship kinds can also use dictionary codes while status labels remain readable: `S1 --d2[unresolved]--> S3`. Expand `d2` using the dictionary and retain the unresolved status.

### Example

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

## Full example

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

## JSON relationships

Use JSON when a script or integration needs structured output:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Each edge includes `from`, `to`, and `kind`. Known source IDs appear in `from_short`; resolved target IDs appear in `to_short`. Uncertain edges add `resolution` and, when available, `candidates` containing full symbol IDs. They omit `to_short`, even when the target has an include metadata symbol.

For example, an ambiguous call can be represented as:

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

JSON and SKT preserve the same uncertainty with different encodings. See [Language Support and API Fidelity](../language-support/) for resolution rules and known limits.
