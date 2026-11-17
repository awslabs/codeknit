---
title: Getting Started
description: Get up and running with codeknit in under 5 minutes.
---

In this tutorial you will build codeknit from source, run your first parse, and read the `.skt` output it writes. You will parse a small package from the codeknit repository, so you do not need a project of your own.

## Prerequisites

- **Go 1.26 or later.** Run `go version` to check.
- **A C compiler.** codeknit uses tree-sitter, which needs CGo.
- **Git** and **Make**.

## 1. Build codeknit

Clone the repository and build:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

You should see the build finish without errors, and a binary at `./bin/codeknit`.

## 2. Add codeknit to your PATH

From the `codeknit` directory, add `bin` to your `PATH` using the line for your shell.

For bash or zsh:

```bash
export PATH="$PATH:$(pwd)/bin"
```

For fish:

```fish
fish_add_path (pwd)/bin
```

This applies to your current bash or zsh session. To keep it for later sessions, add the same line to `~/.bashrc` or `~/.zshrc`, with the absolute path to `bin` in place of `$(pwd)/bin`.

## 3. Verify the installation

```bash
codeknit --version
```

You should see the codeknit version printed.

## 4. Run your first parse

Stay in the `codeknit` directory and parse the `internal/console` package:

```bash
codeknit parse ./internal/console --edges
```

codeknit scans the directory, parses each supported source file, and writes `.skt` files. A *`.skt` file* is a compact text file that lists the symbols in your code and the relationships between them. The `--edges` flag adds the relationships. They are left out by default to save space.

You should see a short summary when the command finishes, and a new `./skeleton` directory in your current directory.

List the output:

```bash
ls ./skeleton
```

You should see a file such as `map_001.skt`.

## 5. Read the output

Open the file:

```bash
cat ./skeleton/map_001.skt
```

You should see a `[symbols]` section first. A *symbol* is a named item in your code, such as a function, type, or field. Symbols are grouped under a `## <source file>` heading. Your output includes lines like these:

```skt
[symbols]
## <path to>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exported}
...
S10 callable/function L28-L41 New() -> *S2 {exported}
...
S12 callable/method L49-L49 Error(msg: string) {exported, receiver=*Console}
```

Each line has the same parts:

| Part | Example | Meaning |
| --- | --- | --- |
| Short ID | `S10` | A short label for the symbol, used elsewhere in the output |
| Category and kind | `callable/function` | What kind of symbol it is |
| Line span | `L28-L41` | The lines the symbol covers in the source file |
| Signature | `New() -> *S2` | The name, parameters, and return type |
| Properties | `{exported}` | Extra details in braces |

In `New() -> *S2`, `S2` is the short ID of the `Console` type. codeknit replaces type names with short IDs when it knows the symbol. Line numbers can differ slightly if your checkout is newer.

Scroll down to the `[edges]` section. An *edge* is a relationship between two symbols. Each edge line names the source symbol, the kind of relationship, and the target symbol or symbols:

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

This example shows the shape of an edge line. It means "`S2` contains `S4`" and "`S4` calls `S5`". You can look up each short ID in `[symbols]` to see what it refers to.

## What you achieved

You built codeknit, put it on your `PATH`, ran `codeknit parse`, and read the `[symbols]` and `[edges]` sections of the `.skt` output.

## Next steps

- Parse your own project, choose a different output layout, and re-run safely: see the [parse command guide](../guides/parse-command/).
- Visualize and analyze your code structure: see the [graph commands guide](../guides/graph-commands/).
- Find duplicate and near-duplicate code: see the [fingerprint command guide](../guides/fingerprint-command/).
- Look up every part of the `.skt` format: see the [output format reference](../reference/output-format/).
- Look up every command-line flag: see the [CLI flags reference](../reference/cli-flags/).