---
title: คู่มืออ้างอิงรูปแบบเอาต์พุต
description: คู่มืออ้างอิงฉบับสมบูรณ์สำหรับรูปแบบเอาต์พุต .skt ที่ใช้โดย codeknit
---

รูปแบบ `.skt` (skeleton) เป็นรูปแบบข้อความที่กระชับ อ่านง่าย ที่ `codeknit` ใช้เพื่อแสดงโครงสร้างโค้ดที่ถูกดึงออกมา ประกอบด้วยสัญลักษณ์ ความสัมพันธ์ และข้อมูลเมตาในรูปแบบที่เล็กที่สุด เหมาะสำหรับการใช้งานโดย LLM และการวิเคราะห์โครงสร้าง

ไฟล์ `.skt` ถูกแบ่งออกเป็นส่วนๆ แต่ละส่วนเริ่มต้นด้วยส่วนหัวในวงเล็บก้ามปู ส่วนต่างๆ อาจปรากฏในลำดับใดก็ได้ แม้ว่า `[symbols]` มักจะปรากฏเป็นส่วนแรก

## [symbols]

ส่วน `[symbols]` จะแสดงรายการสัญลักษณ์ทั้งหมดที่ถูกดึงออกมา โดยจัดกลุ่มตามไฟล์ต้นทาง แต่ละไฟล์จะถูกแนะนำด้วยส่วนหัว `##` ตามด้วยพาธของไฟล์

### รูปแบบบรรทัด

แต่ละสัญลักษณ์จะแสดงในบรรทัดเดียวด้วยโครงสร้างดังนี้:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### ฟิลด์

- **ShortID**: ตัวระบุตามลำดับที่กำหนดให้กับแต่ละสัญลักษณ์ (เช่น `S1`, `S2`, `S3`) ใช้เป็นการอ้างอิงในเส้นเชื่อมและส่วนอื่นๆ
- **category/kind**: คู่ที่คั่นด้วยเครื่องหมายทับระบุหมวดหมู่และประเภทเฉพาะของสัญลักษณ์
- **Lstart-Lend**: ช่วงบรรทัดในไฟล์ต้นทางที่สัญลักษณ์ถูกกำหนด (เช่น `L10-L15`)
- **signature**: ชื่อและข้อมูลประเภทของสัญลักษณ์ รูปแบบขึ้นอยู่กับสัญลักษณ์:
  - `name` — สำหรับประเภท ค่า โมดูล
  - `name(params)` — สำหรับ callable ที่ไม่มีประเภทการคืนกลับ
  - `name(params) -> returnType` — สำหรับ callable ที่มีประเภทการคืนกลับ
- **{properties}**: ข้อมูลเมตาเสริมที่อยู่ในวงเล็บปีกกา คุณสมบัติหลายรายการคั่นด้วยเครื่องหมายจุลภาค

### พารามิเตอร์

- ในภาษาที่ไม่มีการกำหนดประเภท: `paramName`
- ในภาษาที่มีการกำหนดประเภท: `paramName: type`
- การอ้างอิงประเภทที่ตรงกับสัญลักษณ์ที่รู้จักจะถูกแทนที่ด้วย ShortID ของสัญลักษณ์นั้น (เช่น `config: S5` แทนที่ `config: Config`)

### คุณสมบัติ

คุณสมบัติทั่วไป ได้แก่:

- `async`: `true` หรือ `false`
- `exported`: `true` หรือ `false`
- `static`: มีอยู่หากสัญลักษณ์เป็น static
- `visibility=public|private|protected`
- `receiver=*TypeName`: สำหรับเมธอด ระบุประเภทของ receiver

### หมวดหมู่และประเภทของสัญลักษณ์

| หมวดหมู่     | ประเภท                                      | ตัวอย่าง                                           |
| ------------ | ------------------------------------------- | -------------------------------------------------- |
| `callable`   | function, method, constructor               | `callable/function`, `callable/method`             |
| `type`       | class, interface, struct, enum              | `type/class`, `type/interface`                     |
| `value`      | variable, constant, field                   | `value/variable`, `value/constant`                 |
| `module`     | package, namespace                          | `module/package`                                   |
| `meta`       | type parameters, files, includes, metadata  | `meta/type_parameter`, `meta/file`, `meta/include` |

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

ส่วน `[edges]` บันทึกความสัมพันธ์เมื่อเปิดใช้งาน `--edges` จุดปลายที่รู้จักจะใช้ ShortID ของตนเอง เป้าหมายที่ไม่สามารถแก้ไขเป็นสัญลักษณ์ได้อาจปรากฏเป็นชื่อ

### รูปแบบบรรทัด

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

การละสถานะหมายความว่าตัววิเคราะห์สามารถแก้ไขความสัมพันธ์ได้ `unresolved` หมายความว่าไม่สามารถระบุการพึ่งพาได้ `ambiguous` หมายความว่ายังมีการตีความที่แข่งขันกันอยู่ Candidate ID ระบุเป้าหมายที่เป็นไปได้ ไม่ใช่การเชื่อมต่อที่ยืนยัน การแก้ไขจะขึ้นอยู่กับไวยากรณ์และขอบเขต ไม่ใช่การรับประกันจากคอมไพเลอร์

