---
title: Using with AI Assistants
description: Install the codeknit plugin for Claude Code, Codex, and Kiro.
---

Install the codeknit plugin in Claude Code, Codex, or Kiro, keep it current, and ask the assistant structural and duplicate-code questions. The plugin contains one skill, `codeknit`. The skill tells the assistant when to run `codeknit parse`, `codeknit graph analyze`, and `codeknit fingerprint`, and how to read the `.skt` output.

You don't need to install the `codeknit` binary yourself. The first time the skill runs, it runs its bundled `scripts/install-codeknit.sh`, which does the following:

1. Exits immediately if `codeknit` is already on your `PATH` or in `~/.local/bin`.
2. Otherwise, downloads the latest GitHub release for your OS and architecture.
3. Verifies the download against the release `checksums.txt`.
4. Installs the binary to `~/.local/bin` and prints its absolute path on the last line of output.

If `~/.local/bin` is not on your `PATH`, the assistant calls codeknit by that full path for the rest of the session.

To control the install, set these environment variables:

| Variable | Effect |
| --- | --- |
| `CODEKNIT_VERSION` | Pins a release, for example `0.6.0`. |
| `CODEKNIT_INSTALL_DIR` | Changes the install location. |

To install the binary yourself instead, see [Installation](../../installation/).

## Install the plugin

Pick the section for your assistant.

### Claude Code

Run these commands in your shell. Inside a Claude Code session, run the same commands as `/plugin ...`.

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

1. Open the Powers panel.
2. Choose **Add Custom Power**, then **Import power from GitHub**.
3. Enter `https://github.com/awslabs/codeknit`.

### Kiro CLI

The Kiro CLI installs from a local path, so clone the repository first:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### About `--sparse`

The `--sparse` option is optional. It limits the clone to the plugin manifests and `skills/` instead of the whole source tree.

### Install the skill by hand

If your assistant reads Agent Skills but has no plugin support, copy `skills/codeknit/` into the skills directory for your assistant:

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## Update the plugin

| Assistant | Command or action |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Choose **Check for updates** in the Powers panel. |
| Kiro CLI | Run `git -C ~/.codeknit-power pull`, then run `kiro-cli powers install ~/.codeknit-power` again. |

Updating the plugin does not update a binary that is already installed. To change the binary version, set `CODEKNIT_VERSION` as described above.

If you upgrade codeknit and the extraction changes, symbol IDs can change too. Ask the assistant to regenerate all output rather than mixing chunks from different runs.

## What the skill teaches the assistant

The skill gives the assistant these rules, so you don't have to repeat them in every request:

- **Pick the command by question.**
  - `parse` answers what exists, where it is, and what calls, inherits, or contains what.
  - `graph analyze` finds cycles, hubs, dead code, god classes, and bottlenecks.
  - `fingerprint` finds copy-paste and refactor candidates.
- **Default to `directory-flat` output.** The assistant writes chunked files to `./skeleton/` and reads only the chunks it needs. It uses `inline` output only for a single file or a handful of files, because `inline` dumps everything into the context window.
- **Read `.skt` output, not raw source, for structural questions.** The assistant opens source files only to inspect or change implementation details, using the line spans from the output. The skill's `references/output-format.md` holds the full format. See also the [output format reference](../../reference/output-format/).
- **Choose similarity ranges and treat matches as signals.** Fingerprints compare structural shape, not intent. The assistant reads both symbols of a pair before proposing a refactor.

## Ask for a structural analysis

1. Ask the assistant to analyze your codebase structure. For example: "Map the structure of `./src` and tell me which modules depend on the auth package."
2. The assistant ensures the binary is installed, then runs `codeknit parse` with edges enabled and reads the `.skt` output.
3. Ask follow-up questions about dependencies, call chains, or dead code. The assistant answers from the output.
4. For deeper analysis, ask for a structural review. For example: "Find cyclic dependencies and god classes in `./src`." The assistant runs `codeknit graph analyze` and reads the report.

Relationships marked `[unresolved]` or `[ambiguous]` in the output are uncertain, not confirmed dependencies. Treat answers that rest on them as candidates to verify.

For the commands behind these requests, see [the parse command guide](../parse-command/) and [the graph commands guide](../graph-commands/).

## Ask for duplicate detection

1. Ask the assistant to find duplicated code. For example: "Find copy-pasted functions in `./src`."
2. The assistant runs `codeknit fingerprint` and reads the `[duplicates]` section of `./skeleton/fingerprints.skt`.
3. The assistant opens both symbols of each flagged pair, then proposes a refactor.
4. If the results are noisy, ask it to adjust the similarity range or to add semantic reranking. Reranking requires a local Ollama server with an embedding model pulled.

For the options behind these requests, see [the fingerprint command guide](../fingerprint-command/).

## Tips

- Remind the assistant to read `.skt` files, not raw source, for structural questions. This keeps context small, since the format is compact and holds symbols, relationships, and line spans.
- Use `graph analyze` for code quality insights. Findings are review candidates, because external callers and dynamic dispatch can change the result.
- Run fingerprint before a large refactor to spot copy-paste that should be consolidated.
- If the assistant re-runs `parse` into a directory that already holds `.skt` output, codeknit refuses to write. Ask it to pass `--clean` to replace the earlier output.
- If the binary download fails (offline or unsupported platform), get a release manually from the [releases page](https://github.com/awslabs/codeknit/releases), or see [Installation](../../installation/).
- If a run reports partial results, check stderr and the `[errors]` section. See [analysis reliability](../../reference/analysis-reliability/) for the details.