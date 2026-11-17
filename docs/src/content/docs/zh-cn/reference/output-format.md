---
title: 输出格式参考
description: codeknit 使用的 .skt 输出格式的完整参考。
---

本页面描述了 `.skt` 输出的语法以及等效的 JSON 输出。有关关系解析规则的详细信息，请参阅[语言支持和 API 保真度](../language-support/)。

## 概述

`.skt` 文件是由多个部分组成的紧凑文本文件。每个部分以方括号中的标题行开头。

| 部分        | 出现条件                                       | 内容                                          |
| ----------- | ---------------------------------------------- | --------------------------------------------- |
| `[symbols]` | 始终存在                                       | 按源文件分组的符号定义                        |
| `[edges]`   | 设置了 `--edges`                               | 符号之间的关系                                |
| `[errors]`  | 发生扫描或解析诊断                             | 无法完全处理的文件                            |
| `[dict]`    | 设置了 `--minify`                              | 将短代码映射到重复标记的字典                  |

在目录输出模式下，`[errors]` 和 `[dict]` 可以写入辅助文件 `warnings.skt` 和 `dict.skt`。在 `inline` 模式下，它们会出现在与其他部分相同的输出中。

`--format json` 生成包含相同信息的 JSON 文档。请参阅 [JSON 输出](#json-输出)。

## [symbols]

符号按源文件分组。每个组以 `## <path>` 标题行开头，后跟每个符号一行。

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### 符号行

```text
ShortID category/kind Lstart-Lend signature {properties}
```

| 字段          | 描述                                                                                               |
| ------------- | -------------------------------------------------------------------------------------------------- |
| `ShortID`     | 顺序标识符：`S1`、`S2`、`S3` 等。边通过该值引用符号。                                              |
| `category/kind` | 符号的分类。请参阅[类别和种类](#类别和种类)。                                                      |
| `Lstart-Lend` | 符号在源文件中的起始和结束行。                                                                     |
| `signature`   | 符号名称，对于可调用对象，包括参数和返回类型。请参阅[签名格式](#签名格式)。                       |
| `{properties}` | 大括号中的元数据。请参阅[属性](#属性)。                                                            |

符号行保持在一行内。签名内的换行符转义为 `\n` 和 `\r`。

符号 ID 可能会在提取发生变化时改变。请重新生成所有输出块，而不是混合来自不同运行的块。这在升级到 0.5.0 时尤其适用，因为符号 ID 和关系解析结果可能会发生变化。

### 类别和种类

| 类别      | 种类                                      | 示例                                           |
| --------- | ----------------------------------------- | ---------------------------------------------- |
| `callable` | function, method, constructor             | `callable/function`, `callable/method`        |
| `type`    | class, interface, struct, enum            | `type/class`, `type/interface`                 |
| `value`   | variable, constant, field                 | `value/variable`, `value/constant`             |
| `module`  | package, namespace                        | `module/package`                               |
| `meta`    | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### 签名格式

| 格式                         | 用于                                  |
| ---------------------------- | ------------------------------------- |
| `name`                       | 类型、值和模块                        |
| `name(params)`               | 无返回类型的可调用对象                |
| `name(params) -> returnType` | 有返回类型的可调用对象                |

参数的写法如下：

| 格式          | 用于                  |
| ------------- | --------------------- |
| `name`        | 无类型的语言          |
| `name: type`  | 有类型的语言          |

当类型引用与已知符号匹配时，类型名称会被替换为该符号的短 ID。例如，`-> *S2` 引用指向符号 `S2` 的指针。

### 属性

属性出现在签名后的大括号中，以逗号分隔。空大括号（`{}`）表示符号没有属性。

| 属性形式        | 示例                 |
| --------------- | -------------------- |
| 标志            | `exported`, `async`, `static` |
| `key=value`     | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

### 示例

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

`[edges]` 部分列出了符号之间的关系。仅在使用 `--edges` 请求边时才会写入。

### 边行

```text
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

| 部分                     | 描述                                                                           |
| ------------------------ | ------------------------------------------------------------------------------ |
| `FromID`                 | 源端点。                                                                       |
| `kind`                   | 关系种类。请参阅[边种类](#边种类)。                                            |
| `[unresolved]`, `[ambiguous]` | 可选的解析标记。请参阅[解析标记](#解析标记)。                              |
| `ToID1, ToID2`           | 一个或多个目标，以逗号分隔。                                                   |
| `[candidates=...]`       | 模糊边的可能目标，以括号内逗号分隔的列表形式给出。                            |

仅当源、种类、解析标记和候选集完全相同时，目标才会共享一行。边引用的定义可能出现在后续文件或输出块中。在解释边之前，请加载所有块。

### 解析标记

| 标记         | 含义                                                                              |
| ------------ | --------------------------------------------------------------------------------- |
| 无           | 分析器解析了该关系。                                                              |
| `[unresolved]` | 分析器无法建立依赖关系。                                                          |
| `[ambiguous]` | 存在多种可能的解释。候选 ID 列出了可能的目标。                                    |

缺少标记并不意味着编译器保证。候选 ID 标识可能的目标，而非确认的连接。不确定性始终以 SKT 语法表示，而非嵌入的 JSON。

### 端点形式

| 形式                      | 使用时机                                                                                         | 示例                     |
| ------------------------- | ------------------------------------------------------------------------------------------------ | ------------------------ |
| Short ID                  | 端点或候选是已知符号。这包括未解析边的端点和包含符号。                                           | `S4`, `S8`               |
| 裸名称                    | 目标没有符号。名称相对于源文件。                                                                  | `Refresh`, `string`      |
| 引用的完整 ID             | 标识复杂，或字面名称类似于短 ID。                                                                 | `"app.go::S2"`           |

引用的完整 ID 转义引号、反斜杠和换行符。诸如 `S2` 之类的字面名称必须写成引用的完整 ID，以免被读取为短 ID。

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### 边种类

| 种类         | 含义                                         |
| ------------ | -------------------------------------------- |
| `calls`      | 函数/方法调用                                |
| `contains`   | 类包含方法，模块包含函数                     |
| `inherits`   | 类继承另一个类                               |
| `implements` | 类实现接口                                   |
| `overrides`  | 方法覆盖父类方法                             |
| `references` | 符号引用另一个符号                           |
| `imports`    | 模块导入另一个模块                           |
| `decorates`  | 装饰器应用于符号                             |

返回的可调用对象和回调参数记录为 `references`，而非单独的种类。

### C 和 C++ 文件和包含符号

C 和 C++ 源代码会生成 `meta/file` 和 `meta/include` 符号，因此包含关系具有短 ID。每个头文件拼写在符号列表中仅出现一次，边使用短 ID。文件内重复的包含共享一个端点。

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

包含符号记录的是指令的原始写法。它不代表已解析的头文件声明。

### 对分析和可视化的影响

- 不确定的边和文件/包含元数据不参与依赖度量。报告会显示排除的关系数量。
- HTML 图绘制已解析的链接，并显示不确定关系的计数。它不会绘制单个不确定的边。

### 示例

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

`[errors]` 部分列出了未完全处理的文件。仅在存在诊断信息时才会出现。

```text
- <path>: <message>
```

```skt
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict]

`[dict]` 部分仅在压缩输出（`--minify`）中出现。它将短代码映射到重复的标记。

| 部分       | 描述                                                                 |
| ---------- | -------------------------------------------------------------------- |
| `d0`, `d1`, … | 字典代码。                                                        |
| 条目行     | `- <code>: <token>`                                                  |

要读取压缩输出，请将每个代码替换为其字典值。代码可以替换 `[symbols]` 中的类别/种类对和属性。它们还可以替换 `[edges]` 中的关系种类。不确定性标记保持可读。例如，`S1 --d2[unresolved]--> S3` 是一个未解析的边，其种类是 `d2` 的值。

### 压缩示例

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

此处 `S2` 是一个 `callable/method`，具有属性 `async=false`。

## 完整示例

以下输出包含 `[symbols]`、`[edges]` 和 `[errors]`。它未被压缩，因此没有 `[dict]` 部分；请参阅[压缩示例](#压缩示例)。

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

## JSON 输出

`--format json`