---
title: Language Support and API Fidelity
description: Tested extraction capabilities, parser selection, and known language limitations.
---

<!-- Maintained in docs/reference/language-support.md; copied verbatim by docs/gen. -->

Codeknit extracts syntax without running a compiler, evaluating macros, or loading a project's dependencies. The matrix below describes tested constructs, not complete language conformance or guaranteed cross-file name resolution.

Every language has an exact basic parameter/return-signature assertion in `internal/plugin/api_fidelity_test.go`. Declaration coverage is also checked in `internal/plugin/coverage_test.go` and the individual language test packages.

| Language   | Extensions                                  | Signature baseline                                                                                           | Declaration and container coverage                                                                         |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Named parameters and return types in definitions and prototypes                                              | Structs, unions, enums, typedefs, macros; declarations inside header guards                                |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Named parameters and return types in functions, prototypes, and method declarations                          | Classes, structs, methods, namespaces, templates; header guards and `extern "C"` wrappers                  |
| C#         | `.cs`                                       | Named typed parameters and explicit return types                                                             | Classes, structs, interfaces, methods, fields, delegates, namespaces                                       |
| Go         | `.go`                                       | Grouped, unnamed, variadic, function-typed and generic parameters; named/multiple results; generic receivers | Struct fields with types and tags, embedded fields, interface method signatures, type parameters, packages |
| Java       | `.java`                                     | Named typed parameters and explicit return types                                                             | Classes, interfaces, constructors, methods, fields, nested classes, packages                               |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Named parameters; no inferred return type                                                                    | Functions, arrow functions, classes, fields, methods, JSX-containing declarations                          |
| PHP        | `.php`                                      | Named typed parameters and explicit return types                                                             | Classes, interfaces, traits, enums, methods, properties, namespaces                                        |
| Python     | `.py`, `.pyi`                               | Annotations, defaults, `*args`, `**kwargs`, positional/keyword separators, return annotations                | Functions, classes, nested classes, decorated methods; conventional bound receivers omitted                |
| Ruby       | `.rb`                                       | Named parameters; no inferred return type                                                                    | Classes, modules, instance/singleton methods, constants                                                    |
| Rust       | `.rs`                                       | Typed patterns, `self`/`&self`/`&mut self`, mutable bindings, explicit return types                          | Inline and nested modules, named struct fields, trait methods and implementation methods                   |
| Scala      | `.scala`, `.sc`                             | Named typed parameters and explicit return types                                                             | Classes, traits, objects, methods, enums, packages                                                         |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Named typed parameters and explicit return annotations                                                       | Functions, arrow functions, interfaces, classes, fields, namespaces; TSX uses its own grammar              |

## Header language selection

All source-processing CLI commands accept `--header-language c|cpp`. The choice applies only to `.h` files. Other extensions retain their usual parsers.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

The default is `c`, preserving C macro, typedef, and union extraction. Use `cpp` for C++ headers named `.h`; files named `.hpp` and `.hxx` already use C++. The choice is explicit because both grammars can accept some of the other language's syntax without extracting its full structure.

Header guards and conditional branches are traversed without evaluating their conditions. Results can therefore include declarations from mutually exclusive branches. Macro expansion, include resolution, and build flags are not applied. The TUI uses the default C selection for `.h` files.

## Preserving API shape

For example, Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

produces:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Go interface methods and Go/Rust named fields appear as symbols in both normal and fingerprint modes. Methods and fields retain their owning type. Nested Rust modules retain qualified names such as `service.storage.read`, including when another module declares a same-named function.

Rust permits fields and methods with the same name. Field scoped names carry a `#field` suffix internally so their IDs and containment edges remain distinct; display names and signatures retain the original field name.

Python signatures retain annotations, defaults, splats, and calling-convention separators. A conventional first `self` or `cls` parameter is omitted from a bound method, but preserved in a free function or static method.

