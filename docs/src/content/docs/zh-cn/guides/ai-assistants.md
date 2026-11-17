---
title: 在 AI 助手中使用 codeknit
description: 为 Claude Code、Codex 和 Kiro 安装 codeknit 插件。
---

在 Claude Code、Codex 或 Kiro 中安装 codeknit 插件，保持其更新，并向助手询问有关代码结构和重复代码的问题。该插件包含一个技能 `codeknit`。该技能会告知助手何时运行 `codeknit parse`、`codeknit graph analyze` 和 `codeknit fingerprint`，以及如何读取 `.skt` 输出。

您无需自行安装 `codeknit` 二进制文件。首次运行该技能时，它会运行内置的 `scripts/install-codeknit.sh` 脚本，该脚本会执行以下操作：

1. 如果 `codeknit` 已经在您的 `PATH` 或 `~/.local/bin` 中，则立即退出。
2. 否则，下载适用于您的操作系统和架构的最新 GitHub 发布版本。
3. 根据发布版的 `checksums.txt` 验证下载。
4. 将二进制文件安装到 `~/.local/bin`，并在输出的最后一行打印其绝对路径。

如果 `~/.local/bin` 不在您的 `PATH` 中，助手将在会话的其余部分通过该完整路径调用 codeknit。

要控制安装过程，请设置以下环境变量：

| 变量 | 作用 |
| --- | --- |
| `CODEKNIT_VERSION` | 锁定发布版本，例如 `0.6.0`。 |
| `CODEKNIT_INSTALL_DIR` | 更改安装位置。 |

如果要自行安装二进制文件，请参阅[安装](../../installation/)。

## 安装插件

选择适用于您的助手的部分。

### Claude Code

在 shell 中运行以下命令。在 Claude Code 会话中，以 `/plugin ...` 形式运行相同的命令。

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

1. 打开 Powers 面板。
2. 选择**添加自定义 Power**，然后选择**从 GitHub 导入 power**。
3. 输入 `https://github.com/awslabs/codeknit`。

### Kiro CLI

Kiro CLI 从本地路径安装，因此请先克隆存储库：

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### 关于 `--sparse`

`--sparse` 选项是可选的。它将克隆限制在插件清单和 `skills/`，而不是整个源代码树。

### 手动安装技能

如果您的助手能读取 Agent Skills 但不支持插件，请将 `skills/codeknit/` 复制到助手的技能目录中：

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## 更新插件

| 助手 | 命令或操作 |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | 在 Powers 面板中选择**检查更新**。 |
| Kiro CLI | 运行 `git -C ~/.codeknit-power pull`，然后再次运行 `kiro-cli powers install ~/.codeknit-power`。 |

更新插件不会更新已安装的二进制文件。要更改二进制文件版本，请按上述方法设置 `CODEKNIT_VERSION`。

如果升级 codeknit 且提取方式发生变化，符号 ID 也可能会变化。要求助手重新生成所有输出，而不是混合来自不同运行的块。

## 技能教给助手的内容

该技能为助手提供了以下规则，因此您无需在每个请求中重复这些规则：

- **根据问题选择命令。**
  - `parse` 回答存在什么、位置在哪里，以及什么调用、继承或包含什么。
  - `graph analyze` 查找循环依赖、中心节点、死代码、god classes 和瓶颈。
  - `fingerprint` 查找复制粘贴和重构候选项。
- **默认使用 `directory-flat` 输出模式。** 助手将分块文件写入 `./skeleton/`，并仅读取所需的块。它仅在单个文件或少量文件时使用 `inline` 输出模式，因为 `inline` 会将所有内容转储到上下文窗口中。
- **读取 `.skt` 输出而不是原始源代码以回答结构问题。** 助手仅在检查或更改实现细节时打开源文件，并使用输出中的行范围。技能的 `references/output-format.md` 包含完整格式。另请参阅[输出格式参考](../../reference/output-format/)。
- **选择相似度范围并将匹配视为信号。** 指纹比较的是结构形状，而不是意图。助手在提出重构建议前会阅读一对符号。

## 请求结构分析

1. 要求助手分析您的代码库结构。例如：“映射 `./src` 的结构，并告诉我哪些模块依赖于 auth 包。”
2. 助手确保二进制文件已安装，然后运行带有边启用的 `codeknit parse` 并读取 `.skt` 输出。
3. 询问有关依赖、调用链或死代码的后续问题。助手将从输出中回答。
4. 要进行更深入的分析，请要求结构审查。例如：“在 `./src` 中查找循环依赖和 god classes。” 助手将运行 `codeknit graph analyze` 并读取报告。

输出中标记为 `[unresolved]` 或 `[ambiguous]` 的关系是不确定的，不是确认的依赖。将基于这些关系的答案视为需要验证的候选项。

有关这些请求背后的命令，请参阅[解析命令指南](../parse-command/)和[图命令指南](../graph-commands/)。

## 请求重复检测

1. 要求助手查找重复代码。例如：“在 `./src` 中查找复制粘贴的函数。”
2. 助手运行 `codeknit fingerprint` 并读取 `./skeleton/fingerprints.skt` 的 `[duplicates]` 部分。
3. 助手打开每个标记对的两个符号，然后提出重构建议。
4. 如果结果有噪声，要求助手调整相似度范围或添加语义重排。重排需要运行带有已拉取嵌入模型的本地 Ollama 服务器。

有关这些请求背后的选项，请参阅[指纹命令指南](../fingerprint-command/)。

## 提示

- 提醒助手在回答结构问题时读取 `.skt` 文件，而不是原始源代码。这可以保持上下文简洁，因为该格式紧凑且包含符号、关系和行范围。
- 使用 `graph analyze` 获取代码质量洞察。发现结果是审查候选项，因为外部调用者和动态分派可能会改变结果。
- 在大型重构之前运行 fingerprint 以发现应合并的复制粘贴代码。
- 如果助手将 `parse` 重新运行到已包含 `.skt` 输出的目录中，codeknit 将拒绝写入。要求助手传递 `--clean` 以替换之前的输出。
- 如果二进制文件下载失败（离线或不支持的平台），请从[发布页面](https://github.com/awslabs/codeknit/releases)手动获取发布版本，或参阅[安装](../../installation/)。
- 如果运行报告部分结果，请检查 stderr 和 `[errors]` 部分。有关详细信息，请参阅[分析可靠性](../../reference/analysis-reliability/)。