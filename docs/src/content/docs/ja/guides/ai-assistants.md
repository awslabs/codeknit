---
title: AIアシスタントとの連携
description: codeknitをKiro、Claude Code、およびその他のAIコーディングアシスタントのスキルとして設定する方法。
---

codeknitには、AIコーディングアシスタントが効果的に使用する方法を教える、すぐに使えるスキルが同梱されています。これらのスキルにより、アシスタントは手動でのプロンプトなしでコード構造の抽出、重複の検出、および構造分析を実行できます。

## スキルの概要

codeknitは2つのスキルを提供します：

- **`codeknit-parse`**：アシスタントにコード構造（関数、クラス、メソッド、変数）と関係性（呼び出し、継承、包含）を`.skt`ファイルに抽出する方法を教えます。
- **`codeknit-fingerprint`**：アシスタントにファジー・ハッシングを使用して重複および近似重複コードを検出する方法を教えます。

各スキルには、アシスタントが使用方法、フラグ、出力形式、およびワークフローを理解するためにオンデマンドで読み取るドキュメントが含まれています。

## インストール

インストーラーヘルパーを使用して、スキルディレクトリをアシスタントのスキルフォルダにコピーします。インストーラーはバンドルされたスキルファイルのみをダウンロードするため、リポジトリをクローンする必要はありません。

**Codex**、**Kiro**、および**Claude Code**用にインストール：

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash
```

特定のアシスタント用にインストール：

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant codex
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant kiro
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant claude
```

ローカルチェックアウトからは、Makefileヘルパーを使用できます：

```bash
make skills-install-dry-run
make skills-install
```

インストーラーは既存のスキルディレクトリをデフォルトでスキップします。置き換えるには、`--force`を追加します：

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant all --force
```

インストール後、アシスタントは自動的にcodeknitコマンドの呼び出し方法、適切なフラグの選択、および`.skt`出力の解釈方法を理解します。

## 各スキルが教える内容

### codeknit-parse

`codeknit-parse`スキルは、アシスタントに以下を教えます：

- 様々なシナリオに適したフラグで`codeknit parse`を実行する
- 適切な出力モードを選択する：
  - `directory-flat`（デフォルト）ほとんどのプロジェクト向け
  - `inline` 単一ファイルまたは小さな入力向け
  - `directory-tree` ソース構造をミラーリングするため
- `.skt`出力ファイルを読み取り、解釈する。これには`[symbols]`、`[edges]`、およびオプションの`[dict]`セクションが含まれる
- 構造データをリファクタリング、依存関係マッピング、およびコードレビューに使用する
- `codeknit graph analyze`を実行して、より深いコード品質の洞察を得る（循環依存、ハブシンボル、god classesなど）

### codeknit-fingerprint

`codeknit-fingerprint`スキルは、アシスタントに以下を教えます：

- `codeknit fingerprint`を使用して重複検出、DRY監査、およびリファクタ識別を行う
- 適切な類似度範囲（`--min-similarity`、`--max-similarity`）を選択する
- `[duplicates]`セクションを読み取り、近似重複コードを識別する
- フィンガープリントは意味的な意図ではなく、構造的な形状を測定することを理解する
- 必要に応じてOllama埋め込みで`--rerank`を使用して偽陽性を減らす

## ワークフローの例

### 構造分析

1. アシスタントにコードベースの構造を分析するよう依頼する
2. アシスタントが`codeknit parse ./src`を実行し、結果の`.skt`ファイルを読み取る
3. アシスタントが構造的な質問に答える：依存関係、呼び出しチェーン、dead code
4. より深い洞察のために、`codeknit graph analyze ./src`を実行し、レポートを解釈する

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### 重複検出

1. アシスタントに重複コードを見つけるよう依頼する
2. アシスタントが`codeknit fingerprint ./src`を実行する
3. アシスタントが出力の`[duplicates]`セクションを読み取る
4. アシスタントがフラグ付けされたペアを調査し、統合を提案する

```skt
[duplicates]
S1, S2: 87% 類似度
S3, S4: 76% 類似度
```

## ヒント

- **構造的な質問には、生のソースではなく`.skt`ファイルを読み取る** — これらは抽出された構造をコンパクトで信頼性の高い形式で含んでいる
- `codeknit graph analyze`を使用して、循環依存、ハブシンボル、深い継承チェーンなどのコード品質の問題を発見する
- 大規模なリファクタリングの前に`codeknit fingerprint`を実行して、統合すべきコピー＆ペーストされたコードを識別する
- `.skt`形式はトークン効率が良く設計されており、LLMのコンテキストウィンドウに理想的
- 大規模なコードベースを処理する際には、`--minify`を使用してトークン使用量をさらに削減する