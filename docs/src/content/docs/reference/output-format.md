---
title: Output Format Reference
description: Complete reference for the .skt output format used by codeknit.
---

This page describes the grammar of `.skt` output and the equivalent JSON output. For the rules that decide how relationships resolve, see [Language Support and API Fidelity](../language-support/).

## Overview

A `.skt` file is a compact text file made of sections. Each section starts with a header line in square brackets.

| Section        | Present when                                       | Contents                                          |
| -------------- | -------------------------------------------------- | ------------------------------------------------- |
| `[symbols]`    | Always                                             | Symbol definitions grouped by source file         |
| `[edges]`      | `--edges` is set                                   | Relationships between symbols                     |
| `[errors]`     | A scan or parse diagnostic occurred                | Files that could not be fully processed           |
| `[dict]`       | `--minify` is set                                  | Dictionary that maps short codes to repeated tokens |

In directory output modes, `[errors]` and `[dict]` can be written to the auxiliary files `warnings.skt` and `dict.skt`. In `inline` mode they appear in the same output as the other sections.

`--format json` produces a JSON document with the same information. See [JSON output](#json-output).

## [symbols]

Symbols are grouped by source file. Each group starts with a `## <path>` header line, followed by one line per symbol.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### Symbol line

```text
ShortID category/kind Lstart-Lend signature {properties}
```

| Field          | Description                                                                                               |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| `ShortID`      | Sequential identifier: `S1`, `S2`, `S3`, and so on. Edges refer to symbols by this value.                 |
| `category/kind` | Classification of the symbol. See [Categories and kinds](#categories-and-kinds).                          |
| `Lstart-Lend`  | First and last line of the symbol in the source file.                                                     |
| `signature`    | Symbol name, with parameters and return type for callables. See [Signature formats](#signature-formats). |
| `{properties}` | Metadata in braces. See [Properties](#properties).                                                        |

A symbol line stays on one physical line. Line breaks inside a signature are escaped as `\n` and `\r`.

Symbol IDs can change when extraction changes. Regenerate all output chunks together instead of mixing chunks from different runs. This applies in particular when upgrading to 0.5.0, because symbol IDs and relationship resolution results can change.

### Categories and kinds

| Category   | Kinds                                      | Examples                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### Signature formats

| Form                         | Used for                                  |
| ---------------------------- | ----------------------------------------- |
| `name`                       | Types, values, and modules                |
| `name(params)`               | Callables without a return type           |
| `name(params) -> returnType` | Callables with a return type              |

Parameters are written as:

| Form          | Used for                  |
| ------------- | ------------------------- |
| `name`        | Languages without types   |
| `name: type`  | Languages with types      |

When a type reference matches a known symbol, the type name is replaced with that symbol's short ID. For example, `-> *S2` refers to a pointer to the symbol `S2`.

### Properties

Properties appear in braces after the signature, separated by commas. Empty braces (`{}`) mean the symbol has no properties.

| Property form        | Example                 |
| -------------------- | ----------------------- |
| Flag                 | `exported`, `async`, `static` |
| `key=value`          | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

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

The `[edges]` section lists relationships between symbols. It is written only when edges are requested with `--edges`.

### Edge line

```text
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

| Part                     | Description                                                                           |
| ------------------------ | ------------------------------------------------------------------------------------- |
| `FromID`                 | Source endpoint.                                                                      |
| `kind`                   | Relationship kind. See [Edge kinds](#edge-kinds).                                     |
| `[unresolved]`, `[ambiguous]` | Optional resolution marker. See [Resolution markers](#resolution-markers).        |
| `ToID1, ToID2`           | One or more targets, separated by commas.                                             |
| `[candidates=...]`       | Possible targets for an ambiguous edge, as a comma-separated list inside brackets.    |

Targets share one line only when the source, kind, resolution marker, and candidate set are identical. A definition referenced by an edge can occur in a later file or output chunk. Load all chunks before interpreting edges.

### Resolution markers

| Marker         | Meaning                                                                              |
| -------------- | ------------------------------------------------------------------------------------ |
| None           | The analyzer resolved the relationship.                                              |
| `[unresolved]` | The analyzer could not establish the dependency.                                     |
| `[ambiguous]`  | More than one interpretation remains possible. Candidate IDs list the possible targets. |

A missing marker is not a compiler guarantee. Candidate IDs identify possible targets, not confirmed connections. Uncertainty is always written in this SKT grammar, never as embedded JSON.

### Endpoint forms

| Form                      | When used                                                                                         | Example                     |
| ------------------------- | ------------------------------------------------------------------------------------------------- | --------------------------- |
| Short ID                  | The endpoint or candidate is a known symbol. This includes endpoints of unresolved edges and include symbols. | `S4`, `S8`       |
| Bare name                 | The target has no symbol. The name is relative to the source file.                                | `Refresh`, `string`         |
| Quoted full ID            | The identity is complex, or a literal name resembles a short ID.                                  | `"app.go::S2"`              |

Quoted full IDs escape quotes, backslashes, and line breaks. A literal name such as `S2` must be written as a quoted full ID so that it is not read as a short ID.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Edge kinds

| Kind         | Meaning                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | function/method invocation                      |
| `contains`   | class contains method, module contains function |
| `inherits`   | class extends another                           |
| `implements` | class implements interface                      |
| `overrides`  | method overrides parent                         |
| `references` | symbol references another                       |
| `imports`    | module imports another                          |
| `decorates`  | decorator applied to symbol                     |

Returned callables and callback arguments are recorded as `references`, not as a separate kind.

### C and C++ file and include symbols

C and C++ sources produce `meta/file` and `meta/include` symbols, so include relationships have short IDs. Each header spelling appears once in the symbol list, and edges use the short IDs. Repeated includes within a file share one endpoint.

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

An include symbol records the directive as written. It does not represent a resolved header declaration.

### Effect on analysis and visualization

- Uncertain edges and file/include metadata do not contribute to dependency metrics. Reports show the number of excluded relationships.
- The HTML graph draws resolved links and displays a count of uncertain relationships. It does not draw individual uncertain edges.

### Example

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

The `[errors]` section lists files that were not fully processed. It is present only when diagnostics exist.

```text
- <path>: <message>
```

```skt
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict]

The `[dict]` section appears only in minified output (`--minify`). It maps short codes to repeated tokens.

| Part       | Description                                                                 |
| ---------- | --------------------------------------------------------------------------- |
| `d0`, `d1`, … | Dictionary codes.                                                        |
| Entry line | `- <code>: <token>`                                                         |

To read minified output, replace each code with its dictionary value. Codes can replace category/kind pairs and properties in `[symbols]`. They can also replace relationship kinds in `[edges]`. Uncertainty markers stay readable. For example, `S1 --d2[unresolved]--> S3` is an unresolved edge whose kind is the value of `d2`.

### Minified example

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

Here `S2` is a `callable/method` with the property `async=false`.

## Full example

The following output contains `[symbols]`, `[edges]`, and `[errors]`. It is not minified, so it has no `[dict]` section; see [the minified example](#minified-example).

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

## JSON output

`--format json`