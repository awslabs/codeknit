---
title: インストール
description: システムに codeknit をインストールする方法。
---

codeknit はソースからインストールできます。以下の手順に従って、システムに codeknit をセットアップしてください。

## ソースから

主なインストール方法はソースからのビルドです。以下が必要です:

- Go 1.26+
- C コンパイラ（tree-sitter 用の CGo に必要）

リポジトリをクローンしてバイナリをビルドします:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

コンパイルされたバイナリは `./bin/codeknit` にあります。

## PATH に追加

任意のディレクトリから `codeknit` を実行するには、バイナリの場所をシステムの PATH に追加します。

**bash** 用 (`~/.bashrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

**zsh** 用 (`~/.zshrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

**fish** 用 (`~/.config/fish/config.fish`):

```fish
fish_add_path /path/to/codeknit
```

シェル設定を更新した後、 `source ~/.bashrc` （または `~/.zshrc`）を実行するか、ターミナルを再起動して設定を反映させます。

## シェル補完

codeknit は人気のシェルに対応した自動補完をサポートしています。以下のコマンドを使用して補完をインストールします:

**bash** 用:

```bash
codeknit completion bash >> ~/.bashrc
```

**zsh** 用:

```bash
codeknit completion zsh >> ~/.zshrc
```

**fish** 用:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

**PowerShell** 用:

```powershell
codeknit completion powershell >> $PROFILE
```

## インストールの確認

インストール後、codeknit が正しくセットアップされていることを確認します:

```bash
codeknit --version
```

## 開発環境のセットアップ

codeknit にコントリビュートする場合は、以下の追加コマンドを実行します:

開発依存関係をインストール:

```bash
make deps
```

Git フックをセットアップ:

```bash
make setup
```

テストスイートを実行:

```bash
make test
```
