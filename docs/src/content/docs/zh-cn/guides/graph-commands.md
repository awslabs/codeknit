---
title: 图命令
description: 使用图算法可视化和分析代码库结构。
---

使用 `graph` 子命令查看代码的连接方式，发现结构性问题，并决定优先重构哪些文件。本指南假定您已[安装 codeknit](../../installation/)并能运行 [`codeknit parse`](../parse-command/)。

| 任务 | 命令 |
| --- | --- |
| 可视化探索图 | `codeknit graph show` |
| 查找结构性问题 | `codeknit graph analyze` |
| 查找易变且重要的文件 | `codeknit graph hotspots` |

## 可视化代码图

`graph show` 解析输入并生成独立的 HTML 文件。符号（函数、类、类型）为节点。关系（调用、包含、实现）为边。

```bash
codeknit graph show ./myproject
```

文件将写入 `./skeleton/codeknit-graph.html` 并在默认浏览器中打开。

选择输出路径或包含测试文件：

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| 标志 | 默认值 | 用途 |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | 输出 HTML 文件路径 |
| `--collect-test` | `false` | 包含测试文件 |
| `--workers` | `0` (`NumCPU`) | 最大并发解析 goroutine 数 |
| `--verbose` | `false` | 打印进度信息 |

### 检查未解析的关系

查看器仅绘制已解析的关系。它会显示不确定（`unresolved` 或 `ambiguous`）关系的计数，但不会绘制这些关系或其候选链接。

要查看单个不确定的边，请使用 `--edges` 运行 `parse`：

```bash
codeknit parse ./src --output-mode inline --edges
```

添加 `--format json` 以获取 JSON 输出。有关边及候选项的写入方式，请参阅[输出格式参考](../../reference/output-format/)。

## 分析结构并查找问题

`graph analyze` 运行 22 种图算法并生成 LLM 可读的 `.skt` 报告。

```bash
codeknit graph analyze ./myproject
```

报告将写入 `./skeleton/graph_analysis.skt`。可直接阅读，或与相关源代码一起提供给 LLM。

### 调整报告

提高阈值以减少发现，或降低阈值以增加发现：

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0` 移除排名部分的上限。

| 标志 | 默认值 | 用途 |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | 输出 `.skt` 文件路径 |
| `--collect-test` | `false` | 包含测试文件 |
| `--workers` | `0` (`NumCPU`) | 最大并发解析 goroutine 数 |
| `--verbose` | `false` | 打印进度信息 |
| `--fan-threshold` | `10` | 标记中心符号的最小扇入或扇出 |
| `--god-threshold` | `15` | 标记 god class 或函数的最小包含边计数 |
| `--max-inheritance-depth` | `5` | 标记超过此深度的继承链 |
| `--top-n` | `30` | 限制排名部分；`0` 表示无限制 |
| `--betweenness-threshold` | `0.001` | 报告的最小介数中心性 |
| `--propagation-cutoff` | `0.05` | 继续变更传播的最小启发式分数 |

### 了解哪个发现回答哪个问题

| 算法 | 用于查找 |
| --- | --- |
| 循环依赖（Tarjan 的 SCC） | 相互依赖形成循环的符号组 |
| 中心检测 | 具有高扇入或扇出耦合的符号 |
| 孤立检测 | 死代码候选 |
| god class/函数检测 | 包含过多子项的符号 |
| 不稳定性度量 | 脆弱或稳定的符号和包，使用 Ce/(Ca+Ce) |
| 深继承链 | 超过 `--max-inheritance-depth` 的层次结构 |
| 介数中心性 | 许多依赖路径通过的瓶颈符号 |
| 割点 | 移除后会分割图的单点故障 |
| PageRank | 递归重要的符号 |
| 传递扇入 | 影响范围：直接或间接依赖某个符号的所有内容 |
| 变更传播模拟 | 更改某个符号可能传播的范围 |
| 循环包依赖 | 包级别的依赖循环 |
| 层违反检测 | 跨层错误方向的依赖 |
| 从入口点可达性 | 入口点无法到达的代码 |
| 弱连通分量 | 代码中的孤立簇 |
| 依赖权重 | 包之间的耦合强度 |
| 主序列距离 | 抽象性和不稳定性失衡的包 |
| 散弹手术检测 | 在包内多处进行的更改 |
| 特性依恋检测 | 使用另一个包多于自身包的符号 |
| 稳定依赖违反 | 稳定的包依赖于不太稳定的包 |
| 接口隔离违反 | 用户大多不使用其成员的接口 |
| 包含深度 | 深度嵌套的符号 |

### 正确阅读结果

以下规则影响您对发现的信任程度：

- **不确定的关系被排除。** 依赖度量仅使用已解析的边。报告会统计排除的 `unresolved` 和 `ambiguous` 关系。
- **文件和包含元数据被忽略。** `meta/file` 和 `meta/include` 符号不影响依赖排名或死代码发现。
- **扇入和扇出计算不同的邻居。** 包含不算作依赖。
- **继承深度涵盖所有基类和接口。** 循环层次结构不分配深度。
- **可达性从 `main`、`Main` 或 `init` 可调用项开始**，并沿依赖边跟踪。可达成员使其容器相关，但不影响其兄弟成员。如果扫描未发现公认的入口点，则省略可达性发现。
- **回调是引用，而非调用。** 回调参数或返回的可调用项并不证明回调会运行。
- **分数是启发式的。** 变更传播使用最强依赖路径。其分数和层规则不是测量概率或编译器保证。

### 安全地采取行动

将每个发现视为审查候选。在更改代码之前，打开报告中提到的符号并检查：

- 扫描目录之外的外部调用者。
- 动态调度。
- 未扫描的依赖。

分析基于语法，因此上述任何情况都可能改变结论。有关解析规则和限制，请参阅[语言支持参考中的关系解析](../../reference/language-support/#relationship-resolution)。

## 查找历史感知的热点

`graph hotspots` 对经常变更且位于结构重要位置的文件进行排名。它还报告时间耦合：在相同提交中重复变更的文件。在 Git 仓库内的目录上运行。

```bash
codeknit graph hotspots ./myproject
```

报告将写入 `./skeleton/hotspots.skt`。

它结合了两组度量：

- **历史：** 提交频率、新增和删除的行变更以及最近性。
- **结构：** 文件级 PageRank、传递扇入和介数。

报告包含基于分析的相关提交数量的置信度。将低置信度报告的排名视为暂定。

### 更改历史窗口和过滤器

```bash
# 分析两年并输出 JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# 允许更大的提交并要求更强的时间耦合
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

