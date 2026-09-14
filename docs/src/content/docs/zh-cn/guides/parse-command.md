---
title: parse 命令
description: 从源代码中提取结构化信息到 .skt 文件或 JSON。
---

`codeknit parse` 命令从代码库中提取结构化信息——如函数、类、方法、变量及其关系——并默认以紧凑的 `.skt` 格式输出。当需要脚本、集成或下游工具使用的机器可读输出时，请使用 JSON。

## 基本用法

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**：要解析的目录或文件路径。
- **`[output-dir]`**：可选的输出目录。如果未提供，默认为 `./skeleton`。

### 示例

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## 输出模式

使用 `--output-mode` 控制输出结构。有三种模式可用：

| 模式               | 描述                                                                                     | 最适用场景                                      |
| ------------------ | ---------------------------------------------------------------------------------------- | ----------------------------------------------- |
| `directory-flat`   | 将分块的 `.skt` 文件（如 `map_001.skt`、`map_002.skt`）写入输出目录。                     | ✅ **大多数项目** —— 默认且推荐模式             |
| `directory-tree`   | 镜像源目录结构，为每个源文件创建一个 `.skt` 文件。                                        | 与源代码一起导航输出                            |
| `inline`           | 将所有输出转储到 stdout。                                                                | 单个文件或传输到其他工具                        |

> **提示**：除非处理单个文件，否则默认使用 `directory-flat`。避免在大型输入时使用 `inline`，因为可能会超出上下文窗口限制。

## 参数

| 参数               | 默认值           | 描述                                                                                     |
| ------------------ | ---------------- | ---------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | 输出模式：`inline`、`directory-flat` 或 `directory-tree`                                 |
| `--format`         | `skt`            | 输出格式：`skt` 或 `json`                                                                |
| `--max-lines`      | `500`            | 在 flat/tree 模式下每个输出文件的最大行数                                                |
| `--collect-test`   | `false`          | 在分析中包含测试文件                                                                     |
| `--minify`         | `false`          | 启用基于字典的压缩以减少 token 使用量                                                    |
| `--edges`          | `false`          | 包含带有关系数据的 `[edges]` 部分（调用、包含等）                                        |
| `--clean`          | `false`          | 在写入前删除输出目录中现有的 `.skt` 文件                                                 |
| `--workers`        | `NumCPU`         | 并发解析 goroutine 的最大数量（0 = 使用所有 CPU 核心）                                   |
| `--verbose`        | `false`          | 在处理过程中打印进度和计时信息                                                           |

## 常见模式

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

示例 JSON 输出：

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

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## 关系不确定性

使用 `--edges` 时，Codeknit 会保留关系，即使语法和作用域无法确定唯一目标：

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

省略状态表示已被分析器解析。`unresolved` 表示无法建立依赖关系；`ambiguous` 表示存在竞争解释。候选项是可能的目标，而非确认的连接。已知端点使用 ShortID；未解析的目标可能以名称形式出现，而无符号。

JSON 以 `resolution` 和 `candidates` 携带相同信息，并省略不确定目标的 `to_short`。C/C++ 包含具有 `meta/file` 和 `meta/include` 符号，因此边使用 ShortID；这记录了书写的指令，而不声称编译器级别的头文件解析。

依赖度量排除不确定的关系。HTML 图显示其计数，而不绘制单个不确定链接。有关语法的详细信息，请参阅[输出格式参考](../../reference/output-format/)，有关解析限制，请参阅[语言支持参考](../../reference/language-support/)。

## 过时输出保护

如果输出目录已包含上次运行生成的 `.skt` 文件，`codeknit` 将拒绝写入新输出，以防止混合过时和新鲜数据。

要覆盖此行为并清理输出目录后再写入，请使用 `--clean` 参数：

```bash
codeknit parse ./src --clean
```

这确保了输出集的新鲜和一致性。

升级到 `0.5.0` 时，请重新生成整个输出集。符号 ID 和关系结果可能会变化；请勿混合不同运行的块。

## 提示

- ✅ **大多数项目默认使用 `directory-flat`**。它在可读性和可管理性之间取得平衡。
- 🔍 在大型代码库上使用 `--minify` 通过共享字典（`dict.skt`）减少 token 使用量。
- 🔗 `[edges]` 部分**默认被排除**以节省 token。当需要关系数据（如 `calls`、`contains` 或 `inherits`）时，请使用 `--edges`。
- 🧾 当脚本或集成需要结构化数据而非 `.skt` 时，请使用 `--format json`。
- 🧹 在同一输出目录上重新运行时，始终使用 `--clean`。
- 📁 如果希望在编辑器中直接将 `.skt` 文件与源文件关联，请使用 `directory-tree`。