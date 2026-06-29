# codeknit

A static code structure extractor. `codeknit` parses source code and extracts structural information (functions, classes, methods, relationships) into a compact intermediate representation suitable for LLM consumption.

## Why

LLMs are great at generating code, but struggle with large-scale refactoring. They can't hold an entire codebase in context, so they miss duplication, inconsistent patterns, and architectural drift.

`codeknit` solves this by converting your codebase into a compact structural graph that focuses just on the code skeleton: function signatures, class hierarchies, call relationships, and module boundaries. No implementation details, no noise. LLMs can parse this representation directly and produce concrete, well-informed refactoring plans across hundreds of files.

## Supported languages

C, C++, C#, Go, Java, JavaScript, PHP, Python, Ruby, Rust, Scala, TypeScript

## Installation

### From prebuilt binaries

Prebuilt binaries are available on the [releases page](https://github.com/awslabs/codeknit/releases) for Linux, macOS, and Windows (amd64 and arm64).

Download the archive for your platform, extract it, and move the binary onto your `PATH`:

```bash
# macOS (Apple Silicon) — adjust OS/arch and version as needed
VERSION=0.1.0
OS=darwin   # darwin, linux, or windows
ARCH=arm64  # arm64 or amd64

curl -sSL -o codeknit.tar.gz \
  "https://github.com/awslabs/codeknit/releases/download/v${VERSION}/codeknit_${VERSION}_${OS}_${ARCH}.tar.gz"
tar -xzf codeknit.tar.gz
sudo mv codeknit /usr/local/bin/
```

On Windows, download the `..._windows_amd64.zip` archive from the releases page, extract it, and move `codeknit.exe` to a directory on your `PATH`.

Verify the download against the published checksums:

```bash
curl -sSL -O "https://github.com/awslabs/codeknit/releases/download/v${VERSION}/checksums.txt"
sha256sum --ignore-missing -c checksums.txt
```

Then confirm it runs:

```bash
codeknit --version
```

### From source

Requires Go 1.26+ and a C compiler (CGo is needed for tree-sitter).

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# Binary is at ./bin/codeknit
```

### Add to your PATH

If you installed the binary to a custom location instead of `/usr/local/bin/`, add it to your shell:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

Reload your shell or run `source ~/.bashrc` (or `~/.zshrc`) for the change to take effect.

Verify it works:

```bash
codeknit --version
```

### Shell completions

codeknit supports auto-completion for bash, zsh, fish, and PowerShell:

```bash
# bash
codeknit completion bash >> ~/.bashrc

# zsh
codeknit completion zsh >> ~/.zshrc

# fish
codeknit completion fish > ~/.config/fish/completions/codeknit.fish

# PowerShell
codeknit completion powershell >> $PROFILE
```

## Usage

codeknit has three main commands:

- `codeknit parse` — extract structural information into `.skt` files
- `codeknit fingerprint` — detect duplicate and near-duplicate code using fuzzy hashing
- `codeknit graph show` — generate an interactive HTML graph visualization
- `codeknit graph analyze` — run structural analysis algorithms and emit an LLM-readable report

### Parse a codebase

```bash
# Flat directory output (default, writes to ./skeleton)
codeknit parse ./myproject

# Custom output directory
codeknit parse ./myproject ./output

# Tree-mirroring directory output
codeknit parse ./myproject ./output --output-mode directory-tree

# Inline output to stdout
codeknit parse ./myproject --output-mode inline
```

### Interactive mode

Running `codeknit` with no arguments launches the interactive terminal UI, which guides you through all available commands and options.

```bash
codeknit
```

### Parse flags

```
Flags:
      --output-mode string   output mode: inline, directory-flat, directory-tree (default "directory-flat")
      --format string        output format: skt, json (default "skt")
      --max-lines int        maximum lines per output file (default 500)
      --collect-test         include test files in analysis
      --minify               enable dictionary-based output minification
      --edges                include the [edges] section in output (off by default to save tokens)
      --clean                remove stale .skt files from the output directory before writing
      --verbose              print progress information during processing
      --workers int          max concurrent parsing goroutines (0 = NumCPU)
```

The output directory defaults to `./skeleton` when not specified. You can override it by passing a second positional argument. In `inline` mode, no output directory is used — results go to stdout.

If the output directory already contains `.skt` files from a previous run, codeknit will refuse to write to avoid mixing stale and fresh output. Pass `--clean` to remove them automatically.

### Parse examples

```bash
# Include test files and minify output
codeknit parse ./src --collect-test --minify

# Include relationship edges (off by default to save tokens)
codeknit parse ./src --edges

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json

# Re-run and overwrite previous output
codeknit parse ./src --clean

# Custom output directory with tree layout
codeknit parse ./src ./out --output-mode directory-tree

# Limit output file size and parallelism
codeknit parse ./src --max-lines 500 --workers 4
```

### Detect duplicate code

```bash
# Find near-duplicates (65-95% similarity, default)
codeknit fingerprint ./myproject

# Find only exact duplicates
codeknit fingerprint ./myproject --min-similarity 100

# Find moderately similar code (50-80%)
codeknit fingerprint ./myproject --min-similarity 50 --max-similarity 80

# Semantic reranking — filters false positives via Ollama embeddings
# requires: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./myproject --rerank

# Semantic reranking with a different model
codeknit fingerprint ./myproject --rerank --model qwen3-embedding:4b

# Include raw fingerprint listing
codeknit fingerprint ./myproject --show-all

# Custom output file
codeknit fingerprint ./myproject -o duplicates.skt
```

`fingerprint` computes fuzzy hashes from a normalized intermediate representation of each function, method, variable, and type — capturing semantic operations (assignments, calls, comparisons, control flow) while ignoring variable names, string literals, and type annotations. This enables duplicate detection across different programming languages.

```
Flags:
  -o, --output string        output file path (default: ./skeleton/fingerprints.skt)
      --min-similarity int   minimum similarity percentage to report (0-100) (default 65)
      --max-similarity int   maximum similarity percentage to report (0-100) (default 95)
      --show-all             include the [fingerprints] section with raw token data
      --rerank               rerank CTPH candidates with semantic embeddings via Ollama to
                             eliminate false positives (requires: ollama serve && ollama pull
                             qwen3-embedding:0.6b)
      --model string         Ollama embedding model to use with --rerank (default: qwen3-embedding:0.6b)
      --collect-test         include test files in analysis
      --workers int          max concurrent parsing goroutines (0 = NumCPU)
      --verbose              print progress information during processing
```

### Visualize codebase structure

```bash
# Generate an interactive HTML graph (opens in browser)
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

`graph show` parses the codebase and produces a self-contained HTML file with an interactive graph visualization. Symbols (functions, classes, types) appear as nodes and their relationships (calls, contains, implements) as edges. The file opens automatically in your default browser.

```
Flags:
  -o, --output string   output HTML file path (default: ./skeleton/codeknit-graph.html)
      --collect-test    include test files in analysis
      --workers int     max concurrent parsing goroutines (0 = NumCPU)
      --verbose         print progress information during processing
```

### Analyze codebase structure

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Include test files
codeknit graph analyze ./src --collect-test
```

`graph analyze` runs structural graph algorithms on the codebase and emits an LLM-readable `.skt` report. It detects code quality issues such as cyclic dependencies, hub symbols, dead code, god classes, deep inheritance chains, bottleneck functions, and more.

Algorithms include: cyclic dependency detection (Tarjan's SCC), hub detection (fan-in/fan-out coupling), orphan detection, god class/function detection, instability metric, deep inheritance chains, betweenness centrality, articulation points, PageRank, transitive fan-in (blast radius), change propagation simulation, circular package dependencies, layer violation detection, reachability from entry points, weakly connected components, dependency weight, and distance from main sequence.

```
Flags:
  -o, --output string                    output .skt file path (default: ./skeleton/graph_analysis.skt)
      --collect-test                     include test files in analysis
      --workers int                      max concurrent parsing goroutines (0 = NumCPU)
      --verbose                          print progress information during processing
      --fan-threshold int                minimum fan-in or fan-out to flag a hub symbol (default 10)
      --god-threshold int                minimum contains-edge count to flag a god class/function (default 15)
      --max-inheritance-depth int        flag inheritance chains deeper than this (default 5)
      --top-n int                        cap ranked output sections; 0 = no limit (default 30)
      --betweenness-threshold float64    minimum betweenness centrality value to report (default 0.001)
      --propagation-cutoff float64       minimum probability to continue change propagation (default 0.05)
```

## Using with AI coding assistants

`codeknit` ships with ready-made skills in `skills/` that teach AI coding assistants how to use codeknit effectively. Install them to your home directory so they're available across all projects:

```bash
# Install for Codex, Kiro, and Claude Code without cloning this repository
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash

# Install for one assistant
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant codex
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant kiro
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant claude
```

From a local checkout, you can use Makefile helpers:

```bash
make skills-install-dry-run
make skills-install
```

The installer skips existing skill directories by default. To replace them, add `--force`:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant all --force
```

Once installed, the assistant knows how to invoke codeknit, pick the right output mode, read `.skt` files, use the structural graph for refactoring tasks, and detect duplicate code. No extra prompting needed.

The `codeknit-parse` skill includes:

- `SKILL.md` — usage guide, flags, output mode selection, and workflow
- `OUTPUT-FORMAT.md` — complete reference for the `.skt` output format

The `codeknit-fingerprint` skill teaches the assistant how to use the `fingerprint` command to find duplicate and near-duplicate code.

## Development

```bash
# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Lint and format
make lint

# Release (requires a git tag)
git tag v1.0.0
git push --tags
make release
```

## Disclaimer

This project was built almost entirely using [Kiro](https://kiro.dev), an AI-powered IDE. The codebase was then iteratively refactored using codeknit itself to identify structural improvements.

It's not meant to be used for production, leverage it to experiment and planning building features and refactoring with LLMs.

## License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.
