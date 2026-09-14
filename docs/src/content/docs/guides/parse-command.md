---
title: Parse Command
description: Extract structural information from source code into .skt files or JSON.
---

The `codeknit parse` command extracts structural information from your codebase — such as functions, classes, methods, variables, and their relationships — and outputs it in compact `.skt` format by default. Use JSON when you need machine-readable output for scripts, integrations, or downstream tools.

## Basic Usage

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Path to the directory or file you want to parse.
- **`[output-dir]`**: Optional output directory. If not provided, defaults to `./skeleton`.

### Examples

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## Output Modes

Use `--output-mode` to control how output is structured. Three modes are available:

| Mode             | Description                                                                              | Best For                                            |
| ---------------- | ---------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `directory-flat` | Writes chunked `.skt` files (e.g. `map_001.skt`, `map_002.skt`) to the output directory. | ✅ **Most projects** — default and recommended mode |
| `directory-tree` | Mirrors the source directory structure, creating one `.skt` file per source file.        | Navigating output alongside source code             |
| `inline`         | Dumps all output to stdout.                                                              | Single files or piping to other tools               |

> **Tip**: Default to `directory-flat` unless you're working with a single file. Avoid `inline` for large inputs as it can overwhelm context windows.

## Flags

| Flag             | Default          | Description                                                                  |
| ---------------- | ---------------- | ---------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Output mode: `inline`, `directory-flat`, or `directory-tree`                 |
| `--format`       | `skt`            | Output format: `skt` or `json`                                               |
| `--max-lines`    | `500`            | Tree source-chunk limit (minimum 3); flat keeps source blocks intact                             |
| `--collect-test` | `false`          | Include test files in analysis                                               |
| `--minify`       | `false`          | Enable dictionary-based compression to reduce token usage                    |
| `--edges`        | `false`          | Include the `[edges]` section with relationship data (calls, contains, etc.) |
| `--clean`        | `false`          | Replace previous generated output only after successful generation          |
| `--workers`      | `NumCPU`         | Maximum number of concurrent parsing goroutines (0 = use all CPU cores)      |
| `--verbose`      | `false`          | Print progress and timing information during processing                      |

## Common Patterns

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

Example JSON output:

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

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## Relationship uncertainty

With `--edges`, Codeknit preserves relationships even when syntax and scope do not establish a unique target:

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

An omitted status means resolved by the analyzer. `unresolved` means the dependency could not be established; `ambiguous` means competing interpretations remain possible. Candidates are possible targets, not confirmed connections. Known endpoints use short IDs; unresolved targets without symbols may appear as names.

JSON carries the same information in `resolution` and `candidates`, and omits `to_short` for uncertain targets. C/C++ includes have `meta/file` and `meta/include` symbols so edges use short IDs; this records written directives without claiming compiler-level header resolution.

Dependency metrics exclude uncertain relationships. The HTML graph shows their count without drawing individual uncertain links. See the [output format reference](../../reference/output-format/) for the grammar and the [language support reference](../../reference/language-support/) for resolution limits.

## Stale Output Protection

If the output directory already contains `.skt` files from a previous run, `codeknit` will refuse to write new output to prevent mixing stale and fresh data.

To replace previous generated output only after successful generation, use the `--clean` flag:

```bash
codeknit parse ./src --clean
```

Generation is staged, with backup/rollback and next-run recovery for interrupted publication. Unrelated output-directory entries are preserved unless they conflict with generated paths. JSON mode applies stale-output protection to `codeknit.json`.

When upgrading to `0.5.0`, regenerate the entire output set. Symbol IDs and relationship results can change; do not mix chunks from different runs.

## Tips

- ✅ **Default to `directory-flat`** for most projects. It balances readability and manageability.
- 🔍 Use `--minify` on large codebases to reduce token usage via a shared dictionary (`dict.skt`).
- 🔗 The `[edges]` section is **excluded by default** to save tokens. Use `--edges` when you need relationship data like `calls`, `contains`, or `inherits`.
- 🧾 Use `--format json` when a script or integration needs structured data instead of `.skt`.
- 🧹 Always use `--clean` when re-running on the same output directory.
- 📁 Use `directory-tree` if you want to correlate `.skt` files directly with source files in your editor.

## Scan and output reliability

Unreleased changes after 0.5.0 add repeatable `--include GLOB` / `--exclude GLOB`, `--max-file-size BYTES` (10 MiB default, `0` unlimited), and `--strict`. Quote input-relative globs. Includes do not override ignore rules or test filtering. Symlinks are skipped and reported. Partial results include diagnostics; strict mode fails before output publication. Tree files retain source extensions (`foo.c.skt`, `foo.h.skt`); parts use `foo.c_partN.skt` and repeat source headers. See [analysis reliability](../../reference/analysis-reliability/) for the full contract.
