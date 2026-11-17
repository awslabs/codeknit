---
title: 安装
description: 如何在您的系统上安装 codeknit。
---

本指南将从源代码安装 codeknit，将其添加到 `PATH`，启用 shell 补全，并检查其运行情况。

:::tip[使用 AI 编码助手？]
Claude Code、Codex 和 Kiro 用户可以改为安装 codeknit 插件。该插件会在首次运行时下载二进制文件。请参阅[在 AI 助手中使用 codeknit](../guides/ai-assistants/)。
:::

## 从源代码安装

**前提条件**

- Go 1.26 或更高版本
- C 编译器。codeknit 使用 CGo 构建其 tree-sitter 解析器。

**步骤**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

二进制文件将被写入 `./bin/codeknit`。

### 改用预构建二进制文件

Linux、macOS 和 Windows（amd64 和 arm64）的预构建二进制文件位于[发布页面](https://github.com/awslabs/codeknit/releases)。下载适用于您平台的压缩包，解压后将二进制文件移动到 `PATH` 上的目录。每个发布版本都包含一个 `checksums.txt` 文件，可用于验证下载。

## 将 codeknit 添加到 PATH

如果您已将二进制文件移动到已在 `PATH` 上的目录（如 `/usr/local/bin`），可跳过本节。

将 `/path/to/codeknit` 替换为包含二进制文件的目录。如果您是从源代码构建，则该目录是克隆仓库内的 `bin` 目录。

将适用于您 shell 的行添加到其配置文件：

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

重新加载 shell，或导入您编辑的文件：

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

对于 fish，请打开新终端。

## 启用 shell 补全

codeknit 可为 bash、zsh、fish 和 PowerShell 生成补全脚本。为您的 shell 运行一次以下命令：

```bash
# bash
codeknit completion bash >> ~/.bashrc

# zsh
codeknit completion zsh >> ~/.zshrc

# fish
codeknit completion fish > ~/.config/fish/completions/codeknit.fish

# PowerShell
codeknit completion powershell >> $PROFILE
```

bash、zsh 和 PowerShell 命令会附加到您的配置文件。再次运行会导致脚本被添加两次，因此每个命令仅需运行一次。

打开新 shell，或导入您的配置文件，以激活补全。

## 验证安装

```bash
codeknit --version
```

如果命令打印版本号，则表示 codeknit 已安装。如果您的 shell 报告 `command not found`，请检查 `PATH` 行中的目录是否包含 `codeknit` 二进制文件，然后重新加载 shell。

## 设置开发环境

仅在您计划为 codeknit 做出贡献时执行此步骤。从克隆仓库的根目录运行以下命令：

```bash
make deps    # 安装依赖
make setup   # 配置 git 钩子
make test    # 运行测试套件
```

`make lint` 会运行 lint 检查和格式化工具。

## 后续步骤

- 在[入门](../getting-started/)中运行您的第一个命令。
- 浏览[CLI 标志参考](../reference/cli-flags/)中的所有标志。