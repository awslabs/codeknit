---
title: グラフコマンド
description: グラフアルゴリズムを使用してコードベースの構造を可視化および分析します。
---

codeknitは、コードベースの構造を理解し改善するための2つの強力なグラフコマンドを提供します：インタラクティブな可視化のための`graph show`と、自動構造分析のための`graph analyze`です。

## graph show

コードベースのインタラクティブなHTMLグラフ可視化を生成します。

```bash
codeknit graph show <input-path>
```

このコマンドはコードベースを解析し、インタラクティブなグラフ可視化を含む自己完結型のHTMLファイルを生成します。シンボル（関数、クラス、型）はノードとして表示され、それらの関係（呼び出し、包含、実装）はエッジとして表示されます。可視化は自動的にデフォルトのブラウザで開きます。

### フラグ

| フラグ             | デフォルト                          | 説明                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | 出力HTMLファイルのパス                        |
| `--collect-test` | `false`                          | 分析にテストファイルを含める               |
| `--workers`      | `NumCPU`                         | 最大同時解析ゴルーチン数            |
| `--verbose`      | `false`                          | 処理中の進捗情報を表示する |

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

コードベースに対して構造的なグラフアルゴリズムを実行し、コード品質の洞察を含むLLMが読み取り可能な`.skt`レポートを出力します。

```bash
codeknit graph analyze <input-path>
```

このコマンドは、循環依存、ハブシンボル、デッドコード、ゴッドクラス、アーキテクチャのボトルネックなどの一般的なコード品質の問題を検出します。

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
- エントリーポイントからの到達可能性
- 弱連結成分
- 依存度（パッケージ結合強度）
- Main Sequenceからの距離（A+Iバランス）
- ショットガンサージェリー検出
- 機能横恋慕（Feature Envy）検出
- 安定依存の原則違反
- インターフェース分離の原則違反
- 包含階層の深さ

### フラグ

| フラグ                      | デフォルト                         | 説明                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | 出力`.skt`ファイルのパス                                  |
| `--collect-test`          | `false`                         | 分析にテストファイルを含める                           |
| `--workers`               | `NumCPU`                        | 最大同時解析ゴルーチン数                        |
| `--verbose`               | `false`                         | 処理中の進捗情報を表示する             |
| `--fan-threshold`         | `10`                            | ハブシンボルとしてフラグを立てる最小ファンインまたはファンアウト           |
| `--god-threshold`         | `15`                            | ゴッドクラス/関数としてフラグを立てる最小contains-エッジ数 |
| `--max-inheritance-depth` | `5`                             | これより深い継承チェーンをフラグする                 |
| `--top-n`                 | `30`                            | ランキング出力セクションの上限；0 = 無制限                 |
| `--betweenness-threshold` | `0.001`                         | 報告する最小中間中心性値           |
| `--propagation-cutoff`    | `0.05`                          | 変更伝播を続ける最小確率       |

### 例

```skt
# デフォルト設定で構造分析を実行
codeknit graph analyze ./myproject

# カスタム出力と閾値
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# セクションごとに表示する結果を増やす
codeknit graph analyze ./myproject --top-n 50

# テストファイルを含める
codeknit graph analyze ./src --collect-test
```
