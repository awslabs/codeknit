---
title: 图命令
description: 使用图算法可视化和分析代码库结构。
---

codeknit 提供图命令，用于可视化结构、运行自动化分析，并将当前依赖图与 Git 变更历史相结合。

## graph show

生成代码库的交互式 HTML 图可视化。

```bash
codeknit graph show <input-path>
```

该命令解析代码库并生成一个独立的 HTML 文件，包含交互式图可视化。符号（函数、类、类型）显示为节点，其关系（调用、包含、实现）显示为边。可视化会自动在默认浏览器中打开。

### Flags

| Flag             | Default                          | Description                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | 输出 HTML 文件路径                           |
| `--collect-test` | `false`                          | 在分析中包含测试文件                         |
| `--workers`      | `NumCPU`                         | 最大并发解析 goroutine 数                    |
| `--verbose`      | `false`                          | 在处理过程中打印进度信息                     |

### 示例

```skt
# 生成默认可视化
codeknit graph show ./myproject

# 自定义输出文件
codeknit graph show ./myproject -o graph.html

# 包含测试文件
codeknit graph show ./src --collect-test
```

## graph analyze

对代码库运行结构图算法，并生成 LLM 可读的 `.skt` 报告，包含代码质量洞察。

```bash
codeknit graph analyze <input-path>
```

该命令检测常见的代码质量问题，如循环依赖、枢纽符号、死代码、god classes 以及架构瓶颈。

### 算法

分析包括 22 种结构图算法：

- 循环依赖（Tarjan 的 SCC）
- 枢纽检测（高扇入/扇出耦合）
- 孤立检测（死代码候选）
- god class/function 检测（过多子节点）
- 不稳定性度量（Robert C. Martin 的 Ce/(Ca+Ce)）
- 深继承链
- 中介中心性（瓶颈检测）
- 割点（单点故障）
- PageRank（递归重要性）
- 传递扇入（影响范围）
- 变更传播模拟
- 循环包依赖
- 层违反检测
- 从入口点的可达性
- 弱连通分量
- 依赖权重（包耦合强度）
- 与主序列的距离（A+I 平衡）
- 散弹手术检测
- 特性依恋检测
- 稳定依赖违反
- 接口隔离违反
- 包含深度

### Flags

| Flag                      | Default                         | Description                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | 输出 `.skt` 文件路径                                     |
| `--collect-test`          | `false`                         | 在分析中包含测试文件                                     |
| `--workers`               | `NumCPU`                        | 最大并发解析 goroutine 数                                |
| `--verbose`               | `false`                         | 在处理过程中打印进度信息                                 |
| `--fan-threshold`         | `10`                            | 标记枢纽符号的最小扇入或扇出值                           |
| `--god-threshold`         | `15`                            | 标记 god class/function 的最小包含边数                   |
| `--max-inheritance-depth` | `5`                             | 标记超过此深度的继承链                                   |
| `--top-n`                 | `30`                            | 限制排名输出部分；0 = 无限制                             |
| `--betweenness-threshold` | `0.001`                         | 报告的最小中介中心性值                                   |
| `--propagation-cutoff`    | `0.05`                          | 继续变更传播的最小概率                                   |

### 示例

```skt
# 使用默认设置运行结构分析
codeknit graph analyze ./myproject

# 自定义输出和阈值
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# 每个部分显示更多结果
codeknit graph analyze ./myproject --top-n 50

# 包含测试文件
codeknit graph analyze ./src --collect-test
```

## graph hotspots

对频繁变更且结构重要的文件进行排名：

```bash
codeknit graph hotspots <input-path>
```

得分结合提交频率、代码行变更和最近性，以及文件级 PageRank、传递扇入和中介中心性。报告还识别在相同提交中重复变更的文件之间的时间耦合。

默认情况下排除合并提交。变更超过 50 个文件的提交也会被排除，以避免生成、供应或机械批量变更扭曲结果。

### Flags

| Flag                     | Default                   | Description                                      |
| ------------------------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | 输出文件路径                                     |
| `--format`               | `skt`                     | 输出格式：`skt` 或 `json`                        |
| `--since`                | `12mo`                    | 历史窗口，如 `180d`、`12mo` 或 `2y`             |
| `--max-commits`          | `2000`                    | 检查的最大提交数                                 |
| `--max-files-per-commit` | `50`                      | 排除变更更多文件的提交                           |
| `--min-cochanges`        | `3`                       | 时间耦合的最小共享提交数                         |
| `--top-n`                | `30`                      | 每个报告部分的最大结果数                         |
| `--include-merges`       | `false`                   | 包含合并提交                                     |
| `--collect-test`         | `false`                   | 包含测试文件                                     |
| `--workers`              | `NumCPU`                  | 最大并发解析 goroutine 数                        |
| `--verbose`              | `false`                   | 打印进度信息                                     |

### 示例

```bash
# 分析最近 12 个月
codeknit graph hotspots ./myproject

# 分析两年并输出 JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# 包含更大的提交并要求更强的耦合
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```