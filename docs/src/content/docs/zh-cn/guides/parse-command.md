---
title: parse 命令
description: 从源代码中提取结构化信息并输出为 .skt 文件。
---

`codeknit parse` 命令从代码库中提取结构化信息——如函数、类、方法、变量及其关系——并以紧凑的 `.skt` 格式输出，专为高效供 LLM 和分析工具使用而设计。

## 基本用法

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**：要解析的目录或文件路径。
- **`[output-dir]`**：可选的输出目录。若未提供，默认为 `./skeleton`。

### 示例

```bash
# 解析项目，输出到默认目录 ./skeleton
codeknit parse ./src

# 解析并输出到自定义目录
codeknit parse ./src ./output

# 解析单个文件并输出到 stdout
codeknit parse ./src/main.go --output-mode inline
```

## 输出模式

使用 `--output-mode` 控制输出结构。提供三种模式：

| 模式             | 描述                                                                  | 适用场景                          |
| ---------------- | --------------------------------------------------------------------- | --------------------------------- |
| `directory-flat` | 将分块的 `.skt` 文件（如 `map_001.skt`、`map_002.skt`）写入输出目录。 | ✅ **大多数项目**——默认且推荐模式 |
| `directory-tree` | 镜像源目录结构，为每个源文件创建一个 `.skt` 文件。                    | 与源代码一同浏览输出              |
| `inline`         | 将所有输出转储到 stdout。                                             | 单个文件或管道传输至其他工具      |

> **提示**：除非处理单个文件，否则默认使用 `directory-flat`。避免在大型输入时使用 `inline`，以免超出上下文窗口限制。

## 参数

| 参数             | 默认值           | 描述                                                     |
| ---------------- | ---------------- | -------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | 输出模式：`inline`、`directory-flat` 或 `directory-tree` |
| `--max-lines`    | `500`            | 平铺/树状模式下每个输出文件的最大行数                    |
| `--collect-test` | `false`          | 在分析中包含测试文件                                     |
| `--minify`       | `false`          | 启用基于字典的压缩以减少 token 使用量                    |
| `--edges`        | `false`          | 包含 `[edges]` 部分，记录关系数据（调用、包含等）        |
| `--clean`        | `false`          | 写入前清除输出目录中现有的 `.skt` 文件                   |
| `--workers`      | `NumCPU`         | 并发解析 goroutine 的最大数量（0 = 使用所有 CPU 核心）   |
| `--verbose`      | `false`          | 处理过程中打印进度和计时信息                             |

## 常见用法

```bash
# 首次运行项目
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
# 包含关系边（如用于依赖分析）
codeknit parse ./src --edges
```

```bash
# 在输出中镜像源树结构
codeknit parse ./src --output-mode directory-tree
```

## 过期输出保护

如果输出目录已包含之前运行生成的 `.skt` 文件，`codeknit` 将拒绝写入新输出，以防止混合过期和新鲜数据。

要覆盖此行为并在写入前清理输出目录，请使用 `--clean` 参数：

```bash
codeknit parse ./src --clean
```

这将确保输出集的新鲜和一致性。

## 提示

- ✅ **大多数项目默认使用 `directory-flat`**。它在可读性和可管理性之间取得平衡。
- 🔍 在大型代码库上使用 `--minify`，通过共享字典（`dict.skt`）减少 token 使用量。
- 🔗 `[edges]` 部分**默认被排除**以节省 token。当需要关系数据（如 `calls`、`contains` 或 `inherits`）时，请使用 `--edges`。
- 🧹 重新运行同一输出目录时，始终使用 `--clean`。
- 📁 如果希望在编辑器中直接关联 `.skt` 文件与源文件，请使用 `directory-tree`。