Multiline signatures retain source line breaks in JSON. SKT escapes line breaks as `\n` and `\r` so that each symbol stays on one physical output line.

## Relationship resolution

Relationships retain receiver qualification and lexical ownership. Relative imports and renamed JavaScript/TypeScript and Python imports provide target context. Go and Java package declarations can resolve across files in the same directory. An unmatched import never falls back to an unrelated declaration, and competing candidates remain ambiguous.

Local receiver bindings take precedence over namespace imports, including when a parameter shadows an imported namespace. Python package imports consider declarations in the package's `__init__` module, not arbitrary descendant modules. Re-exports that cannot be established from that module remain unresolved.

Java, C++, C#, Scala, and TypeScript callable extraction retains declaration byte ranges so overloaded bodies can have distinct callers and alias scopes, including declarations on the same line. Selecting an overloaded target remains conservative. Explicit override relationships in Java, C++, C#, Scala, and TypeScript look for the nearest inherited declaration; missing bases, cycles, or competing candidates retain uncertainty instead of producing self-links.

This is syntax-based analysis, not a compiler or runtime call graph. Bare JavaScript package specifiers, package re-exports, build configurations, complex receiver types, and dynamic dispatch may remain unresolved. A unique name elsewhere in the repository is not sufficient evidence of a dependency.

JSON preserves uncertain edges with `resolution: "unresolved"` or `"ambiguous"` and includes candidate IDs when available. These edges have no target short ID.

SKT uses the same edge notation for all relationships, with an optional resolution marker:

```text
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

Known endpoints and candidates use short IDs, including endpoints of unresolved relationships. Bare uncertain targets without a symbol are names relative to the source file. Literal names that look like short IDs, and other complex identities, use quoted full IDs to avoid ambiguity. Edges with the same source, kind, resolution, and candidate set share one target list. The SKT reader preserves resolution and candidates, including references to symbols in later output chunks. Minified output encodes the relationship kind while keeping the resolution marker readable.

C/C++ includes have explicit `meta/file` and `meta/include` symbols. Header spellings appear once in the symbol list; edges use their short IDs:

```text
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

An include symbol describes the written directive, not a resolved header declaration. Repeated includes within a file share an endpoint. File/include metadata is excluded from dependency rankings and dead-code findings; adding IDs does not claim that a header has been found through the compiler's include search configuration.

Dependency metrics and HTML graph links use resolved edges only. Reports show excluded relationship counts, and the HTML graph displays an uncertainty count. Fan-in and fan-out count distinct neighbors. Inheritance analysis considers all bases and interfaces; cyclic inheritance has no reported depth. Reachability starts from named `main`, `Main`, or `init` callables and follows dependency edges. Reaching a member makes its container relevant without making every sibling reachable. Without recognized entry points, reachability findings are omitted.

Callback arguments and returned callables are references, not proof of invocation. Alias resolution retains file, function, and object identity; conflicting assignments remain uncertain. Graph findings are review candidates: external callers and dynamic behavior can change the result. Directory-based layer rules and change-propagation weights are heuristics, not measured probabilities.

## Known limits

- Complex C/C++ pointer, reference, array, and function-pointer declarators can still have incomplete signatures. Macro-generated declarations are not expanded.
- TypeScript/JavaScript constructor handling and some optional, rest, or destructured parameter forms are incomplete.
- Advanced generic constraints and type forms outside the tested cases vary by language. Return types are not inferred.
- Rust macros are not expanded. Out-of-line module declarations are recorded; their files must be included in the scan. Tuple fields, enum variant details, and associated items have partial coverage.
- Syntax extraction does not resolve dynamic dispatch, overloads, imported dependencies, or build-configuration semantics with compiler precision.
- Partial syntax errors can still omit declarations. Inspect reported warnings and verify source before applying a refactor.

More complete extraction changes symbol counts and short IDs. Regenerate existing skeleton output when upgrading instead of mixing it with older output.
