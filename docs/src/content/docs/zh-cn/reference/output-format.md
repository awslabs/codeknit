---
title: 输出格式参考
description: codeknit 使用的 .skt 输出格式的完整参考。
---

`.skt`（skeleton）格式是一种紧凑、人类可读的文本格式，由 `codeknit` 用于表示提取的代码结构。它以最小形式包含符号、关系和元数据，适合 LLM 消费和结构分析。

`.skt` 文件分为多个部分。每个部分以方括号中的标题开头。各部分可以按任意顺序出现，但 `[symbols]` 通常位于首位。

## [symbols]

`[symbols]` 部分列出按源文件分组的所有提取的符号。每个文件以 `##` 标题开头，后跟文件路径。

### 行格式

每个符号在单独一行中表示，结构如下：

```
ShortID category/kind Lstart-Lend signature {properties}
```

### 字段

- **ShortID**：分配给每个符号的顺序标识符（例如 `S1`、`S2`、`S3`）。在边和其他部分中用作引用。
- **category/kind**：用斜杠分隔的对，表示符号的类别和具体种类。
- **Lstart-Lend**：符号在源文件中定义的行范围（例如 `L10-L15`）。
- **signature**：符号的名称和类型信息。格式取决于符号：
  - `name` — 适用于类型、值、模块
  - `name(params)` — 适用于无返回类型的可调用对象
  - `name(params) -> returnType` — 适用于有返回类型的可调用对象
- **{properties}**：用花括号括起来的可选元数据。多个属性用逗号分隔。

### 参数

- 在无类型语言中：`paramName`
- 在有类型语言中：`paramName: type`
- 匹配已知符号的类型引用会被替换为其 ShortID（例如 `config: S5` 而不是 `config: Config`）。

### 属性

常见属性包括：

- `async`：`true` 或 `false`
- `exported`：`true` 或 `false`
- `static`：如果符号是静态的，则出现
- `visibility=public|private|protected`
- `receiver=*TypeName`：适用于方法，表示接收器类型

### 符号类别和种类

| 类别       | 种类                          | 示例                               |
| ---------- | ------------------------------ | ---------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

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

`[edges]` 部分使用 ShortID 定义符号之间的关系。

### 行格式

```
FromID --kind--> ToID1, ToID2
```

多个目标 ID 用逗号分隔。每行表示一个单向关系。

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

`[dict]` 部分仅在使用 `--minify` 标志时出现。它将短字典代码映射到重复的字符串标记，以减小输出大小。

### 格式

每行将字典代码（`d0`、`d1` 等）映射到其展开值：

```
- d0: async=false
- d1: callable/method
- d2: exported
```

在文件的其余部分中，这些代码替换其完整值。

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
S3 --returns--> S2

[errors]
- utils/broken.go: syntax error at line 5
```