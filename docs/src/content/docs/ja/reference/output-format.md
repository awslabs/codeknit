---
title: 出力フォーマットリファレンス
description: codeknitが使用する.skt出力フォーマットの完全なリファレンスです。
---

`.skt`（skeleton）フォーマットは、`codeknit`が抽出したコード構造を表現するためのコンパクトで人間が読めるテキストフォーマットです。シンボル、関係性、メタデータを最小限の形式で含み、LLMの消費や構造分析に適しています。

`.skt`ファイルはセクションに分かれています。各セクションは角括弧で囲まれたヘッダーで始まります。セクションは任意の順序で現れることがありますが、`[symbols]`が通常最初に来ます。

## [symbols]

`[symbols]`セクションには、抽出されたすべてのシンボルがソースファイルごとにグループ化されてリストされています。各ファイルは`##`ヘッダーとそれに続くファイルパスで紹介されます。

### 行フォーマット

各シンボルは以下の構造を持つ1行で表現されます：

```
ShortID category/kind Lstart-Lend signature {properties}
```

### フィールド

- **ShortID**: 各シンボルに割り当てられた連続識別子（例：`S1`、`S2`、`S3`）。エッジや他のセクションで参照するために使用されます。
- **category/kind**: シンボルのカテゴリと特定の種類をスラッシュで区切ったペア。
- **Lstart-Lend**: シンボルが定義されているソースファイル内の行範囲（例：`L10-L15`）。
- **signature**: シンボルの名前と型情報。シンボルの種類によってフォーマットが異なります：
  - `name` — 型、値、モジュール用
  - `name(params)` — 戻り値の型がないcallable用
  - `name(params) -> returnType` — 戻り値の型があるcallable用
- **{properties}**: 波括弧で囲まれたオプションのメタデータ。複数のプロパティはカンマで区切られます。

### パラメータ

- 型のない言語：`paramName`
- 型のある言語：`paramName: type`
- 既知のシンボルと一致する型参照は、そのShortIDに置き換えられます（例：`config: S5`の代わりに`config: Config`）。

### プロパティ

一般的なプロパティには以下が含まれます：

- `async`: `true`または`false`
- `exported`: `true`または`false`
- `static`: シンボルが静的である場合に存在
- `visibility=public|private|protected`
- `receiver=*TypeName`: メソッドの場合、レシーバーの型を示す

### シンボルのカテゴリと種類

| カテゴリ   | 種類                                      | 例                                           |
| ---------- | ------------------------------------------ | -------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`       |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`               |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`           |
| `module`   | package, namespace                         | `module/package`                             |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### 例

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
```

## [edges]

`[edges]`セクションは、`--edges`フラグが有効な場合に関係性を記録します。既知のエンドポイントはShortIDを使用します。シンボルに解決できないターゲットは名前として表示されることがあります。

### 行フォーマット

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

ステータスが省略されている場合、アナライザーが関係性を解決したことを意味します。`unresolved`は依存関係を確立できなかったことを、`ambiguous`は競合する解釈が可能であることを意味します。候補IDは可能性のあるターゲットを識別し、確認された接続ではありません。解決は構文とスコープに基づいており、コンパイラの保証ではありません。

ターゲットは、ソース、種類、解決、候補セットが一致する場合にのみグループ化されます。シンボルのない名前はソースファイルに対して相対的です。複雑な識別子は、エスケープされた引用符、バックスラッシュ、改行を含む引用符付きのフルIDを使用します。ShortIDに似たリテラル名は、フルIDとして引用する必要があります（例：`"app.go::S2"`）。シンボル定義は後続のファイルや出力チャンクに現れることがあります。

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### ファイルとインクルードのエンドポイント

C/C++のインクルードには明示的なメタデータシンボルがあります。ヘッダーのスペルはシンボルリストに一度だけ現れ、エッジはShortIDを使用しますが、ヘッダーの解決は未解決のままです：

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

