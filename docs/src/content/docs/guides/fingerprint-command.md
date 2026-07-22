---
title: Fingerprint Command
description: Detect duplicate and near-duplicate code across files and languages using fuzzy hashing.
---

The `codeknit fingerprint` command detects duplicate and near-duplicate code across your codebase using **Context-Triggered Piecewise Hashing (CTPH)**. It works across files and even across programming languages by normalizing away variable names, string literals, and type annotations before computing structural fingerprints.

## What it does

`codeknit fingerprint` analyzes every function, method, variable, and type in your codebase and computes a **normalized structural fingerprint** based on:

- Control flow (`if`, `for`, `while`, `switch`)
- Operations (`=`, `+`, `==`, `&&`, `||`)
- Calls, returns, assignments, and object creation
- Language constructs like `try/catch`, `yield`, `await`, `defer`

This normalization means that **renamed copy-paste**, **trivial refactors**, and **equivalent logic in different languages** can still be detected as duplicates.

The algorithm uses **CTPH** (a rolling hash variant) to efficiently find near-duplicates. Similar code produces similar fingerprints, enabling fuzzy matching even when code has been slightly modified.

## Basic usage

```bash
codeknit fingerprint ./src
```

This command:

- Parses all source files in `./src`
- Computes structural fingerprints
- Outputs results to `./skeleton/fingerprints.skt`
- Reports matches with similarity between **65% and 95%** (default range)

## Flags

| Flag               | Default                       | Description                                                                                                                                                |
| ------------------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Output `.skt` file path                                                                                                                                    |
| `--min-similarity` | `65`                          | Minimum similarity percentage to report (0–100)                                                                                                            |
| `--max-similarity` | `95`                          | Maximum similarity percentage to report (0–100)                                                                                                            |
| `--show-all`       | `false`                       | Include the `[fingerprints]` section with raw token data                                                                                                   |
| `--rerank`         | `false`                       | Find semantic neighbors and rerank candidates using Ollama embeddings (requires: `ollama serve` and `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Ollama embedding model to use with `--rerank`                                                                                                              |
| `--collect-test`   | `false`                       | Include test files in analysis                                                                                                                             |
| `--workers`        | `NumCPU`                      | Max concurrent parsing goroutines (0 = use all CPU cores)                                                                                                  |
| `--verbose`        | `false`                       | Print progress information during processing                                                                                                               |

## Output format

The output is a `.skt` file with the following sections:

### `[duplicates]` (always present)

Lists pairs of symbols with similarity above the threshold:

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

Each line shows:

- Similarity percentage
- Left symbol (file path, scope, name)
- Right symbol (file path, scope, name)

### `[fingerprints]` (only with `--show-all`)

Contains raw fingerprint data for each symbol:

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Fields:

- Symbol name
- `FP:<version>:<hash1>:<hash2>` — CTPH fingerprint
- `tokens:<hex>` — normalized body token stream

This section is useful for debugging or building downstream tools.

## Common patterns

```bash
# Default scan
codeknit fingerprint ./src
```

```bash
# Find only exact duplicates
codeknit fingerprint ./src --min-similarity 100
```

```bash
# Find moderately similar code (e.g. same algorithm, different names)
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

```bash
# Use semantic matching to find additional candidates and reduce false positives
# Requires: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank
```

```bash
# Use a different embedding model for semantic matching
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

```bash
# Output full fingerprint listing (for analysis tools)
codeknit fingerprint ./src --show-all
```

```bash
# Custom output file
codeknit fingerprint ./src -o duplicates.skt
```

## Choosing a similarity range

| Range   | Guidance                                                                                 |
| ------- | ---------------------------------------------------------------------------------------- |
| 96–100% | Exact or near-exact structural duplicates. Almost certainly copy-paste.                  |
| 85–95%  | Near-duplicates. Usually copy-paste with minor edits (e.g. renamed vars, added logging). |
| 65–84%  | Default range. Strong structural similarity. Good candidates for refactoring.            |
| 50–64%  | Moderate similarity. Same algorithmic shape but different details. Review manually.      |
| < 50%   | Usually noise. Not meaningful duplication.                                               |

## Tips

- **Fingerprints measure structure, not meaning**: A high similarity score means the code _looks_ similar, not that it _does_ the same thing. Always review both symbols.
- **Use `--rerank` for semantic matching**: Embeddings add semantic neighbors that structural retrieval can miss and filter candidates that disagree semantically.
- **Short bodies are skipped**: Symbols with fewer than 4 normalized tokens (e.g. simple getters) are ignored to avoid noise.
- **Cross-language matching works**: Equivalent constructs (e.g., a Python function and a Go function with the same logic) can match, but language-specific patterns may produce spurious low-similarity matches.
- **A match is a signal, not a verdict**: Treat each match as a prompt to investigate — not automatic proof of duplication.
