---
title: Installation
description: How to install codeknit on your system.
---

codeknit can be installed from source. The following steps will guide you through setting up codeknit on your system.

## From source

The primary installation method is building from source. You'll need:

- Go 1.26+
- A C compiler (required for tree-sitter via CGo)

Clone the repository and build the binary:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

The compiled binary will be available at `./bin/codeknit`.

## Add to PATH

To run `codeknit` from any directory, add the binary location to your system's PATH.

For **bash** (`~/.bashrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

For **zsh** (`~/.zshrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

For **fish** (`~/.config/fish/config.fish`):

```fish
fish_add_path /path/to/codeknit
```

After updating your shell configuration, reload it by running `source ~/.bashrc` (or `~/.zshrc`) or restart your terminal.

## Shell completions

codeknit supports auto-completion for popular shells. Install completions using these commands:

For **bash**:

```bash
codeknit completion bash >> ~/.bashrc
```

For **zsh**:

```bash
codeknit completion zsh >> ~/.zshrc
```

For **fish**:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

For **PowerShell**:

```powershell
codeknit completion powershell >> $PROFILE
```

## Verify installation

After installation, verify codeknit is correctly set up:

```bash
codeknit --version
```

## Development setup

If you're contributing to codeknit, run these additional commands:

Install development dependencies:

```bash
make deps
```

Set up git hooks:

```bash
make setup
```

Run the test suite:

```bash
make test
```
