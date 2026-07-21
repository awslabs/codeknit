---
title: CLI 参考
description: codeknit 所有命令和标志的完整参考。
---

## codeknit

启动交互式终端 UI（TUI），引导您完成可用的命令和选项。

```bash
codeknit
```

## codeknit parse

从源代码中提取结构信息，输出为 `.skt` 文件或 JSON。

```bash
codeknit parse <input-path> [output-dir]
```

| 标志               | 类型    | 默认值            | 描述                                                                                     |
| ------------------ | ------- | ----------------- | ---------------------------------------------------------------------------------------- |
| `--output-mode`    | string  | `directory-flat`  | 输出模式：`inline`、`directory-flat` 或 `directory-tree`                                 |
| `--format`         | string  | `skt`             | 输出格式：`skt` 或 `json`                                                               |
| `--max-lines`      | int     | `500`             | 每个输出文件的最大行数（适用于 `directory-flat` 和 `directory-tree` 模式）               |
| `--collect-test`   | bool    | `false`           | 在分析中包含测试文件                                                                     |
| `--minify`         | bool    | `false`           | 启用基于字典的输出压缩                                                                   |
| `--edges`          | bool    | `false`           | 在输出中包含 `[edges]` 部分（默认关闭以节省 token）                                      |
| `--clean`          | bool    | `false`           | 在写入前从输出目录中删除过时的 `.skt` 文件                                               |
| `--workers`        | int     | `0` (NumCPU)      | 最大并发解析 goroutine 数量                                                              |
| `--verbose`        | bool    | `false`           | 在处理过程中打印进度信息                                                                 |

当未指定输出目录时，默认为 `./skeleton`。在 `inline` 模式下，输出写入 stdout 且不使用目录。使用 `--format json` 时，目录输出将写入 `codeknit.json`。

## codeknit graph show

生成代码库结构的交互式 HTML 图可视化。

```bash
codeknit graph show <input-path>
```

| 标志               | 类型    | 默认值                              | 描述                                  |
| ------------------ | ------- | ----------------------------------- | ------------------------------------- |
| `-o`, `--output`   | string  | `./skeleton/codeknit-graph.html`   | 输出 HTML 文件路径                    |
| `--collect-test`   | bool    | `false`                             | 在分析中包含测试文件                  |
| `--workers`        | int     | `0` (NumCPU)                        | 最大并发解析 goroutine 数量           |
| `--verbose`        | bool    | `false`                             | 在处理过程中打印进度信息              |

生成的 HTML 文件是自包含的，并会在默认浏览器中自动打开。

## codeknit graph analyze

运行结构分析算法并生成 LLM 可读的 `.skt` 报告。

```bash
codeknit graph analyze <input-path>
```

| 标志                          | 类型     | 默认值                            | 描述                                                   |
| ----------------------------- | -------- | --------------------------------- | ------------------------------------------------------ |
| `-o`, `--output`              | string   | `./skeleton/graph_analysis.skt`   | 输出 `.skt` 文件路径                                   |
| `--collect-test`              | bool     | `false`                           | 在分析中包含测试文件                                   |
| `--workers`                   | int      | `0` (NumCPU)                      | 最大并发解析 goroutine 数量                            |
| `--verbose`                   | bool     | `false`                           | 在处理过程中打印进度信息                               |
| `--fan-threshold`             | int      | `10`                              | 标记枢纽符号的最小扇入或扇出值                         |
| `--god-threshold`             | int      | `15`                              | 标记 god class/function 的最小包含边数量               |
| `--max-inheritance-depth`     | int      | `5`                               | 标记超过此深度的继承链                                 |
| `--top-n`                     | int      | `30`                              | 限制排名输出部分的数量；`0` 表示无限制                 |
| `--betweenness-threshold`     | float64  | `0.001`                           | 报告的最小中介中心性值                                 |
| `--propagation-cutoff`        | float64  | `0.05`                            | 继续变更传播模拟的最小概率                             |

## codeknit graph hotspots

使用 Git 历史和结构重要性对文件进行排名，并报告重复一起变更的文件之间的时间耦合。

```bash
codeknit graph hotspots <input-path>
```

| 标志                          | 类型    | 默认值                     | 描述                                      |
| ----------------------------- | ------- | -------------------------- | ----------------------------------------- |
| `-o`, `--output`              | string  | `./skeleton/hotspots.skt`  | 输出文件路径                              |
| `--format`                    | string  | `skt`                      | 输出格式：`skt` 或 `json`                 |
| `--since`                     | string  | `12mo`                     | 历史窗口，例如 `180d`、`12mo` 或 `2y`    |
| `--max-commits`               | int     | `2000`                     | 检查的最大提交数                          |
| `--max-files-per-commit`      | int     | `50`                       | 排除变更文件数超过此值的提交              |
| `--min-cochanges`             | int     | `3`                        | 时间耦合的最小共享提交数                  |
| `--top-n`                     | int     | `30`                       | 每个报告部分的最大结果数                  |
| `--include-merges`            | bool    | `false`                    | 包含合并提交                              |
| `--collect-test`              | bool    | `false`                    | 包含测试文件                              |
| `--workers`                   | int     | `0` (NumCPU)               | 最大并发解析 goroutine 数量               |
| `--verbose`                   | bool    | `false`                    | 打印进度信息                              |

## codeknit fingerprint

使用模糊哈希检测重复和近似重复的代码。

```bash
codeknit fingerprint <input-path>
```

| 标志                     | 类型    | 默认值                          | 描述                                                                                                                  |
| ------------------------ | ------- | ------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`         | string  | `./skeleton/fingerprints.skt`   | 输出 `.skt` 文件路径                                                                                                  |
| `--min-similarity`       | int     | `65`                            | 报告的最小相似度百分比（0–100）                                                                                       |
| `--max-similarity`       | int     | `95`                            | 报告的最大相似度百分比（0–100）                                                                                       |
| `--show-all`             | bool    | `false`                         | 包含带有原始 token 数据的 `[fingerprints]` 部分                                                                       |
| `--rerank`               | bool    | `false`                         | 使用 Ollama 通过语义嵌入重新排序 CTPH 候选（需要 `ollama serve` 和 `ollama pull qwen3-embedding:0.6b`）              |
| `--model`                | string  | `qwen3-embedding:0.6b`          | 与 `--rerank` 一起使用的 Ollama 嵌入模型                                                                              |
| `--collect-test`         | bool    | `false`                         | 在分析中包含测试文件                                                                                                  |
| `--workers`              | int     | `0` (NumCPU)                    | 最大并发解析 goroutine 数量                                                                                           |
| `--verbose`              | bool    | `false`                         | 在处理过程中打印进度信息                                                                                              |

## codeknit completion

为支持的 shell 生成 shell 补全脚本。

```bash
codeknit completion <shell>
```

支持的 shell：`bash`、`zsh`、`fish`、`powershell`。

## 全局标志

| 标志           | 描述               |
| -------------- | ------------------ |
| `--version`    | 打印版本信息       |
| `--help`, `-h` | 显示当前命令的帮助 |