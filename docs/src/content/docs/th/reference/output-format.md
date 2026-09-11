---
title: คู่มืออ้างอิงรูปแบบเอาต์พุต
description: คู่มืออ้างอิงฉบับสมบูรณ์สำหรับรูปแบบเอาต์พุต .skt ที่ใช้โดย codeknit
---

รูปแบบ `.skt` (skeleton) เป็นรูปแบบข้อความที่กระชับและอ่านได้ง่ายที่ใช้โดย `codeknit` เพื่อแสดงโครงสร้างโค้ดที่ถูกดึงออกมา ประกอบด้วยสัญลักษณ์ ความสัมพันธ์ และข้อมูลเมตาในรูปแบบที่เล็กที่สุด เหมาะสำหรับการใช้งานโดย LLM และการวิเคราะห์โครงสร้าง

ไฟล์ `.skt` จะแบ่งออกเป็นส่วนต่าง ๆ แต่ละส่วนจะเริ่มต้นด้วยส่วนหัวในวงเล็บเหลี่ยม ส่วนต่าง ๆ อาจปรากฏในลำดับใดก็ได้ แม้ว่า `[symbols]` มักจะปรากฏเป็นส่วนแรก

## [symbols]

ส่วน `[symbols]` จะแสดงรายการสัญลักษณ์ทั้งหมดที่ถูกดึงออกมา โดยจัดกลุ่มตามไฟล์ต้นทาง แต่ละไฟล์จะเริ่มต้นด้วยส่วนหัว `##` ตามด้วยพาธของไฟล์

### รูปแบบบรรทัด

แต่ละสัญลักษณ์จะแสดงในบรรทัดเดียวด้วยโครงสร้างดังนี้:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### ฟิลด์

- **ShortID**: ตัวระบุตามลำดับที่กำหนดให้กับแต่ละสัญลักษณ์ (เช่น `S1`, `S2`, `S3`) ใช้เป็นการอ้างอิงในเส้นเชื่อมและส่วนอื่น ๆ
- **category/kind**: คู่ที่คั่นด้วยเครื่องหมายทับระบุหมวดหมู่และประเภทเฉพาะของสัญลักษณ์
- **Lstart-Lend**: ช่วงบรรทัดในไฟล์ต้นทางที่สัญลักษณ์ถูกกำหนด (เช่น `L10-L15`)
- **signature**: ชื่อและข้อมูลประเภทของสัญลักษณ์ รูปแบบขึ้นอยู่กับสัญลักษณ์:
  - `name` — สำหรับประเภท ค่า โมดูล
  - `name(params)` — สำหรับ callable ที่ไม่มีประเภทการคืนกลับ
  - `name(params) -> returnType` — สำหรับ callable ที่มีประเภทการคืนกลับ
- **{properties}**: ข้อมูลเมตาเสริมที่อยู่ในวงเล็บปีกกา คุณสมบัติหลายรายการจะคั่นด้วยเครื่องหมายจุลภาค

### พารามิเตอร์

- ในภาษาที่ไม่มีการกำหนดประเภท: `paramName`
- ในภาษาที่มีการกำหนดประเภท: `paramName: type`
- การอ้างอิงประเภทที่ตรงกับสัญลักษณ์ที่รู้จักจะถูกแทนที่ด้วย ShortID ของสัญลักษณ์นั้น (เช่น `config: S5` แทนที่ `config: Config`)

### คุณสมบัติ

คุณสมบัติทั่วไป ได้แก่:

- `async`: `true` หรือ `false`
- `exported`: `true` หรือ `false`
- `static`: มีอยู่หากสัญลักษณ์เป็นแบบ static
- `visibility=public|private|protected`
- `receiver=*TypeName`: สำหรับเมธอด ระบุประเภทของ receiver

### หมวดหมู่และประเภทของสัญลักษณ์

| หมวดหมู่     | ประเภท                          | ตัวอย่าง                               |
| ------------ | -------------------------------- | -------------------------------------- |
| `callable`   | function, method, constructor    | `callable/function`, `callable/method` |
| `type`       | class, interface, struct, enum   | `type/class`, `type/interface`         |
| `value`      | variable, constant, field        | `value/variable`, `value/constant`     |
| `module`     | package, namespace               | `module/package`                       |
| `meta`       | type parameters, metadata        | `meta/type_parameter`                  |

### ตัวอย่าง

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

ส่วน `[edges]` กำหนดความสัมพันธ์ระหว่างสัญลักษณ์โดยใช้ ShortID ของสัญลักษณ์

### รูปแบบบรรทัด

```
FromID --kind--> ToID1, ToID2
```

ID ของเป้าหมายหลายรายการจะคั่นด้วยเครื่องหมายจุลภาค แต่ละบรรทัดแสดงถึงความสัมพันธ์แบบมีทิศทางหนึ่งความสัมพันธ์

### ประเภทของเส้นเชื่อม

| ประเภท        | ความหมาย                                         |
| -------------- | ------------------------------------------------ |
| `calls`        | การเรียกใช้ฟังก์ชัน/เมธอด                        |
| `contains`     | คลาสมีเมธอด โมดูลมีฟังก์ชัน                     |
| `inherits`     | คลาสสืบทอดจากคลาสอื่น                            |
| `implements`   | คลาสดำเนินการอินเทอร์เฟซ                         |
| `overrides`    | เมธอดแทนที่เมธอดของคลาสแม่                      |
| `references`   | สัญลักษณ์อ้างอิงถึงสัญลักษณ์อื่น                   |
| `imports`      | โมดูลนำเข้าโมดูลอื่น                              |
| `decorates`    | เดคอเรเตอร์ถูกนำไปใช้กับสัญลักษณ์                 |

### ตัวอย่าง

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

ส่วน `[errors]` แสดงรายการไฟล์ที่ไม่สามารถแยกวิเคราะห์ได้อย่างสมบูรณ์

### รูปแบบ

แต่ละบรรทัดเริ่มต้นด้วย `-` ตามด้วยพาธของไฟล์และข้อความแสดงข้อผิดพลาด:

```
- path/to/file.go: syntax error at line 42
```

### ตัวอย่าง

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

ส่วน `[dict]` จะปรากฏเมื่อใช้แฟล็ก `--minify` เท่านั้น จะทำการแมปโค้ดพจนานุกรมสั้น ๆ กับโทเค็นสตริงที่ซ้ำกันเพื่อลดขนาดเอาต์พุต

### รูปแบบ

แต่ละบรรทัดจะแมปโค้ดพจนานุกรม (`d0`, `d1` ฯลฯ) กับค่าที่ขยายออก:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

ในส่วนอื่น ๆ ของไฟล์ โค้ดเหล่านี้จะแทนที่ค่าที่สมบูรณ์

### ตัวอย่าง

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

## ตัวอย่างเต็มรูปแบบ

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