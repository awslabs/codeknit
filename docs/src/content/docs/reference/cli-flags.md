---
title: CLI Reference
description: Complete reference for all codeknit commands and flags.
---

This page lists every codeknit command, its arguments, and its flags. For task-oriented instructions, see the [parse command guide](../../guides/parse-command/), the [graph commands guide](../../guides/graph-commands/), and the [fingerprint command guide](../../guides/fingerprint-command/).

## Command summary

| Command                                       | Purpose                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | Launch the interactive terminal UI.                            |
| `codeknit parse <input-path> [output-dir]`    | Extract symbols and relationships into `.skt` or JSON output.  |
| `codeknit graph show <input-path>`            | Generate an interactive HTML graph.                            |
| `codeknit graph analyze <input-path>`         | Run structural graph algorithms and write a `.skt` report.     |
| `codeknit graph hotspots <input-path>`        | Rank change hotspots using Git history and graph structure.    |
| `codeknit fingerprint <input-path>`           | Detect duplicate and near-duplicate code.                      |
| `codeknit completion <shell>`                 | Generate a shell completion script.                            |

Five commands process source code: `parse`, `graph show`, `graph analyze`, `graph hotspots`, and `fingerprint`. They share the flags described in [Shared source-processing flags](#shared-source-processing-flags).

## Global flags

| Flag              | Type | Default | Description                  |
| ----------------- | ---- | ------- | ---------------------------- |
| `--version`       | bool | `false` | Print version information.   |
| `-h`, `--help`    | bool | `false` | Show help for the command.   |

## Shared source-processing flags

These flags appear in the table of each source-processing command.

| Flag                    | Type                | Default               | Allowed values       | Description                                                                          |
| ----------------------- | ------------------- | --------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                   | `c`, `cpp`           | Parser used for `.h` files.                                                          |
| `--include GLOB`        | string, repeatable  | All supported sources | Gitignore-style glob | Include matching source paths. Multiple patterns form a union.                       |
| `--exclude GLOB`        | string, repeatable  | None                  | Gitignore-style glob | Exclude matching paths or directories.                                               |
| `--max-file-size BYTES` | int                 | `10485760` (10 MiB)   | `0` or greater       | Skip source files larger than this size. `0` disables the limit.                     |
| `--strict`              | bool                | `false`               | —                    | Fail before output is written if scan or parse diagnostics occur.                    |

### `--header-language`

Applies to `.h` files only. Other extensions keep their usual parser (`.hpp` and `.hxx` already use C++).

- `c` (default) preserves C macro, typedef, and union extraction.
- `cpp` selects the C++ parser. Use it for C++ projects whose headers are named `.h`.

The TUI uses the default `c` selection for `.h` files.

```bash
codeknit parse ./src --header-language cpp
```

For the tested extraction capabilities and known limits per language, see [Language Support and API Fidelity](../language-support/). That page describes tested constructs, not complete language or compiler-semantic coverage.

### Source selection

- Patterns use slash-separated, gitignore-style globs relative to the input directory. For a single-file input, patterns match the file's basename.
- Quote patterns so the shell does not expand them.
- Excludes take precedence over includes.
- Includes do not override `.gitignore` rules, test-file filtering, or built-in exclusions.
- Negation patterns (`!pattern`) are not accepted in `--include` or `--exclude`.
- Directory scans honor ancestor and nested `.gitignore` files independently of the current working directory.
- Explicitly selecting a regular source file bypasses ignore and test filtering. Include, exclude, and size limits still apply.
- Test files are excluded unless `--collect-test` is set.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### Diagnostics and `--strict`

- Symlinks, directory symlinks, and non-regular files are skipped and reported.
- Unreadable paths, failed ignore-file reads, oversized sources, and syntax warnings make an analysis incomplete.
- By default, usable partial results are kept. Diagnostics appear on stderr. `parse` also records them in JSON `errors` or in SKT `[errors]`.
- With `--strict`, an incomplete analysis is rejected before output is written.
- A run with no selected supported sources, or in which every source read fails, always exits unsuccessfully.
- The TUI applies the same symlink policy and 10 MiB default. Its forms do not expose the selection and strict flags.

See [Analysis Reliability](../analysis-reliability/) for exact guarantees.

## `codeknit`

```bash
codeknit
```

Running `codeknit` with no subcommand launches the interactive terminal UI. The UI offers the `parse`, `graph show`, `graph analyze`, `graph hotspots`, and `fingerprint` commands and their options.

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| Argument       | Required | Default      | Description                                                              |
| -------------- | -------- | ------------ | ------------------------------------------------------------------------ |
| `input-path`   | Yes      | —            | Source directory or single source file.                                  |
| `output-dir`   | No       | `./skeleton` | Output directory. Not used when `--output-mode` is `inline`.             |

| Flag                    | Type                | Default            | Allowed values                                   | Description                                                                                         |
| ----------------------- | ------------------- | ------------------ | ------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`   | `inline`, `directory-flat`, `directory-tree`     | Where and how output is written.                                                                    |
| `--format`              | string              | `skt`              | `skt`, `json`                                    | Output format.                                                                                      |
| `--max-lines`           | int                 | `500`              | For `directory-tree`, 3 or greater               | Line target for flat output. Hard source-chunk limit for tree output.                               |
| `--collect-test`        | bool                | `false`            | —                                                | Include test files.                                                                                 |
| `--minify`              | bool                | `false`            | —                                                | Enable dictionary-based output compression.                                                         |
| `--edges`               | bool                | `false`            | —                                                | Include the `[edges]` section. Off by default to save tokens.                                       |
| `--clean`               | bool                | `false`            | —                                                | Replace prior generated output, only after successful generation.                                   |
| `--workers`             | int                 | `0` (`NumCPU`)     | `0` or greater                                   | Maximum concurrent parsing goroutines.                                                              |
| `--verbose`             | bool                | `false`            | —                                                | Print progress information.                                                                         |
| `--header-language`     | string              | `c`                | `c`, `cpp`                                       | Parser used for `.h` files.                                                                         |
| `--include GLOB`        | string, repeatable  | All supported sources | —                                             | Include matching source paths.                                                                      |
| `--exclude GLOB`        | string, repeatable  | None               | —                                                | Exclude matching paths or directories.                                                              |
| `--max-file-size BYTES` | int                 | `10485760`         | `0` disables the limit                           | Skip larger source files.                                                                           |
| `--strict`              | bool                | `false`            | —                                                | Fail before output if scan or parse diagnostics occur.                                              |

### Output modes

| Value             | Behavior                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| `inline`          | Write to stdout. No output directory is used.                                                    |
| `directory-flat`  | Write chunked files to the output directory. Source blocks stay intact, so `--max-lines` is a splitting target rather than a hard bound. |
| `directory-tree`  | Mirror the source tree. Names keep the source extension (`foo.c.skt`, `foo.h.skt`). Split files use `foo.c_part1.skt`. |

### Output behavior

- Without `--clean`, existing generated output for the selected format in the output directory causes an error.
- With `--clean`, obsolete `.skt` files (SKT format) or `codeknit.json` (JSON format) are replaced only after successful generation. Unrelated entries are preserved.
- In JSON format, directory modes write `codeknit.json`. JSON is a single document and is not split by line count.
- `dict.skt` and `warnings.skt` are reserved auxiliary file names.
- Do not use a source directory or one of its ancestors as the output directory.

See [Output modes](../../guides/output-modes/) and the [output format reference](../output-format/) for details.

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| Argument       | Required | Description                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Yes      | Source directory or single source file.  |

| Flag                    | Type                | Default                            | Allowed values          | Description                                                       |
| ----------------------- | ------------------- | ---------------------------------- | ----------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`   | File path               | Output HTML file.                                                 |
| `--collect-test`        | bool                | `false`                            | —                       | Include test files.                                               |
| `--workers`             | int                 | `0` (`NumCPU`)                     | `0` or greater          | Maximum concurrent parsing goroutines.                            |
| `--verbose`             | bool                | `false`                            | —                       | Print progress information.                                       |
| `--header-language`     | string              | `c`                                | `c`, `cpp`              | Parser used for `.h` files.                                       |
| `--include GLOB`        | string, repeatable  | All supported sources              | —                       | Include matching source paths.                                    |
| `--exclude GLOB`        | string, repeatable  | None                               | —                       | Exclude matching paths or directories.                            |
| `--max-file-size BYTES` | int                 | `10485760`                         | `0` disables the limit  | Skip larger source files.                                         |
| `--strict`              | bool                | `false`                            | —                       | Fail before output if scan or parse diagnostics occur.            |

The command writes a self-contained HTML file and opens it in the default browser. The viewer draws resolved relationships and shows a count of uncertain relationships.

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| Argument       | Required | Description                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Yes      | Source directory or single source file.  |

| Flag                       | Type                | Default                            | Allowed values          | Description                                                  |
| -------------------------- | ------------------- | ---------------------------------- | ----------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`    | File path               | Output `.skt` file.                                          |
| `--collect-test`           | bool                | `false`                            | —                       | Include test files.                                          |
| `--workers`                | int                 | `0` (`NumCPU`)                     | `0` or greater          | Maximum concurrent parsing goroutines.                       |
| `--verbose`                | bool                | `false`                            | —                       | Print progress information.                                 |
| `--fan-threshold`          | int                 | `10`                               | —                       | Minimum fan-in or fan-out to flag a hub symbol.              |
| `--god-threshold`          | int                 | `15`                               | —                       | Minimum `contains` edge count to flag a god class or function. |
| `--max-inheritance-depth`  | int                 | `5`                                | —                       | Flag inheritance chains deeper than this value.              |
| `--top-n`                  | int                 | `30`                               | `0` = no limit          | Cap ranked output sections.                                  |
| `--betweenness-threshold`  | float               | `0.001`                            | —                       | Minimum betweenness centrality value to report.              |
| `--propagation-cutoff`     | float               | `0.05`                             | —                       | Minimum probability to continue change propagation.          |
| `--header-language`        | string              | `c`                                | `c`, `cpp`              | Parser used for `.h` files.                                  |
| `--include GLOB`           | string, repeatable  | All supported sources              | —                       | Include matching source paths.                               |
| `--exclude GLOB`           | string, repeatable  | None                               | —                       | Exclude matching paths or directories.                       |
| `--max-file-size BYTES`    | int                 | `10485760`                         | `0` disables the limit  | Skip larger source files.                                    |
| `--strict`                 | bool                | `false`                            | —                       | Fail before output if scan or parse diagnostics occur.       |

Dependency metrics exclude uncertain relationships and file/include metadata. The report records the excluded relationship counts.

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| Argument       | Required | Description                                                  |
| -------------- | -------- | ------------------------------------------------------------ |
| `input-path`   | Yes      | Source directory inside a Git repository.                    |

| Flag                       | Type                | Default                    | Allowed values                              | Description                                                  |
| -------------------------- | ------------------- | -------------------------- | ------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`  | File path                                   | Output file.                                                 |
| `--format`                 | string              | `skt`                      | `skt`, `json`                               | Output format.                                               |
| `--since`                  | string              | `12mo`                     | Window such as `180d`, `12mo`, or `2y`      | History window.                                              |
| `--max-commits`            | int                 | `2000`                     | —                                           | Maximum commits to inspect.                                  |
| `--max-files-per-commit`   | int                 | `50`                       | —                                           | Exclude commits that touch more files than this value.       |
| `--min-cochanges`          | int                 | `3`                        | —                                           | Minimum shared commits to report temporal coupling.          |
| `--top-n`                  | int                 | `30`                       | —                                           | Maximum results per section.                                 |
| `--include-merges`         | bool                | `false`                    | —                                           | Include merge commits in history metrics.                    |
| `--collect-test`           | bool                | `false`                    | —                                           | Include test files.                                          |
| `--workers`                | int                 | `0` (`NumCPU`)             | `0` or greater                              | Maximum concurrent parsing goroutines.                       |
| `--verbose`                | bool                | `false`                    | —                                           | Print progress information.                                 |
| `--header-language`        | string              | `c`                        | `c`, `cpp`                                  | Parser used for `.h` files.                                  |
| `--include GLOB`           | string, repeatable  | All supported sources      | —                                           | Include matching source paths.                               |
| `--exclude GLOB`           | string, repeatable  | None                       | —                                           | Exclude matching paths or directories.                       |
| `--max-file-size BYTES`    | int                 | `10485760`                 | `0` disables the limit                      | Skip larger source files.                                    |
| `--strict`                 | bool                | `false`                    | —                                           | Fail before output if scan or parse diagnostics occur.       |

The command combines commit frequency, line churn, and recency with file-level PageRank, transitive fan-in, and betweenness centrality. It also reports temporal coupling between files that change together. Merge commits are excluded by default.

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| Argument       | Required | Description                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Yes      | Source directory or single source file.  |

| Flag                    | Type                | Default                          | Allowed values          | Description                                                                                                      |
| ----------------------- | ------------------- | -------------------------------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`    | File path               | Output `.skt` file.                                                                                              |
| `--min-similarity`      | int                 | `65`                             | `0`–`100`               | Minimum similarity percentage to report.                                                                         |
| `--max-similarity`      | int                 | `95`                             | `0`–`100`               | Maximum similarity percentage to report.                                                                         |
| `--show-all`            | bool                | `false`                          | —                       | Also emit the `[fingerprints]` section with raw token data.                                                      |
| `--rerank`              | bool                | `false`                          | —                       | Find semantic neighbors and rerank candidates with Ollama embeddings. Requires `ollama serve` and `ollama pull qwen3-embedding:0.6b`. |
| `--model`               | string              | `qwen3-embedding:0.6b`           | Ollama model name       | Ollama embedding model to use with `--rerank`.                                                                   |
| `--collect-test`        | bool                | `false`                          | —                       | Include test files.                                                                                              |
| `--workers`             | int                 | `0` (`NumCPU`)                   | `0` or greater          | Maximum concurrent parsing goroutines.                                                                           |
| `--verbose`             | bool                | `false`                          | —                       | Print progress information.                                                                                      |
| `--header-language`     | string              | `c`                              | `c`, `cpp`              | Parser used for `.h` files.                                                                                      |
| `--include GLOB`        | string, repeatable  | All supported sources            | —                       | Include matching source paths.                                                                                   |
| `--exclude GLOB`        | string, repeatable  | None                             | —                       | Exclude matching paths or directories.                                                                           |
| `--max-file-size BYTES` | int                 | `10485760`                       | `0` disables the limit  | Skip larger source files.                                                                                        |
| `--strict`              | bool                | `false`                          | —                       | Fail before output if scan or parse diagnostics occur.                                                           |

### Similarity range

The default range, 65–95%, excludes 100% matches. To report exact normalized-token matches, set both bounds to `100`:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Exact normalized-token candidates are complete. With 500 or more eligible symbols, other near-duplicate candidates use approximate top-50 neighbor retrieval. `--rerank` can remove a structurally exact candidate. See [Analysis Reliability](../analysis-reliability/#duplicate-detection-guarantees) for details.

## `codeknit completion`

```bash
codeknit completion <shell>
```

| Argument  | Required | Allowed values                         | Description                                   |
| --------- | -------- | -------------------------------------- | --------------------------------------------- |
| `shell`   | Yes      | `bash`, `zsh`, `fish`, `powershell`    | Shell for which to generate a completion script. |

The script is written to stdout.

```bash
codeknit completion zsh >> ~/.zshrc
```

## Related pages

- [Output format reference](../output-format/)
- [Language Support and API Fidelity](../language-support/)
- [Analysis Reliability](../analysis-reliability/)