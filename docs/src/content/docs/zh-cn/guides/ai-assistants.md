---
title: 与 AI 助手配合使用
description: 将 codeknit 设置为 Kiro、Claude Code 及其他 AI 编码助手的技能。
---

codeknit 内置了预制的技能，可教会 AI 编码助手如何有效使用它。这些技能使助手能够提取代码结构、检测重复代码并执行结构分析，而无需手动提示。

## 技能概览

codeknit 提供两个技能：

- **`codeknit-parse`**：教会助手提取代码结构（函数、类、方法、变量）及关系（调用、继承、包含）到 `.skt` 文件。
- **`codeknit-fingerprint`**：教会助手使用模糊哈希检测重复和近似重复代码。

每个技能都包含文档，助手可按需阅读以了解用法、标志、输出模式及工作流程。

## 安装

将技能目录复制到助手的技能文件夹。

对于 **Kiro**：

```bash
cp -r skills/codeknit-parse ~/.kiro/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.kiro/skills/codeknit-fingerprint
```

对于 **Claude Code**：

```bash
cp -r skills/codeknit-parse ~/.claude/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.claude/skills/codeknit-fingerprint
```

安装后，助手将自动知晓如何调用 codeknit 命令、选择适当的标志并解析 `.skt` 输出。

## 每个技能的作用

### codeknit-parse

`codeknit-parse` 技能教会助手：

- 根据不同场景运行带有适当标志的 `codeknit parse`
- 选择正确的输出模式：
  - `directory-flat`（默认）适用于大多数项目
  - `inline` 适用于单个文件或小型输入
  - `directory-tree` 用于镜像源代码结构
- 读取并解析 `.skt` 输出文件，包括 `[symbols]`、`[edges]` 和可选的 `[dict]` 部分
- 使用结构化数据进行重构、依赖映射和代码审查
- 运行 `codeknit graph analyze` 以获取更深入的代码质量洞察（循环依赖、枢纽符号、god classes 等）

### codeknit-fingerprint

`codeknit-fingerprint` 技能教会助手：

- 使用 `codeknit fingerprint` 进行重复检测、DRY 审计和重构识别
- 选择适当的相似度范围（`--min-similarity`、`--max-similarity`）
- 读取 `[duplicates]` 部分以识别近似重复代码
- 理解 fingerprints 测量的是结构形状，而非语义意图
- 在需要时使用 `--rerank` 与 Ollama 嵌入来减少误报

## 工作流示例

### 结构分析

1. 要求助手分析代码库结构
2. 助手运行 `codeknit parse ./src` 并读取生成的 `.skt` 文件
3. 助手回答结构相关问题：依赖关系、调用链、死代码
4. 如需更深入的洞察，助手运行 `codeknit graph analyze ./src` 并解析报告

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### 重复检测

1. 要求助手查找重复代码
2. 助手运行 `codeknit fingerprint ./src`
3. 助手读取输出中的 `[duplicates]` 部分
4. 助手调查标记的代码对并提出合并建议

```skt
[duplicates]
S1, S2: 87% 相似度
S3, S4: 76% 相似度
```

## 提示

- **对于结构相关问题，始终读取 `.skt` 文件而非原始源代码**——它们以紧凑、可靠的格式包含提取的结构
- 使用 `codeknit graph analyze` 揭示代码质量问题，如循环依赖、枢纽符号和深度继承链
- 在大型重构前运行 `codeknit fingerprint` 以识别应合并的复制粘贴代码
- `.skt` 格式专为令牌高效设计，非常适合 LLM 上下文窗口
- 在处理大型代码库时使用 `--minify` 进一步减少令牌使用