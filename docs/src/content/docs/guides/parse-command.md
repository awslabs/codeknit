---
title: Parse Command
description: Extract structural information from source code into .skt files or JSON.
---

`codeknit parse` reads source files and writes their symbols (and, optionally, relationships between them) as `.skt` files or JSON. This guide covers the situations you are most likely to hit. For the complete flag list shared by all commands, see the [CLI flags reference](../../reference/cli-flags/). For the output grammar, see the [output format reference](../../reference/output-format/).

## Parse a project for the first time

```bash
codeknit parse ./src
```

codeknit writes chunked `.skt` files to `./skeleton/`: `map_001.skt`, `map_002.skt`, and so on. Open the chunks selectively instead of loading everything at once.

To write somewhere else, pass the output directory as the second argument:

```bash
codeknit parse ./src ./output
```

Do not use the source directory, or one of its ancestors, as the output directory.

## Choose an output mode

Select a mode with `--output-mode`.

| Mode             | Use it when                                    | Result                                                                         |
| ---------------- | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `directory-flat` | Parsing most projects (default)                | `map_001.skt`, `map_002.skt`, ...                                              |
| `directory-tree` | You want to look up one file's structure       | One `.skt` per source file, mirroring the source tree                          |
| `inline`         | Parsing a single file, or piping to another tool | Output on stdout; no output directory is used                                  |

In tree mode the file names keep the source extension: `src/foo.c` becomes `src/foo.c.skt`, and `src/foo.h` becomes `src/foo.h.skt`. Large files are split into `foo.c_part1.skt`, `foo.c_part2.skt`, and so on.

```bash
codeknit parse ./src --output-mode directory-tree
```

Avoid `inline` on a whole project. It writes everything to stdout and can flood your terminal or an LLM context window.

For more on choosing a mode, see [Output modes](../output-modes/).

## Re-run on the same project

codeknit refuses to write into a directory that already holds generated output for the selected format. This prevents stale and fresh chunks from mixing. Add `--clean` to replace the previous output:

```bash
codeknit parse ./src --clean
```

`--clean` does not delete anything up front. codeknit generates the new output in a staging directory and replaces the old generated files only after generation succeeds. If generation fails, or the process is interrupted, the previous output is kept or restored on the next run. Files in the output directory that codeknit did not generate are preserved. If one of them collides with a generated path, the run fails.

In JSON mode, the same rules apply to `codeknit.json`.

**Pitfall: mixed chunks.** Symbol IDs (`S1`, `S2`, ...) and relationship results can change between runs and between versions. After upgrading to 0.5.0, regenerate the whole output set with `--clean`. Never combine chunks from different runs.

## Parse a single file

Point `parse` at the file and use `inline` mode:

```bash
codeknit parse ./src/main.go --output-mode inline
```

When you select a regular source file explicitly, codeknit skips ignore-file and test-file filtering for it. Include/exclude patterns and the size limit still apply.

## Get JSON for scripts

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

The JSON document has the top-level fields `files`, `symbols`, `edges` (omitted when there are none), and `errors` (omitted when there are none). A small example:

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

To write the JSON to disk instead, drop `--output-mode inline`. In the directory modes, JSON goes to `codeknit.json`. JSON is a single document and is not split by line count.

## Include relationships

By default, output has only the `[symbols]` section, which saves tokens. Add `--edges` when you need to know what calls, contains, inherits from, or references what:

```bash
codeknit parse ./src --edges
```

### Read uncertain relationships

Codeknit analyzes syntax without a compiler, so some relationships cannot be established. Each edge says how certain it is:

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->` is a resolved relationship.
- `--kind[unresolved]-->` means codeknit could not establish the target.
- `--kind[ambiguous]-->` means more than one target remains possible. Candidate IDs are listed when available.

Candidates are possible targets, not confirmed dependencies. Known endpoints use short IDs. An unresolved target with no matching symbol may appear as a name, relative to the source file.

In JSON, uncertain edges carry `resolution` (`"unresolved"` or `"ambiguous"`) and, when available, `candidates` as full symbol IDs. They omit `to_short`.

For C and C++, `#include` directives become `meta/file` and `meta/include` symbols, so their edges have short IDs. This records the directive as written. It does not mean codeknit found the header.

