---
title: Parse 命令
description: 从源代码中提取结构信息到 .skt 文件或 JSON。
---

`codeknit parse` 读取源文件并将其符号（以及可选的符号之间的关系）写入 `.skt` 文件或 JSON。本指南涵盖了您最可能遇到的情况。有关所有命令共享的完整标志列表，请参阅 [CLI 标志参考](../../reference/cli-flags/)。有关输出语法，请参阅 [输出格式参考](../../reference/output-format/)。

## 首次解析项目

```bash
codeknit parse ./src
```

codeknit 将分块的 `.skt` 文件写入 `./skeleton/`：`map_001.skt`、`map_002.skt` 等。选择性地打开这些块，而不是一次性加载所有内容。

要写入其他位置，请将输出目录作为第二个参数传递：

```bash
codeknit parse ./src ./output
```

不要将源目录或其祖先目录用作输出目录。

## 选择输出模式

使用 `--output-mode` 选择模式。

| 模式               | 使用时机                                      | 结果                                                                           |
| ------------------ | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `directory-flat`   | 解析大多数项目（默认）                        | `map_001.skt`、`map_002.skt`、...                                              |
| `directory-tree`   | 您希望查找单个文件的代码结构                  | 每个源文件生成一个 `.skt` 文件，镜像源树                                       |
| `inline`           | 解析单个文件，或通过管道传输到其他工具        | 输出到 stdout；不使用输出目录                                                 |

在树模式下，文件名保留源扩展名：`src/foo.c` 变为 `src/foo.c.skt`，`src/foo.h` 变为 `src/foo.h.skt`。大文件会被拆分为 `foo.c_part1.skt`、`foo.c_part2.skt` 等。

```bash
codeknit parse ./src --output-mode directory-tree
```

避免在整个项目上使用 `inline`。它会将所有内容写入 stdout，可能会淹没您的终端或 LLM 上下文窗口。

有关选择模式的更多信息，请参阅 [输出模式](../output-modes/)。

## 在同一项目上重新运行

codeknit 拒绝写入已包含所选格式生成输出的目录。这可以防止陈旧和新鲜的块混合。添加 `--clean` 以替换之前的输出：

```bash
codeknit parse ./src --clean
```

`--clean` 不会预先删除任何内容。codeknit 在临时目录中生成新输出，仅在生成成功后才替换旧的生成文件。如果生成失败或进程被中断，之前的输出将被保留或在下次运行时恢复。codeknit 未生成的输出目录中的文件将被保留。如果其中一个文件与生成路径冲突，运行将失败。

在 JSON 模式下，相同的规则适用于 `codeknit.json`。

**陷阱：混合块。** 符号 ID（`S1`、`S2` 等）和关系结果可能在不同运行和版本之间变化。升级到 0.5.0 后，使用 `--clean` 重新生成整个输出集。切勿将不同运行的块混合使用。

## 解析单个文件

将 `parse` 指向文件并使用 `inline` 模式：

```bash
codeknit parse ./src/main.go --output-mode inline
```

当明确选择常规源文件时，codeknit 会跳过忽略文件和测试文件过滤。包含/排除模式和大小限制仍然适用。

## 获取用于脚本的 JSON

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

JSON 文档具有顶级字段 `files`、`symbols`、`edges`（无边时省略）和 `errors`（无错误时省略）。一个小例子：

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

要将 JSON 写入磁盘，请去掉 `--output-mode inline`。在目录模式下，JSON 写入 `codeknit.json`。JSON 是单个文档，不会按行数拆分。

## 包含关系

默认情况下，输出仅包含 `[symbols]` 部分，以节省令牌。当您需要知道什么调用、包含、继承或引用什么时，添加 `--edges`：

```bash
codeknit parse ./src --edges
```

### 读取不确定的关系

Codeknit 分析语法时不使用编译器，因此无法建立某些关系。每条边都说明其确定性：

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->` 是已解析的关系。
- `--kind[unresolved]-->` 表示 codeknit 无法确定目标。
- `--kind[ambiguous]-->` 表示存在多个可能的目标。候选 ID 会在可用时列出。

候选是可能的目标，而不是确认的依赖项。已知端点使用短 ID。未解析的目标在没有匹配符号时可能显示为相对于源文件的名称。

在 JSON 中，不确定的边带有 `resolution`（`"unresolved"` 或 `"ambiguous"`），并在可用时带有 `candidates` 作为完整符号 ID。它们省略 `to_short`。

对于 C 和 C++，`#include` 指令变为 `meta/file` 和 `meta/include` 符号，因此它们的边具有短 ID。这记录了指令的书写形式。这并不意味着 codeknit 找到了头文件。

