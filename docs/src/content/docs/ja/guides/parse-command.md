---
title: parse コマンド
description: ソースコードから構造情報を抽出し、.skt ファイルまたは JSON として出力します。
---

`codeknit parse` コマンドは、コードベースから関数、クラス、メソッド、変数、およびそれらの関係といった**コード構造**を抽出し、デフォルトではコンパクトな `.skt` 形式で出力します。スクリプト、インテグレーション、またはダウンストリームツールで機械可読な出力が必要な場合は、JSON を使用してください。

## 基本的な使用方法

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: パースするディレクトリまたはファイルへのパス。
- **`[output-dir]`**: オプションの出力ディレクトリ。指定しない場合、デフォルトは `./skeleton` です。

### 例

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## 出力モード

`--output-mode` を使用して、出力の構造を制御します。3つのモードが利用可能です:

| モード           | 説明                                                                                                   | 最適な用途                                               |
| ---------------- | ------------------------------------------------------------------------------------------------------ | -------------------------------------------------------- |
| `directory-flat` | チャンク化された `.skt` ファイル（例: `map_001.skt`、`map_002.skt`）を出力ディレクトリに書き込みます。 | ✅ **ほとんどのプロジェクト** — デフォルトかつ推奨モード |
| `directory-tree` | ソースディレクトリ構造をミラーリングし、ソースファイルごとに1つの `.skt` ファイルを作成します。        | ソースコードと並行して出力をナビゲートする場合           |
| `inline`         | すべての出力を stdout にダンプします。                                                                 | 単一ファイルまたは他のツールへのパイプ処理               |

> **ヒント**: 単一ファイルを扱う場合を除き、`directory-flat` をデフォルトとして使用してください。`inline` は大量の入力に対してコンテキストウィンドウを圧迫する可能性があります。

## フラグ

| フラグ           | デフォルト       | 説明                                                                |
| ---------------- | ---------------- | ------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | 出力モード: `inline`、`directory-flat`、または `directory-tree`     |
| `--format`       | `skt`            | 出力形式: `skt` または `json`                                       |
| `--max-lines`    | `500`            | フラット/ツリーモードでの出力ファイルごとの最大行数                 |
| `--collect-test` | `false`          | 解析にテストファイルを含める                                        |
| `--minify`       | `false`          | トークン使用量を削減するための辞書ベースの圧縮を有効化              |
| `--edges`        | `false`          | 関係データ（呼び出し、包含など）を含む `[edges]` セクションを含める |
| `--clean`        | `false`          | 書き込み前に出力ディレクトリ内の既存の `.skt` ファイルを削除        |
| `--workers`      | `NumCPU`         | 並列パースゴルーチンの最大数（0 = すべての CPU コアを使用）         |
| `--verbose`      | `false`          | 処理中の進捗とタイミング情報を表示                                  |

## 一般的なパターン

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
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
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## 関係の不確実性

`--edges` を使用すると、Codeknit は構文やスコープが一意のターゲットを確定できない場合でも関係を保持します:

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

ステータスが省略されている場合は、アナライザーによって解決されたことを意味します。`unresolved` は依存関係が確立できなかったことを、`ambiguous` は競合する解釈が残っていることを意味します。候補は可能性のあるターゲットであり、確定された接続ではありません。既知のエンドポイントは ShortID を使用します。シンボルのない未解決のターゲットは名前として表示される場合があります。

JSON は `resolution` と `candidates` に同じ情報を含み、`to_short` は不確実なターゲットでは省略されます。C/C++ のインクルードには `meta/file` と `meta/include` シンボルが含まれるため、エッジは ShortID を使用します。これにより、コンパイラレベルのヘッダー解決を主張することなく、書かれたディレクティブを記録します。

依存関係メトリクスは不確実な関係を除外します。HTML グラフは個々の不確実なリンクを描画せず、その数を表示します。文法については [出力形式リファレンス](../../reference/output-format/) を、解決の制限については [言語サポートリファレンス](../../reference/language-support/) を参照してください。

## 古い出力の保護

出力ディレクトリに以前の実行からの `.skt` ファイルがすでに存在する場合、`codeknit` は古いデータと新しいデータが混在するのを防ぐため、新しい出力の書き込みを拒否します。

この動作を上書きし、書き込み前に出力ディレクトリをクリーンするには、`--clean` フラグを使用します:

```bash
codeknit parse ./src --clean
```

これにより、新鮮で一貫性のある出力セットが確保されます。

`0.5.0` にアップグレードする際は、出力セット全体を再生成してください。シンボル ID と関係の結果は変更される可能性があります。異なる実行からのチャンクを混在させないでください。

## ヒント

- ✅ **ほとんどのプロジェクトでは `directory-flat` をデフォルトとして使用**します。これは可読性と管理のバランスが取れています。
- 🔍 大規模なコードベースでは `--minify` を使用して、共有辞書 (`dict.skt`) によるトークン使用量を削減します。
- 🔗 `[edges]` セクションは **デフォルトでは除外** され、トークンを節約します。`calls`、`contains`、`inherits` などの関係データが必要な場合は `--edges` を使用してください。
- 🧾 スクリプトやインテグレーションが構造化データを必要とする場合は、`--format json` を使用してください。
- 🧹 同じ出力ディレクトリで再実行する場合は、常に `--clean` を使用してください。
- 📁 エディターでソースファイルと直接 `.skt` ファイルを関連付けたい場合は、`directory-tree` を使用してください。
