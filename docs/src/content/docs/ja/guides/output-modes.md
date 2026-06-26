---
title: 出力モード
description: プロジェクトのサイズとワークフローに適した出力モードを選択してください。
---

`codeknit` は、`--output-mode` フラグで制御される3つの出力モードをサポートしています。各モードは、抽出された**コード構造**をディスク（またはstdout）に書き込む方法を決定します。

出力モードは出力形式とは別です。デフォルトの形式は `.skt` です。`--format json` を渡すと、同じ解析結果を機械可読のJSONとして出力します。ディレクトリモードでは、JSONは `codeknit.json` に書き込まれます。`inline` モードでは、JSONはstdoutに書き込まれます。

### directory-flat（デフォルト、推奨）

- **動作**: `map_001.skt`、`map_002.skt` などのチャンク化された `.skt` ファイルを書き込みます。
- **出力ディレクトリ**: デフォルトで `./skeleton/`
- **分割**: `--max-lines` 制限（デフォルト: 500行）を超えるファイルは分割されます。
- **ユースケース**: ほとんどのプロジェクトに最適です。ファイルサイズを制限することで、出力を整理し、読みやすく保ちます。タスクに関連するチャンクのみを読むことができます。
- **ミニフィケーション**: `--minify` が有効な場合、出力ディレクトリに `dict.skt` ファイルも生成され、圧縮された値のトークンマッピングが含まれます。

例:

```bash
codeknit parse ./src
# 出力: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **動作**: ソースディレクトリ構造を正確にミラーリングします。
- **出力ディレクトリ**: デフォルトで `./skeleton/`
- **マッピング**: ソースファイルごとに対応するパスに `.skt` ファイルが1つ作成されます。
- **ユースケース**: 特定のファイルの構造を素早く調べたい場合に最適です。元のコードベースと並行してナビゲーションするのに便利です。

例:

```bash
codeknit parse ./src --output-mode directory-tree
# 出力: ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt, など
```

### inline

- **動作**: すべての出力をstdoutにダンプします。
- **出力ディレクトリ**: 作成されません
- **ユースケース**: 単一ファイルまたは非常に小さなプロジェクト（5ファイル未満）にのみ推奨されます。他のツールにパイプしたり、単一ファイルを対話的に検査する場合に便利です。

例:

```bash
codeknit parse ./src/main.go --output-mode inline
# 出力: ターミナルに直接出力
```

### JSON形式

- **動作**: `files`、`symbols`、オプションの `edges`、およびオプションの `errors` を含む単一のJSONドキュメントを出力します。
- **出力場所**: ディレクトリモードでは `codeknit.json`、または `inline` モードではstdout。
- **ユースケース**: スクリプト、エディタ統合、CIチェック、および構造化データを必要とするツールに最適です。

例:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

サンプル出力:

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

### 決定表

| モード             | 最適な用途                                | 出力場所                                     |
| ---------------- | --------------------------------------- | ------------------------------------------- |
| `directory-flat` | ほとんどのプロジェクト（デフォルト、推奨）    | `./skeleton/map_001.skt`、`map_002.skt`、... |
| `directory-tree` | ソースコードと並行して出力をナビゲーションする場合 | `./skeleton/<ミラーリングされたパス>.skt`    |
| `inline`         | 単一ファイル、他のツールへのパイプ処理     | stdout — 単一ファイルまたは非常に小さなプロジェクトにのみ使用 |

| 形式   | 最適な用途                           | 出力                                                   |
| ------ | ---------------------------------- | ----------------------------------------------------- |
| `skt`  | LLMコンテキストと人間による検査   | `.skt` ファイルまたはstdout                           |
| `json` | スクリプトと構造化統合             | ディレクトリモードでは `codeknit.json`、または `inline` モードではstdout |

### 経験則

- **迷ったとき** → `directory-flat`（デフォルト）を使用
- **単一ファイルの検査** → `inline` も許容
- **数ファイル以上** → `directory-flat` または `directory-tree` を推奨
- **大規模なコードベース** → トークン使用量を減らすために `--minify` を追加
- **同じ出力に再実行** → 古い `.skt` ファイルを削除するために `--clean` を使用

### ミニフィケーション

`--minify` フラグは、繰り返されるトークン（`exported`、`async`、一般的な型名などのプロパティキー）の辞書ベースの圧縮を有効にします。有効にすると:

- 繰り返される値は短いコード（`d0`、`d1`、`d2`、...）に置き換えられます。
- 出力ディレクトリに `dict.skt` ファイルが書き込まれ、コードから元の値へのマッピングが含まれます。
- 大規模なコードベースの出力サイズを大幅に削減します。
- `directory-flat` と `directory-tree` の両モードで機能します。

ミニフィケーションされた出力例:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```

この形式は、完全な情報を保持しながらトークンの使用量を最小限に抑えるため、LLMベースの分析に最適です。