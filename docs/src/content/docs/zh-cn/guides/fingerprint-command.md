---
title: fingerprint 命令
description: 使用模糊哈希检测跨文件和跨语言的重复和近似重复代码。
---

`codeknit fingerprint` 命令使用 **上下文触发的分段哈希（CTPH）** 在代码库中检测重复和近似重复代码。它通过在计算结构指纹前规范化变量名、字符串字面量和类型注解，实现跨文件甚至跨编程语言的检测。

## 功能概述

`codeknit fingerprint` 分析代码库中的每个函数、方法、变量和类型，并基于以下内容计算 **规范化结构指纹**：

- 控制流（`if`、`for`、`while`、`switch`）
- 操作（`=`、`+`、`==`、`&&`、`||`）
- 调用、返回、赋值和对象创建
- 语言结构，如 `try/catch`、`yield`、`await`、`defer`

这种规范化意味着 **重命名的复制粘贴**、**简单的重构** 以及 **不同语言中的等效逻辑** 仍然可以被检测为重复。

该算法使用 **CTPH**（一种滚动哈希变体）来高效查找近似重复。相似的代码会产生相似的指纹，即使代码经过轻微修改也能实现模糊匹配。

## 基本用法

```bash
codeknit fingerprint ./src
```

该命令将：

- 解析 `./src` 中的所有源文件
- 计算结构指纹
- 将结果输出到 `./skeleton/fingerprints.skt`
- 报告相似度在 **65% 到 95%** 之间的匹配（默认范围）

## 参数

| 参数               | 默认值                        | 描述                                                                                                                 |
| ------------------ | ----------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | 输出 `.skt` 文件路径                                                                                                 |
| `--min-similarity` | `65`                          | 报告的最小相似度百分比（0–100）                                                                                      |
| `--max-similarity` | `95`                          | 报告的最大相似度百分比（0–100）                                                                                      |
| `--show-all`       | `false`                       | 包含 `[fingerprints]` 部分，其中包含原始标记数据                                                                     |
| `--rerank`         | `false`                       | 使用 Ollama 的语义嵌入重新排序 CTPH 候选项，以消除误报（要求：`ollama serve` 和 `ollama pull qwen3-embedding:0.6b`） |
| `--model`          | `qwen3-embedding:0.6b`        | 与 `--rerank` 一起使用的 Ollama 嵌入模型                                                                             |
| `--collect-test`   | `false`                       | 在分析中包含测试文件                                                                                                 |
| `--workers`        | `NumCPU`                      | 最大并发解析 goroutine 数量（0 = 使用所有 CPU 核心）                                                                 |
| `--verbose`        | `false`                       | 在处理过程中打印进度信息                                                                                             |

## 输出格式

输出是一个 `.skt` 文件，包含以下部分：

### `[duplicates]`（始终存在）

列出相似度超过阈值的符号对：

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

每行显示：

- 相似度百分比
- 左符号（文件路径、作用域、名称）
- 右符号（文件路径、作用域、名称）

### `[fingerprints]`（仅在 `--show-all` 时出现）

包含每个符号的原始指纹数据：

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

字段：

- 符号名称
- `FP:<version>:<hash1>:<hash2>` — CTPH 指纹
- `tokens:<hex>` — 规范化的主体标记流

此部分适用于调试或构建下游工具。

## 常见模式

```bash
# 默认扫描
codeknit fingerprint ./src
```

```bash
# 查找完全重复的代码
codeknit fingerprint ./src --min-similarity 100
```

```bash
# 查找中等相似度的代码（例如相同算法，不同名称）
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

```bash
# 使用语义重新排序减少误报
# 要求：ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank
```

```bash
# 使用不同的嵌入模型进行重新排序
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

```bash
# 输出完整的指纹列表（用于分析工具）
codeknit fingerprint ./src --show-all
```

```bash
# 自定义输出文件
codeknit fingerprint ./src -o duplicates.skt
```

## 选择相似度范围

| 范围    | 指导建议                                                             |
| ------- | -------------------------------------------------------------------- |
| 96–100% | 完全或近乎完全的结构重复。几乎可以确定是复制粘贴。                   |
| 85–95%  | 近似重复。通常是经过少量编辑的复制粘贴（例如重命名变量、添加日志）。 |
| 65–84%  | 默认范围。结构高度相似。适合重构的候选项。                           |
| 50–64%  | 中等相似度。算法形状相同，但细节不同。需手动审查。                   |
| < 50%   | 通常为噪声。无意义的重复。                                           |

## 提示

- **指纹衡量的是结构而非含义**：高相似度得分意味着代码*看起来*相似，而不是*做*相同的事情。始终审查两个符号。
- **对噪声结果使用 `--rerank`**：如果出现大量误报，请启用语义重新排序，使用嵌入过滤匹配。
- **跳过短主体**：少于 4 个规范化标记的符号（例如简单的 getter）会被忽略，以避免噪声。
- **支持跨语言匹配**：等效结构（例如 Python 函数和具有相同逻辑的 Go 函数）可以匹配，但特定于语言的模式可能会产生低相似度的虚假匹配。
- **匹配是信号而非结论**：将每个匹配视为调查的提示，而非自动证明重复。
