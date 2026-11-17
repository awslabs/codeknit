---
title: CLI 参考
description: codeknit 所有命令及标志的完整参考。
---

本页面列出了每个 codeknit 命令、其参数及标志。有关面向任务的说明，请参阅[解析命令指南](../../guides/parse-command/)、[图命令指南](../../guides/graph-commands/)和[指纹命令指南](../../guides/fingerprint-command/)。

## 命令概览

| 命令                                       | 用途                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | 启动交互式终端 UI。                            |
| `codeknit parse <input-path> [output-dir]`    | 提取符号和关系到 `.skt` 或 JSON 输出。  |
| `codeknit graph show <input-path>`            | 生成交互式 HTML 图。                            |
| `codeknit graph analyze <input-path>`         | 运行结构化图算法并写入 `.skt` 报告。     |
| `codeknit graph hotspots <input-path>`        | 使用 Git 历史和图结构对变更热点进行排名。    |
| `codeknit fingerprint <input-path>`           | 检测重复和近似重复代码。                      |
| `codeknit completion <shell>`                 | 生成 shell 补全脚本。                            |

有五个命令处理源代码：`parse`、`graph show`、`graph analyze`、`graph hotspots` 和 `fingerprint`。它们共享[共享源代码处理标志](#共享源代码处理标志)中描述的标志。

## 全局标志

| 标志              | 类型 | 默认值 | 描述                  |
| ----------------- | ---- | ------- | ---------------------------- |
| `--version`       | bool | `false` | 打印版本信息。   |
| `-h`, `--help`    | bool | `false` | 显示命令的帮助信息。   |

## 共享源代码处理标志

这些标志出现在每个源代码处理命令的表格中。

| 标志                    | 类型                | 默认值               | 允许值       | 描述                                                                          |
| ----------------------- | ------------------- | --------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                   | `c`, `cpp`           | 用于 `.h` 文件的解析器。                                                          |
| `--include GLOB`        | string, repeatable  | 所有支持的源文件 | Gitignore 风格的 glob | 包含匹配的源文件路径。多个模式形成并集。                       |
| `--exclude GLOB`        | string, repeatable  | 无                  | Gitignore 风格的 glob | 排除匹配的路径或目录。                                               |
| `--max-file-size BYTES` | int                 | `10485760` (10 MiB)   | `0` 或更大       | 跳过大于此大小的源文件。`0` 禁用限制。                     |
| `--strict`              | bool                | `false`               | —                    | 如果扫描或解析诊断出现，则在输出写入前失败。                    |

### `--header-language`

仅适用于 `.h` 文件。其他扩展名保留其通常的解析器（`.hpp` 和 `.hxx` 已经使用 C++）。

- `c`（默认）保留 C 宏、typedef 和 union 提取。
- `cpp` 选择 C++ 解析器。适用于头文件名为 `.h` 的 C++ 项目。

TUI 对 `.h` 文件使用默认的 `c` 选择。

```bash
codeknit parse ./src --header-language cpp
```

有关每种语言测试的提取能力和已知限制，请参阅[语言支持和 API 保真度](../language-support/)。该页面描述的是测试的构造，而非完整的语言或编译器语义覆盖。

### 源文件选择

- 模式使用相对于输入目录的斜杠分隔的 gitignore 风格的 glob。对于单文件输入，模式匹配文件的 basename。
- 引用模式以防止 shell 展开它们。
- 排除优先于包含。
- 包含不会覆盖 `.gitignore` 规则、测试文件过滤或内置排除。
- `--include` 或 `--exclude` 不接受否定模式（`!pattern`）。
- 目录扫描独立于当前工作目录，遵循祖先和嵌套的 `.gitignore` 文件。
- 明确选择常规源文件会绕过忽略和测试过滤。包含、排除和大小限制仍然适用。
- 除非设置 `--collect-test`，否则测试文件被排除。

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### 诊断和 `--strict`

- 符号链接、目录符号链接和非常规文件被跳过并报告。
- 无法读取的路径、忽略文件读取失败、过大的源文件和语法警告会导致分析不完整。
- 默认情况下，保留可用的部分结果。诊断出现在 stderr 上。`parse` 还会在 JSON `errors` 或 SKT `[errors]` 中记录它们。
- 使用 `--strict` 时，不完整的分析在输出写入前被拒绝。
- 未选择任何支持的源文件，或每个源文件读取失败的运行，始终以失败退出。
- TUI 应用相同的符号链接策略和 10 MiB 默认值。其表单不公开选择和严格标志。

有关精确保证，请参阅[分析可靠性](../analysis-reliability/)。

## `codeknit`

```bash
codeknit
```

不带子命令运行 `codeknit` 会启动交互式终端 UI。该 UI 提供 `parse`、`graph show`、`graph analyze`、`graph hotspots` 和 `fingerprint` 命令及其选项。

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| 参数       | 必填 | 默认值      | 描述                                                              |
| -------------- | -------- | ------------ | ------------------------------------------------------------------------ |
| `input-path`   | 是      | —            | 源目录或单个源文件。                                  |
| `output-dir`   | 否       | `./skeleton` | 输出目录。当 `--output-mode` 为 `inline` 时不使用。             |

| 标志                    | 类型                | 默认值            | 允许值                                   | 描述                                                                                         |
| ----------------------- | ------------------- | ------------------ | ------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`   | `inline`, `directory-flat`, `directory-tree`     | 输出位置和方式。                                                                    |
| `--format`              | string              | `skt`              | `skt`, `json`                                    | 输出格式。                                                                                      |
| `--max-lines`           | int                 | `500`              | 对于 `directory-tree`，3 或更大               | 平面输出的行目标。树输出的硬源块限制。                               |
| `--collect-test`        | bool                | `false`            | —                                                | 包含测试文件。                                                                                 |
| `--minify`              | bool                | `false`            | —                                                | 启用基于字典的输出压缩。                                                         |
| `--edges`               | bool                | `false`            | —                                                | 包含 `[edges]` 部分。默认关闭以节省 token。                                       |
| `--clean`               | bool                | `false`            | —                                                | 仅在成功生成后替换之前生成的输出。                                   |
| `--workers`             | int                 | `0` (`NumCPU`)     | `0` 或更大                                   | 最大并发解析 goroutine 数量。                                                              |
| `--verbose`             | bool                | `false`            | —                                                | 打印进度信息。                                                                         |
| `--header-language`     | string              | `c`                | `c`, `cpp`                                       | 用于 `.h` 文件的解析器。                                                                         |
| `--include GLOB`        | string, repeatable  | 所有支持的源文件 | —                                             | 包含匹配的源文件路径。                                                                      |
| `--exclude GLOB`        | string, repeatable  | 无               | —                                                | 排除匹配的路径或目录。                                                              |
| `--max-file-size BYTES` | int                 | `10485760`         | `0` 禁用限制                           | 跳过更大的源文件。                                                                           |
| `--strict`              | bool                | `false`            | —                                                | 如果扫描或解析诊断出现，则在输出前失败。                                              |

### 输出模式

| 值             | 行为                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| `inline`          | 写入到 stdout。不使用输出目录。                                                    |
| `directory-flat`  | 将分块文件写入输出目录。源块保持完整，因此 `--max-lines` 是分割目标而非硬限制。 |
| `directory-tree`  | 镜像源树。名称保留源扩展名（`foo.c.skt`、`foo.h.skt`）。分割文件使用 `foo.c_part1.skt`。 |

### 输出行为

- 未使用 `--clean` 时，输出目录中选定格式的现有生成输出会导致错误。
- 使用 `--clean` 时，过时的 `.skt` 文件（SKT 格式）或 `codeknit.json`（JSON 格式）仅在成功生成后被替换。无关条目被保留。
- 在 JSON 格式中，目录模式写入 `codeknit.json`。JSON 是单个文档，不按行数分割。
- `dict.skt` 和 `warnings.skt` 是保留的辅助文件名。
- 不要将源目录或其祖先之一用作输出目录。

有关详细信息，请参阅[输出模式](../../guides/output-modes/)和[输出格式参考](../output-format/)。

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| 参数       | 必填 | 描述                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | 是      | 源目录或单个源文件。  |

| 标志                    | 类型                | 默认值                            | 允许值          | 描述                                                       |
| ----------------------- | ------------------- | ---------------------------------- | ----------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`   | 文件路径               | 输出 HTML 文件。                                                 |
| `--collect-test`        | bool                | `false`                            | —                       | 包含测试文件。                                               |
| `--workers`             | int                 | `0` (`NumCPU`)                     | `0` 或更大          | 最大并发解析 goroutine 数量。                            |
| `--verbose`             | bool                | `false`                            | —                       | 打印进度信息。                                       |
| `--header-language`     | string              | `c`                                | `c`, `cpp`              | 用于 `.h` 文件的解析器。                                       |
| `--include GLOB`        | string, repeatable  | 所有支持的源文件              | —                       | 包含匹配的源文件路径。                                    |
| `--exclude GLOB`        | string, repeatable  | 无                               | —                       | 排除匹配的路径或目录。                            |
| `--max-file-size BYTES` | int                 | `10485760`                         | `0` 禁用限制  | 跳过更大的源文件。                                         |
| `--strict`              | bool                | `false`                            | —                       | 如果扫描或解析诊断出现，则在输出前失败。            |

该命令写入一个独立的 HTML 文件并用默认浏览器打开它。查看器绘制已解析的关系并显示不确定关系的计数。

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| 参数       | 必填 | 描述                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | 是      | 源目录或单个源文件。  |

| 标志                       | 类型                | 默认值                            | 允许值          | 描述                                                  |
| -------------------------- | ------------------- | ---------------------------------- | ----------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`    | 文件路径               | 输出 `.skt` 文件。                                          |
| `--collect-test`           | bool                | `false`                            | —                       | 包含测试文件。                                          |
| `--workers`                | int                 | `0` (`NumCPU`)                     | `0` 或更大          | 最大并发解析 goroutine 数量。                       |
| `--verbose`                | bool                | `false`                            | —                       | 打印进度信息。                                 |
| `--fan-threshold`          | int                 | `10`                               | —                       | 标记中心符号的最小扇入或扇出。              |
| `--god-threshold`          | int                 | `15`                               | —                       | 标记 god 类或函数的最小 `contains` 边数量。 |
| `--max-inheritance-depth`  | int                 | `5`                                | —                       | 标记超过此值的继承链。              |
| `--top-n`                  | int                 | `30`                               | `0` = 无限制          | 限制排名输出部分。                                  |
| `--betweenness-threshold`  | float               | `0.001`                            | —                       | 报告的最小中介中心性值。              |
| `--propagation-cutoff`     | float               | `0.05`                             | —                       | 继续变更传播的最小概率。          |
| `--header-language`        | string              | `c`                                | `c`, `cpp`              | 用于 `.h` 文件的解析器。                                  |
| `--include GLOB`           | string, repeatable  | 所有支持的源文件              | —                       | 包含匹配的源文件路径。                               |
| `--exclude GLOB`           | string, repeatable  | 无                               | —                       | 排除匹配的路径或目录。                       |
| `--max-file-size BYTES`    | int                 | `10485760`                         | `0` 禁用限制  | 跳过更大的源文件。                                    |
| `--strict`                 | bool                | `false`                            | —                       | 如果扫描或解析诊断出现，则在输出前失败。       |

依赖度量排除了不确定的关系和文件/包含元数据。报告记录了排除的关系计数。

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| 参数       | 必填 | 描述                                                  |
| -------------- | -------- | ------------------------------------------------------------ |
| `input-path`   | 是      | Git 仓库内的源目录。                    |

| 标志                       | 类型                | 默认值                    | 允许值                              | 描述                                                  |
| -------------------------- | ------------------- | -------------------------- | ------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`  | 文件路径                                   | 输出文件。                                                 |
| `--format`                 | string              | `skt`                      | `skt`, `json`                               | 输出格式。                                               |
| `--since`                  | string              | `12mo`                     | 窗口如 `180d`、`12mo` 或 `2y`      | 历史窗口。                                              |
| `--max-commits`            | int                 | `2000`                     | —                                           | 检查的最大提交数。                                  |
| `--max-files-per-commit`   | int                 | `50`                       | —                                           | 排除触及文件数超过此值的提交。       |
| `--min-cochanges`          | int                 | `3`                        | —                                           | 报告时间耦合的最小共享提交数。          |
| `--top-n`                  | int                 | `30`                       | —                                           | 每个部分的最大结果数。                                 |
| `--include-merges`         | bool                | `false`                    | —                                           | 在历史度量中包含合并提交。                    |
| `--collect-test`           | bool                | `false`                    | —                                           | 包含测试文件。                                          |
| `--workers`                | int                 | `0` (`NumCPU`)             | `0` 或更大                              | 最大并发解析 goroutine 数量。                       |
| `--verbose`                | bool                | `false`                    | —                                           | 打印进度信息。                                 |
| `--header-language`        | string              | `c`                        | `c`, `cpp`                                  | 用于 `.h` 文件的解析器。                                  |
| `--include GLOB`           | string, repeatable  | 所有支持的源文件      | —                                           | 包含匹配的源文件路径。                               |
| `--exclude GLOB`           | string, repeatable  | 无                       | —                                           | 排除匹配的路径或目录。                       |
| `--max-file-size BYTES`    | int                 | `10485760`                 | `0` 禁用限制                      | 跳过更大的源文件。                                    |
| `--strict`                 | bool                | `false`                    | —                                           | 如果扫描或解析诊断出现，则在输出前失败。       |

该命令结合提交频率、行变更和最近性与文件级 PageRank、传递扇入和中介中心性。它还报告一起变更的文件之间的时间耦合。默认情况下排除合并提交。

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| 参数       | 必填 | 描述                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | 是      | 源目录或单个源文件。  |

| 标志                    | 类型                | 默认值                          | 允许值          | 描述                                                                                                      |
| ----------------------- | ------------------- | -------------------------------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`    | 文件路径               | 输出 `.skt` 文件。                                                                                              |
| `--min-similarity`      | int                 | `65`                             | `0`–`100`               | 报告的最小相似度百分比。                                                                         |
| `--max-similarity`      | int                 | `95`                             | `0`–`100`               | 报告的最大相似度百分比。                                                                         |
| `--show-all`            | bool                | `false`                          | —                       | 同时输出 `[fingerprints]` 部分，包含原始 token 数据。                                                      |
| `--rerank`              | bool                | `false`                          | —                       | 查找语义邻居并使用 Ollama 嵌入重新排序候选。需要 `ollama serve` 和 `ollama pull qwen3-embedding:0.6b`。 |
| `--model`               | string              | `qwen3-embedding:0.6b`           | Ollama 模型名称       | 与 `--rerank` 一起使用的 Ollama 嵌入模型。                                                                   |
| `--collect-test`        | bool                | `false`                          | —                       | 包含测试文件。                                                                                              |
| `--workers`             | int                 | `0` (`NumCPU`)                   | `0` 或更大          | 最大并发解析 goroutine 数量。                                                                           |
| `--verbose`             | bool                | `false`                          | —                       | 打印进度信息。                                                                                      |
| `--header-language`     | string              | `c`                              | `c`, `cpp`              | 用于 `.h` 文件的解析器。                                                                                      |
| `--include GLOB`        | string, repeatable  | 所有支持的源文件            | —                       | 包含匹配的源文件路径。                                                                                   |
| `--exclude GLOB`        | string, repeatable  | 无                             | —                       | 排除匹配的路径或目录。                                                                           |
| `--max-file-size BYTES` | int                 | `10485760`                       | `0` 禁用限制  | 跳过更大的源文件。                                                                                        |
| `--strict`              | bool                | `false`                          | —                       | 如果扫描或解析诊断出现，则在输出前失败。                                                           |

### 相似度范围

默认范围 65–95% 排除了 100% 匹配。要报告精确的归一化 token 匹配，请将两个边界设置为 `100`：

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

精确归一化 token 候选是完整的。当符合条件的符号有 500 个或更多时，其他近似重复候选使用近似的前 50 个邻居检索。`--rerank` 可能会移除结构上精确的候选。有关详细信息，请参阅[分析可靠性](../analysis-reliability/#duplicate-detection-guarantees)。

## `codeknit completion`

```bash
codeknit completion <shell>
```

| 参数  | 必填 | 允许值                         | 描述                                   |
| --------- | -------- | -------------------------------------- | --------------------------------------------- |
| `shell`   | 是      | `bash`, `zsh`, `fish`, `powershell`    | 生成补全脚本的 shell。 |

脚本被写入到 stdout。

```bash
codeknit completion zsh >> ~/.zshrc
```

## 相关页面

- [输出格式参考](../output-format/)
- [语言支持和 API 保真度](../language-support/)
- [分析可靠性](../analysis-reliability/)