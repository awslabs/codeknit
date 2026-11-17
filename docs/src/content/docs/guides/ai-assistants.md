---
title: Using with AI Assistants
description: Install the codeknit plugin for Claude Code, Codex, and Kiro.
---

codeknit ships as a plugin for Claude Code, Codex, and Kiro. The plugin contains one skill, `codeknit`, that teaches the assistant when to run `parse`, `graph analyze`, or `fingerprint`, and how to read the `.skt` output.

You do not need to install the codeknit binary yourself. On first use, the skill runs its bundled `scripts/install-codeknit.sh`. If `codeknit` is not on `PATH`, the script downloads the latest release for your platform, verifies it against the release `checksums.txt`, and installs it to `~/.local/bin`. Set `CODEKNIT_VERSION` to pin a release, or `CODEKNIT_INSTALL_DIR` to change the location.

## Installation

### Claude Code

Run in your shell, or inside a session as `/plugin ...`:

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

Open the Powers panel, choose **Add Custom Power → Import power from GitHub**, and enter `https://github.com/awslabs/codeknit`.

### Kiro CLI

Kiro CLI installs powers from a catalog or a local path:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

`--sparse` is optional. It limits the clone to the plugin manifests and `skills/` instead of the whole source tree.

### Updating

- Claude Code: `claude plugin marketplace update codeknit`, or enable auto-update under **Marketplaces** in `/plugin`
- Codex: `codex plugin marketplace upgrade`
- Kiro IDE: **Check for updates** on the power
- Kiro CLI: `git -C ~/.codeknit-power pull`, then run `kiro-cli powers install ~/.codeknit-power` again

### Manual install

For any assistant that reads Agent Skills, copy `skills/codeknit/` into its skills directory, such as `~/.claude/skills/`, `~/.codex/skills/`, or `~/.kiro/skills/`.

## What the skill teaches

The `codeknit` skill teaches assistants to:

- Pick the command by question: `parse` for what exists and what calls what, `graph analyze` for cycles, hubs, dead code, and god classes, and `fingerprint` for copy-paste and refactor candidates
- Default to `directory-flat` output and read chunks selectively, using `inline` only for single files
- Read `.skt` output instead of raw source for structural questions. `references/output-format.md` holds the full format.
- Choose fingerprint similarity ranges (`--min-similarity`, `--max-similarity`), use `--rerank` to cut false positives, and treat matches as signals of structural shape, not semantic intent

## Workflow examples

### Structural analysis

1. Ask the assistant to analyze your codebase structure
2. It runs `codeknit parse ./src` and reads the resulting `.skt` files
3. It answers structural questions: dependencies, call chains, dead code
4. For deeper insights, it runs `codeknit graph analyze ./src` and interprets the report

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### Duplicate detection

1. Ask the assistant to find duplicated code
2. It runs `codeknit fingerprint ./src`
3. It reads the `[duplicates]` section in the output
4. It investigates flagged pairs and proposes consolidation

```skt
[duplicates]
S1, S2: 87% similarity
S3, S4: 76% similarity
```

## Tips

- **Always read `.skt` files, not raw source, for structural questions** — they contain the extracted structure in a compact, reliable format
- Use `codeknit graph analyze` to uncover code quality issues like cyclic dependencies, hub symbols, and deep inheritance chains
- Run `codeknit fingerprint` before large refactors to identify copy-pasted code that should be consolidated
- The `.skt` format is designed to be token-efficient, making it ideal for LLM context windows
- Use `--minify` to further reduce token usage when processing large codebases
