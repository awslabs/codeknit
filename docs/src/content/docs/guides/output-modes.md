---
title: Output Modes
description: Choose the right output mode for your project size and workflow.
---

codeknit separates two decisions: where output goes (the **output mode**) and how it is encoded (the **output format**). This page explains why each option exists, what it costs, and when it fits. For the `.skt` grammar, see the [output format reference](../../reference/output-format/). For commands, see the [parse command guide](../parse-command/).

## Two independent choices

The `--output-mode` flag selects one of three modes: `directory-flat`, `directory-tree`, or `inline`. The `--format` flag selects `skt` (the default) or `json`. Any mode can be combined with either format.

The two choices are separate because they answer different questions. The mode decides how the output is split and located on disk, which matters to a person or an LLM reading it. The format decides who consumes it: an LLM or human reader, or a script.

## directory-flat

`directory-flat` is the default. It writes numbered chunk files, `map_001.skt`, `map_002.skt`, and so on, into `./skeleton/` unless you pass another output directory.

**Why it exists.** A large codebase produces more structure than fits in one context window. Chunking lets a reader open only the files it needs instead of loading everything.

**How splitting works.** codeknit keeps each source file's block intact and starts a new chunk between blocks once a chunk reaches roughly `--max-lines` (default 500). The limit is a target, not a hard bound: one source file with a very large block can produce a chunk longer than 500 lines. Keeping blocks intact means a symbol is never separated from its siblings in the same file.

**Trade-off.** Chunks are not organized by source path. To find one file's structure, you search the chunks for its `## source-path` header. If you often look up specific files, `directory-tree` fits better.

**Minification.** With `--minify`, the dictionary is written to `dict.skt` next to the chunks. See [Minification](#minification).

## directory-tree

`directory-tree` mirrors your source tree under the output directory. Each source file gets a `.skt` file at the matching path.

**Naming.** codeknit appends `.skt` to the complete source filename rather than replacing the extension. `src/foo.c` becomes `src/foo.c.skt`, and `src/foo.h` becomes `src/foo.h.skt`. Without this rule, both would map to `foo.skt` and collide. When a file's output is split, the parts are named `foo.c_part1.skt`, `foo.c_part2.skt`, and so on. Conflicting output paths, including case-only collisions, fail visibly instead of overwriting each other.

**Splitting differs from flat mode.** In tree mode, `--max-lines` is a hard limit on each source chunk, with a minimum of 3. Each symbol chunk repeats the `[symbols]` and `## source-path` headers so it can be read alone. Edge-only chunks carry `[edges]`. The limit applies to source chunks, not to auxiliary files such as `dict.skt` and `warnings.skt`.

**Trade-off.** Tree output is easy to navigate alongside the source, but it produces many small files. A large project can yield thousands of them, and a reader that wants a broad overview must open many files.

## inline

`inline` writes everything to stdout. It creates no output directory and no staging area.

**Why it exists.** It suits two cases: inspecting a single file quickly, and piping output into another tool. It is the only mode where nothing is written to disk.

**Trade-off.** Nothing is split. The whole result lands in the terminal or in the consumer's context at once, so it overwhelms a reader on anything beyond a handful of files. Treat it as appropriate for a single file or a very small project, roughly fewer than five files. Because stdout cannot be rolled back, a failing consumer can see partial output. Directory modes do not have this problem.

## Output formats

| Format | Best for | Output |
| --- | --- | --- |
| `skt` | LLM context and human inspection | `.skt` files or stdout |
| `json` | Scripts and structured integration | `codeknit.json` in directory modes, or stdout in inline mode |

**skt** is compact and line-oriented. Each symbol sits on one physical line, with line breaks in signatures escaped, so a reader can load part of the output without breaking its structure. That is why chunking works for `.skt`.

**json** exists for consumers that need typed fields rather than text to read. It is a single document with `files`, `symbols`, and, when edges are included, `edges`. Because it is one document, it is not split by line count. In directory modes `--max-lines` therefore has no effect on JSON, and a large project produces one large `codeknit.json`.

A compact sample:

```json
{
  "files": ["app.go"],
  "symbols": [
    {"id": "app.go::User", "short_id": "S1", "name": "User", "file": "app.go",
     "category": "type", "kind": "struct", "signature": "type User struct", "span": [3, 3]},
    {"id": "app.go::Save", "short_id": "S2", "name": "Save", "file": "app.go",
     "category": "callable", "kind": "function", "signature": "Save(u: S1)", "span": [5, 5]}
  ],
  "edges": [
    {"from": "app.go::Save", "from_short": "S2", "to": "app.go::User",
     "to_short": "S1", "kind": "references"}
  ]
}
```

## Choosing a mode

| Mode | Best for | Output |
| --- | --- | --- |
| `directory-flat` | Most projects (default) | `./skeleton/map_001.skt`, `map_002.skt`, ... |
| `directory-tree` | Navigating alongside source | `./skeleton/<mirrored path>.skt` |
| `inline` | Single file, piping | stdout |

Rules of thumb:

- **When unsure**, use `directory-flat`, the default.
- **A single file**: `inline` is fine.
- **Anything beyond a few files**: use `directory-flat` or `directory-tree`.
- **Large codebases**: add `--minify`.
- **Re-running into the same directory**: add `--clean`.

## Staging, re-runs, and `--clean`

Directory modes do not write into the output directory as they go. codeknit generates output in a sibling staging directory and publishes it only after generation succeeds. A failed run therefore leaves your previous output in place.

By default, codeknit refuses to write when generated output for the selected format already exists. The reason is that symbol IDs (`S1`, `S2`, ...) can change whenever extraction changes. Mixing chunks from different runs would give edges that point at the wrong symbols. `--clean` opts in to replacement: after successful generation, obsolete `.skt` files (in SKT mode) or `codeknit.json` (in JSON mode) are replaced. Unrelated entries in the directory are preserved. If one of them collides with a generated path, the run fails instead of overwriting it.

Publication takes a directory-specific lock, backs up the previous output, and restores it if replacement fails. If the process is interrupted between the two directory renames, the next run restores the backup before generating anything. This recovers from process failure. It does not make the directory update instantaneous: there is a brief interval between the renames, so a reader that needs uninterrupted access should coordinate with generation. It also does not protect against storage failure or power loss.

Do not choose a source directory, or an ancestor of it, as the output directory. For the full publication guarantees, see the [analysis reliability reference](../../reference/analysis-reliability/).

## Minification

`--minify` applies dictionary-based compression. codeknit replaces repeated tokens, such as `callable/method` or `exported`, with short codes (`d0`, `d1`, ...). The mapping is stored in a `[dict]` section, which in directory modes is the `dict.skt` file.

The saving comes from repetition. Most symbol lines share the same categories, kinds, and properties, so a large codebase repeats these strings thousands of times. The cost is indirection: a reader must expand codes using the dictionary, and a single chunk is no longer self-contained. The dictionary is separate auxiliary output, so it does not count against `--max-lines`.

Minified output encodes the relationship kind but keeps resolution markers readable, so `unresolved` and `ambiguous` edges stay distinguishable. For small projects the savings are modest and the extra step is rarely worth it. For large ones the savings in tokens can be significant.

## Related pages

- [Parse command guide](../parse-command/): the commands for each mode.
- [Output format reference](../../reference/output-format/): the `.skt` and JSON grammar.
- [CLI flags reference](../../reference/cli-flags/): every flag and default.
- [Analysis reliability reference](../../reference/analysis-reliability/): output naming and publication guarantees.