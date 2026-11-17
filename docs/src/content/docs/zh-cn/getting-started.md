---
title: 快速入门
description: 在 5 分钟内开始使用 codeknit。
---

在本教程中，您将从源代码构建 codeknit，运行首次解析，并阅读其生成的 `.skt` 输出。您将解析来自 codeknit 仓库的一个小型包，因此无需拥有自己的项目。

## 前提条件

- **Go 1.26 或更高版本。** 运行 `go version` 进行检查。
- **C 编译器。** codeknit 使用 tree-sitter，需要 CGo。
- **Git** 和 **Make**。

## 1. 构建 codeknit

克隆仓库并构建：

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

您应该看到构建成功完成，且在 `./bin/codeknit` 生成二进制文件。

## 2. 将 codeknit 添加到 PATH

在 `codeknit` 目录中，使用适合您 shell 的命令将 `bin` 添加到 `PATH`。

对于 bash 或 zsh：

```bash
export PATH="$PATH:$(pwd)/bin"
```

对于 fish：

```fish
fish_add_path (pwd)/bin
```

此操作仅适用于当前的 bash 或 zsh 会话。要在后续会话中保持，请将相同命令添加到 `~/.bashrc` 或 `~/.zshrc`，并将 `$(pwd)/bin` 替换为 `bin` 的绝对路径。

## 3. 验证安装

```bash
codeknit --version
```

您应该看到打印出的 codeknit 版本。

## 4. 运行首次解析

保持在 `codeknit` 目录中，解析 `internal/console` 包：

```bash
codeknit parse ./internal/console --edges
```

codeknit 会扫描目录，解析每个受支持的源文件，并生成 `.skt` 文件。`.skt` 文件是一种紧凑的文本文件，列出了代码中的符号及其之间的关系。`--edges` 标志会添加这些关系。默认情况下，为了节省空间，这些关系会被省略。

命令完成后，您应该看到简短的摘要，并在当前目录中看到新的 `./skeleton` 目录。

列出输出：

```bash
ls ./skeleton
```

您应该看到类似 `map_001.skt` 的文件。

## 5. 阅读输出

打开文件：

```bash
cat ./skeleton/map_001.skt
```

您应该首先看到 `[symbols]` 部分。符号是代码中命名的项，例如函数、类型或字段。符号按 `## <源文件>` 标题分组。您的输出包含类似以下的行：

```skt
[symbols]
## <path to>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exported}
...
S10 callable/function L28-L41 New() -> *S2 {exported}
...
S12 callable/method L49-L49 Error(msg: string) {exported, receiver=*Console}
```

每行包含相同的部分：

| 部分 | 示例 | 含义 |
| --- | --- | --- |
| Short ID | `S10` | 符号的简短标签，用于输出中的其他位置 |
| 类别和种类 | `callable/function` | 符号的种类 |
| 行范围 | `L28-L41` | 符号在源文件中覆盖的行 |
| 签名 | `New() -> *S2` | 名称、参数和返回类型 |
| 属性 | `{exported}` | 大括号中的额外详细信息 |

在 `New() -> *S2` 中，`S2` 是 `Console` 类型的 Short ID。当 codeknit 知道符号时，会用 Short ID 替换类型名称。如果您的检出版本较新，行号可能会略有不同。

向下滚动到 `[edges]` 部分。边是两个符号之间的关系。每条边行列出了源符号、关系类型以及目标符号：

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

此示例展示了边行的结构。它表示“`S2` 包含 `S4`”和“`S4` 调用 `S5`”。您可以在 `[symbols]` 中查找每个 Short ID 以了解其含义。

## 您的成果

您构建了 codeknit，将其添加到 `PATH`，运行了 `codeknit parse`，并阅读了 `.skt` 输出中的 `[symbols]` 和 `[edges]` 部分。

## 后续步骤

- 解析您自己的项目，选择不同的输出布局，并安全地重新运行：请参阅[解析命令指南](../guides/parse-command/)。
- 可视化和分析代码结构：请参阅[图命令指南](../guides/graph-commands/)。
- 查找重复和近似重复的代码：请参阅[指纹命令指南](../guides/fingerprint-command/)。
- 查阅 `.skt` 格式的每个部分：请参阅[输出格式参考](../reference/output-format/)。
- 查阅每个命令行标志：请参阅[CLI 标志参考](../reference/cli-flags/)。