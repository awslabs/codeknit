---
title: Graph Commands
description: Visualize and analyze your codebase structure with graph algorithms.
---

Use the `graph` subcommands to see how your code connects, find structural problems, and decide which files to refactor first. This guide assumes you have [installed codeknit](../../installation/) and can run [`codeknit parse`](../parse-command/).

| Task | Command |
| --- | --- |
| Explore the graph visually | `codeknit graph show` |
| Find structural problems | `codeknit graph analyze` |
| Find volatile, important files | `codeknit graph hotspots` |

## Visualize the code graph

`graph show` parses the input and writes a self-contained HTML file. Symbols (functions, classes, types) are nodes. Relationships (calls, contains, implements) are edges.

```bash
codeknit graph show ./myproject
```

The file is written to `./skeleton/codeknit-graph.html` and opens in your default browser.

To choose the output path or include test files:

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| Flag | Default | Purpose |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Output HTML file path |
| `--collect-test` | `false` | Include test files |
| `--workers` | `0` (`NumCPU`) | Maximum concurrent parsing goroutines |
| `--verbose` | `false` | Print progress information |

### Inspect unresolved relationships

The viewer draws resolved relationships only. It shows a count of uncertain (`unresolved` or `ambiguous`) relationships but does not draw them or their candidate links.

To see individual uncertain edges, run `parse` with `--edges`:

```bash
codeknit parse ./src --output-mode inline --edges
```

Add `--format json` for JSON output. See the [output format reference](../../reference/output-format/) for how edges and candidates are written.

## Analyze structure and find problems

`graph analyze` runs 22 graph algorithms and writes an LLM-readable `.skt` report.

```bash
codeknit graph analyze ./myproject
```

The report is written to `./skeleton/graph_analysis.skt`. Read it directly, or give it to an LLM along with the relevant source.

### Tune the report

Raise a threshold to see fewer findings, or lower it to see more:

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0` removes the cap on ranked sections.

| Flag | Default | Purpose |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | Output `.skt` file path |
| `--collect-test` | `false` | Include test files |
| `--workers` | `0` (`NumCPU`) | Maximum concurrent parsing goroutines |
| `--verbose` | `false` | Print progress information |
| `--fan-threshold` | `10` | Minimum fan-in or fan-out to flag a hub symbol |
| `--god-threshold` | `15` | Minimum contains-edge count to flag a god class or function |
| `--max-inheritance-depth` | `5` | Flag inheritance chains deeper than this |
| `--top-n` | `30` | Cap ranked sections; `0` means no limit |
| `--betweenness-threshold` | `0.001` | Minimum betweenness centrality to report |
| `--propagation-cutoff` | `0.05` | Minimum heuristic score for continuing change propagation |

### Know which finding answers which question

| Algorithm | Use it to find |
| --- | --- |
| Cyclic dependencies (Tarjan's SCC) | Groups of symbols that depend on each other in a loop |
| Hub detection | Symbols with high fan-in or fan-out coupling |
| Orphan detection | Dead-code candidates |
| God class/function detection | Symbols that contain an excessive number of children |
| Instability metric | Symbols and packages that are fragile or stable, using Ce/(Ca+Ce) |
| Deep inheritance chains | Hierarchies deeper than `--max-inheritance-depth` |
| Betweenness centrality | Bottleneck symbols that many dependency paths pass through |
| Articulation points | Single points of failure whose removal splits the graph |
| PageRank | Symbols that are recursively important |
| Transitive fan-in | Blast radius: everything that depends on a symbol, directly or indirectly |
| Change propagation simulation | How far a change to a symbol is likely to spread |
| Circular package dependencies | Package-level dependency loops |
| Layer violation detection | Dependencies that cross layers in the wrong direction |
| Reachability from entry points | Code that no entry point reaches |
| Weakly connected components | Isolated clusters of code |
| Dependency weight | How strongly packages are coupled |
| Distance from Main Sequence | Packages whose abstractness and instability are out of balance |
| Shotgun surgery detection | Changes that would touch many places in a package |
| Feature envy detection | Symbols that use another package more than their own |
| Stable dependency violations | Stable packages that depend on less stable ones |
| Interface segregation violations | Interfaces whose members are mostly unused by their users |
| Containment depth | Deeply nested symbols |

### Read the results correctly

These rules affect how far you can trust a finding:

- **Uncertain relationships are excluded.** Dependency metrics use resolved edges only. The report counts the `unresolved` and `ambiguous` relationships it left out.
- **File and include metadata is ignored.** `meta/file` and `meta/include` symbols do not affect dependency rankings or dead-code findings.
- **Fan-in and fan-out count distinct neighbors.** Containment does not count as a dependency.
- **Inheritance depth covers all bases and interfaces.** Cyclic hierarchies are not assigned a depth.
- **Reachability starts at `main`, `Main`, or `init` callables** and follows dependency edges. A reachable member makes its container relevant, but not its sibling members. If the scan finds no recognized entry point, reachability findings are omitted.
- **Callbacks are references, not calls.** A callback argument or returned callable does not prove the callback runs.
- **Scores are heuristics.** Change propagation uses the strongest dependency path. Its scores and the layer rules are not measured probabilities or compiler guarantees.

### Act on findings safely

Treat every finding as a candidate for review. Before changing code, open the symbols the report names and check for:

- External callers outside the scanned directory.
- Dynamic dispatch.
- Dependencies you did not scan.

Analysis is syntax-based, so any of these can change the conclusion. See [relationship resolution in the language support reference](../../reference/language-support/#relationship-resolution) for the resolution rules and limits.

## Find history-aware hotspots

`graph hotspots` ranks files that change often and sit in structurally important positions. It also reports temporal coupling: files that repeatedly change in the same commits. Run it on a directory inside a Git repository.

```bash
codeknit graph hotspots ./myproject
```

The report is written to `./skeleton/hotspots.skt`.

It combines two groups of metrics:

- **History:** commit frequency, added and deleted line churn, and recency.
- **Structure:** file-level PageRank, transitive fan-in, and betweenness.

The report includes a confidence level based on the number of relevant commits analyzed. Treat rankings from a low-confidence report as tentative.

### Change the history window and filters

```bash
# Analyze two years and write JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Allow larger commits and require stronger temporal coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

