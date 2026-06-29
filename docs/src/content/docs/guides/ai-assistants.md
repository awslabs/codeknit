---
title: Using with AI Assistants
description: Set up codeknit as a skill for Kiro, Claude Code, and other AI coding assistants.
---

codeknit ships with ready-made skills that teach AI coding assistants how to use it effectively. These skills enable assistants to extract code structure, detect duplicates, and perform structural analysis without manual prompting.

## Skills overview

codeknit provides two skills:

- **`codeknit-parse`**: Teaches assistants to extract code structure (functions, classes, methods, variables) and relationships (calls, inheritance, containment) into `.skt` files.
- **`codeknit-fingerprint`**: Teaches assistants to detect duplicate and near-duplicate code using fuzzy hashing.

Each skill includes documentation that the assistant reads on demand to understand usage, flags, output formats, and workflows.

## Installation

Use the installer helper to copy the skill directories to your assistant's skills folder. The installer downloads only the bundled skill files, so you do not need to clone the repository.

Install for **Codex**, **Kiro**, and **Claude Code**:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash
```

Install for one assistant:

```bash
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

After installation, the assistant automatically knows how to invoke codeknit commands, select appropriate flags, and interpret `.skt` output.

## What each skill teaches

### codeknit-parse

The `codeknit-parse` skill teaches assistants to:

- Run `codeknit parse` with appropriate flags for different scenarios
- Choose the right output mode:
  - `directory-flat` (default) for most projects
  - `inline` for single files or small inputs
  - `directory-tree` to mirror source structure
- Read and interpret `.skt` output files, including `[symbols]`, `[edges]`, and optional `[dict]` sections
- Use structural data for refactoring, dependency mapping, and code review
- Run `codeknit graph analyze` for deeper code quality insights (cyclic dependencies, hub symbols, god classes, etc.)

### codeknit-fingerprint

The `codeknit-fingerprint` skill teaches assistants to:

- Use `codeknit fingerprint` for duplicate detection, DRY audits, and refactor identification
- Select appropriate similarity ranges (`--min-similarity`, `--max-similarity`)
- Read the `[duplicates]` section to identify near-duplicate code
- Understand that fingerprints measure structural shape, not semantic intent
- Use `--rerank` with Ollama embeddings to reduce false positives when needed

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
