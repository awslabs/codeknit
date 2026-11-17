---
name: codeknit
description: "This skill should be used when the user asks to \"map the codebase\", \"show the code structure\", \"what calls X\", \"find dependencies\", \"find dead code\", \"find cyclic dependencies\", \"analyze the architecture\", \"find duplicate code\", \"find copy-paste\", \"find similar functions\", \"check for DRY violations\", or \"find refactoring candidates\". Runs codeknit to extract symbols and relationships (parse), report structural issues (graph analyze), and detect near-duplicate code across files and languages (fingerprint) for C, C++, C#, Go, Java, JavaScript, PHP, Python, Ruby, Rust, Scala, and TypeScript."
---

# codeknit — code structure, graph analysis, and duplicate detection

codeknit parses source code with tree-sitter and writes compact `.skt` files (or JSON) designed for LLM context. Pick the command by question:

| Question                                                         | Command                    |
| ---------------------------------------------------------------- | -------------------------- |
| What exists, where, and what calls/inherits/contains what        | `codeknit parse`           |
| Cycles, hubs, dead code, god classes, bottlenecks, change impact | `codeknit graph analyze`   |
| Copy-paste, near-duplicates, DRY violations, refactor candidates | `codeknit fingerprint`     |

## Setup: ensure the binary is available

Before the first codeknit command, run the bundled installer with the absolute path of this skill's directory:

```bash
bash <this-skill-dir>/scripts/install-codeknit.sh
```

It exits immediately when `codeknit` is already on `PATH` or in `~/.local/bin`. Otherwise it downloads the latest GitHub release for the current OS/arch, verifies the SHA-256 against the release `checksums.txt`, and installs it to `~/.local/bin`. The **last line of stdout is the absolute binary path**; when that directory is not on `PATH`, invoke codeknit by that path for the rest of the session. Tell the user when a download happens. Set `CODEKNIT_VERSION=0.6.0` to pin a release or `CODEKNIT_INSTALL_DIR` to change the location. If the download fails (offline, unsupported platform), point the user to https://github.com/awslabs/codeknit/releases.

## Read codeknit output, not source, for structural questions

For dependency mapping, refactoring plans, or architecture review, read the `.skt` / JSON output instead of raw source files: it holds symbols, relationships, and line spans in far fewer tokens. Open source files only to inspect or change implementation details, using the `L<start>-L<end>` spans from the output.

## Parse: extract structure

**Default to `directory-flat` (no flag) unless the input is a single file or a confirmed handful of files.** `inline` dumps everything to stdout and can overwhelm the context window; `directory-flat` writes chunked files to `./skeleton/` that can be read selectively.

```bash
codeknit parse ./src                       # chunked .skt files in ./skeleton/ (default)
codeknit parse ./src --clean               # re-run, replacing previous output
codeknit parse ./src/main.go --output-mode inline           # single file to stdout
codeknit parse ./src --output-mode inline --format json --edges   # structured data for scripts
codeknit parse ./src --output-mode directory-tree           # mirror the source tree
codeknit parse ./src ./output              # custom output directory
```

| Flag             | Default          | Description                                                           |
| ---------------- | ---------------- | --------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | `inline`, `directory-flat`, or `directory-tree`                       |
| `--format`       | `skt`            | `skt` or `json` (`codeknit.json` in directory modes)                  |
| `--edges`        | `false`          | Include the `[edges]` section (off by default to save tokens)         |
| `--minify`       | `false`          | Dictionary-compress repeated tokens into `dict.skt`                   |
| `--max-lines`    | `500`            | Tree source-chunk limit (minimum 3); flat keeps source blocks intact  |
| `--clean`        | `false`          | Replace prior output after successful generation                      |
| `--collect-test` | `false`          | Include test files                                                    |
| `--workers`      | `NumCPU`         | Max concurrent parsing goroutines                                     |
| `--verbose`      | `false`          | Print progress and timing info                                        |

| Mode             | Use for                                   | Output                                                         |
| ---------------- | ----------------------------------------- | -------------------------------------------------------------- |
| `inline`         | One file, or piping to another tool       | stdout                                                         |
| `directory-flat` | Most projects (default)                   | `./skeleton/map_001.skt`, `map_002.skt`, ...                   |
| `directory-tree` | Looking up a specific file's structure    | `./skeleton/<source name>.skt`; splits use `<source>_partN.skt` |

codeknit refuses to write into a directory that already holds `.skt` output; pass `--clean` (failed runs keep the previous output). Pass `--edges` whenever relationships matter — "what calls X", inheritance, containment.

### Reading parse output

