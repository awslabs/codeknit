---
title: インストール
description: システムに codeknit をインストールする方法。
---

codeknit はソースからインストールできます。以下の手順でシステムに codeknit をセットアップする方法を説明します。

## ソースから

主なインストール方法はソースからのビルドです。以下が必要です：

- Go 1.26 以上
- C コンパイラ（tree-sitter 用の CGo が必要）

リポジトリをクローンしてバイナリをビルドします：

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

コンパイルされたバイナリは `./bin/codeknit` に作成されます。

## PATH に追加

`codeknit` を任意のディレクトリから実行できるようにするには、バイナリの場所をシステムの PATH に追加します。

**bash** の場合 (`~/.bashrc`)：

```bash
export PATH="$PATH:/path/to/codeknit"
```

**zsh** の場合 (`~/.zshrc`)：

```bash
export PATH="$PATH:/path/to/codeknit"
```

**fish** の場合 (`~/.config/fish/config.fish`)：

```fish
fish_add_path /path/to/codeknit
```

シェル設定を更新した後、 `source ~/.bashrc` （または `~/.zshrc`）を実行するか、ターミナルを再起動して設定を反映します。

## シェル補完

codeknit は人気のシェルに対応した自動補完をサポートしています。以下のコマンドで補完をインストールします：

**bash** の場合：

```bash
codeknit completion bash >> ~/.bashrc
```

**zsh** の場合：

```bash
codeknit completion zsh >> ~/.zshrc
```

**fish** の場合：

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

**PowerShell** の場合：

```powershell
codeknit completion powershell >> $PROFILE
```

## インストールの確認

インストール後、codeknit が正しくセットアップされていることを確認します：

```bash
codeknit --version
```

## 開発環境のセットアップ

codeknit にコントリビュートする場合は、以下の追加コマンドを実行します：

開発依存関係をインストール：

```bash
make deps
```

git フックをセットアップ：

```bash
make setup
```

テストスイートを実行：

```bash
make test
```