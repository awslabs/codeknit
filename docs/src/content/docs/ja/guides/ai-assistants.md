---
title: AIアシスタントとの連携
description: Claude Code、Codex、Kiro用のcodeknitプラグインをインストールします。
---

Claude Code、Codex、またはKiroにcodeknitプラグインをインストールし、最新の状態に保ち、構造や重複コードに関する質問をアシスタントに行います。このプラグインには、`codeknit`という1つのスキルが含まれています。このスキルは、アシスタントに`codeknit parse`、`codeknit graph analyze`、`codeknit fingerprint`をいつ実行するか、そして`.skt`出力をどのように読み取るかを指示します。

`codeknit`バイナリを自分でインストールする必要はありません。スキルが初めて実行されると、バンドルされた`scripts/install-codeknit.sh`が実行され、以下の処理を行います：

1. `codeknit`がすでに`PATH`または`~/.local/bin`に存在する場合は、即座に終了します。
2. それ以外の場合は、お使いのOSとアーキテクチャに対応した最新のGitHubリリースをダウンロードします。
3. ダウンロードをリリースの`checksums.txt`と照合して検証します。
4. バイナリを`~/.local/bin`にインストールし、出力の最後の行に絶対パスを表示します。

`~/.local/bin`が`PATH`に含まれていない場合、アシスタントはセッションの残りの部分でそのフルパスを使用してcodeknitを呼び出します。

インストールを制御するには、以下の環境変数を設定します：

| 変数 | 効果 |
| --- | --- |
| `CODEKNIT_VERSION` | リリースを固定します。例：`0.6.0`。 |
| `CODEKNIT_INSTALL_DIR` | インストール場所を変更します。 |

代わりに自分でバイナリをインストールする場合は、[インストール](../../installation/)を参照してください。

## プラグインのインストール

お使いのアシスタントに応じたセクションを選択してください。

### Claude Code

シェルで以下のコマンドを実行します。Claude Codeセッション内では、`/plugin ...`として同じコマンドを実行します。

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

1. Powersパネルを開きます。
2. **Add Custom Power**を選択し、**Import power from GitHub**を選択します。
3. `https://github.com/awslabs/codeknit`を入力します。

### Kiro CLI

Kiro CLIはローカルパスからインストールするため、まずリポジトリをクローンします：

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### `--sparse`について

`--sparse`オプションは任意です。このオプションは、クローンをプラグインマニフェストと`skills/`に限定し、ソースツリー全体をクローンしません。

### スキルを手動でインストール

アシスタントがAgent Skillsを読み込むがプラグイン機能がない場合は、`skills/codeknit/`をアシスタントのスキルディレクトリにコピーします：

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## プラグインの更新

| アシスタント | コマンドまたは操作 |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Powersパネルで**Check for updates**を選択します。 |
| Kiro CLI | `git -C ~/.codeknit-power pull`を実行し、再度`kiro-cli powers install ~/.codeknit-power`を実行します。 |

プラグインの更新では、すでにインストールされているバイナリは更新されません。バイナリのバージョンを変更するには、上記のように`CODEKNIT_VERSION`を設定します。

codeknitをアップグレードして抽出内容が変更された場合、シンボルIDも変更される可能性があります。異なる実行からのチャンクを混在させるのではなく、アシスタントにすべての出力を再生成させてください。

## スキルがアシスタントに教えること

このスキルは、アシスタントに以下のルールを教えるため、リクエストごとに繰り返す必要はありません：

- **質問に応じてコマンドを選択。**
  - `parse`は、何が存在し、どこにあり、何が何を呼び出し、継承し、または含んでいるかを回答します。
  - `graph analyze`は、循環依存、ハブ、デッドコード、ゴッドクラス、ボトルネックを見つけます。
  - `fingerprint`は、コピー＆ペーストやリファクタリングの候補を見つけます。
