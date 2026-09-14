---
title: Output Modes
description: Choose the right output mode for your project size and workflow.
---

codeknit supports three output modes, controlled by the `--output-mode` flag. Each mode determines how the extracted code structure is written to disk (or stdout).

Output mode is separate from output format. The default format is `.skt`; pass `--format json` to emit the same parse result as machine-readable JSON. In directory modes, JSON is written to `codeknit.json`. In `inline` mode, JSON is written to stdout.

### directory-flat (default, recommended)

- **Behavior**: Writes chunked `.skt` files such as `map_001.skt`, `map_002.skt`, etc.
- **Output directory**: `./skeleton/` by default
- **Splitting**: Files split between source blocks near `--max-lines` (default: 500); one large source block can exceed this target
- **Use case**: Best for most projects. Keeps output organized and readable by limiting file size. You can read only the chunks relevant to your task.
- **Minification**: When `--minify` is enabled, a `dict.skt` file is also generated in the output directory, containing token mappings for compressed values.

Example:

```bash
codeknit parse ./src
# Output: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **Behavior**: Mirrors the source directory structure exactly.
- **Output directory**: `./skeleton/` by default
- **Mapping**: Append `.skt` to the complete source name: `foo.c.skt` and `foo.h.skt` stay distinct. Large files split into `foo.c_part1.skt`, `foo.c_part2.skt`, and so on.
- **Use case**: Ideal when you want to quickly look up the structure of a specific file. Useful for navigation alongside the original codebase.

Example:

```bash
codeknit parse ./src --output-mode directory-tree
# Source: ./src/handler.go and ./src/pkg/db.go
# Output: ./skeleton/handler.go.skt, ./skeleton/pkg/db.go.skt
```

Tree SKT requires `--max-lines` of at least 3. Each symbol chunk repeats its source header; edge-only chunks contain `[edges]`. Source chunks respect the limit, while dictionaries and warning files are separate auxiliary output. These naming and splitting changes follow 0.5.0 and are unreleased.

Directory generation uses staging and replaces old generated output with `--clean` only after successful generation. Existing results survive failed writes; interrupted publication can recover on the next run. See [analysis reliability](../../reference/analysis-reliability/) for the precise guarantees.

### inline

- **Behavior**: Dumps all output to stdout.
- **Output directory**: None created
- **Use case**: Only recommended for single files or very small projects (fewer than 5 files). Useful when piping output to another tool or inspecting a single file interactively.

Example:

```bash
codeknit parse ./src/main.go --output-mode inline
# Output: printed directly to terminal
```

### JSON format

- **Behavior**: Emits a single JSON document containing `files`, `symbols`, optional `edges`, and optional `errors`.
- **Output location**: `codeknit.json` in directory modes, or stdout in `inline` mode.
- **Use case**: Best for scripts, editor integrations, CI checks, and tools that need structured data.

Example:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Sample output:

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

### Decision Table

| Mode             | Best for                                | Output location                                     |
| ---------------- | --------------------------------------- | --------------------------------------------------- |
| `directory-flat` | Most projects (default, recommended)    | `./skeleton/map_001.skt`, `map_002.skt`, ...        |
| `directory-tree` | Navigating output alongside source code | `./skeleton/<mirrored path>.skt`                    |
| `inline`         | Single file, piping to another tool     | stdout — only use for single files or tiny projects |

| Format | Best for                           | Output                                                   |
| ------ | ---------------------------------- | -------------------------------------------------------- |
| `skt`  | LLM context and human inspection   | `.skt` files or stdout                                   |
| `json` | Scripts and structured integration | `codeknit.json` in directory modes, or stdout in `inline` |

### Rules of Thumb

- **When unsure** → use `directory-flat` (the default)
- **Single file inspection** → `inline` is acceptable
- **More than a few files** → prefer `directory-flat` or `directory-tree`
- **Large codebases** → add `--minify` to reduce token usage
- **Re-running on same output** → use `--clean` to remove stale `.skt` files

### Minification

The `--minify` flag enables dictionary-based compression of repeated tokens (e.g., property keys like `exported`, `async`, or common type names). When enabled:

- Repeated values are replaced with short codes (`d0`, `d1`, `d2`, ...)
- A `dict.skt` file is written to the output directory, mapping codes to original values
- Significantly reduces output size for large codebases
- Works in both `directory-flat` and `directory-tree` modes

Example minified output:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```

This format preserves full information while minimizing token footprint, making it ideal for LLM-based analysis.
