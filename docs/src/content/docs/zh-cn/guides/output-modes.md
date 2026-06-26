---
title: 输出模式
description: 为您的项目规模和工作流程选择合适的输出模式。
---

codeknit 支持三种输出模式，由 `--output-mode` 标志控制。每种模式决定了提取的代码结构如何写入磁盘（或 stdout）。

输出模式与输出格式是分开的。默认格式为 `.skt`；传递 `--format json` 以将相同的解析结果以机器可读的 JSON 形式输出。在目录模式下，JSON 会写入 `codeknit.json`。在 `inline` 模式下，JSON 会写入 stdout。

### directory-flat（默认，推荐）

- **行为**：写入分块的 `.skt` 文件，如 `map_001.skt`、`map_002.skt` 等。
- **输出目录**：默认为 `./skeleton/`
- **分割**：当文件超过 `--max-lines` 限制（默认：500 行）时进行分割
- **使用场景**：适用于大多数项目。通过限制文件大小保持输出的组织性和可读性。您可以仅阅读与任务相关的分块。
- **压缩**：启用 `--minify` 时，输出目录中还会生成一个 `dict.skt` 文件，包含压缩值的标记映射。

示例：

```bash
codeknit parse ./src
# 输出：./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **行为**：精确镜像源目录结构。
- **输出目录**：默认为 `./skeleton/`
- **映射**：每个源文件在对应路径下创建一个 `.skt` 文件。
- **使用场景**：当您希望快速查找特定文件的结构时非常理想。适用于与原始代码库一起导航。

示例：

```bash
codeknit parse ./src --output-mode directory-tree
# 输出：./skeleton/src/handler.skt, ./skeleton/pkg/db.skt 等。
```

### inline

- **行为**：将所有输出转储到 stdout。
- **输出目录**：不创建
- **使用场景**：仅推荐用于单个文件或非常小的项目（少于 5 个文件）。适用于将输出传输到其他工具或交互式检查单个文件。

示例：

```bash
codeknit parse ./src/main.go --output-mode inline
# 输出：直接打印到终端
```

### JSON 格式

- **行为**：输出一个包含 `files`、`symbols`、可选 `edges` 和可选 `errors` 的单个 JSON 文档。
- **输出位置**：目录模式下为 `codeknit.json`，`inline` 模式下为 stdout。
- **使用场景**：最适合脚本、编辑器集成、CI 检查以及需要结构化数据的工具。

示例：

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

示例输出：

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

### 决策表

| 模式             | 最适合                                  | 输出位置                                     |
| ---------------- | --------------------------------------- | -------------------------------------------- |
| `directory-flat` | 大多数项目（默认，推荐）                | `./skeleton/map_001.skt`、`map_002.skt` 等   |
| `directory-tree` | 与源代码一起导航输出                    | `./skeleton/<镜像路径>.skt`                  |
| `inline`         | 单个文件，传输到其他工具                | stdout — 仅用于单个文件或极小型项目          |

| 格式 | 最适合                           | 输出                                                   |
| ---- | -------------------------------- | ------------------------------------------------------ |
| `skt`| LLM 上下文和人工检查             | `.skt` 文件或 stdout                                   |
| `json`| 脚本和结构化集成                 | 目录模式下为 `codeknit.json`，`inline` 模式下为 stdout |

### 经验法则

- **不确定时** → 使用 `directory-flat`（默认）
- **检查单个文件** → 可接受 `inline`
- **多于几个文件** → 优先选择 `directory-flat` 或 `directory-tree`
- **大型代码库** → 添加 `--minify` 以减少标记使用
- **在相同输出上重新运行** → 使用 `--clean` 以删除过时的 `.skt` 文件

### 压缩

`--minify` 标志启用基于字典的重复标记压缩（例如属性键如 `exported`、`async` 或常见类型名称）。启用时：

- 重复值被替换为短代码（`d0`、`d1`、`d2` 等）
- 输出目录中会写入一个 `dict.skt` 文件，将代码映射到原始值
- 显著减少大型代码库的输出大小
- 在 `directory-flat` 和 `directory-tree` 模式下均有效

示例压缩输出：

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```

这种格式在最小化标记占用的同时保留了完整信息，非常适合基于 LLM 的分析。