- **デフォルトで`directory-flat`出力モードを使用。** アシスタントはチャンク化されたファイルを`./skeleton/`に書き込み、必要なチャンクのみを読み込みます。`inline`出力モードは、単一のファイルまたは少数のファイルに対してのみ使用します。これは、`inline`がすべてをコンテキストウィンドウにダンプするためです。
- **構造に関する質問には、生のソースではなく`.skt`出力を読み取る。** アシスタントは、実装の詳細を検査または変更する場合にのみソースファイルを開き、出力の行範囲を使用します。スキルの`references/output-format.md`に完全なフォーマットが記載されています。[出力フォーマットリファレンス](../../reference/output-format/)も参照してください。
- **類似度の範囲を選択し、マッチをシグナルとして扱う。** フィンガープリントは、意図ではなく構造的な形状を比較します。アシスタントは、リファクタリングを提案する前に、ペアの両方のシンボルを読み取ります。

## 構造分析を依頼する

1. アシスタントにコードベースの構造を分析させます。例：「`./src`の構造をマップし、authパッケージに依存しているモジュールを教えてください。」
2. アシスタントはバイナリがインストールされていることを確認し、`codeknit parse`をエッジを有効にして実行し、`.skt`出力を読み取ります。
3. 依存関係、呼び出しチェーン、またはデッドコードについてのフォローアップの質問をします。アシスタントは出力から回答します。
4. より深い分析のために、構造レビューを依頼します。例：「`./src`の循環依存とゴッドクラスを見つけてください。」アシスタントは`codeknit graph analyze`を実行し、レポートを読み取ります。

出力で`[unresolved]`または`[ambiguous]`とマークされた関係は不確定であり、確認された依存関係ではありません。これらに基づく回答は、検証が必要な候補として扱ってください。

これらのリクエストの背後にあるコマンドについては、[parseコマンドガイド](../parse-command/)と[graphコマンドガイド](../graph-commands/)を参照してください。

## 重複検出を依頼する

1. アシスタントに重複コードを見つけさせます。例：「`./src`内のコピー＆ペーストされた関数を見つけてください。」
2. アシスタントは`codeknit fingerprint`を実行し、`./skeleton/fingerprints.skt`の`[duplicates]`セクションを読み取ります。
3. アシスタントは、フラグが付けられた各ペアの両方のシンボルを開き、リファクタリングを提案します。
4. 結果がノイズの場合は、類似度の範囲を調整するか、セマンティックな再ランキングを追加するように依頼します。再ランキングには、プルされた埋め込みモデルを持つローカルのOllamaサーバーが必要です。

これらのリクエストの背後にあるオプションについては、[fingerprintコマンドガイド](../fingerprint-command/)を参照してください。

## ヒント

- 構造に関する質問には、生のソースではなく`.skt`ファイルを読み取るようにアシスタントにリマインドしてください。このフォーマットはコンパクトで、シンボル、関係、行範囲を含んでいるため、コンテキストを小さく保つことができます。
- コード品質の洞察には`graph analyze`を使用します。発見事項はレビューの候補です。外部の呼び出し元や動的ディスパッチによって結果が変わる可能性があるためです。
- 大規模なリファクタリングの前にfingerprintを実行し、統合すべきコピー＆ペーストを見つけます。
- アシスタントがすでに`.skt`出力を含むディレクトリに`parse`を再実行する場合、codeknitは書き込みを拒否します。以前の出力を置き換えるために`--clean`を渡すように依頼してください。
- バイナリのダウンロードが失敗した場合（オフラインまたはサポートされていないプラットフォーム）、[リリースページ](https://github.com/awslabs/codeknit/releases)から手動でリリースを取得するか、[インストール](../../installation/)を参照してください。
- 実行が部分的な結果を報告した場合は、stderrと`[errors]`セクションを確認してください。詳細は[分析の信頼性](../../reference/analysis-reliability/)を参照してください。