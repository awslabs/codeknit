---
title: 安装
description: 如何在您的系统上安装 codeknit。
---

codeknit 可以通过源代码安装。以下步骤将指导您在系统上设置 codeknit。

## 从源代码安装

主要的安装方法是从源代码构建。您需要：

- Go 1.26+
- C 编译器（tree-sitter 通过 CGo 需要）

克隆仓库并构建二进制文件：

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

编译后的二进制文件将位于 `./bin/codeknit`。

## 添加到 PATH

要在任何目录下运行 `codeknit`，请将二进制文件位置添加到系统的 PATH 中。

对于 **bash** (`~/.bashrc`)：

```bash
export PATH="$PATH:/path/to/codeknit"
```

对于 **zsh** (`~/.zshrc`)：

```bash
export PATH="$PATH:/path/to/codeknit"
```

对于 **fish** (`~/.config/fish/config.fish`)：

```fish
fish_add_path /path/to/codeknit
```

更新 shell 配置后，通过运行 `source ~/.bashrc`（或 `~/.zshrc`）或重新启动终端来重新加载配置。

## Shell 自动补全

codeknit 支持主流 shell 的自动补全功能。使用以下命令安装补全：

对于 **bash**：

```bash
codeknit completion bash >> ~/.bashrc
```

对于 **zsh**：

```bash
codeknit completion zsh >> ~/.zshrc
```

对于 **fish**：

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

对于 **PowerShell**：

```powershell
codeknit completion powershell >> $PROFILE
```

## 验证安装

安装完成后，验证 codeknit 是否正确设置：

```bash
codeknit --version
```

## 开发环境设置

如果您要为 codeknit 贡献代码，请运行以下额外命令：

安装开发依赖：

```bash
make deps
```

设置 git 钩子：

```bash
make setup
```

运行测试套件：

```bash
make test
```