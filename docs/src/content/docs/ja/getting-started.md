---
title: はじめに
description: 5分以内にcodeknitを使い始めましょう。
---

このチュートリアルでは、codeknitをソースからビルドし、最初のパースを実行して、生成される`.skt`出力を読み取ります。codeknitリポジトリ内の小さなパッケージをパースするため、独自のプロジェクトは必要ありません。

## 前提条件

- **Go 1.26以降。** `go version`を実行して確認してください。
- **Cコンパイラ。** codeknitはtree-sitterを使用しており、CGoが必要です。
- **Git**と**Make**。

## 1. codeknitのビルド

リポジトリをクローンしてビルドします：

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

ビルドがエラーなく完了し、`./bin/codeknit`にバイナリが生成されるはずです。

## 2. codeknitをPATHに追加

`codeknit`ディレクトリから、`bin`を`PATH`に追加します。使用しているシェルに応じた行を使用してください。

bashまたはzshの場合：

```bash
export PATH="$PATH:$(pwd)/bin"
```

fishの場合：

```fish
fish_add_path (pwd)/bin
```

これは現在のbashまたはzshセッションに適用されます。後続のセッションでも維持するには、`~/.bashrc`または`~/.zshrc`に同じ行を追加し、`$(pwd)/bin`の代わりに`bin`への絶対パスを使用してください。

## 3. インストールの確認

```bash
codeknit --version
```

codeknitのバージョンが表示されるはずです。

## 4. 最初のパースを実行

`codeknit`ディレクトリに留まり、`internal/console`パッケージをパースします：

```bash
codeknit parse ./internal/console --edges
```

codeknitはディレクトリをスキャンし、サポートされている各ソースファイルをパースして、`.skt`ファイルを生成します。*`.skt`ファイル*は、コード内のシンボルとそれらの間の関係をリストしたコンパクトなテキストファイルです。`--edges`フラグは関係を追加します。デフォルトではスペースを節約するために省略されています。

コマンドが完了すると短いサマリーが表示され、カレントディレクトリに新しい`./skeleton`ディレクトリが作成されるはずです。

出力をリストします：

```bash
ls ./skeleton
```

`map_001.skt`のようなファイルが表示されるはずです。

## 5. 出力を読む

ファイルを開きます：

```bash
cat ./skeleton/map_001.skt
```

最初に`[symbols]`セクションが表示されるはずです。*シンボル*は、関数、型、フィールドなど、コード内の名前付きアイテムです。シンボルは`## <ソースファイル>`見出しの下にグループ化されています。以下のような行が出力に含まれているはずです：

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

各行は同じ部分で構成されています：

| 部分 | 例 | 意味 |
| --- | --- | --- |
| Short ID | `S10` | 出力内の他の場所で使用されるシンボルの短いラベル |
| カテゴリと種類 | `callable/function` | シンボルの種類 |
| 行範囲 | `L28-L41` | ソースファイル内でシンボルがカバーする行 |
| シグネチャ | `New() -> *S2` | 名前、パラメータ、戻り値の型 |
| プロパティ | `{exported}` | 中括弧内の追加の詳細 |

`New() -> *S2`では、`S2`は`Console`型のShort IDです。codeknitはシンボルを認識すると、型名をShort IDに置き換えます。チェックアウトが新しい場合、行番号が若干異なることがあります。

`[edges]`セクションまでスクロールします。*エッジ*は2つのシンボル間の関係です。各エッジ行は、ソースシンボル、関係の種類、およびターゲットシンボルまたはシンボルを示します：

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

この例はエッジ行の形式を示しています。これは「`S2`は`S4`を含む」および「`S4`は`S5`を呼び出す」という意味です。`[symbols]`で各Short IDを調べて、それが何を指しているかを確認できます。

## 達成したこと

codeknitをビルドし、`PATH`に追加し、`codeknit parse`を実行して、`.skt`出力の`[symbols]`および`[edges]`セクションを読み取りました。

## 次のステップ

- 自分のプロジェクトをパースし、異なる出力モードを選択して安全に再実行する方法：[parseコマンドガイド](../guides/parse-command/)を参照してください。
- コード構造を可視化および分析する方法：[グラフコマンドガイド](../guides/graph-commands/)を参照してください。
- 重複および近似重複コードを見つける方法：[fingerprintコマンドガイド](../guides/fingerprint-command/)を参照してください。
- `.skt`形式のすべての部分を確認する方法：[出力形式リファレンス](../reference/output-format/)を参照してください。
- すべてのコマンドラインフラグを確認する方法：[CLIフラグリファレンス](../reference/cli-flags/)を参照してください。