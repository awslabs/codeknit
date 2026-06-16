---
title: 出力フォーマットリファレンス
description: codeknit で使用される .skt 出力フォーマットの完全なリファレンスです。
---

`.skt`（skeleton）フォーマットは、`codeknit` が抽出したコード構造を表現するためのコンパクトで人間が読みやすいテキストフォーマットです。シンボル、関係性、メタデータを LLM 消費や構造分析に適した最小限の形式で含んでいます。

`.skt` ファイルはセクションに分かれています。各セクションは角括弧で囲まれたヘッダーで始まります。セクションは任意の順序で現れることがありますが、`[symbols]` が通常最初に来ます。

## [symbols]

`[symbols]` セクションには、抽出されたすべてのシンボルがソースファイルごとにグループ化されてリストされています。各ファイルは `##` ヘッダーとそれに続くファイルパスで導入されます。

### 行フォーマット

各シンボルは以下の構造を持つ1行で表現されます：

```
ShortID category/kind Lstart-Lend signature {properties}
```

### フィールド

- **ShortID**: 各シンボルに割り当てられた連続した識別子（例：`S1`、`S2`、`S3`）。エッジや他のセクションで参照するために使用されます。
- **category/kind**: シンボルのカテゴリと具体的な種類をスラッシュで区切ったペア。
- **Lstart-Lend**: シンボルが定義されているソースファイル内の行範囲（例：`L10-L15`）。
- **signature**: シンボルの名前と型情報。シンボルの種類に応じてフォーマットが異なります：
  - `name` — 型、値、モジュールの場合
  - `name(params)` — 戻り値の型がない callable の場合
  - `name(params) -> returnType` — 戻り値の型がある callable の場合
- **{properties}**: 中括弧で囲まれたオプションのメタデータ。複数のプロパティはカンマで区切られます。

### パラメータ

- 型のない言語の場合：`paramName`
- 型のある言語の場合：`paramName: type`
- 既知のシンボルと一致する型参照は、その短縮 ID に置き換えられます（例：`config: Config` の代わりに `config: S5`）。

### プロパティ

一般的なプロパティには以下が含まれます：

- `async`: `true` または `false`
- `exported`: `true` または `false`
- `static`: シンボルが static の場合に存在
- `visibility=public|private|protected`
- `receiver=*TypeName`: メソッドの場合、レシーバーの型を示します

### シンボルのカテゴリと種類

| Category   | Kinds                          | Examples                               |
| ---------- | ------------------------------ | -------------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

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

`[edges]` セクションは、ShortID を使用してシンボル間の関係を定義します。

### 行フォーマット

```
FromID --kind--> ToID1, ToID2
```

複数のターゲット ID はカンマで区切られます。各行は一方向の関係を表します。

### エッジの種類

| Kind         | Meaning                                        |
| ------------ | ---------------------------------------------- |
| `calls`      | 関数/メソッドの呼び出し                        |
| `contains`   | クラスがメソッドを含む、モジュールが関数を含む |
| `inherits`   | クラスが他のクラスを継承する                   |
| `implements` | クラスがインターフェースを実装する             |
| `overrides`  | メソッドが親メソッドをオーバーライドする       |
| `references` | シンボルが他のシンボルを参照する               |
| `imports`    | モジュールが他のモジュールをインポートする     |
| `decorates`  | デコレータがシンボルに適用される               |

### 例

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

`[errors]` セクションには、完全にパースできなかったファイルがリストされます。

### フォーマット

各行は `-` で始まり、ファイルパスとエラーメッセージが続きます：

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

`[dict]` セクションは、`--minify` フラグが使用された場合にのみ表示されます。出力サイズを削減するために、繰り返される文字列トークンを短い辞書コードにマッピングします。

### フォーマット

各行は辞書コード（`d0`、`d1` など）を展開された値にマッピングします：

```
- d0: async=false
- d1: callable/method
- d2: exported
```

ファイルの残りの部分では、これらのコードが完全な値の代わりに使用されます。

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
S3 --returns--> S2

[errors]
- utils/broken.go: syntax error at line 5
```
