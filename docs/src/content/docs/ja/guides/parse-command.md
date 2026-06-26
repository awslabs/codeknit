---
title: parse コマンド
description: ソースコードから構造情報を抽出し、.skt ファイルまたは JSON 形式で出力します。
---

`codeknit parse` コマンドは、コードベースから関数、クラス、メソッド、変数、およびそれらの関係などのコード構造情報を抽出し、デフォルトではコンパクトな `.skt` 形式で出力します。スクリプト、統合、またはダウンストリームツールで機械可読な出力が必要な場合は、JSON を使用してください。

## 基本的な使用方法

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: パースするディレクトリまたはファイルへのパス。
- **`[output-dir]`**: オプションの出力ディレクトリ。指定しない場合、デフォルトは `./skeleton` です。

### 例

```bash
# プロジェクトをパースし、デフォルトのディレクトリ ./skeleton に出力
codeknit parse ./src

# パースしてカスタム出力ディレクトリに書き込む
codeknit parse ./src ./output

# 単一ファイルをパースして stdout に出力
codeknit parse ./src/main.go --output-mode inline

# 機械可読な JSON を stdout に出力
codeknit parse ./src --output-mode inline --format json
```

## 出力モード

`--output-mode` を使用して出力の構造を制御します。3つのモードが利用可能です:

| モード             | 説明                                                                              | 最適な用途                                            |
| ---------------- | ---------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `directory-flat` | チャンク化された `.skt` ファイル（例: `map_001.skt`、`map_002.skt`）を出力ディレクトリに書き込みます。 | ✅ **ほとんどのプロジェクト** — デフォルトかつ推奨モード |
| `directory-tree` | ソースディレクトリ構造をミラーリングし、ソースファイルごとに1つの `.skt` ファイルを作成します。        | ソースコードと並行して出力をナビゲートする場合             |
| `inline`         | すべての出力を stdout にダンプします。                                                              | 単一ファイルまたは他のツールへのパイプ処理               |

> **ヒント**: 単一ファイルで作業しない限り、`directory-flat` をデフォルトにしてください。`inline` は大規模な入力に対してコンテキストウィンドウを圧迫する可能性があります。

## フラグ

| フラグ             | デフォルト          | 説明                                                                  |
| ---------------- | ---------------- | ---------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | 出力モード: `inline`、`directory-flat`、または `directory-tree`                 |
| `--format`       | `skt`            | 出力形式: `skt` または `json`                                               |
| `--max-lines`    | `500`            | フラット/ツリーモードでの出力ファイルごとの最大行数                             |
| `--collect-test` | `false`          | テストファイルを分析に含める                                               |
| `--minify`       | `false`          | トークン使用量を削減するための辞書ベースの圧縮を有効にする                    |
| `--edges`        | `false`          | 関係データ（呼び出し、含むなど）を含む `[edges]` セクションを含める |
| `--clean`        | `false`          | 書き込み前に出力ディレクトリ内の既存の `.skt` ファイルを削除する          |
| `--workers`      | `NumCPU`         | 同時パースゴルーチンの最大数（0 = すべてのCPUコアを使用）      |
| `--verbose`      | `false`          | 処理中の進捗とタイミング情報を出力する                      |

## 一般的なパターン

```bash
# プロジェクトで初回実行
codeknit parse ./src
```

```bash
# 前回の出力をクリーンして再実行
codeknit parse ./src --clean
```

```bash
# 単一ファイルをパースして stdout に出力
codeknit parse ./src/main.go --output-mode inline
```

```bash
# 大規模なコードベースの出力を圧縮
codeknit parse ./src --minify
```

```bash
# 依存関係分析用の関係エッジを含める
codeknit parse ./src --edges
```

```bash
# 他のツール用に JSON を出力
codeknit parse ./src --output-mode inline --format json --edges
```

JSON 出力の例:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

```bash
# 出力でソースツリー構造をミラーリング
codeknit parse ./src --output-mode directory-tree
```

## 古い出力の保護

出力ディレクトリに以前の実行からの `.skt` ファイルが既に存在する場合、`codeknit` は古いデータと新しいデータが混在するのを防ぐため、新しい出力の書き込みを拒否します。

この動作を上書きして、書き込み前に出力ディレクトリをクリーンにするには、`--clean` フラグを使用します:

```bash
codeknit parse ./src --clean
```

これにより、新鮮で一貫性のある出力セットが確保されます。

## ヒント

- ✅ **ほとんどのプロジェクトでは `directory-flat` をデフォルトにする** — 読みやすさと管理のバランスが取れています。
- 🔍 大規模なコードベースでは `--minify` を使用して、共有辞書 (`dict.skt`) によるトークン使用量を削減します。
- 🔗 `[edges]` セクションはトークンを節約するために**デフォルトでは除外されています**。`calls`、`contains`、`inherits` などの関係データが必要な場合は `--edges` を使用してください。
- 🧾 スクリプトや統合が `.skt` の代わりに構造化データを必要とする場合は `--format json` を使用します。
- 🧹 同じ出力ディレクトリで再実行する場合は、常に `--clean` を使用してください。
- 📁 エディタで `.skt` ファイルをソースファイルと直接関連付けたい場合は `directory-tree` を使用します。