# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.1] - 2026-06-26

### Added

- JSON output for `codeknit parse` via `--format json`, including CLI, TUI, docs, and skill documentation.

## [0.1.0] - 2026-06-16

### Added

#### Core extraction

- Static code structure extraction for 12 languages: C, C++, C#, Go, Java, JavaScript, PHP, Python, Ruby, Rust, Scala, TypeScript
- Three output modes for `codeknit parse`: `inline`, `directory-flat`, `directory-tree`
- Compact `.skt` output format with `[symbols]`, `[edges]`, `[errors]`, and `[dict]` sections
- Type-aware signatures: parameter types and return types included for statically typed languages
- Type references in signatures resolved to symbol IDs (e.g., `Environment` → `S3`)
- Cross-file symbol resolution with directory proximity heuristics and import-aware disambiguation
- Configurable `--max-lines` per output file with automatic splitting
- Default output directory `./skeleton` when no output path is specified
- Stale output detection: refuses to overwrite existing `.skt` files without `--clean`
- `--minify` flag for dictionary-based token compression (shared `dict.skt`)
- `--edges` flag to include the `[edges]` section (omitted by default to save tokens)
- `--clean` flag to remove stale `.skt` files before writing
- `--collect-test` flag to include test files in analysis
- Parallel file parsing with configurable `--workers`

#### Graph visualization and analysis

- `codeknit graph show` subcommand: generates a self-contained interactive HTML visualization of the codebase graph, opened automatically in the default browser
- `codeknit graph analyze` subcommand: runs 17 structural graph algorithms and emits an LLM-readable `.skt` report, including:
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
- Configurable thresholds for hubs, god classes, inheritance depth, top-N ranking, betweenness, and change propagation

#### Fuzzy duplicate detection

- `codeknit fingerprint` subcommand: detects duplicate and near-duplicate code across files and languages using Context-Triggered Piecewise Hashing (CTPH) over normalized body tokens
- Structural type fingerprints that capture class/struct/interface shape (field count, method signatures, inheritance) independently from method-body logic
- `--min-similarity` and `--max-similarity` flags to bound the reported similarity range
- `--show-all` flag to emit the full per-symbol `[fingerprints]` section in addition to the `[duplicates]` pairs

#### Interactive terminal UI

- `codeknit` with no arguments launches an interactive TUI with command selection, directory completion, and field navigation
- Per-command screens for `parse`, `graph show`, `graph analyze`, and `fingerprint`

#### Tooling and distribution

- Per-command configuration architecture (Docker-style options structs) following the patterns used in kubectl, docker, gh, and hugo
- Deterministic output: planner iterates module maps in sorted key order so rerunning the tool produces byte-identical output
- Skill files for AI coding assistants in `skills/codeknit-parse/` and `skills/codeknit-fingerprint/`
- Generated documentation pipeline (`docs/gen`) using Amazon Bedrock for English and translations into Spanish, French, Italian, German, Japanese, Korean, Vietnamese, and Simplified Chinese
- Third-party license generation via `make third-party-licenses`
- Cross-plugin property test enforcing uniform signature format across all supported languages

[0.1.1]: https://github.com/awslabs/codeknit/releases/tag/v0.1.1
[0.1.0]: https://github.com/awslabs/codeknit/releases/tag/v0.1.0
