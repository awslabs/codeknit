---
title: インストール
description: codeknitをシステムにインストールする方法。
---

このガイドでは、codeknitをソースからインストールし、`PATH`に追加し、シェル補完を有効にして、正常に動作するか確認します。

:::tip[AIコーディングアシスタントを使用していますか？]
Claude Code、Codex、Kiroのユーザーは、代わりにcodeknitプラグインをインストールできます。プラグインは初回実行時にバイナリをダウンロードします。詳細は[AIアシスタントとcodeknitを使用する](../guides/ai-assistants/)を参照してください。
:::

## ソースからインストール

**前提条件**

- Go 1.26以降
- Cコンパイラ。codeknitはtree-sitterパーサーをビルドするためにCGoを使用します。

**手順**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

バイナリは`./bin/codeknit`に書き込まれます。

### 代わりにプレビルドバイナリを使用する

Linux、macOS、Windows（amd64およびarm64）用のプレビルドバイナリは[リリースページ](https://github.com/awslabs/codeknit/releases)にあります。プラットフォーム用のアーカイブをダウンロードし、展開して、バイナリを`PATH`上のディレクトリに移動します。各リリースには、`checksums.txt`ファイルが含まれており、ダウンロードを検証できます。

## codeknitをPATHに追加する

`PATH`上のディレクトリ（例：`/usr/local/bin`）にバイナリを移動した場合は、このセクションをスキップしてください。

`/path/to/codeknit`をバイナリが含まれるディレクトリに置き換えてください。ソースからビルドした場合は、クローンしたリポジトリ内の`bin`ディレクトリです。

シェルの設定ファイルに、該当するシェルの行を追加します：

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

シェルを再読み込みするか、編集したファイルをソースします：

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

fishの場合は、新しいターミナルを開きます。

## シェル補完を有効にする

codeknitはbash、zsh、fish、PowerShell用の補完スクリプトを生成します。該当するシェルのコマンドを一度実行します：

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

bash、zsh、PowerShellのコマンドはプロファイルに追記します。再度実行するとスクリプトが二重に追加されるため、各コマンドは一度だけ実行してください。

新しいシェルを開くか、プロファイルをソースして補完を有効にします。

## インストールの確認

```bash
codeknit --version
```

コマンドがバージョンを表示すれば、codeknitはインストールされています。シェルが`command not found`と報告する場合は、`PATH`行のディレクトリに`codeknit`バイナリが含まれているか確認し、シェルを再読み込みしてください。

## 開発環境のセットアップ

これは、codeknitに貢献する予定がある場合にのみ行ってください。クローンしたリポジトリのルートからこれらのコマンドを実行します：

```bash
make deps    # 依存関係をインストール
make setup   # gitフックを設定
make test    # テストスイートを実行
```

`make lint`はリンターとフォーマッターを実行します。

## 次のステップ

- [はじめに](../getting-started/)で最初のコマンドを実行。
- [CLIフラグリファレンス](../reference/cli-flags/)で全てのフラグを確認。