---
title: 输出模式
description: 为项目规模和工作流选择合适的输出模式。
---

codeknit 将两个决策分开：输出的位置（**输出模式**）和输出的编码方式（**输出格式**）。本页解释每个选项的存在原因、成本及适用场景。有关 `.skt` 语法，请参阅[输出格式参考](../../reference/output-format/)。有关命令，请参阅[解析命令指南](../parse-command/)。

## 两个独立的选择

`--output-mode` 标志选择三种模式之一：`directory-flat`、`directory-tree` 或 `inline`。`--format` 标志选择 `skt`（默认）或 `json`。任何模式都可以与任一格式组合。

这两个选择是独立的，因为它们回答不同的问题。模式决定输出如何在磁盘上分割和定位，这对阅读者（人或 LLM）很重要。格式决定由谁消费输出：LLM 或人类阅读者，或脚本。

## directory-flat

`directory-flat` 是默认模式。它将编号的块文件 `map_001.skt`、`map_002.skt` 等写入 `./skeleton/`，除非指定其他输出目录。

**存在原因。** 大型代码库生成的结构超出单个上下文窗口的容量。分块允许阅读者仅打开所需的文件，而无需加载所有内容。

**分割方式。** codeknit 保持每个源文件的块完整，并在块达到大约 `--max-lines`（默认 500）时在块之间开始新的分块。该限制是目标值，而非硬性约束：一个包含非常大块的源文件可能生成超过 500 行的块。保持块完整意味着符号不会与同一文件中的兄弟符号分离。

**权衡。** 块不按源路径组织。要查找某个文件的结构，需在块中搜索其 `## source-path` 标头。如果经常查找特定文件，`directory-tree` 更合适。

