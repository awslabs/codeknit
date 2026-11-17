---
title: Fingerprint Command
description: Detect duplicate and near-duplicate code across files and languages using fuzzy hashing.
---

`codeknit fingerprint` computes a Context-Triggered Piecewise Hash (CTPH) over the normalized body tokens of every function, method, variable, and type. Variable names, string literals, and type annotations are normalized away before hashing. Renamed copy-paste and ports to other languages can therefore still match.

This guide shows how to find duplicates, tune the similarity range, reduce false positives, and decide what to refactor.

## Before you start

- `codeknit` is installed. See [Installation](../../installation/).
- You have a source directory to scan, such as `./src`.

## Find duplicates in a project

Run the command on the directory you want to check:

```bash
codeknit fingerprint ./src
```

The command writes `./skeleton/fingerprints.skt` and reports pairs with similarity from 65% to 95%.

Open the file. The `[duplicates]` section is always present, with the strongest matches first:

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

Each line has the form `similarity:<N>%  <file>::<symbol> <-> <file>::<symbol>`. If nothing falls in the range, the section contains only `# no duplicates found`.

For the full output grammar, see the [output format reference](../../reference/output-format/).

## Find exact duplicates only

The default range tops out at 95%, so 100% matches are excluded. To list exact matches, set both bounds to 100:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

"Exact" means the complete normalized token streams are identical within the same symbol category. It does not mean the source text is identical, or that the two symbols are proven equivalent.

## Widen the range to find looser matches

Lower the minimum to include code that shares an algorithmic shape but differs in detail:

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

Results below 50% are usually noise.

## Reduce false positives with semantic reranking

If the structural results are noisy, rerank candidates with embeddings from a local Ollama server.

1. Start Ollama and pull the default embedding model:

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. Run the command with `--rerank`:

   ```bash
   codeknit fingerprint ./src --rerank
   ```

To use another embedding model, pass `--model`:

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

Reranking adds semantic candidates and filters on a combined score. It reduces false positives but does not remove all of them. It can also remove a candidate that is structurally exact, and the similarity range still applies. Embeddings are computed from the same parsed source snapshot used for extraction.

## Inspect the raw fingerprints

To see the fingerprint of every symbol, add `--show-all`:

```bash
codeknit fingerprint ./src --show-all
```

This adds a `[fingerprints]` section, grouped by source file, one line per symbol:

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

The line format is `<symbol-name>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`. `tokens` is the hex-encoded normalized token stream.

## Write the report somewhere else

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## Control which files are scanned

Test files are skipped by default. Include them with `--collect-test`:

```bash
codeknit fingerprint ./src --collect-test
```

To narrow the scan, use the source-selection flags. Quote the patterns so your shell does not expand them:

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

Other selection controls:

- `--max-file-size BYTES` skips larger files. The default is `10485760` (10 MiB). `0` disables the limit.
- `--strict` fails before writing output if the scan or parse is incomplete.
- `--header-language c|cpp` selects the parser for ambiguous `.h` files. See [Language support](../../reference/language-support/#header-language-selection).

Skipped symlinks, oversized files, and syntax warnings appear on stderr. Includes do not override `.gitignore` or test filtering, and excludes take precedence over includes. See the [analysis reliability reference](../../reference/analysis-reliability/) for exact guarantees.

## Choose a similarity range

| Range     | What it usually surfaces                                      |
| --------- | ------------------------------------------------------------- |
| `96-100%` | Exact or near-exact structural duplicates; clear copy-paste   |
| `85-95%`  | Near-duplicates; usually copy-paste with trivial edits        |
| `65-84%`  | Default range; strong structural similarity, often refactor candidates |
| `50-64%`  | Same algorithmic shape with different details                 |
| Below 50% | Usually noise                                                 |

To see the top two rows, raise `--max-similarity` to `100`.

## Decide what to refactor

1. Start with the default range and read the top pairs.
2. Open both symbols of each pair and compare them.
3. Treat a match as a signal, not proof. Fingerprints measure structural shape, not identifier names or intent. The same logic written iteratively and recursively scores low, and unrelated code with the same shape can score high.
4. Pairs in the 85–95% range are usually the safest candidates to consolidate. Pairs in the 65–84% range often need a closer look.
5. If the list is noisy, narrow the range or add `--rerank`.

Cross-language matches work for equivalent constructs. Language-specific features can produce spurious low-similarity matches.

Similarity percentages are a heuristic ranking, not an edit distance. The project's labeled calibration fixture reports precision and recall separately from similarity scores. It is a regression check, not an estimate of accuracy on your code.

## Large projects

- Below 500 eligible symbols, codeknit scores every same-category pair.
- At 500 or more, complete exact normalized-token groups are always retained. Other pairs use approximate top-50 neighbors per symbol, so near-duplicate results can be incomplete.
- For exact duplicates, use the 100/100 command above. It is not affected by the approximation.
- Ordering is deterministic for stable inputs and scores.

## Troubleshooting

| Symptom | Cause and fix |
| ------- | ------------- |
| `# no duplicates found` | No pair falls in the range. Lower `--min-similarity` or raise `--max-similarity`. |
| A known short function is missing | Symbols with fewer than 4 normalized tokens are skipped. |
| Exact copies are missing | The default maximum is 95%. Use `--min-similarity 100 --max-similarity 100`. With `--rerank`, the semantic filter can also remove them. |
| `--rerank` fails | Check that `ollama serve` is running and that the model is pulled. |
| Partial results with warnings on stderr | Fix the reported files, or add `--strict` to fail instead of publishing incomplete output. |

## Flags used in this guide

| Flag               | Default                       | Description                                                                 |
| ------------------ | ----------------------------- | --------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Output `.skt` file path.                                                    |
| `--min-similarity` | `65`                          | Minimum similarity percentage to report (0–100).                            |
| `--max-similarity` | `95`                          | Maximum similarity percentage to report (0–100).                            |
| `--show-all`       | `false`                       | Also emit the `[fingerprints]` section with raw token data.                 |
| `--rerank`         | `false`                       | Rerank candidates with semantic embeddings via Ollama. Requires `ollama serve` and `ollama pull qwen3-embedding:0.6b`. |
| `--model`          | `qwen3-embedding:0.6b`        | Ollama embedding model to use with `--rerank`.                              |
| `--collect-test`   | `false`                       | Include test files in analysis.                                             |
| `--workers`        | `NumCPU`                      | Maximum concurrent parsing goroutines.                                      |
| `--verbose`        | `false`                       | Print progress information.                                                 |

For the complete list, including `--include`, `--exclude`, `--max-file-size`, `--strict`, and `--header-language`, see the [CLI flags reference](../../reference/cli-flags/).

## Related

- [Parse command](../parse-command/)
- [Graph commands](../graph-commands/)