インクルードシンボルは、解決されたヘッダー宣言ではなく、書かれたディレクティブを記録します。ファイル内の繰り返しインクルードはエンドポイントを共有します。ファイル/インクルードのメタデータは依存性ランキングやデッドコードの発見から除外されます。

不確実な関係性はパース出力に保持されますが、依存性メトリクスからは除外されます。グラフレポートには除外された関係性のカウントが表示され、HTMLビューアーでは個々の不確実なエッジを描画せずにそのカウントを表示します。

`0.5.0`にアップグレードする際は、すべての出力チャンクを一緒に再生成してください。シンボルIDと関係性の結果は変更される可能性があるため、異なる実行からのチャンクを混ぜてはいけません。

### エッジの種類

| 種類         | 意味                                         |
| ------------ | -------------------------------------------- |
| `calls`      | 関数/メソッドの呼び出し                      |
| `contains`   | クラスがメソッドを含む、モジュールが関数を含む |
| `inherits`   | クラスが他のクラスを継承する                 |
| `implements` | クラスがインターフェースを実装する           |
| `overrides`  | メソッドが親メソッドをオーバーライドする     |
| `references` | シンボルが他のシンボルを参照する             |
| `imports`    | モジュールが他のモジュールをインポートする   |
| `decorates`  | デコレーターがシンボルに適用される           |

### 例

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

`[errors]`セクションには、完全にパースできなかったファイルがリストされます。

### フォーマット

各行は`-`で始まり、ファイルパスとエラーメッセージが続きます：

```
- path/to/file.go: syntax error at line 42
```

### 例

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

`[dict]`セクションは、`--minify`フラグが使用された場合にのみ表示されます。出力サイズを削減するために、繰り返し使用される文字列トークンを短い辞書コードにマッピングします。

### フォーマット

各行は辞書コード（`d0`、`d1`など）を展開された値にマッピングします：

```
- d0: async=false
- d1: callable/method
- d2: exported
```

ファイルの残りの部分では、これらのコードが完全な値の代わりに使用されます。

関係性の種類も辞書コードを使用できますが、ステータスラベルは読み取り可能なままです：`S1 --d2[unresolved]--> S3`。`d2`を辞書を使用して展開し、unresolvedステータスを保持します。

### 例

```skt
[dict]
- d0: async=false
- d1: callable/method
- d2: exported

[symbols]
## src/handler.py
S1 type/class L1-L6 Handler {}
S2 d1 L2-L3 __init__(name) {d0}
S3 d1 L5-L6 handle(request) {d0}

[edges]
S1 --contains--> S2, S3
```

## 完全な例

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {d0}
S3 d1 L10-L12 NewServer(addr: string) -> *S2 {d0}
S4 callable/method L14-L20 Serve() {d0, receiver=*Server}
S5 callable/function L22-L25 handleError(err: error) -> bool {}

[edges]
S2 --contains--> S4
S4 --calls--> S5
S3 --references--> S2

[errors]
- utils/broken.go: syntax error at line 5
```

## JSON関係性

スクリプトや統合が構造化された出力を必要とする場合は、JSONを使用します：

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

各エッジには`from`、`to`、`kind`が含まれます。既知のソースIDは`from_short`に表示され、解決されたターゲットIDは`to_short`に表示されます。不確実なエッジは`resolution`を追加し、利用可能な場合は`candidates`にフルシンボルIDを含みます。ターゲットにインクルードメタデータシンボルがあっても、`to_short`は省略されます。

例えば、曖昧な呼び出しは次のように表現できます：

```json
{
  "from": "app.py::run",
  "from_short": "S1",
  "to": "app.py::Helper",
  "kind": "calls",
  "resolution": "ambiguous",
  "candidates": ["helpers/a.py::Helper", "helpers/b.py::Helper"]
}
```

JSONとSKTは、異なるエンコーディングで同じ不確実性を保持します。解決ルールと既知の制限については、[言語サポートとAPI忠実度](../language-support/)を参照してください。