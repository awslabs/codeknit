---
title: Parse 命令
description: 从源代码中提取结构化信息到 .skt 文件或 JSON。
---

`codeknit parse` 命令从代码库中提取结构化信息——例如函数、类、方法、变量及其关系——并默认以紧凑的 `.skt` 格式输出。当需要脚本、集成或下游工具使用的机器可读输出时，请使用 JSON。

## 基本用法

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**：要解析的目录或文件路径。
- **`[output-dir]`**：可选的输出目录。如果未提供，默认为 `./skeleton`。

### 示例

```bash
# 解析项目，输出到默认目录 ./skeleton
codeknit parse ./src

# 解析并写入到自定义输出目录
codeknit parse ./src ./output

# 解析单个文件并输出到 stdout
codeknit parse ./src/main.go --output-mode inline

# 输出机器可读的 JSON 到 stdout
codeknit parse ./src --output-mode inline --format json
```

## 输出模式

使用 `--output-mode` 控制输出的结构方式。有三种模式可用：

| 模式             | 描述                                                                                     | 适用场景                                            |
| ---------------- | ---------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `directory-flat` | 将分块的 `.skt` 文件（例如 `map_001.skt`、`map_002.skt`）写入输出目录。                  | ✅ **大多数项目**——默认且推荐的模式                 |
| `directory-tree` | 镜像源目录结构，为每个源文件创建一个 `.skt` 文件。                                       | 在源代码旁导航输出                                  |
| `inline`         | 将所有输出转储到 stdout。                                                                | 单个文件或传输到其他工具                            |

> **提示**：除非处理单个文件，否则默认使用 `directory-flat`。避免在大型输入时使用 `inline`，因为可能会超出上下文窗口限制。

## 标志

| 标志             | 默认值          | 描述                                                                                     |
| ---------------- | ---------------- | ---------------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | 输出模式：`inline`、`directory-flat` 或 `directory-tree`                                 |
| `--format`       | `skt`            | 输出格式：`skt` 或 `json`                                                                |
| `--max-lines`    | `500`            | 在 flat/tree 模式下每个输出文件的最大行数                                               |
| `--collect-test` | `false`          | 在分析中包含测试文件                                                                     |
| `--minify`       | `false`          | 启用基于字典的压缩以减少 token 使用量                                                   |
| `--edges`        | `false`          | 包含带有关系数据的 `[edges]` 部分（调用、包含等）                                       |
| `--clean`        | `false`          | 写入前删除输出目录中现有的 `.skt` 文件                                                   |
| `--workers`      | `NumCPU`         | 并发解析 goroutine 的最大数量（0 = 使用所有 CPU 核心）                                  |
| `--verbose`      | `false`          | 处理过程中打印进度和计时信息                                                             |

## 常见模式

```bash
# 首次在项目上运行
codeknit parse ./src
```

```bash
# 重新运行并清理之前的输出
codeknit parse ./src --clean
```

```bash
# 解析单个文件到 stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# 压缩大型代码库的输出
codeknit parse ./src --minify
```

```bash
# 包含关系边（例如用于依赖分析）
codeknit parse ./src --edges
```

```bash
# 为其他工具输出 JSON
codeknit parse ./src --output-mode inline --format json --edges
```

JSON 输出示例：

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
# 在输出中镜像源树结构
codeknit parse ./src --output-mode directory-tree
```

## 过期输出保护

如果输出目录已包含上次运行生成的 `.skt` 文件，`codeknit` 将拒绝写入新输出，以防止混合过期和新鲜数据。

要覆盖此行为并清理输出目录后再写入，请使用 `--clean` 标志：

```bash
codeknit parse ./src --clean
```

这确保了输出集的新鲜和一致。

## 提示

- ✅ **大多数项目默认使用 `directory-flat`**。它在可读性和可管理性之间取得平衡。
- 🔍 在大型代码库上使用 `--minify` 通过共享字典（`dict.skt`）减少 token 使用量。
- 🔗 `[edges]` 部分**默认被排除**以节省 token。当需要关系数据（如 `calls`、`contains` 或 `inherits`）时，请使用 `--edges`。
- 🧾 当脚本或集成需要结构化数据而非 `.skt` 时，请使用 `--format json`。
- 🧹 在同一输出目录上重新运行时，始终使用 `--clean`。
- 📁 如果希望在编辑器中直接将 `.skt` 文件与源文件关联，请使用 `directory-tree`。