`.skt` sections: `[symbols]`, `[edges]` (with `--edges`), optional `[errors]` and `[dict]`. JSON has top-level `files`, `symbols`, optional `edges`, and `errors`. Edges marked `[unresolved]` or `[ambiguous]` are uncertain, not confirmed dependencies; their candidate IDs are possible targets. C/C++ `meta/file` and `meta/include` symbols give readable endpoints without proving a header was resolved.

For the complete grammar, symbol categories, edge kinds, and examples, read [references/output-format.md](references/output-format.md).

## Graph analyze: find structural issues

```bash
codeknit graph analyze ./src                          # writes ./skeleton/graph_analysis.skt
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15
codeknit graph analyze ./src --top-n 50
```

| Flag                      | Default                         | Description                                          |
| ------------------------- | ------------------------------- | ---------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Output path                                          |
| `--fan-threshold`         | `10`                            | Min fan-in/fan-out to flag a hub symbol              |
| `--god-threshold`         | `15`                            | Min contains-edge count to flag a god class/function |
| `--max-inheritance-depth` | `5`                             | Flag deeper inheritance chains                       |
| `--top-n`                 | `30`                            | Cap ranked sections; `0` = no limit                  |
| `--betweenness-threshold` | `0.001`                         | Min betweenness centrality to report                 |
| `--propagation-cutoff`    | `0.05`                          | Min probability to continue change propagation       |

Also accepts `--collect-test`, `--workers`, `--verbose`. The report is `.skt`; read it like parse output. Uncertain edges are excluded from the algorithms and only counted.

## Fingerprint: find duplicate code

`codeknit fingerprint` hashes a normalized token stream (CTPH) for every function, method, variable, and type. Identifiers, string literals, and type annotations are normalized away, so renamed copies and cross-language ports still match.

```bash
codeknit fingerprint ./src                                     # 65-95% similarity → ./skeleton/fingerprints.skt
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100   # exact structural duplicates
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80     # conceptual duplicates
codeknit fingerprint ./src --rerank        # filter false positives via Ollama embeddings
codeknit fingerprint ./src --show-all -o duplicates.skt
```

| Flag               | Default                       | Description                                                                  |
| ------------------ | ----------------------------- | ---------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Output path                                                                  |
| `--min-similarity` | `65`                          | Minimum similarity percentage (0-100)                                        |
| `--max-similarity` | `95`                          | Maximum similarity percentage (0-100)                                        |
| `--show-all`       | `false`                       | Also emit the per-symbol `[fingerprints]` section                            |
| `--rerank`         | `false`                       | Semantic reranking via Ollama (`ollama serve && ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Ollama embedding model for `--rerank`                                        |

Also accepts `--collect-test`, `--workers`, `--verbose`. Output lines look like `similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin`, strongest first; see the fingerprint section of [references/output-format.md](references/output-format.md).

| Range     | Surfaces                                                |
| --------- | ------------------------------------------------------- |
| `96-100%` | Exact or near-exact copy-paste                          |
| `85-95%`  | Copy-paste with trivial edits (renamed variables)       |
| `65-84%`  | Strong structural similarity, often refactor candidates |
| `50-64%`  | Same algorithmic shape, different details               |
| `<50%`    | Usually noise                                           |

Fingerprints measure structural shape, not intent: the same logic written iteratively vs recursively scores low, and unrelated code with identical shape can score high. Treat a match as a signal and read both symbols before proposing a refactor. Bodies under 4 normalized tokens are skipped. Exact normalized-token candidates are always complete; above 500 eligible symbols other pairs use approximate top-50 retrieval. Percentages are a ranking, not an edit distance.

## Workflow

1. Ensure the binary (Setup above).
2. Structure: `codeknit parse ./src --edges`, then read the relevant `.skt` chunks — `[symbols]` for what exists, `[edges]` for relationships.
3. Quality: `codeknit graph analyze ./src` and read the report.
4. Duplicates: `codeknit fingerprint ./src`, read `[duplicates]`, open both symbols of each pair, then adjust the range or add `--rerank` if results are noisy.
5. Re-runs into the same directory need `--clean` (parse).

## Scan reliability (all commands)

All source commands accept repeatable `--include GLOB` / `--exclude GLOB` (quoted, relative to the input root; includes do not override ignore rules or test filtering), `--max-file-size BYTES` (10 MiB default, `0` unlimited), `--header-language c|cpp` for ambiguous `.h` files, and `--strict` to fail before writing output when the scan or parse is incomplete. Symlinks are skipped and reported. Check stderr and JSON `errors` / SKT `[errors]` for partial analyses. The maintained contract is the [analysis reliability reference](https://awslabs.github.io/codeknit/reference/analysis-reliability/).
