---
title: 出力フォーマットリファレンス
description: codeknitで使用される.skt出力フォーマットの完全なリファレンスです。
---

このページでは、`.skt`出力の文法と同等のJSON出力について説明します。関係性がどのように解決されるかについてのルールは、[言語サポートとAPI忠実度](../language-support/)を参照してください。

## 概要

`.skt`ファイルはセクションで構成されるコンパクトなテキストファイルです。各セクションは角括弧で囲まれたヘッダ行で始まります。

| セクション    | 存在する条件                                      | 内容                                          |
| ------------- | ------------------------------------------------- | --------------------------------------------- |
| `[symbols]`   | 常に                                              | ソースファイルごとにグループ化されたシンボル定義 |
| `[edges]`     | `--edges`が設定されている場合                     | シンボル間の関係性                            |
| `[errors]`    | スキャンまたはパースの診断が発生した場合          | 完全に処理できなかったファイル                |
| `[dict]`      | `--minify`が設定されている場合                    | 繰り返しトークンに短縮コードをマッピングする辞書 |

ディレクトリ出力モードでは、`[errors]`と`[dict]`は補助ファイル`warnings.skt`と`dict.skt`に書き込まれることがあります。`inline`モードでは、これらのセクションは他のセクションと同じ出力に表示されます。

