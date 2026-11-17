---
title: Installation
description: How to install codeknit on your system.
---

This guide installs codeknit from source, adds it to your `PATH`, enables shell completions, and checks that it runs.

:::tip[Using an AI coding assistant?]
Claude Code, Codex, and Kiro users can install the codeknit plugin instead. The plugin downloads the binary the first time it runs. See [Use codeknit with AI assistants](../guides/ai-assistants/).
:::

## Install from source

**Prerequisites**

- Go 1.26 or later
- A C compiler. codeknit uses CGo to build its tree-sitter parsers.

**Steps**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

The binary is written to `./bin/codeknit`.

### Use a prebuilt binary instead

Prebuilt binaries for Linux, macOS, and Windows (amd64 and arm64) are on the [releases page](https://github.com/awslabs/codeknit/releases). Download the archive for your platform, extract it, and move the binary onto your `PATH`. Each release includes a `checksums.txt` file you can verify the download against.

## Add codeknit to your PATH

Skip this section if you moved the binary into a directory that is already on your `PATH`, such as `/usr/local/bin`.

Replace `/path/to/codeknit` with the directory that contains the binary. If you built from source, that is the `bin` directory inside the cloned repository.

Add the line for your shell to its configuration file:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

Reload your shell, or source the file you edited:

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

For fish, open a new terminal.

## Enable shell completions

codeknit generates completion scripts for bash, zsh, fish, and PowerShell. Run the command for your shell once:

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

The bash, zsh, and PowerShell commands append to your profile. Running them again adds the script a second time, so run each one only once.

Open a new shell, or source your profile, to activate completions.

## Verify the installation

```bash
codeknit --version
```

If the command prints a version, codeknit is installed. If your shell reports `command not found`, check that the directory in your `PATH` line contains the `codeknit` binary, then reload your shell.

## Set up a development environment

Do this only if you plan to contribute to codeknit. Run these commands from the root of the cloned repository:

```bash
make deps    # install dependencies
make setup   # configure git hooks
make test    # run the test suite
```

`make lint` runs the linter and formatter.

## Next steps

- Run your first command in [Getting started](../getting-started/).
- Browse every flag in the [CLI flags reference](../reference/cli-flags/).