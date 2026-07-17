---
title: Graph Commands
description: Visualize and analyze your codebase structure with graph algorithms.
---

codeknit provides two powerful graph commands to help you understand and improve your codebase structure: `graph show` for interactive visualization and `graph analyze` for automated structural analysis.

## graph show

Generates an interactive HTML graph visualization of your codebase.

```bash
codeknit graph show <input-path>
```

This command parses your codebase and produces a self-contained HTML file with an interactive graph visualization. Symbols (functions, classes, types) appear as nodes, and their relationships (calls, contains, implements) as edges. The visualization opens automatically in your default browser.

### Flags

| Flag             | Default                          | Description                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Output HTML file path                        |
| `--collect-test` | `false`                          | Include test files in analysis               |
| `--workers`      | `NumCPU`                         | Max concurrent parsing goroutines            |
| `--verbose`      | `false`                          | Print progress information during processing |

### Examples

```skt
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

Runs structural graph algorithms on your codebase and emits an LLM-readable `.skt` report containing code quality insights.

```bash
codeknit graph analyze <input-path>
```

This command detects common code quality issues such as cyclic dependencies, hub symbols, dead code, god classes, and architectural bottlenecks.

### Algorithms

The analysis includes 22 structural graph algorithms:

- Cyclic dependencies (Tarjan's SCC)
- Hub detection (high fan-in/fan-out coupling)
- Orphan detection (dead code candidates)
- God class/function detection (excessive children)
- Instability metric (Robert C. Martin's Ce/(Ca+Ce))
- Deep inheritance chains
- Betweenness centrality (bottleneck detection)
- Articulation points (single points of failure)
- PageRank (recursive importance)
- Transitive fan-in (blast radius)
- Change propagation simulation
- Circular package dependencies
- Layer violation detection
- Reachability from entry points
- Weakly connected components
- Dependency weight (package coupling strength)
- Distance from Main Sequence (A+I balance)
- Shotgun surgery detection
- Feature envy detection
- Stable dependency violations
- Interface segregation violations
- Containment depth

### Flags

| Flag                      | Default                         | Description                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Output `.skt` file path                                  |
| `--collect-test`          | `false`                         | Include test files in analysis                           |
| `--workers`               | `NumCPU`                        | Max concurrent parsing goroutines                        |
| `--verbose`               | `false`                         | Print progress information during processing             |
| `--fan-threshold`         | `10`                            | Minimum fan-in or fan-out to flag a hub symbol           |
| `--god-threshold`         | `15`                            | Minimum contains-edge count to flag a god class/function |
| `--max-inheritance-depth` | `5`                             | Flag inheritance chains deeper than this                 |
| `--top-n`                 | `30`                            | Cap ranked output sections; 0 = no limit                 |
| `--betweenness-threshold` | `0.001`                         | Minimum betweenness centrality value to report           |
| `--propagation-cutoff`    | `0.05`                          | Minimum probability to continue change propagation       |

### Examples

```skt
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Include test files
codeknit graph analyze ./src --collect-test
```