`--format json`は同じ情報を含むJSONドキュメントを生成します。詳細は[JSON出力](#json出力)を参照してください。

## [symbols]

シンボルはソースファイルごとにグループ化されます。各グループは`## <path>`ヘッダ行で始まり、その後にシンボルごとに1行が続きます。

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### シンボル行

```text
ShortID カテゴリ/種類 L開始-終了 シグネチャ {プロパティ}
```

| フィールド          | 説明                                                                                               |
| ------------------- | -------------------------------------------------------------------------------------------------- |
| `ShortID`           | 連続した識別子: `S1`、`S2`、`S3`など。エッジはこの値でシンボルを参照します。                       |
| `カテゴリ/種類`     | シンボルの分類。詳細は[カテゴリと種類](#カテゴリと種類)を参照してください。                        |
| `L開始-終了`        | ソースファイル内のシンボルの最初と最後の行。                                                       |
| `シグネチャ`        | シンボル名。呼び出し可能なシンボルにはパラメータと戻り値の型が含まれます。詳細は[シグネチャ形式](#シグネチャ形式)を参照してください。 |
| `{プロパティ}`      | 中括弧内のメタデータ。詳細は[プロパティ](#プロパティ)を参照してください。                          |

シンボル行は1つの物理行に収まります。シグネチャ内の改行は`\n`と`\r`としてエスケープされます。

シンボルIDは抽出が変更されると変わる可能性があります。異なる実行からのチャンクを混ぜるのではなく、すべての出力チャンクを一緒に再生成してください。これは特に0.5.0へのアップグレード時には重要です。シンボルIDと関係性解決の結果が変わる可能性があるためです。

### カテゴリと種類

| カテゴリ     | 種類                                      | 例                                           |
| ------------ | ----------------------------------------- | -------------------------------------------- |
| `callable`   | function, method, constructor             | `callable/function`、`callable/method`       |
| `type`       | class, interface, struct, enum            | `type/class`、`type/interface`               |
| `value`      | variable, constant, field                 | `value/variable`、`value/constant`           |
| `module`     | package, namespace                        | `module/package`                             |
| `meta`       | type parameters, files, includes, metadata | `meta/type_parameter`、`meta/file`、`meta/include` |

### シグネチャ形式

| 形式                         | 使用対象                                  |
| ---------------------------- | ----------------------------------------- |
| `name`                       | 型、値、モジュール                        |
| `name(params)`               | 戻り値の型がない呼び出し可能なシンボル    |
| `name(params) -> returnType` | 戻り値の型がある呼び出し可能なシンボル    |

パラメータは次のように記述されます:

| 形式          | 使用対象                  |
| ------------- | ------------------------- |
| `name`        | 型がない言語              |
| `name: type`  | 型がある言語              |

型参照が既知のシンボルと一致する場合、型名はそのシンボルのShortIDに置き換えられます。例えば、`-> *S2`はシンボル`S2`へのポインタを参照します。

### プロパティ

プロパティはシグネチャの後に中括弧で表示され、カンマで区切られます。空の中括弧(`{}`)はシンボルにプロパティがないことを意味します。

| プロパティ形式        | 例                     |
| -------------------- | ----------------------- |
| フラグ               | `exported`、`async`、`static` |
| `キー=値`            | `visibility=public`、`receiver=*AuthService`、`exported=false`、`resolution=unresolved` |

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

`[edges]`セクションはシンボル間の関係性をリストします。このセクションは`--edges`が指定された場合にのみ書き込まれます。

### エッジ行

```text
FromID --種類--> ToID1, ToID2
FromID --種類[unresolved]--> Target1, Target2
FromID --種類[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

| 部分                     | 説明                                                                           |
| ------------------------ | ------------------------------------------------------------------------------ |
| `FromID`                 | ソースエンドポイント。                                                        |
| `種類`                   | 関係性の種類。詳細は[エッジの種類](#エッジの種類)を参照してください。         |
| `[unresolved]`、`[ambiguous]` | オプションの解決マーカー。詳細は[解決マーカー](#解決マーカー)を参照してください。 |
| `ToID1, ToID2`           | カンマで区切られた1つ以上のターゲット。                                       |
| `[candidates=...]`       | あいまいなエッジの可能性のあるターゲット。括弧内にカンマ区切りのリストとして表示されます。 |

ソース、種類、解決マーカー、候補セットが同一の場合にのみ、ターゲットは1行を共有します。エッジによって参照される定義は、後のファイルや出力チャンクで発生することがあります。すべてのチャンクを読み込んでからエッジを解釈してください。

### 解決マーカー

| マーカー         | 意味                                                                              |
| -------------- | --------------------------------------------------------------------------------- |
| なし           | アナライザが関係性を解決しました。                                                |
| `[unresolved]` | アナライザが依存関係を確立できませんでした。                                      |
| `[ambiguous]`  | 複数の解釈が可能なまま残っています。候補IDは可能性のあるターゲットをリストします。 |

マーカーがないことはコンパイラの保証ではありません。候補IDは可能性のあるターゲットを識別し、確認された接続ではありません。不確実性は常にこのSKT文法で記述され、埋め込みJSONとしては記述されません。

### エンドポイント形式

| 形式                      | 使用時機                                                                                         | 例                     |
| ------------------------- | ------------------------------------------------------------------------------------------------ | ----------------------- |
| Short ID                  | エンドポイントまたは候補が既知のシンボルである場合。これには未解決のエッジのエンドポイントやインクルードシンボルが含まれます。 | `S4`、`S8`       |
| ベア名                    | ターゲットにシンボルがない場合。名前はソースファイルに対して相対的です。                        | `Refresh`、`string`     |
| 引用されたフルID          | IDが複雑であるか、リテラル名がShort IDに似ている場合。                                          | `"app.go::S2"`          |

引用されたフルIDは引用符、バックスラッシュ、改行をエスケープします。`S2`のようなリテラル名はShort IDとして読み取られないように引用されたフルIDとして記述する必要があります。

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### エッジの種類

| 種類         | 意味                                         |
| ------------ | ------------------------------------------- |
| `calls`      | 関数/メソッドの呼び出し                     |
| `contains`   | クラスがメソッドを含む、モジュールが関数を含む |
| `inherits`   | クラスが他のクラスを継承する                 |
| `implements` | クラスがインターフェースを実装する           |
| `overrides`  | メソッドが親をオーバーライドする             |
| `references` | シンボルが他のシンボルを参照する             |
| `imports`    | モジュールが他のモジュールをインポートする   |
| `decorates`  | デコレータがシンボルに適用される             |

返された呼び出し可能なシンボルやコールバック引数は`references`として記録され、別の種類としては記録されません。

### CおよびC++のファイルとインクルードシンボル

CおよびC++のソースは`meta/file`および`meta/include`シンボルを生成するため、インクルード関係性にはShort IDが使用されます。各ヘッダのスペルはシンボルリストに一度だけ表示され、エッジはShort IDを使用します。ファイル内の繰り返しインクルードは1つのエンドポイントを共有します。

```skt
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

インクルードシンボルは書かれたディレクティブを記録します。これは解決されたヘッダ宣言を表すものではありません。

### 分析と可視化への影響

- 不確実なエッジおよびファイル/インクルードのメタデータは依存性メトリクスに寄与しません。レポートには除外された関係性の数が表示されます。
- HTMLグラフは解決されたリンクを描画し、不確実な関係性の数を表示します。個々の不確実なエッジは描画されません。

### 例

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

`[errors]`セクションは完全に処理できなかったファイルをリストします。このセクションは診断が存在する場合にのみ表示されます。

```text
- <path>: <message>
```

```skt
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict]

`[dict]`セクションは圧縮出力(`--minify`)でのみ表示されます。短縮コードを繰り返しトークンにマッピングします。

| 部分       | 説明                                                                 |
| ---------- | -------------------------------------------------------------------- |
| `d0`、`d1`、… | 辞書コード。                                                        |
| エントリ行 | `- <code>: <token>`                                                  |

圧縮出力を読むには、各コードを辞書の値で置き換えます。コードは`[symbols]`内のカテゴリ/種類のペアやプロパティを置き換えることができます。また、`[edges]`内の関係性の種類も置き換えることができます。不確実性マーカーは読み取り可能なままです。例えば、`S1 --d2[unresolved]--> S3`は種類が`d2`の値である未解決のエッジです。

### 圧縮例

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

ここで`S2`は`callable/method`であり、プロパティ`async=false`を持っています。

## 完全な例

以下の出力には`[symbols]`、`[edges]`、および`[errors]`が含まれています。圧縮されていないため、`[dict]`セクションはありません。圧縮例は[圧縮例](#圧縮例)を参照してください。

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
## pkg/util/hash.go
S6 callable/function L3-L5 Hash(data: string) -> string {exported}
## pkg/util/hash_alt.go
S7 callable/function L3-L5 Hash(data: string) -> string {exported}

[edges]
S1 --contains--> S2, S3, S4, S5
S2 --contains--> S4
S3 --references--> S2
S4 --calls--> S5
S4 --calls[unresolved]--> Refresh
S5 --calls[ambiguous]--> Hash [candidates=S6, S7]

[errors]
- pkg/legacy/broken.go: syntax error at line 42
```

## JSON出力

`--format json`