---
title: 输出格式参考
description: codeknit 使用的 .skt 输出格式的完整参考。
---

`.skt`（skeleton）格式是一种紧凑、人类可读的文本格式，`codeknit` 用其表示提取的代码结构。它以最小形式包含符号、关系和元数据，适用于 LLM 消费和结构分析。

`.skt` 文件分为多个部分。每个部分以方括号中的标题开头。部分可以按任意顺序出现，但 `[symbols]` 通常位于首位。

## [symbols]

`[symbols]` 部分列出按源文件分组的所有提取的符号。每个文件以 `##` 标题开头，后跟文件路径。

### 行格式

每个符号在单独一行中表示，结构如下：

```
ShortID category/kind Lstart-Lend signature {properties}
```

### 字段

- **ShortID**：分配给每个符号的顺序标识符（例如 `S1`、`S2`、`S3`）。在边和其他部分中用作引用。
- **category/kind**：斜线分隔的对，表示符号的类别和具体种类。
- **Lstart-Lend**：符号在源文件中定义的行范围（例如 `L10-L15`）。
- **signature**：符号的名称和类型信息。格式取决于符号：
  - `name` — 适用于类型、值、模块
  - `name(params)` — 适用于无返回类型的可调用对象
  - `name(params) -> returnType` — 适用于有返回类型的可调用对象
- **{properties}**：大括号内的可选元数据。多个属性以逗号分隔。

### 参数

- 在无类型语言中：`paramName`
- 在有类型语言中：`paramName: type`
- 匹配已知符号的类型引用会被替换为其 ShortID（例如 `config: S5` 而不是 `config: Config`）。

### 属性

常见属性包括：

- `async`：`true` 或 `false`
- `exported`：`true` 或 `false`
- `static`：如果符号是静态的，则存在
- `visibility=public|private|protected`
- `receiver=*TypeName`：适用于方法，表示接收器类型

### 符号类别和种类

| 类别       | 种类                                      | 示例                                           |
| ---------- | ------------------------------------------ | ---------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`        |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                 |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`             |
| `module`   | package, namespace                         | `module/package`                               |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

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

`[edges]` 部分在启用 `--edges` 时记录关系。已知端点使用其 ShortID；无法解析为符号的目标可能以名称形式出现。

### 行格式

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

省略状态表示分析器已解析该关系。`unresolved` 表示无法建立依赖关系；`ambiguous` 表示存在竞争解释。候选 ID 标识可能的目标，而非确认的连接。解析基于语法和作用域，而非编译器保证。

仅在源、种类、解析和候选集匹配时，目标才会分组。无符号的名称相对于源文件。复杂标识使用带转义引号、反斜杠和换行符的引用完整 ID。类似 ShortID 的字面名称必须作为完整 ID 引用，例如 `"app.go::S2"`。符号定义可能出现在后续文件或输出块中。

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### 文件和包含端点

C/C++ 包含具有显式元数据符号。头文件拼写在符号列表中出现一次，边使用 ShortID，即使头文件解析仍未解决：

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

包含符号记录书写的指令，而非已解析的头文件声明。文件内重复包含共享一个端点。文件/包含元数据不包括在依赖排名和死代码发现中。

不确定的关系保留在解析输出中，但不包括在依赖指标中。图报告显示排除的关系计数；HTML 查看器显示其计数但不绘制单个不确定边。

升级到 `0.5.0` 时，请同时重新生成所有输出块。符号 ID 和关系结果可能会变化，因此来自不同运行的块不能混合使用。

### 边种类

| 种类         | 含义                                         |
| ------------ | -------------------------------------------- |
| `calls`      | 函数/方法调用                                |
| `contains`   | 类包含方法，模块包含函数                     |
| `inherits`   | 类继承另一个类                               |
| `implements` | 类实现接口                                   |
| `overrides`  | 方法覆盖父方法                               |
| `references` | 符号引用另一个符号                           |
| `imports`    | 模块导入另一个模块                           |
| `decorates`  | 装饰器应用于符号                             |

### 示例

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

`[errors]` 部分列出无法完全解析的文件。

### 格式

每行以 `-` 开头，后跟文件路径和错误消息：

```
- path/to/file.go: syntax error at line 42
```

### 示例

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

`[dict]` 部分仅在使用 `--minify` 标志时出现。它将短字典代码映射到重复的字符串标记，以减少输出大小。

### 格式

每行将字典代码（`d0`、`d1` 等）映射到其展开值：

```
- d0: async=false
- d1: callable/method
- d2: exported
```

在文件的其余部分，这些代码替换其完整值。

关系种类也可以使用字典代码，而状态标签保持可读：`S1 --d2[unresolved]--> S3`。使用字典展开 `d2` 并保留未解析状态。

### 示例

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

## 完整示例

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

## JSON 关系

当脚本或集成需要结构化输出时，使用 JSON：

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

每个边包括 `from`、`to` 和 `kind`。已知源 ID 出现在 `from_short` 中；已解析的目标 ID 出现在 `to_short` 中。不确定的边添加 `resolution`，并在可用时包含 `candidates`，其中包含完整符号 ID。它们省略 `to_short`，即使目标具有包含元数据符号。

例如，一个模糊调用可以表示为：

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

JSON 和 SKT 以不同的编码方式保留相同的不确定性。有关解析规则和已知限制，请参阅 [语言支持和 API 保真度](../language-support/)。