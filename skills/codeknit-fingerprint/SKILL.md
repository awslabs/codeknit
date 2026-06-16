---
name: codeknit-fingerprint
description: "Detects duplicate and near-duplicate code across a codebase using fuzzy hashing with codeknit. Use when finding copy-paste, refactoring candidates, merging similar implementations, auditing for DRY violations, or locating semantically equivalent code across different files or languages. Supports C, C++, C#, Go, Java, JavaScript, PHP, Python, Ruby, Rust, Scala, and TypeScript."
---

# codeknit — fuzzy duplicate code detection

codeknit fingerprint computes Context-Triggered Piecewise Hashes (CTPH) over a normalized token stream for every function, method, variable, and type in a codebase. Similar code produces similar fingerprints, so near-duplicates are detected even across files and across programming languages.

Token normalization strips variable names, string literals, and type annotations before hashing, so `getUserById(id int)` and `fetch_user(user_id uint64)` compare structurally.

## When to use this skill

Use `codeknit fingerprint` when the user asks to:

- Find duplicate or near-duplicate code
- Identify refactoring candidates (copy-paste, DRY violations)
- Spot semantically equivalent code across files or languages
- Audit a codebase before a large refactor
- Compare two implementations for structural similarity

Do not use it for: "what does this function do", "find callers of X", or other structural questions. Use `codeknit-parse` for those.

## Quick start

```bash
# Default: find near-duplicates with 65-95% similarity
codeknit fingerprint ./src

# Exact duplicates only
codeknit fingerprint ./src --min-similarity 100

# Moderately similar code (50-80%)
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80

# Semantic reranking — filters false positives via Ollama embeddings
# requires: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank

# Semantic reranking with a different model
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b

# Include the full per-symbol fingerprint listing in output
codeknit fingerprint ./src --show-all

# Custom output path
codeknit fingerprint ./src -o duplicates.skt
```

Output is written to `./skeleton/fingerprints.skt` by default.

## Command reference

### `codeknit fingerprint <input-path>`

| Flag               | Default                       | Description                                                                                                                            |
| ------------------ | ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Output `.skt` file path                                                                                                                |
| `--min-similarity` | `65`                          | Minimum similarity percentage to report (0-100)                                                                                        |
| `--max-similarity` | `95`                          | Maximum similarity percentage to report (0-100)                                                                                        |
| `--show-all`       | `false`                       | Also emit the `[fingerprints]` section with raw token data                                                                             |
| `--rerank`         | `false`                       | Rerank CTPH candidates with semantic embeddings via Ollama to eliminate false positives (requires: `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Ollama embedding model to use with `--rerank`                                                                                          |
| `--collect-test`   | `false`                       | Include test files in analysis                                                                                                         |
| `--workers`        | `NumCPU`                      | Max concurrent parsing goroutines                                                                                                      |
| `--verbose`        | `false`                       | Print progress and timing info                                                                                                         |

## Output format

The `.skt` file always contains a `[duplicates]` section, and optionally a `[fingerprints]` section when `--show-all` is passed.

### `[duplicates]` section

One line per duplicate pair, sorted by similarity descending:

```
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

Format: `similarity:<N>%  <file>::<symbol> <-> <file>::<symbol>`

When no duplicates match the similarity range, the section contains a single comment: `# no duplicates found`.

### `[fingerprints]` section (with `--show-all`)

Grouped by source file. One line per symbol:

```
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
refreshToken   FP:3:b2c3d4...:e5f6a7...  tokens:8e0f1a3c...
```

Format: `<symbol-name>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`

- `FP` is the CTPH fingerprint — two hashes at two block sizes for similarity comparison
- `tokens` is the hex-encoded normalized token stream

## Choosing a similarity range

| Range     | What it surfaces                                                           |
| --------- | -------------------------------------------------------------------------- |
| `96-100%` | Exact or near-exact structural duplicates — clear copy-paste candidates    |
| `85-95%`  | Near-duplicates — likely copy-paste with trivial edits (renamed variables) |
| `65-84%`  | Default range — strong structural similarity, often refactor candidates    |
| `50-64%`  | Moderate similarity — same algorithmic shape, different details            |
| `<50%`    | Weak similarity — usually noise, rarely actionable                         |

**Rules of thumb:**

- Start with the default range (65-95%) — it's tuned to surface real duplicates without noise
- Use `--rerank` to filter false positives via semantic embeddings when the CTPH results are noisy
- Drop to 50-80% only when hunting for conceptual duplicates across divergent implementations
- Use `--min-similarity 100` for a quick sanity check before merging two branches or after a bulk copy

## Workflow

1. Run `codeknit fingerprint ./src` to produce `./skeleton/fingerprints.skt`
2. Read the `[duplicates]` section — pairs are sorted with the strongest matches first
3. For each flagged pair, open both files at the reported symbol names and compare
4. If the similarity range looks too strict or too loose, re-run with adjusted `--min-similarity` / `--max-similarity`
5. If you see too many false positives, re-run with `--rerank` to filter via semantic embeddings
6. For an audit of the full fingerprint set (e.g. custom tooling that compares specific pairs), re-run with `--show-all`

## Important: what fingerprints measure

Fingerprints capture **structural shape**, not identifier names or semantic intent.

- Two functions with identical structure but different names and types get **high similarity**
- Two functions with the same intent but different control flow (iterative vs recursive) get **low similarity**
- Docstrings, comments, and whitespace are ignored — they don't appear in the token stream
- Variable names and string literals are normalized away — `x = "foo"` and `name = "bar"` are indistinguishable

This makes fingerprint detection robust against rename-refactors and cross-language copy-paste, but it means a match is a **signal to investigate**, not a proof of duplication. Always read both symbols before making a refactoring decision.

## Limitations

- Symbols with very short bodies (fewer than 4 tokens after normalization) are skipped — trivial getters and one-liners don't produce useful fingerprints
- Cross-language matches work for equivalent constructs (a loop is a loop), but language-specific constructs (Go channels, Python decorators) can produce spurious low-similarity matches
- The similarity score is a heuristic over a rolling hash, not an edit distance on source — treat percentages as a ranking, not an exact metric