Graph metrics exclude uncertain edges and report how many were excluded. The HTML graph viewer shows a count of uncertain relationships and does not draw them. To inspect individual uncertain edges, use `--edges` on `parse`. See [Graph commands](../graph-commands/) for those reports.

The edge grammar is in the [output format reference](../../reference/output-format/). Resolution rules and limits are in the [language support reference](../../reference/language-support/#relationship-resolution).

## Reduce output size

To compress repeated tokens, add `--minify`:

```bash
codeknit parse ./src --minify
```

This writes a dictionary to `dict.skt` (or a `[dict]` section inline). Replace codes such as `d0` and `d1` with their dictionary values when reading. Relationship kinds are also encoded, but resolution labels such as `[unresolved]` stay readable.

To control chunk size, use `--max-lines` (default `500`):

- In `directory-flat` mode, it is a target. A source block is never split, so a chunk can exceed it by one block.
- In `directory-tree` mode, it is a hard limit for each source chunk. The minimum is `3`. Each chunk repeats its `[symbols]` and `## source-path` headers.

## Parse C++ headers named `.h`

Codeknit parses `.h` files as C by default. For C++ headers with that extension, set `--header-language cpp`:

```bash
codeknit parse ./src --header-language cpp
```

The flag affects only `.h` files. Files named `.hpp` and `.hxx` are already parsed as C++, and other extensions keep their usual parsers.

## Control which files are scanned

All five source-processing commands accept these flags. They are listed in full in the [CLI flags reference](../../reference/cli-flags/).

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` and `--exclude GLOB` are repeatable, gitignore-style patterns relative to the input directory. Quote them so your shell does not expand them. Excludes win. Includes do not override `.gitignore`, test-file filtering, or built-in exclusions.
- `--max-file-size BYTES` skips larger files. The default is `10485760` (10 MiB). `0` disables the limit.
- `--strict` fails before writing any output if the scan or parse is incomplete. The default is `false`.

Codeknit skips and reports symlinks and non-regular files. Ancestor `.gitignore` rules apply independently of your working directory.

**Pitfall: partial results.** Unreadable paths, oversized files, skipped symlinks, and syntax warnings make an analysis incomplete. Without `--strict`, codeknit keeps the usable results and reports the problems:

- on stderr,
- in the JSON `errors` field, or
- in the SKT `[errors]` section (inline, or in `warnings.skt` in directory modes).

An input with no selected supported sources, or a run in which every source read fails, always exits unsuccessfully. For exact guarantees, see [Analysis reliability](../../reference/analysis-reliability/).

## Parse flags

These flags are specific to `parse`. Shared flags are in the [CLI flags reference](../../reference/cli-flags/).

| Flag               | Default          | Description                                                                                       |
| ------------------ | ---------------- | ------------------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | Output mode: `inline`, `directory-flat`, or `directory-tree`.                                     |
| `--format`         | `skt`            | Output format: `skt` or `json`.                                                                   |
| `--header-language`| `c`              | Parser for `.h` files: `c` or `cpp`. Other extensions are unchanged.                              |
| `--max-lines`      | `500`            | Line target for flat output; hard limit (minimum 3) for each tree source chunk.                   |
| `--collect-test`   | `false`          | Include test files.                                                                               |
| `--minify`         | `false`          | Use dictionary-based compression.                                                                 |
| `--edges`          | `false`          | Include the `[edges]` section (off by default to save tokens).                                    |
| `--clean`          | `false`          | Replace prior generated output, only after generation succeeds.                                   |
| `--workers`        | `NumCPU`         | Maximum concurrent parsing goroutines.                                                            |
| `--verbose`        | `false`          | Print progress information.                                                                       |

## Tips

- Use `directory-flat` for most projects, and read only the chunks you need.
- Use `--minify` on large codebases to save tokens.
- Pass `--edges` only when you need relationship data.
- Check the [tested capability matrix and known limits](../../reference/language-support/) before relying on extraction for a specific language.

## See also

- [Output modes](../output-modes/)
- [Output format reference](../../reference/output-format/)
- [Analysis reliability](../../reference/analysis-reliability/)
- [Graph commands](../graph-commands/)