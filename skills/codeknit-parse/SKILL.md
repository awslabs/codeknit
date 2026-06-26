---
name: codeknit-parse
description: "Extracts code structure (functions, classes, methods, variables) and relationships (calls, inheritance, containment) from source files into compact .skt or JSON output using codeknit. Use when analyzing a codebase, understanding code structure, mapping dependencies, or preparing context for code review and refactoring. Supports C, C++, C#, Go, Java, JavaScript, PHP, Python, Ruby, Rust, Scala, and TypeScript."
---

# codeknit — static code structure extraction

codeknit parses source code and produces a structural map of symbols and relationships designed for LLM consumption or structured tool integration.

## Important: choosing the right output mode

**Default to `directory-flat` (no flag needed) unless you are certain the input is a single file or just a few files.** Most projects have more files than you expect, and `inline` mode dumps everything to stdout which can overwhelm context windows and become unreadable. When in doubt, use the default — it writes chunked `.skt` files to `./skeleton/` that you can read selectively.

Only use `--output-mode inline` when:

- The input is a single file
- You explicitly need stdout output for piping
- You have confirmed the project has very few source files (< 5)

## Quick start

```bash
# Default: chunked .skt files in ./skeleton/ (best for most projects)
codeknit parse ./src

# Re-run on the same project (cleans previous output)
codeknit parse ./src --clean

# Single file to stdout
codeknit parse ./src/main.go --output-mode inline

# Machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json --edges

# Custom output directory
codeknit parse ./src ./output

# Mirror source tree structure
codeknit parse ./src --output-mode directory-tree
```

## Commands

### `codeknit parse <input-path> [output-dir]`

The output directory defaults to `./skeleton` for `directory-flat` and `directory-tree` modes. Pass a second argument to override it. In `inline` mode no directory is created. The default output format is `.skt`; pass `--format json` for machine-readable JSON. In directory modes, JSON is written as `codeknit.json`.

If the output directory already contains `.skt` files from a previous run, codeknit will refuse to write to avoid mixing stale and fresh output. Pass `--clean` to remove them automatically.

| Flag             | Default          | Description                                                           |
| ---------------- | ---------------- | --------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | `inline`, `directory-flat`, or `directory-tree`                       |
| `--format`       | `skt`            | Output format: `skt` or `json`                                        |
| `--max-lines`    | `500`            | Max lines per output file (flat/tree modes)                           |
| `--collect-test` | `false`          | Include test files in analysis                                        |
| `--minify`       | `false`          | Dictionary-based compression of repeated tokens                       |
| `--edges`        | `false`          | Include the [edges] section in output (off by default to save tokens) |
| `--clean`        | `false`          | Remove stale .skt files from output dir before writing                |
| `--workers`      | `NumCPU`         | Max concurrent parsing goroutines                                     |
| `--verbose`      | `false`          | Print progress and timing info                                        |

## Choosing an output mode

| Mode             | Best for                                | Output location                                                                       |
| ---------------- | --------------------------------------- | ------------------------------------------------------------------------------------- |
| `inline`         | A single file, piping to another tool   | Dumps everything to stdout — only use for single files or very small projects         |
| `directory-flat` | Most projects (default, recommended)    | `./skeleton/map_001.skt`, `map_002.skt`, ... — chunked, read only what you need       |
| `directory-tree` | Navigating output alongside source code | `./skeleton/<mirrored path>.skt` — one `.skt` per source file, mirrors the input tree |

| Format | Best for                           | Output                                                   |
| ------ | ---------------------------------- | -------------------------------------------------------- |
| `skt`  | LLM context and human inspection   | `.skt` files or stdout                                   |
| `json` | Scripts and structured integration | `codeknit.json` in directory modes, or stdout in `inline` |

Rules of thumb:

- **When unsure about project size → use the default (`directory-flat`), never `inline`**
- Analyzing a single file → `inline` is fine
- Anything beyond a few files → `directory-flat` or `directory-tree`
- When you want to look up a specific file's structure quickly → `directory-tree`
- When another tool needs structured data → `--format json`
- Add `--minify` on any mode to compress repeated tokens via a `dict.skt` dictionary
- By default the `[edges]` section is omitted to save tokens; pass `--edges` when you need relationship data (contains, calls, inherits, etc.)
- Add `--clean` when re-running on the same output directory
- Omit the output directory to use the default `./skeleton`; pass a path to override

## Reading output

The default `.skt` output format has sections: `[symbols]`, `[edges]`, optionally `[errors]` and `[dict]`. JSON output contains top-level `files`, `symbols`, optional `edges`, and optional `errors` arrays.

See [OUTPUT-FORMAT.md](OUTPUT-FORMAT.md) for the complete format reference with examples.

## Important: use codeknit output for analysis, not source code

When performing structural analysis (dependency mapping, refactoring planning, architecture review, etc.), **always read codeknit output** instead of the raw source code. The `.skt` skeleton files or JSON output contain the extracted structural information — symbols, relationships, line spans — in a compact format designed for exactly this purpose. Reading source files directly for structural questions wastes context and is less reliable.

Only read the actual source code when you need to inspect or modify implementation details (e.g. fixing a bug, reviewing logic, writing new code). For anything structural — "what calls what", "where is this class used", "show me the dependency graph" — codeknit output is the right tool.

## Graph analysis

Use `codeknit graph analyze` to run structural graph algorithms and get an LLM-readable report of code quality issues (cyclic dependencies, hub symbols, dead code, god classes, bottlenecks, and more).

```bash
# Run structural analysis
codeknit graph analyze ./src

# Custom output and thresholds
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15

# Show more results
codeknit graph analyze ./src --top-n 50
```

| Flag                      | Default                         | Description                                          |
| ------------------------- | ------------------------------- | ---------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Output `.skt` file path                              |
| `--collect-test`          | `false`                         | Include test files in analysis                       |
| `--workers`               | `NumCPU`                        | Max concurrent parsing goroutines                    |
| `--verbose`               | `false`                         | Print progress and timing info                       |
| `--fan-threshold`         | `10`                            | Min fan-in or fan-out to flag a hub symbol           |
| `--god-threshold`         | `15`                            | Min contains-edge count to flag a god class/function |
| `--max-inheritance-depth` | `5`                             | Flag inheritance chains deeper than this             |
| `--top-n`                 | `30`                            | Cap ranked output sections; 0 = no limit             |
| `--betweenness-threshold` | `0.001`                         | Min betweenness centrality value to report           |
| `--propagation-cutoff`    | `0.05`                          | Min probability to continue change propagation       |

The output is a `.skt` file — read it the same way you read parse output.

## Workflow

1. Run `codeknit parse ./src` to extract structure into `./skeleton/`
2. Read the `.skt` files selectively — start with the ones relevant to your task (do not read source files for structural questions)
3. Read `[symbols]` to see what exists and where (file, line span)
4. Read `[edges]` to trace relationships — calls, inheritance, containment
5. Use line spans (e.g. `L5-L20`) to locate exact source code only when you need to inspect or modify implementation
6. For deeper structural insights, run `codeknit graph analyze ./src` and read the resulting `.skt` report
7. For re-runs, add `--clean` to remove previous output: `codeknit parse ./src --clean`
8. For single-file quick inspection only, use `codeknit parse ./file.ts --output-mode inline`
9. For scripts or integrations, use `codeknit parse ./src --output-mode inline --format json --edges`
