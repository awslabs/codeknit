---
title: グラフコマンド
description: グラフアルゴリズムを使用してコードベースの構造を可視化および分析します。
---

codeknitは、グラフコマンドを提供し、構造の可視化、自動化された分析の実行、現在の依存関係グラフとGitの変更履歴の組み合わせをサポートします。

## graph show

コードベースのインタラクティブなHTMLグラフ可視化を生成します。

```bash
codeknit graph show <input-path>
```

このコマンドはコードベースを解析し、インタラクティブなグラフ可視化を含む自己完結型のHTMLファイルを生成します。シンボル（関数、クラス、型）はノードとして表示され、それらの関係（呼び出し、包含、実装）はエッジとして表示されます。可視化はデフォルトのブラウザで自動的に開きます。

ビューアは解決された関係を描画します。不確定な関係のカウントは表示しますが、個々の未解決または曖昧な関係やその候補リンクは描画しません。これらの関係を調べるには、parse出力を使用します：

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### フラグ

| フラグ             | デフォルト                          | 説明                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | 出力HTMLファイルのパス                        |
| `--collect-test` | `false`                          | 分析にテストファイルを含める               |
| `--workers`      | `NumCPU`                         | 最大同時解析ゴルーチン数            |
| `--verbose`      | `false`                          | 処理中の進捗情報を表示             |

### 例

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

コードベースに対して構造的なグラフアルゴリズムを実行し、コード品質の洞察を含むLLMが読み取れる`.skt`レポートを出力します。

```bash
codeknit graph analyze <input-path>
```

このコマンドは、循環依存、ハブシンボル、デッドコード、ゴッドクラス、アーキテクチャのボトルネックなど、レビュー候補を報告します。

### アルゴリズム

分析には22の構造的グラフアルゴリズムが含まれます：

- 循環依存（TarjanのSCC）
- ハブ検出（高ファンイン/ファンアウト結合）
- 孤立検出（デッドコード候補）
- ゴッドクラス/関数検出（過剰な子要素）
- 不安定性メトリック（Robert C. MartinのCe/(Ca+Ce)）
- 深い継承チェーン
- 中間中心性（ボトルネック検出）
- 接続点（単一障害点）
- PageRank（再帰的重要度）
- 推移的ファンイン（影響範囲）
- 変更伝播シミュレーション
- 循環パッケージ依存
- レイヤー違反検出
- エントリポイントからの到達可能性
- 弱連結成分
- 依存度重み（パッケージ結合強度）
- Main Sequenceからの距離（A+Iバランス）
- 散弾銃手術検出
- フィーチャー嫉妬検出
- 安定依存違反
- インターフェース分離違反
- 包含深度

### レポートの解釈

依存メトリックは解決された関係を使用します。レポートは除外された未解決および曖昧な関係のカウントを含みます。ファイル/インクルードメタデータは依存度ランキングやデッドコードの発見から除外されます。

- ファンインとファンアウトは異なる隣接ノードをカウントします。繰り返しの関係はこれらのカウントを増加させず、包含は依存関係ではありません。
- 継承深度はすべての基底およびインターフェースを考慮します。循環階層には深度は割り当てられません。
- 到達可能性は、`main`、`Main`、または`init`という名前のcallableから開始し、依存エッジをたどります。メンバーに到達すると、そのコンテナは関連性があるとみなされますが、すべての兄弟が到達可能とマークされるわけではありません。認識されたエントリポイントがない場合、到達可能性の発見は省略されます。
- コールバック引数と返されたcallableは参照であり、コールバックが実行される証拠ではありません。
- 変更伝播は最強の依存パスを使用します。そのスコアとディレクトリベースのレイヤールールはヒューリスティックであり、測定された確率ではありません。

外部の呼び出し元、動的ディスパッチ、およびスキャン外の依存関係は結論を変える可能性があります。解決ルールと制限については、[言語サポートとAPI忠実度](../../reference/language-support/)を参照してください。

### フラグ

| フラグ                      | デフォルト                         | 説明                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | 出力`.skt`ファイルのパス                                  |
| `--collect-test`          | `false`                          | 分析にテストファイルを含める                           |
| `--workers`               | `NumCPU`                        | 最大同時解析ゴルーチン数                        |
| `--verbose`               | `false`                          | 処理中の進捗情報を表示             |
| `--fan-threshold`         | `10`                            | ハブシンボルとしてフラグを立てる最小ファンインまたはファンアウト           |
| `--god-threshold`         | `15`                            | ゴッドクラス/関数としてフラグを立てる最小包含エッジ数 |
| `--max-inheritance-depth` | `5`                             | これより深い継承チェーンにフラグを立てる                 |
| `--top-n`                 | `30`                            | ランク付けされた出力セクションの上限；0 = 無制限                 |
| `--betweenness-threshold` | `0.001`                          | 報告する最小中間中心性値           |
| `--propagation-cutoff`    | `0.05`                          | 変更伝播を続ける最小ヒューリスティックスコア   |

### 例

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

頻繁に変更され、構造的に重要なファイルをランク付けします：

```bash
codeknit graph hotspots <input-path>
```

スコアはコミット頻度、行の変更量、および新しさをファイルレベルのPageRank、推移的ファンイン、および中間中心性と組み合わせます。レポートは、同じコミットで繰り返し変更されるファイル間の時間的結合も特定します。

マージコミットはデフォルトで除外されます。50を超えるファイルを変更するコミットも除外されるため、生成されたファイル、ベンダー提供のファイル、または機械的なバルク変更が結果を歪めることはありません。

### フラグ

| フラグ                     | デフォルト                   | 説明                                     |
| ------------------------ | ------------------------- | ----------------------------------------------- |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | 出力ファイルのパス                                |
| `--format`               | `skt`                     | 出力形式：`skt`または`json`                  |
| `--since`                | `12mo`                    | 履歴ウィンドウ（例：`180d`、`12mo`、`2y`） |
| `--max-commits`          | `2000`                    | 調査する最大コミット数                      |
| `--max-files-per-commit` | `50`                      | 変更ファイル数が多いコミットを除外             |
| `--min-cochanges`        | `3`                       | 時間的結合のための最小共有コミット数    |
| `--top-n`                | `30`                      | レポートセクションごとの最大結果数              |
| `--include-merges`       | `false`                   | マージコミットを含める                           |
| `--collect-test`         | `false`                   | テストファイルを含める                              |
| `--workers`              | `NumCPU`                  | 最大同時解析ゴルーチン数           |
| `--verbose`              | `false`                   | 進捗情報を表示                      |

### 例

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```