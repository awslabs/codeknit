---
title: 输出模式
description: 根据项目规模和工作流选择合适的输出模式。
---

codeknit 支持三种输出模式，通过 `--output-mode` 标志控制。每种模式决定了提取的代码结构如何写入磁盘（或标准输出）。

### directory-flat（默认，推荐）

- **行为**：将分块的 `.skt` 文件写入，如 `map_001.skt`、`map_002.skt` 等。
- **输出目录**：默认为 `./skeleton/`
- **分割**：当文件超过 `--max-lines` 限制（默认：500 行）时进行分割。
- **使用场景**：适用于大多数项目。通过限制文件大小保持输出的组织性和可读性。您可以仅阅读与任务相关的分块。
- **压缩**：启用 `--minify` 时，输出目录中还会生成一个 `dict.skt` 文件，包含用于压缩值的标记映射。

示例：

```bash
codeknit parse ./src
# 输出：./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **行为**：精确镜像源目录结构。
- **输出目录**：默认为 `./skeleton/`
- **映射**：每个源文件在对应路径下生成一个 `.skt` 文件。
- **使用场景**：在希望快速查找特定文件结构时非常理想。适用于与原始代码库一起导航。

示例：

```bash
codeknit parse ./src --output-mode directory-tree
# 输出：./skeleton/src/handler.skt, ./skeleton/pkg/db.skt 等
```

### inline

- **行为**：将所有输出转储到标准输出。
- **输出目录**：不创建
- **使用场景**：仅建议用于单个文件或非常小的项目（少于 5 个文件）。适用于将输出传输到其他工具或交互式检查单个文件。

示例：

```bash
codeknit parse ./src/main.go --output-mode inline
# 输出：直接打印到终端
```

### 决策表

| 模式             | 最适用于                 | 输出位置                                   |
| ---------------- | ------------------------ | ------------------------------------------ |
| `directory-flat` | 大多数项目（默认，推荐） | `./skeleton/map_001.skt`、`map_002.skt` 等 |
| `directory-tree` | 与源代码一起导航输出     | `./skeleton/<镜像路径>.skt`                |
| `inline`         | 单个文件，传输到其他工具 | 标准输出 — 仅适用于单个文件或极小型项目    |

### 经验法则

- **不确定时** → 使用 `directory-flat`（默认）
- **检查单个文件** → 可接受使用 `inline`
- **超过几个文件** → 优先选择 `directory-flat` 或 `directory-tree`
- **大型代码库** → 添加 `--minify` 以减少标记使用量
- **在相同输出上重新运行** → 使用 `--clean` 以删除过时的 `.skt` 文件

### 压缩

`--minify` 标志启用基于字典的重复标记压缩（例如属性键如 `exported`、`async` 或常见类型名称）。启用时：

- 重复值被替换为短代码（`d0`、`d1`、`d2` 等）
- 输出目录中会写入一个 `dict.skt` 文件，将代码映射到原始值
- 显著减少大型代码库的输出大小
- 适用于 `directory-flat` 和 `directory-tree` 两种模式

示例压缩输出：

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```

此格式在最小化标记占用的同时保留完整信息，非常适合基于 LLM 的分析。
