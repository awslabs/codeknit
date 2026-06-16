---
title: Getting Started
description: Get up and running with codeknit in under 5 minutes.
---

# Getting Started

Get up and running with codeknit in under 5 minutes.

## 1. Prerequisites

You'll need:

- Go 1.26+
- A C compiler (CGo is required for tree-sitter)

## 2. Installation from source

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# Binary is at ./bin/codeknit
```

## 3. Add to PATH

Add the binary to your shell's PATH:

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

Reload your shell or run `source ~/.bashrc` (or `~/.zshrc`) for the change to take effect.

## 4. Verify installation

Check that codeknit is working:

```bash
codeknit --version
```

## 5. First parse

Run your first parse on a codebase:

```bash
codeknit parse ./myproject
```

This command:

- Parses all source files in `./myproject`
- Extracts structural information (functions, classes, relationships)
- Writes chunked `.skt` files to `./skeleton/` (default output directory)

If you re-run this command, use `--clean` to remove previous output:

```bash
codeknit parse ./myproject --clean
```

## 6. Reading the output

The `.skt` files contain structured code information. Here's a small example:

```skt
[symbols]
## src/main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {exported}
S3 callable/function L10-L12 NewServer(addr: string) -> *S2 {exported}
S4 callable/method L14-L19 Start() {receiver=*Server}

[edges]
S2 --contains--> S4
S3 --returns--> S2
```

Key sections:

- `[symbols]`: Definitions grouped by file, showing name, line span, and metadata
- `[edges]`: Relationships like `contains`, `calls`, `inherits`, or `returns`

## 7. Next steps

Now that you've run your first parse:

- Learn more about the parse command: [Parse command guide](../guides/parse-command/)
- Explore structural analysis: [Graph commands guide](../guides/graph-commands/)
- Understand duplicate detection: [Fingerprint command guide](../guides/fingerprint-command/)
- Read the full output format: [Output format reference](../reference/output-format/)
- See all available flags: [CLI flags reference](../reference/cli-flags/)