合并提交默认被排除，除非传递 `--include-merges`。触及超过 `--max-files-per-commit` 文件的提交被跳过，视为批量提交，因此生成或机械变更不会扭曲排名。

| 标志 | 默认值 | 用途 |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | 输出文件路径 |
| `--format` | `skt` | 输出格式：`skt` 或 `json` |
| `--since` | `12mo` | 历史窗口，如 `180d`、`12mo` 或 `2y` |
| `--max-commits` | `2000` | 最大检查提交数 |
| `--max-files-per-commit` | `50` | 排除变更文件超过此数量的提交 |
| `--min-cochanges` | `3` | 报告时间耦合的最小共享提交数 |
| `--top-n` | `30` | 每个部分的最大结果数 |
| `--include-merges` | `false` | 在历史度量中包含合并提交 |
| `--collect-test` | `false` | 包含测试文件 |
| `--workers` | `0` (`NumCPU`) | 最大并发解析 goroutine 数 |
| `--verbose` | `false` | 打印进度信息 |

### 阅读 JSON 输出

使用 `--format json` 时，每个热点包含 `score`、`history_score`、`structure_score`、`recency_score`、`pagerank`、`betweenness`、`commits`、`churn` 和 `transitive_fan_in`。每个时间耦合条目包含 `left`、`right`、`strength` 和 `co_changes`。顶层还记录 `commits_visited`、`commits_analyzed`、`skipped_merges` 和 `skipped_bulk_commits`。使用这些计数检查排名基于多少历史。

## 控制扫描的源文件

所有五个源处理命令（`parse`、`fingerprint`、`graph show`、`graph analyze` 和 `graph hotspots`）接受相同的选择标志：

| 标志 | 默认值 | 用途 |
| --- | --- | --- |
| `--include GLOB` | 所有支持的源文件 | 包含匹配的路径；可重复 |
| `--exclude GLOB` | 无 | 排除匹配的路径或目录；可重复。排除优先于包含 |
| `--max-file-size BYTES` | `10485760`（10 MiB） | 跳过更大的文件；`0` 禁用限制 |
| `--strict` | `false` | 如果扫描或解析不完整，则在写入输出前失败 |

模式是相对于输入目录的 gitignore 风格 glob。用引号括起来以防止 shell 展开。包含不会覆盖 `.gitignore` 规则或测试过滤。

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

符号链接和非常规文件会被跳过并报告。扫描和解析问题会显示在 stderr 上。没有 `--strict` 时，codeknit 会保留可用的部分结果，因此在信任报告前请检查 stderr。使用 `--strict` 时，不完整的分析会在写入任何输出前停止。每个源读取都失败的运行总是会失败。

有关确切保证，请参阅[分析可靠性参考](../../reference/analysis-reliability/)。

## 相关页面

- [CLI 标志参考](../../reference/cli-flags/)：每个命令的所有标志。
- [输出格式参考](../../reference/output-format/)：`.skt` 和 JSON 语法。
- [语言支持参考](../../reference/language-support/)：提取和解析限制。
- [解析命令指南](../parse-command/)：提取单个文件的符号和边。