**精简化。** 使用 `--minify` 时，字典会写入块旁边的 `dict.skt`。请参阅[精简化](#minification)。

## directory-tree

`directory-tree` 在输出目录下镜像源代码树。每个源文件在对应路径生成一个 `.skt` 文件。

**命名。** codeknit 将 `.skt` 附加到完整的源文件名，而非替换扩展名。`src/foo.c` 变为 `src/foo.c.skt`，`src/foo.h` 变为 `src/foo.h.skt`。若不遵循此规则，两者都会映射到 `foo.skt` 并发生冲突。当文件的输出被分割时，各部分命名为 `foo.c_part1.skt`、`foo.c_part2.skt` 等。输出路径冲突（包括仅大小写不同的冲突）会可见地失败，而非相互覆盖。

**分割与 flat 模式不同。** 在 tree 模式中，`--max-lines` 是每个源块的硬性限制，最小值为 3。每个符号块重复 `[symbols]` 和 `## source-path` 标头，以便单独阅读。仅包含边的块携带 `[edges]`。该限制适用于源块，而非辅助文件（如 `dict.skt` 和 `warnings.skt`）。

**权衡。** Tree 输出易于与源代码并行导航，但会生成大量小文件。大型项目可能生成数千个文件，需要广泛概览的阅读者必须打开多个文件。

## inline

`inline` 将所有内容写入 stdout。它不创建输出目录或临时区域。

**存在原因。** 适用于两种情况：快速检查单个文件，以及将输出通过管道传输到其他工具。这是唯一不将任何内容写入磁盘的模式。

**权衡。** 输出未分割。整个结果一次性进入终端或消费者的上下文，因此在超过少量文件时会使阅读者不堪重负。将其视为适用于单个文件或非常小的项目（大约少于五个文件）。由于 stdout 无法回滚，失败的消费者可能看到部分输出。目录模式无此问题。

## 输出格式

| 格式 | 最适用于 | 输出 |
| --- | --- | --- |
| `skt` | LLM 上下文和人工检查 | `.skt` 文件或 stdout |
| `json` | 脚本和结构化集成 | 目录模式下的 `codeknit.json`，或 inline 模式下的 stdout |

**skt** 紧凑且面向行。每个符号位于一行，签名中的换行符被转义，因此阅读者可以加载部分输出而不破坏其结构。这就是 `.skt` 支持分块的原因。

**json** 适用于需要类型化字段而非文本阅读的消费者。它是一个包含 `files`、`symbols` 以及（包含边时）`edges` 的单一文档。由于它是单一文档，因此不按行数分割。在目录模式中，`--max-lines` 对 JSON 无效，大型项目会生成一个大型的 `codeknit.json`。

简洁示例：

```json
{
  "files": ["app.go"],
  "symbols": [
    {"id": "app.go::User", "short_id": "S1", "name": "User", "file": "app.go",
     "category": "type", "kind": "struct", "signature": "type User struct", "span": [3, 3]},
    {"id": "app.go::Save", "short_id": "S2", "name": "Save", "file": "app.go",
     "category": "callable", "kind": "function", "signature": "Save(u: S1)", "span": [5, 5]}
  ],
  "edges": [
    {"from": "app.go::Save", "from_short": "S2", "to": "app.go::User",
     "to_short": "S1", "kind": "references"}
  ]
}
```

## 选择模式

| 模式 | 最适用于 | 输出 |
| --- | --- | --- |
| `directory-flat` | 大多数项目（默认） | `./skeleton/map_001.skt`、`map_002.skt`、... |
| `directory-tree` | 与源代码并行导航 | `./skeleton/<镜像路径>.skt` |
| `inline` | 单个文件、管道传输 | stdout |

经验法则：

- **不确定时**，使用 `directory-flat`（默认）。
- **单个文件**：`inline` 即可。
- **超过少量文件**：使用 `directory-flat` 或 `directory-tree`。
- **大型代码库**：添加 `--minify`。
- **重新运行到同一目录**：添加 `--clean`。

## 临时区域、重新运行和 `--clean`

目录模式不会在生成过程中直接写入输出目录。codeknit 在兄弟临时目录中生成输出，仅在生成成功后发布。因此，失败的运行会保留之前的输出。

默认情况下，当所选格式的生成输出已存在时，codeknit 拒绝写入。原因是符号 ID（`S1`、`S2` 等）在提取变化时可能改变。混合不同运行的块会导致边指向错误的符号。`--clean` 选择替换：生成成功后，过时的 `.skt` 文件（SKT 模式）或 `codeknit.json`（JSON 模式）会被替换。目录中不相关的条目会被保留。如果其中之一与生成路径冲突，运行会失败而非覆盖。

发布过程会获取目录特定的锁，备份之前的输出，并在替换失败时恢复。如果进程在两次目录重命名之间被中断，下次运行会在生成任何内容前恢复备份。这可从进程失败中恢复，但不会使目录更新瞬时完成：重命名之间存在短暂间隔，因此需要不间断访问的阅读者应与生成协调。它也无法防止存储故障或断电。

不要选择源目录或其祖先目录作为输出目录。有关完整的发布保证，请参阅[分析可靠性参考](../../reference/analysis-reliability/)。

## 精简化

`--minify` 应用基于字典的压缩。codeknit 将重复的标记（如 `callable/method` 或 `exported`）替换为短代码（`d0`、`d1` 等）。映射存储在 `[dict]` 部分，在目录模式中为 `dict.skt` 文件。

节省来自重复。大多数符号行共享相同的类别、种类和属性，因此大型代码库会重复这些字符串数千次。成本是间接性：阅读者必须使用字典展开代码，单个块不再自包含。字典是单独的辅助输出，因此不计入 `--max-lines`。

精简化输出编码关系类型，但保持解析标记可读，因此 `unresolved` 和 `ambiguous` 边仍可区分。对于小型项目，节省有限，额外步骤通常不值得。对于大型项目，节省的令牌可能非常可观。

## 相关页面

- [解析命令指南](../parse-command/)：每种模式的命令。
- [输出格式参考](../../reference/output-format/)：`.skt` 和 JSON 语法。
- [CLI 标志参考](../../reference/cli-flags/)：所有标志及默认值。
- [分析可靠性参考](../../reference/analysis-reliability/)：输出命名和发布保证。