图分析排除不确定的边，并报告排除了多少。HTML 图查看器显示不确定关系的计数，但不绘制它们。要检查单个不确定边，请在 `parse` 上使用 `--edges`。有关这些报告的信息，请参阅 [图命令](../graph-commands/)。

边的语法在 [输出格式参考](../../reference/output-format/) 中。解析规则和限制在 [语言支持参考](../../reference/language-support/#relationship-resolution) 中。

## 减少输出大小

要压缩重复的令牌，请添加 `--minify`：

```bash
codeknit parse ./src --minify
```

这会将字典写入 `dict.skt`（或内联的 `[dict]` 部分）。在阅读时，将代码如 `d0` 和 `d1` 替换为其字典值。关系类型也会被编码，但解析标签如 `[unresolved]` 保持可读。

要控制块大小，请使用 `--max-lines`（默认 `500`）：

- 在 `directory-flat` 模式下，它是一个目标。源块永远不会被拆分，因此块可能超过该值一个块。
- 在 `directory-tree` 模式下，它是每个源块的硬限制。最小值为 `3`。每个块重复其 `[symbols]` 和 `## source-path` 头。

## 解析命名为 `.h` 的 C++ 头文件

Codeknit 默认将 `.h` 文件解析为 C。对于使用该扩展名的 C++ 头文件，请设置 `--header-language cpp`：

```bash
codeknit parse ./src --header-language cpp
```

该标志仅影响 `.h` 文件。命名为 `.hpp` 和 `.hxx` 的文件已被解析为 C++，其他扩展名保持其通常的解析器。

## 控制扫描哪些文件

所有五个源处理命令都接受这些标志。它们在 [CLI 标志参考](../../reference/cli-flags/) 中有完整列表。

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` 和 `--exclude GLOB` 是可重复的、相对于输入目录的 gitignore 风格模式。请用引号括起来，以免 shell 展开它们。排除优先。包含不会覆盖 `.gitignore`、测试文件过滤或内置排除。
- `--max-file-size BYTES` 跳过更大的文件。默认值为 `10485760`（10 MiB）。`0` 禁用该限制。
- `--strict` 如果扫描或解析不完整，则在写入任何输出之前失败。默认值为 `false`。

Codeknit 会跳过并报告符号链接和非常规文件。祖先 `.gitignore` 规则独立于您的工作目录应用。

**陷阱：部分结果。** 无法读取的路径、过大的文件、跳过的符号链接和语法警告会导致分析不完整。没有 `--strict` 时，codeknit 会保留可用的结果并报告问题：

- 在 stderr 上，
- 在 JSON 的 `errors` 字段中，或
- 在 SKT 的 `[errors]` 部分（内联，或在目录模式下的 `warnings.skt` 中）。

没有选定受支持的源的输入，或每个读取的源都失败的运行，总是以失败退出。有关精确保证，请参阅 [分析可靠性](../../reference/analysis-reliability/)。

## 解析标志

以下标志特定于 `parse`。共享标志在 [CLI 标志参考](../../reference/cli-flags/) 中。

| 标志               | 默认值          | 描述                                                                                       |
| ------------------ | ---------------- | ------------------------------------------------------------------------------------------ |
| `--output-mode`    | `directory-flat` | 输出模式：`inline`、`directory-flat` 或 `directory-tree`。                                 |
| `--format`         | `skt`            | 输出格式：`skt` 或 `json`。                                                                |
| `--header-language`| `c`              | `.h` 文件的解析器：`c` 或 `cpp`。其他扩展名保持不变。                                      |
| `--max-lines`      | `500`            | 平面输出的行目标；每个树源块的硬限制（最小值为 3）。                                       |
| `--collect-test`   | `false`          | 包含测试文件。                                                                             |
| `--minify`         | `false`          | 使用基于字典的压缩。                                                                       |
| `--edges`          | `false`          | 包含 `[edges]` 部分（默认关闭以节省令牌）。                                                |
| `--clean`          | `false`          | 仅在生成成功后替换之前生成的输出。                                                         |
| `--workers`        | `NumCPU`         | 最大并发解析 goroutine 数量。                                                              |
| `--verbose`        | `false`          | 打印进度信息。                                                                             |

## 提示

- 对于大多数项目，请使用 `directory-flat`，并仅读取所需的块。
- 在大型代码库上使用 `--minify` 以节省令牌。
- 仅在需要关系数据时传递 `--edges`。
- 在依赖特定语言的提取之前，请检查 [测试能力矩阵和已知限制](../../reference/language-support/)。

## 另请参阅

- [输出模式](../output-modes/)
- [输出格式参考](../../reference/output-format/)
- [分析可靠性](../../reference/analysis-reliability/)
- [图命令](../graph-commands/)