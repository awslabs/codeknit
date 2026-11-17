---
title: Analysis Reliability
description: Source selection, incomplete analyses, output publication, and duplicate-detection guarantees.
---

<!-- Maintained in docs/reference/analysis-reliability.md; copied verbatim by docs/gen. -->

These behaviors describe the unreleased Go/CLI changes following 0.5.0.

## Source selection and diagnostics

All five source-processing commands (`parse`, `fingerprint`, `graph show`, `graph analyze`, and `graph hotspots`) accept:

| Flag                    | Default               | Meaning                                                             |
| ----------------------- | --------------------- | ------------------------------------------------------------------- |
| `--include GLOB`        | All supported sources | Include matching source paths; repeat to select several patterns.   |
| `--exclude GLOB`        | None                  | Exclude matching paths or directories; repeat for several patterns. |
| `--max-file-size BYTES` | `10485760` (10 MiB)   | Skip larger source files; `0` disables this limit.                  |
| `--strict`              | `false`               | Fail before emitting results if scan or parse diagnostics occur.    |

Patterns use slash-separated, gitignore-style globs relative to the selected input directory. For a single-file input, patterns match its basename. Quote patterns so the shell does not expand them. Includes form a union; excludes take precedence. Includes do not override `.gitignore`, test filtering, or built-in exclusions. Negation patterns (`!pattern`) are not accepted in these flags; use the separate include and exclude controls.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
codeknit fingerprint ./src --max-file-size 2097152 --include '**/*.ts'
codeknit parse ./src/app.go --output-mode inline --format json
```

Directory scans honor ancestor and nested `.gitignore` files relative to their project location, independently of the process working directory. The nearest `.git` directory or worktree marker establishes the repository boundary. Outside a repository, ancestor ignore files still apply. A selected subdirectory beneath an ignored directory remains ignored. Explicitly selecting a regular source file bypasses ignore and test filtering, but still applies include/exclude and size limits.

Source symlinks, directory symlinks, and non-regular sources are skipped and reported. Reads are confined to the scan root, and the size limit is enforced again while reading so a file that grows after discovery cannot bypass it. Built-in directory exclusions include `.git`, `.hg`, `.svn`, `node_modules`, Python caches, and common Python virtual environments. Test files remain excluded unless `--collect-test` is set.

Unreadable directories/files, failed ignore-file reads, skipped symlinks, oversized sources, and syntax warnings make an analysis incomplete. By default, usable partial results are retained and diagnostics appear on stderr. Parse JSON includes path/reason records in `errors`; SKT includes `[errors]` inline or in `warnings.skt`. Graph/fingerprint commands also report diagnostics on stderr. Deliberately ignored or filtered files are not warnings. `--strict` rejects incomplete results before output publication. An input with no selected supported sources, or a run in which every source read fails, always exits unsuccessfully.

The TUI uses the same symlink policy and 10 MiB default. Its forms do not yet expose the new selection and strict flags.

## Output names and publication

Tree output appends `.skt` to the complete source name: `src/foo.c` becomes `src/foo.c.skt`, while `src/foo.h` becomes `src/foo.h.skt`. Split output uses `src/foo.c_part1.skt`, `src/foo.c_part2.skt`, and so on. Dictionary and diagnostic files retain the reserved names `dict.skt` and `warnings.skt`. Conflicting output paths, including case-only collisions, fail visibly.

Tree SKT requires `--max-lines` of at least 3. Each symbol chunk repeats its `[symbols]` and `## source-path` headers. Edge-only chunks carry `[edges]`. Every source chunk respects the limit; dictionaries and diagnostic files are separate auxiliary output. Flat SKT keeps source blocks intact, so its line limit is a splitting target rather than a hard bound for a large source block. JSON is a single document and is not split by line count.

Directory generation completes in a sibling staging directory. Without `--clean`, existing generated output for the selected format causes an error. With `--clean`, obsolete `.skt` files (SKT mode) or `codeknit.json` (JSON mode) are replaced only after successful generation. Unrelated entries are preserved; a collision with one causes an error. Do not use a source directory or its ancestor as the output directory.

Publication holds a directory-specific lock, backs up the previous output, and restores it if replacement fails. If a process terminates between directory renames, the next invocation restores the backup before generating anything. The sibling `.NAME.codeknit.lock` file intentionally remains to coordinate future publishers; do not remove an active lock or recovery backup. There is a brief interval between the two directory renames, so readers requiring uninterrupted access should coordinate with generation. This provides process-failure recovery, not a guarantee against storage failure or power loss.

Individual reports are written and synced to a temporary sibling before replacing the destination. Existing files are not truncated during generation. Inline stdout cannot be rolled back if its consumer fails.

## Duplicate-detection guarantees

Exact matching means identical complete normalized token streams in the same symbol category among declarations eligible for fingerprinting. It does not mean identical source text or proven semantic equivalence. Exact candidate groups use the token streams themselves, rather than fuzzy-hash equality.

Below 500 eligible symbols, structural detection compares all same-category pairs. At 500 or more, all exact token pairs are retained independently of candidate retrieval, and other pairs use approximate top-50 structural neighbors per symbol. Near-duplicate retrieval is therefore not exhaustive on large inputs. Retrieval still scans all vectors; it does not make the overall algorithm subquadratic.

Similarity range filtering still applies. The default range is 65–95%, which excludes 100% matches. To include exact matches, use:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Optional `--rerank` adds semantic candidates, applies cosine agreement, and filters on the combined score; it can remove a structurally exact candidate. Embedding text comes from the same bounded source snapshot used for extraction. Stable inputs and fixed scores produce deterministic ordering, including ties; model responses themselves are external inputs.

`internal/emitter/calibration_test.go` evaluates a small labeled source fixture at a 70% threshold, reporting true/false positives and negatives, precision, and recall separately from similarity scores. The fixture has one exact pair, two near-clone pairs, and three unrelated pairs, exercised below and across the 500-symbol boundary. This is a regression check, not a representative estimate of real-world accuracy. Larger multilingual calibration and performance budgets remain roadmap work.