Merge commits are excluded unless you pass `--include-merges`. Commits that touch more files than `--max-files-per-commit` are skipped as bulk commits, so generated or mechanical changes do not distort the ranking.

| Flag | Default | Purpose |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | Output file path |
| `--format` | `skt` | Output format: `skt` or `json` |
| `--since` | `12mo` | History window, such as `180d`, `12mo`, or `2y` |
| `--max-commits` | `2000` | Maximum commits to inspect |
| `--max-files-per-commit` | `50` | Exclude commits that change more files than this |
| `--min-cochanges` | `3` | Minimum shared commits to report temporal coupling |
| `--top-n` | `30` | Maximum results per section |
| `--include-merges` | `false` | Include merge commits in history metrics |
| `--collect-test` | `false` | Include test files |
| `--workers` | `0` (`NumCPU`) | Maximum concurrent parsing goroutines |
| `--verbose` | `false` | Print progress information |

### Read the JSON output

With `--format json`, each hotspot has `score`, `history_score`, `structure_score`, `recency_score`, `pagerank`, `betweenness`, `commits`, `churn`, and `transitive_fan_in`. Each temporal coupling entry has `left`, `right`, `strength`, and `co_changes`. The top level also records `commits_visited`, `commits_analyzed`, `skipped_merges`, and `skipped_bulk_commits`. Use these counts to check how much history the ranking is based on.

## Control which sources are scanned

All five source-processing commands (`parse`, `fingerprint`, `graph show`, `graph analyze`, and `graph hotspots`) accept the same selection flags:

| Flag | Default | Purpose |
| --- | --- | --- |
| `--include GLOB` | All supported sources | Include matching paths; repeatable |
| `--exclude GLOB` | None | Exclude matching paths or directories; repeatable. Excludes win over includes |
| `--max-file-size BYTES` | `10485760` (10 MiB) | Skip larger files; `0` disables the limit |
| `--strict` | `false` | Fail before writing output if the scan or parse is incomplete |

Patterns are gitignore-style globs relative to the input directory. Quote them so your shell does not expand them. Includes do not override `.gitignore` rules or test filtering.

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

Symlinks and non-regular files are skipped and reported. Scan and parse problems appear on stderr. Without `--strict`, codeknit keeps usable partial results, so check stderr before trusting a report. With `--strict`, an incomplete analysis stops before any output is written. A run where every source read fails always fails.

For the exact guarantees, see the [analysis reliability reference](../../reference/analysis-reliability/).

## Related pages

- [CLI flags reference](../../reference/cli-flags/): every flag for every command.
- [Output format reference](../../reference/output-format/): `.skt` and JSON syntax.
- [Language support reference](../../reference/language-support/): extraction and resolution limits.
- [Parse command guide](../parse-command/): extract symbols and edges for individual files.