---
title: はじめに
description: 5分以内にcodeknitを使い始めましょう。
---

# はじめに

5分以内にcodeknitを使い始めましょう。

## 1. 前提条件

以下が必要です：

- Go 1.26+
- Cコンパイラ（tree-sitter用にCGoが必要）

## 2. ソースからのインストール

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# バイナリは ./bin/codeknit にあります
```

## 3. PATHへの追加

バイナリをシェルのPATHに追加します：

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

変更を反映させるためにシェルを再読み込みするか、`source ~/.bashrc`（または`~/.zshrc`）を実行してください。

## 4. インストールの確認

codeknitが動作しているか確認します：

```bash
codeknit --version
```

## 5. 最初のパース

コードベースで最初のパースを実行します：

```bash
codeknit parse ./myproject
```

このコマンドは：

- `./myproject`内のすべてのソースファイルをパースします
- 構造情報（関数、クラス、関係性）を抽出します
- チャンク化された`.skt`ファイルを`./skeleton/`（デフォルトの出力ディレクトリ）に書き込みます

このコマンドを再実行する場合は、`--clean`を使用して以前の出力を削除します：

```bash
codeknit parse ./myproject --clean
```

## 6. 出力の読み方

`.skt`ファイルには構造化されたコード情報が含まれています。以下は小さな例です：

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

主なセクション：

- `[symbols]`：ファイルごとにグループ化された定義。名前、行範囲、メタデータを表示
- `[edges]`：`contains`、`calls`、`inherits`、`returns`などの関係性

## 7. 次のステップ

最初のパースを実行したので、次は：

- パースコマンドについて詳しく学ぶ：[パースコマンドガイド](/codeknit/ja/guides/parse-command/)
- 構造分析を探索する：[グラフコマンドガイド](/codeknit/ja/guides/graph-commands/)
- 重複検出を理解する：[フィンガープリントコマンドガイド](/codeknit/ja/guides/fingerprint-command/)
- 出力フォーマットの詳細を読む：[出力フォーマットリファレンス](/codeknit/ja/reference/output-format/)
- 利用可能なすべてのフラグを見る：[CLIフラグリファレンス](/codeknit/ja/reference/cli-flags/)