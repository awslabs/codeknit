---
title: グラフコマンド
description: グラフアルゴリズムを使用してコードベースの構造を可視化し、分析します。
---

codeknitは、構造の可視化、自動分析の実行、現在の依存関係グラフとGit変更履歴の組み合わせを行うグラフコマンドを提供します。

## graph show

コードベースのインタラクティブなHTMLグラフ可視化を生成します。

```bash
codeknit graph show <input-path>
```

このコマンドはコードベースを解析し、インタラクティブなグラフ可視化を含む自己完結型のHTMLファイルを生成します。シンボル（関数、クラス、型）はノードとして表示され、それらの関係（呼び出し、包含、実装）はエッジとして表示されます。可視化はデフォルトのブラウザで自動的に開きます。

### フラグ

| フラグ             | デフォルト                          | 説明                                  |
| ---------------- | -------------------------------- | ------------------------------------ |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | 出力HTMLファイルパス                        |
| `--collect-test` | `false`                          | 分析にテストファイルを含める               |
| `--workers`      | `NumCPU`                         | 最大同時解析ゴルーチン数            |
| `--verbose`      | `false`                          | 処理中の進捗情報を表示                     |

### 例

```skt
# デフォルトの可視化を生成
codeknit graph show ./myproject

# カスタム出力ファイル
codeknit graph show ./myproject -o graph.html

# テストファイルを含める
codeknit graph show ./src --collect-test
```

## graph analyze

コードベースに対して構造グラフアルゴリズムを実行し、コード品質に関する洞察を含むLLMが読み取れる`.skt`レポートを出力します。

```bash
codeknit graph analyze <input-path>
```

このコマンドは、循環依存、ハブシンボル、デッドコード、ゴッドクラス、アーキテクチャのボトルネックなどの一般的なコード品質の問題を検出します。

### アルゴリズム

分析には22の構造グラフアルゴリズムが含まれます：

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
- 依存重み（パッケージ結合強度）
- Main Sequenceからの距離（A+Iバランス）
- 散弾銃手術検出
- フィーチャー嫉妬検出
- 安定依存違反
- インターフェース分離違反
- 包含深度

### フラグ

| フラグ                      | デフォルト                         | 説明                                              |
| ------------------------- | ------------------------------- | ------------------------------------------------ |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | 出力`.skt`ファイルパス                                  |
| `--collect-test`          | `false`                         | 分析にテストファイルを含める                           |
| `--workers`               | `NumCPU`                        | 最大同時解析ゴルーチン数                        |
| `--verbose`               | `false`                         | 処理中の進捗情報を表示                     |
| `--fan-threshold`         | `10`                            | ハブシンボルとしてフラグを立てる最小ファンインまたはファンアウト           |
| `--god-threshold`         | `15`                            | ゴッドクラス/関数としてフラグを立てる最小包含エッジ数 |
| `--max-inheritance-depth` | `5`                             | これより深い継承チェーンにフラグを立てる                 |
| `--top-n`                 | `30`                            | ランク付けされた出力セクションの上限。0 = 無制限                 |
| `--betweenness-threshold` | `0.001`                         | 報告する最小中間中心性値           |
| `--propagation-cutoff`    | `0.05`                          | 変更伝播を続ける最小確率       |

### 例

```skt
# デフォルトで構造分析を実行
codeknit graph analyze ./myproject

# カスタム出力と閾値
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# セクションごとにさらに多くの結果を表示
codeknit graph analyze ./myproject --top-n 50

# テストファイルを含める
codeknit graph analyze ./src --collect-test
```

## graph hotspots

頻繁に変更され、構造的に重要なファイルをランク付けします：

```bash
codeknit graph hotspots <input-path>
```

スコアはコミット頻度、行の変更量、新しさとファイルレベルのPageRank、推移的ファンイン、中間中心性を組み合わせたものです。レポートには、同じコミットで繰り返し変更されるファイル間の時間的結合も特定されます。

デフォルトではマージコミットは除外されます。50ファイル以上を変更するコミットも除外されるため、生成されたファイル、ベンダーファイル、または機械的な大量変更が結果を歪めることはありません。

### フラグ

| フラグ                     | デフォルト                   | 説明                                      |
| ------------------------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | 出力ファイルパス                                 |
| `--format`               | `skt`                     | 出力形式：`skt`または`json`                   |
| `--since`                | `12mo`                    | 履歴ウィンドウ。例：`180d`、`12mo`、`2y`  |
| `--max-commits`          | `2000`                    | 検査する最大コミット数                       |
| `--max-files-per-commit` | `50`                      | これより多くのファイルを変更するコミットを除外              |
| `--min-cochanges`        | `3`                       | 時間的結合のための最小共有コミット数     |
| `--top-n`                | `30`                      | レポートセクションごとの最大結果数               |
| `--include-merges`       | `false`                   | マージコミットを含める                            |
| `--collect-test`         | `false`                   | テストファイルを含める                               |
| `--workers`              | `NumCPU`                  | 最大同時解析ゴルーチン数            |
| `--verbose`              | `false`                   | 進捗情報を表示                       |

### 例

```bash
# 過去12か月を分析
codeknit graph hotspots ./myproject

# 2年間を分析し、JSONを出力
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# より大きなコミットを含め、より強い結合を要求
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```