---
title: 快速入门
description: 在5分钟内开始使用 codeknit。
---

# 快速入门

在5分钟内开始使用 codeknit。

## 1. 前提条件

你需要：

- Go 1.26+
- C 编译器（tree-sitter 需要 CGo）

## 2. 从源代码安装

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# 二进制文件位于 ./bin/codeknit
```

## 3. 添加到 PATH

将二进制文件添加到 shell 的 PATH：

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

重新加载 shell 或运行 `source ~/.bashrc`（或 `~/.zshrc`）使更改生效。

## 4. 验证安装

检查 codeknit 是否正常工作：

```bash
codeknit --version
```

## 5. 首次解析

在代码库上运行首次解析：

```bash
codeknit parse ./myproject
```

此命令：

- 解析 `./myproject` 中的所有源文件
- 提取结构信息（函数、类、关系）
- 将分块的 `.skt` 文件写入 `./skeleton/`（默认输出目录）

如果重新运行此命令，请使用 `--clean` 删除之前的输出：

```bash
codeknit parse ./myproject --clean
```

## 6. 阅读输出

`.skt` 文件包含结构化的代码信息。以下是一个小示例：

```skt
[symbols]
## src/main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {exported}
S3 callable/function L10-L12 NewServer(addr: string) -> *S2 {exported}
S4 callable/method L14-L19 Start() {receiver=*Server}

[edges]
S2 --contains--> S4
S3 --returns--> S2
```

关键部分：

- `[symbols]`：按文件分组的定义，显示名称、行范围和元数据
- `[edges]`：关系，如 `contains`、`calls`、`inherits` 或 `returns`

## 7. 下一步

现在你已经运行了首次解析：

- 了解更多关于解析命令的信息：[解析命令指南](/codeknit/zh-cn/guides/parse-command/)
- 探索结构分析：[图命令指南](/codeknit/zh-cn/guides/graph-commands/)
- 了解重复检测：[指纹命令指南](/codeknit/zh-cn/guides/fingerprint-command/)
- 阅读完整的输出格式：[输出格式参考](/codeknit/zh-cn/reference/output-format/)
- 查看所有可用标志：[CLI 标志参考](/codeknit/zh-cn/reference/cli-flags/)