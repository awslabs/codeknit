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

查看器绘制已解析的关系。它会显示不确定性计数，但不会绘制单个未解析或模糊的关系或其候选链接。使用解析输出检查这些关系：

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### 标志

| 标志               | 默认值                            | 描述                                  |
| ------------------ | --------------------------------- | ------------------------------------- |
| `-o`, `--output`   | `./skeleton/codeknit-graph.html`  | 输出 HTML 文件路径                    |
| `--collect-test`   | `false`                           | 在分析中包含测试文件                  |
| `--workers`        | `NumCPU`                          | 最大并发解析 goroutine 数             |
| `--verbose`        | `false`                           | 处理过程中打印进度信息                |

### 示例

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

在代码库上运行结构图算法，并生成 LLM 可读的 `.skt` 报告，包含代码质量洞察。

```bash
codeknit graph analyze <input-path>
```

该命令报告审查候选项，包括循环依赖、中心符号、死代码、上帝类和架构瓶颈。

### 算法

分析包括 22 种结构图算法：

- 循环依赖（Tarjan 的 SCC）
- 中心检测（高扇入/扇出耦合）
- 孤立检测（死代码候选）
- 上帝类/函数检测（过多子节点）
- 不稳定性指标（Robert C. Martin 的 Ce/(Ca+Ce)）
- 深继承链
- 中介中心性（瓶颈检测）
- 割点（单点故障）
- PageRank（递归重要性）
- 传递扇入（影响范围）
- 变更传播模拟
- 循环包依赖
- 层违规检测
- 从入口点的可达性
- 弱连通分量
- 依赖权重（包耦合强度）
- 与主序列的距离（A+I 平衡）
- 散弹手术检测
- 特性依恋检测
- 稳定依赖违规
- 接口隔离违规
- 包含深度

### 解读报告

依赖指标使用已解析的关系。报告会计算排除的未解析和模糊关系。文件/包含元数据不包含在依赖排名和死代码发现中。

- 扇入和扇出计算不同的邻居。重复关系不会增加这些计数，且包含不视为依赖。
- 继承深度考虑所有基类和接口。循环层次结构不会被分配深度。
- 可达性从名为 `main`、`Main` 或 `init` 的可调用对象开始，并沿依赖边跟踪。到达成员会使其容器相关，而无需标记每个兄弟节点为可达。如果没有识别的入口点，可达性发现将被省略。
- 回调参数和返回的可调用对象是引用，而不是回调执行的证明。
- 变更传播使用最强的依赖路径。其分数和基于目录的层规则是启发式的，而不是测量的概率。

外部调用者、动态分派和扫描范围之外的依赖可能会改变结论。有关解析规则和限制，请参阅[语言支持和 API 保真度](../../reference/language-support/)。

### 标志

| 标志                         | 默认值                          | 描述                                              |
| ---------------------------- | ------------------------------- | ------------------------------------------------- |
| `-o`, `--output`             | `./skeleton/graph_analysis.skt` | 输出 `.skt` 文件路径                              |
| `--collect-test`             | `false`                         | 在分析中包含测试文件                              |
| `--workers`                  | `NumCPU`                        | 最大并发解析 goroutine 数                         |
| `--verbose`                  | `false`                         | 处理过程中打印进度信息                            |
| `--fan-threshold`            | `10`                            | 标记中心符号的最小扇入或扇出                      |
| `--god-threshold`            | `15`                            | 标记上帝类/函数的最小包含边计数                   |
| `--max-inheritance-depth`    | `5`                             | 标记超过此深度的继承链                            |
| `--top-n`                    | `30`                            | 限制排名输出部分；0 = 无限制                      |
| `--betweenness-threshold`    | `0.001`                         | 报告的最小中介中心性值                            |
| `--propagation-cutoff`       | `0.05`                          | 继续变更传播的最小启发式分数                      |

### 示例

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

对既频繁变更又结构重要的文件进行排名：

```bash
codeknit graph hotspots <input-path>
```

得分结合提交频率、代码行变更量和最近性，以及文件级 PageRank、传递扇入和中介中心性。报告还识别在相同提交中重复变更的文件之间的时间耦合。

默认情况下排除合并提交。变更超过 50 个文件的提交也会被排除，以避免生成、供应或机械批量变更扭曲结果。

### 标志

| 标志                      | 默认值                    | 描述                                     |
| ------------------------- | ------------------------- | ---------------------------------------- |
| `-o`, `--output`          | `./skeleton/hotspots.skt` | 输出文件路径                             |
| `--format`                | `skt`                     | 输出格式：`skt` 或 `json`                |
| `--since`                 | `12mo`                    | 历史窗口，例如 `180d`、`12mo` 或 `2y`   |
| `--max-commits`           | `2000`                    | 检查的最大提交数                         |
| `--max-files-per-commit`  | `50`                      | 排除变更更多文件的提交                   |
| `--min-cochanges`         | `3`                       | 时间耦合的最小共享提交数                 |
| `--top-n`                 | `30`                      | 每个报告部分的最大结果数                 |
| `--include-merges`        | `false`                   | 包含合并提交                             |
| `--collect-test`          | `false`                   | 包含测试文件                             |
| `--workers`               | `NumCPU`                  | 最大并发解析 goroutine 数                |
| `--verbose`               | `false`                   | 打印进度信息                             |

### 示例

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```