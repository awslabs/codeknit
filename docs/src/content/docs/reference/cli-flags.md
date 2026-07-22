---
title: CLI Reference
description: Complete reference for all codeknit commands and flags.
---

## codeknit

Launches the interactive terminal UI (TUI), which guides you through available commands and options.

```bash
codeknit
```

## codeknit parse

Extract structural information from source code into `.skt` files or JSON.

```bash
codeknit parse <input-path> [output-dir]
```

| Flag             | Type   | Default          | Description                                                                            |
| ---------------- | ------ | ---------------- | -------------------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat` | Output mode: `inline`, `directory-flat`, or `directory-tree`                           |
| `--format`       | string | `skt`            | Output format: `skt` or `json`                                                        |
| `--max-lines`    | int    | `500`            | Maximum lines per output file (applies to `directory-flat` and `directory-tree` modes) |
| `--collect-test` | bool   | `false`          | Include test files in analysis                                                         |
| `--minify`       | bool   | `false`          | Enable dictionary-based output minification                                            |
| `--edges`        | bool   | `false`          | Include the `[edges]` section in output (off by default to save tokens)                |
| `--clean`        | bool   | `false`          | Remove stale `.skt` files from the output directory before writing                     |
| `--workers`      | int    | `0` (NumCPU)     | Maximum concurrent parsing goroutines                                                  |
| `--verbose`      | bool   | `false`          | Print progress information during processing                                           |

The output directory defaults to `./skeleton` when not specified. In `inline` mode, output is written to stdout and no directory is used. With `--format json`, directory output is written as `codeknit.json`.

## codeknit graph show

Generate an interactive HTML graph visualization of the codebase structure.

```bash
codeknit graph show <input-path>
```

| Flag             | Type   | Default                          | Description                                  |
| ---------------- | ------ | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | Output HTML file path                        |
| `--collect-test` | bool   | `false`                          | Include test files in analysis               |
| `--workers`      | int    | `0` (NumCPU)                     | Maximum concurrent parsing goroutines        |
| `--verbose`      | bool   | `false`                          | Print progress information during processing |

The generated HTML file is self-contained and opens automatically in your default browser.

## codeknit graph analyze

Run structural analysis algorithms and emit an LLM-readable `.skt` report.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Type    | Default                         | Description                                                   |
| ------------------------- | ------- | ------------------------------- | ------------------------------------------------------------- |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Output `.skt` file path                                       |
| `--collect-test`          | bool    | `false`                         | Include test files in analysis                                |
| `--workers`               | int     | `0` (NumCPU)                    | Maximum concurrent parsing goroutines                         |
| `--verbose`               | bool    | `false`                         | Print progress information during processing                  |
| `--fan-threshold`         | int     | `10`                            | Minimum fan-in or fan-out to flag a hub symbol                |
| `--god-threshold`         | int     | `15`                            | Minimum contains-edge count to flag a god class/function      |
| `--max-inheritance-depth` | int     | `5`                             | Flag inheritance chains deeper than this value                |
| `--top-n`                 | int     | `30`                            | Cap ranked output sections; `0` means no limit                |
| `--betweenness-threshold` | float64 | `0.001`                         | Minimum betweenness centrality value to report                |
| `--propagation-cutoff`    | float64 | `0.05`                          | Minimum probability to continue change propagation simulation |

## codeknit graph hotspots

Rank files using Git history and structural importance, and report temporal
coupling between files that repeatedly change together.

```bash
codeknit graph hotspots <input-path>
```

| Flag                     | Type   | Default                   | Description                                      |
| ------------------------ | ------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | string | `./skeleton/hotspots.skt` | Output file path                                 |
| `--format`               | string | `skt`                     | Output format: `skt` or `json`                   |
| `--since`                | string | `12mo`                    | History window, such as `180d`, `12mo`, or `2y`  |
| `--max-commits`          | int    | `2000`                    | Maximum commits to inspect                       |
| `--max-files-per-commit` | int    | `50`                      | Exclude commits changing more files              |
| `--min-cochanges`        | int    | `3`                       | Minimum shared commits for temporal coupling     |
| `--top-n`                | int    | `30`                      | Maximum results per report section               |
| `--include-merges`       | bool   | `false`                   | Include merge commits                            |
| `--collect-test`         | bool   | `false`                   | Include test files                               |
| `--workers`              | int    | `0` (NumCPU)              | Maximum concurrent parsing goroutines            |
| `--verbose`              | bool   | `false`                   | Print progress information                       |

## codeknit fingerprint

Detect duplicate and near-duplicate code using fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Type   | Default                       | Description                                                                                                                  |
| ------------------ | ------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | Output `.skt` file path                                                                                                      |
| `--min-similarity` | int    | `65`                          | Minimum similarity percentage to report (0–100)                                                                              |
| `--max-similarity` | int    | `95`                          | Maximum similarity percentage to report (0–100)                                                                              |
| `--show-all`       | bool   | `false`                       | Include the `[fingerprints]` section with raw token data                                                                     |
| `--rerank`         | bool   | `false`                       | Find semantic neighbors and rerank candidates using Ollama embeddings (requires `ollama serve` and `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`        | Ollama embedding model to use with `--rerank`                                                                                |
| `--collect-test`   | bool   | `false`                       | Include test files in analysis                                                                                               |
| `--workers`        | int    | `0` (NumCPU)                  | Maximum concurrent parsing goroutines                                                                                        |
| `--verbose`        | bool   | `false`                       | Print progress information during processing                                                                                 |

## codeknit completion

Generate shell completion scripts for supported shells.

```bash
codeknit completion <shell>
```

Supported shells: `bash`, `zsh`, `fish`, `powershell`.

## Global flags

| Flag           | Description                       |
| -------------- | --------------------------------- |
| `--version`    | Print version information         |
| `--help`, `-h` | Show help for the current command |
