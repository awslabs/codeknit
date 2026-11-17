---
title: 指纹命令
description: 使用模糊哈希检测跨文件和语言的重复和近似重复代码。
---

`codeknit fingerprint` 计算每个函数、方法、变量和类型的规范化主体标记的上下文触发分段哈希（CTPH）。变量名、字符串字面量和类型注解在哈希前被规范化。因此，重命名的复制粘贴和移植到其他语言的代码仍然可以匹配。

本指南展示如何查找重复、调整相似度范围、减少误报，以及决定重构内容。

## 开始之前

- 已安装 `codeknit`。请参阅[安装](../../installation/)。
- 您有一个要扫描的源目录，例如 `./src`。

## 在项目中查找重复

在要检查的目录上运行命令：

```bash
codeknit fingerprint ./src
```

该命令将写入 `./skeleton/fingerprints.skt` 并报告相似度在 65% 到 95% 之间的匹配对。

打开该文件。`[duplicates]` 部分始终存在，最强的匹配排在前面：

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

每行的格式为 `similarity:<N>%  <file>::<symbol> <-> <file>::<symbol>`。如果没有匹配项落在该范围内，该部分仅包含 `# no duplicates found`。

有关完整的输出语法，请参阅[输出格式参考](../../reference/output-format/)。

## 仅查找精确重复

默认范围上限为 95%，因此 100% 的匹配被排除在外。要列出精确匹配，将两个边界设置为 100：

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

“精确”意味着在相同符号类别内，完整的规范化标记流完全相同。这并不意味着源文本完全相同，也不意味着两个符号被证明是等价的。

## 扩大范围以查找更宽松的匹配

降低最小值以包含共享算法形状但细节不同的代码：

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

低于 50% 的结果通常是噪声。

## 使用语义重排减少误报

如果结构化结果有噪声，使用本地 Ollama 服务器的嵌入重新排序候选项。

1. 启动 Ollama 并拉取默认嵌入模型：

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. 使用 `--rerank` 运行命令：

   ```bash
   codeknit fingerprint ./src --rerank
   ```

要使用其他嵌入模型，请传递 `--model`：

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

重排会添加语义候选项并基于组合分数进行过滤。它减少了误报，但并不能消除所有误报。它还可能移除结构上精确的候选项，且相似度范围仍然适用。嵌入是从用于提取的相同解析源快照计算得出的。

## 检查原始指纹

要查看每个符号的指纹，请添加 `--show-all`：

```bash
codeknit fingerprint ./src --show-all
```

这将添加一个 `[fingerprints]` 部分，按源文件分组，每个符号一行：

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

行格式为 `<symbol-name>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`。`tokens` 是十六进制编码的规范化标记流。

## 将报告写入其他位置

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## 控制扫描的文件

默认情况下，测试文件会被跳过。使用 `--collect-test` 包含它们：

```bash
codeknit fingerprint ./src --collect-test
```

要缩小扫描范围，请使用源选择标志。用引号括起模式，以防止 shell 展开：

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

其他选择控制：

- `--max-file-size BYTES` 跳过更大的文件。默认值为 `10485760`（10 MiB）。`0` 禁用限制。
- `--strict` 如果扫描或解析不完整，则在写入输出前失败。
- `--header-language c|cpp` 为模糊的 `.h` 文件选择解析器。请参阅[语言支持](../../reference/language-support/#header-language-selection)。

跳过的符号链接、过大的文件和语法警告会显示在 stderr 上。包含项不会覆盖 `.gitignore` 或测试过滤，且排除项优先于包含项。有关精确保证，请参阅[分析可靠性参考](../../reference/analysis-reliability/)。

## 选择相似度范围

| 范围       | 通常发现的内容                                      |
| ---------- | --------------------------------------------------- |
| `96-100%`  | 精确或近似精确的结构重复；明显的复制粘贴            |
| `85-95%`   | 近似重复；通常是带有微小编辑的复制粘贴              |
| `65-84%`   | 默认范围；强结构相似性，通常是重构候选              |
| `50-64%`   | 算法形状相同但细节不同                              |
| 低于 50%   | 通常是噪声                                          |

要查看前两行，请将 `--max-similarity` 提高到 `100`。

## 决定重构内容

1. 从默认范围开始，阅读最顶部的匹配对。
2. 打开每对符号并进行比较。
3. 将匹配视为信号，而非证明。指纹衡量的是结构形状，而非标识符名称或意图。以迭代和递归方式编写的相同逻辑得分较低，而具有相同形状的无关代码可能得分较高。
4. 85–95% 范围内的匹配对通常是最安全的合并候选。65–84% 范围内的匹配对通常需要更仔细的检查。
5. 如果列表有噪声，请缩小范围或添加 `--rerank`。

跨语言匹配适用于等效结构。特定于语言的特性可能产生虚假的低相似度匹配。

相似度百分比是一种启发式排序，而非编辑距离。项目的标记校准固件单独报告精确度和召回率，与相似度得分无关。这是一种回归检查，而非对您代码准确性的估计。

## 大型项目

- 符号数量少于 500 时，codeknit 会对每个同类别对进行评分。
- 符号数量为 500 或更多时，完全精确的规范化标记组始终保留。其他对使用每个符号的近似 top-50 邻居，因此近似重复结果可能不完整。
- 对于精确重复，请使用上述 100/100 命令。它不受近似影响。
- 对于稳定的输入和分数，排序是确定性的。

## 故障排除

| 症状                     | 原因及解决方法                                      |
| ------------------------ | --------------------------------------------------- |
| `# no duplicates found`  | 没有匹配对落在范围内。降低 `--min-similarity` 或提高 `--max-similarity`。 |
| 已知的短函数缺失         | 少于 4 个规范化标记的符号会被跳过。                 |
| 精确副本缺失             | 默认最大值为 95%。使用 `--min-similarity 100 --max-similarity 100`。使用 `--rerank` 时，语义过滤器也可能移除它们。 |
| `--rerank` 失败          | 检查 `ollama serve` 是否正在运行，以及模型是否已拉取。 |
| 带有 stderr 警告的部分结果 | 修复报告的文件，或添加 `--strict` 以在发布不完整输出前失败。 |

## 本指南中使用的标志

| 标志                     | 默认值                     | 描述                                                                 |
| ------------------------ | -------------------------- | -------------------------------------------------------------------- |
| `-o`, `--output`         | `./skeleton/fingerprints.skt` | 输出 `.skt` 文件路径。                                               |
| `--min-similarity`       | `65`                       | 报告的最小相似度百分比（0–100）。                                   |
| `--max-similarity`       | `95`                       | 报告的最大相似度百分比（0–100）。                                   |
| `--show-all`             | `false`                    | 同时输出包含原始标记数据的 `[fingerprints]` 部分。                  |
| `--rerank`               | `false`                    | 通过 Ollama 使用语义嵌入重新排序候选项。需要 `ollama serve` 和 `ollama pull qwen3-embedding:0.6b`。 |
| `--model`                | `qwen3-embedding:0.6b`     | 与 `--rerank` 一起使用的 Ollama 嵌入模型。                          |
| `--collect-test`         | `false`                    | 在分析中包含测试文件。                                               |
| `--workers`              | `NumCPU`                   | 最大并发解析 goroutine 数量。                                        |
| `--verbose`              | `false`                    | 打印进度信息。                                                       |

有关完整列表，包括 `--include`、`--exclude`、`--max-file-size`、`--strict` 和 `--header-language`，请参阅[CLI 标志参考](../../reference/cli-flags/)。

## 相关

- [解析命令](../parse-command/)
- [图命令](../graph-commands/)