เป้าหมายจะถูกจัดกลุ่มเฉพาะเมื่อแหล่งที่มา ชนิด การแก้ไข และชุดผู้สมัครตรงกัน ชื่อที่ไม่มีสัญลักษณ์จะสัมพันธ์กับไฟล์ต้นทาง รหัสประจำตัวที่ซับซ้อนจะใช้ ID เต็มที่อยู่ในเครื่องหมายคำพูด โดยมีการ escape เครื่องหมายคำพูด แบ็กสแลช และการขึ้นบรรทัดใหม่ ชื่อตามตัวอักษรที่คล้ายกับ ShortID ต้องอยู่ในเครื่องหมายคำพูดเป็น ID เต็ม เช่น `"app.go::S2"` สัญลักษณ์อาจถูกกำหนดในไฟล์หรือชิ้นส่วนเอาต์พุตในภายหลัง

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### จุดปลายไฟล์และ include

ไฟล์ include ของ C/C++ มีสัญลักษณ์เมตาที่ชัดเจน การสะกดชื่อส่วนหัวจะปรากฏครั้งเดียวในรายการสัญลักษณ์ และเส้นเชื่อมจะใช้ ShortID แม้ว่าการแก้ไขส่วนหัวจะยังไม่ได้รับการแก้ไข:

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

สัญลักษณ์ include บันทึกคำสั่งที่เขียน ไม่ใช่การประกาศส่วนหัวที่ถูกแก้ไข การ include ซ้ำภายในไฟล์จะใช้จุดปลายร่วมกัน ข้อมูลเมตาของไฟล์/include จะถูกยกเว้นจากการจัดอันดับการพึ่งพาและการค้นหา dead code

ความสัมพันธ์ที่ไม่แน่นอนจะถูกเก็บรักษาไว้ในเอาต์พุตการแยกวิเคราะห์ แต่จะถูกยกเว้นจากการวัดการพึ่งพา รายงานกราฟจะแสดงจำนวนความสัมพันธ์ที่ถูกยกเว้น ผู้ชม HTML จะแสดงจำนวนโดยไม่วาดเส้นเชื่อมที่ไม่แน่นอนแต่ละรายการ

เมื่ออัปเกรดเป็น `0.5.0` ให้สร้างชิ้นส่วนเอาต์พุตทั้งหมดใหม่ สัญลักษณ์ ID และผลลัพธ์ความสัมพันธ์อาจเปลี่ยนแปลงได้ ดังนั้นชิ้นส่วนจากการรันต่างๆ จึงต้องไม่ถูกผสมกัน

### ประเภทของเส้นเชื่อม

| ประเภท       | ความหมาย                                         |
| ------------- | ------------------------------------------------ |
| `calls`       | การเรียกใช้ฟังก์ชัน/เมธอด                       |
| `contains`    | คลาสมีเมธอด โมดูลมีฟังก์ชัน                     |
| `inherits`    | คลาสขยายคลาสอื่น                                 |
| `implements`  | คลาส implement อินเทอร์เฟซ                      |
| `overrides`   | เมธอด override เมธอดของคลาสแม่                 |
| `references`  | สัญลักษณ์อ้างอิงถึงสัญลักษณ์อื่น                  |
| `imports`     | โมดูล import โมดูลอื่น                           |
| `decorates`   | decorator ถูกนำไปใช้กับสัญลักษณ์                 |

### ตัวอย่าง

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

ส่วน `[errors]` จะแสดงรายการไฟล์ที่ไม่สามารถแยกวิเคราะห์ได้อย่างสมบูรณ์

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

ส่วน `[dict]` จะปรากฏเมื่อใช้แฟล็ก `--minify` เท่านั้น จะจับคู่รหัสพจนานุกรมสั้นกับโทเค็นสตริงที่ซ้ำกันเพื่อลดขนาดเอาต์พุต

### รูปแบบ

แต่ละบรรทัดจะจับคู่รหัสพจนานุกรม (`d0`, `d1` ฯลฯ) กับค่าที่ขยาย:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

ในส่วนอื่นๆ ของไฟล์ รหัสเหล่านี้จะแทนที่ค่าที่สมบูรณ์

ประเภทของเส้นเชื่อมสามารถใช้รหัสพจนานุกรมได้ ในขณะที่ป้ายสถานะยังคงอ่านได้: `S1 --d2[unresolved]--> S3` ขยาย `d2` โดยใช้พจนานุกรมและรักษาสถานะ unresolved ไว้

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
S3 --references--> S2

[errors]
- utils/broken.go: syntax error at line 5
```

## ความสัมพันธ์ JSON

ใช้ JSON เมื่อสคริปต์หรือการผสานรวมต้องการเอาต์พุตที่มีโครงสร้าง:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

แต่ละเส้นเชื่อมประกอบด้วย `from`, `to` และ `kind` ID แหล่งที่มาที่รู้จักจะปรากฏใน `from_short` ID เป้าหมายที่ถูกแก้ไขจะปรากฏใน `to_short` เส้นเชื่อมที่ไม่แน่นอนจะเพิ่ม `resolution` และเมื่อมีให้ `candidates` ที่มี ID สัญลักษณ์เต็ม พวกเขาจะละเว้น `to_short` แม้ว่าเป้าหมายจะมีสัญลักษณ์เมตา include

ตัวอย่างเช่น การเรียกที่คลุมเครือสามารถแสดงเป็น:

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

JSON และ SKT รักษาความไม่แน่นอนเดียวกันด้วยการเข้ารหัสที่แตกต่างกัน ดู [การสนับสนุนภาษาและความเที่ยงตรงของ API](../language-support/) สำหรับกฎการแก้ไขและข้อจำกัดที่ทราบ