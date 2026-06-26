---
title: CLIリファレンス
description: すべてのcodeknitコマンドとフラグの完全なリファレンスです。
---

## codeknit

対話型ターミナルUI（TUI）を起動し、利用可能なコマンドとオプションをガイドします。

```bash
codeknit
```

## codeknit parse

ソースコードから構造情報を抽出し、`.skt`ファイルまたはJSONに出力します。

```bash
codeknit parse <input-path> [output-dir]
```

| フラグ             | タイプ   | デフォルト          | 説明                                                                            |
| ---------------- | ------ | ---------------- | ----------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat` | 出力モード: `inline`、`directory-flat`、または`directory-tree`                           |
| `--format`       | string | `skt`            | 出力形式: `skt`または`json`                                                        |
| `--max-lines`    | int    | `500`            | 出力ファイルごとの最大行数（`directory-flat`および`directory-tree`モードに適用）                     |
| `--collect-test` | bool   | `false`          | 解析にテストファイルを含める                                                             |
| `--minify`       | bool   | `false`          | 辞書ベースの出力圧縮を有効にする                                                            |
| `--edges`        | bool   | `false`          | 出力に`[edges]`セクションを含める（デフォルトではトークン節約のためオフ）                                |
| `--clean`        | bool   | `false`          | 書き込み前に出力ディレクトリから古い`.skt`ファイルを削除する                                           |
| `--workers`      | int    | `0` (NumCPU)     | 最大同時解析goroutine数                                                             |
| `--verbose`      | bool   | `false`          | 処理中に進捗情報を表示する                                                              |

出力ディレクトリは指定されない場合、`./skeleton`がデフォルトです。`inline`モードでは、出力はstdoutに書き込まれ、ディレクトリは使用されません。`--format json`を使用すると、ディレクトリ出力は`codeknit.json`として書き込まれます。

## codeknit graph show

コードベース構造のインタラクティブなHTMLグラフ可視化を生成します。

```bash
codeknit graph show <input-path>
```

| フラグ             | タイプ   | デフォルト                          | 説明                                  |
| ---------------- | ------ | -------------------------------- | ----------------------------------- |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | 出力HTMLファイルパス                        |
| `--collect-test` | bool   | `false`                          | 解析にテストファイルを含める               |
| `--workers`      | int    | `0` (NumCPU)                     | 最大同時解析goroutine数                |
| `--verbose`      | bool   | `false`                          | 処理中に進捗情報を表示する                |

生成されたHTMLファイルは自己完結型で、デフォルトのブラウザで自動的に開きます。

## codeknit graph analyze

構造分析アルゴリズムを実行し、LLMが読み取り可能な`.skt`レポートを出力します。

```bash
codeknit graph analyze <input-path>
```

| フラグ                      | タイプ    | デフォルト                         | 説明                                                   |
| ------------------------- | ------- | ------------------------------- | ---------------------------------------------------- |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | 出力`.skt`ファイルパス                                       |
| `--collect-test`          | bool    | `false`                         | 解析にテストファイルを含める                                |
| `--workers`               | int     | `0` (NumCPU)                    | 最大同時解析goroutine数                                 |
| `--verbose`               | bool    | `false`                         | 処理中に進捗情報を表示する                                  |
| `--fan-threshold`         | int     | `10`                            | ハブシンボルとしてフラグを立てる最小ファンインまたはファンアウト値                |
| `--god-threshold`         | int     | `15`                            | god class/functionとしてフラグを立てる最小contains-edge数      |
| `--max-inheritance-depth` | int     | `5`                             | この値より深い継承チェーンにフラグを立てる                          |
| `--top-n`                 | int     | `30`                            | ランク付けされた出力セクションの上限。`0`は無制限を意味する                |
| `--betweenness-threshold` | float64 | `0.001`                         | 報告する最小媒介中心性値                                      |
| `--propagation-cutoff`    | float64 | `0.05`                          | 変更伝播シミュレーションを続行する最小確率                          |

## codeknit fingerprint

ファジー・ハッシングを使用して重複および近似重複コードを検出します。

```bash
codeknit fingerprint <input-path>
```

| フラグ               | タイプ   | デフォルト                       | 説明                                                                                                                  |
| ------------------ | ------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | 出力`.skt`ファイルパス                                                                                                      |
| `--min-similarity` | int    | `65`                          | 報告する最小類似度（0–100）                                                                                                  |
| `--max-similarity` | int    | `95`                          | 報告する最大類似度（0–100）                                                                                                  |
| `--show-all`       | bool   | `false`                       | 生のトークンデータを含む`[fingerprints]`セクションを含める                                                                     |
| `--rerank`         | bool   | `false`                       | Ollamaを使用してCTPH候補を意味的埋め込みで再ランク付けする（`ollama serve`および`ollama pull qwen3-embedding:0.6b`が必要） |
| `--model`          | string | `qwen3-embedding:0.6b`        | `--rerank`で使用するOllama埋め込みモデル                                                                              |
| `--collect-test`   | bool   | `false`                       | 解析にテストファイルを含める                                                                                               |
| `--workers`        | int    | `0` (NumCPU)                  | 最大同時解析goroutine数                                                                                                |
| `--verbose`        | bool   | `false`                       | 処理中に進捗情報を表示する                                                                                               |

## codeknit completion

サポートされているシェル用の補完スクリプトを生成します。

```bash
codeknit completion <shell>
```

サポートされているシェル: `bash`、`zsh`、`fish`、`powershell`。

## グローバルフラグ

| フラグ           | 説明                       |
| -------------- | ------------------------ |
| `--version`    | バージョン情報を表示する         |
| `--help`, `-h` | 現在のコマンドのヘルプを表示する |