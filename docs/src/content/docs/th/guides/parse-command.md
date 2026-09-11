---
title: คำสั่ง Parse
description: ดึงข้อมูลโครงสร้างจากซอร์สโค้ดไปยังไฟล์ .skt หรือ JSON
---

คำสั่ง `codeknit parse` จะดึงข้อมูลโครงสร้างจากโค้ดเบสของคุณ — เช่น ฟังก์ชัน คลาส เมธอด ตัวแปร และความสัมพันธ์ของพวกมัน — และส่งออกในรูปแบบ `.skt` ที่กระชับโดยค่าเริ่มต้น ใช้ JSON เมื่อคุณต้องการเอาต์พุตที่อ่านได้ด้วยเครื่องสำหรับสคริปต์ การผสานรวม หรือเครื่องมือปลายน้ำ

## การใช้งานพื้นฐาน

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: เส้นทางไปยังไดเรกทอรีหรือไฟล์ที่คุณต้องการแยกวิเคราะห์
- **`[output-dir]`**: ไดเรกทอรีเอาต์พุตที่ไม่บังคับ หากไม่ระบุ จะใช้ค่าเริ่มต้นเป็น `./skeleton`

### ตัวอย่าง

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## โหมดเอาต์พุต

ใช้ `--output-mode` เพื่อควบคุมโครงสร้างของเอาต์พุต มีสามโหมดให้เลือก:

| โหมด               | คำอธิบาย                                                                                     | เหมาะสำหรับ                                      |
| ------------------ | -------------------------------------------------------------------------------------------- | ------------------------------------------------ |
| `directory-flat`   | เขียนไฟล์ `.skt` ที่แบ่งเป็นส่วนย่อย (เช่น `map_001.skt`, `map_002.skt`) ไปยังไดเรกทอรีเอาต์พุต | ✅ **โปรเจ็กต์ส่วนใหญ่** — โหมดเริ่มต้นและแนะนำ |
| `directory-tree`   | สร้างโครงสร้างไดเรกทอรีที่เหมือนกับซอร์สโค้ด สร้างไฟล์ `.skt` หนึ่งไฟล์ต่อไฟล์ซอร์ส            | การนำทางเอาต์พุตควบคู่ไปกับซอร์สโค้ด             |
| `inline`           | ส่งออกทั้งหมดไปยัง stdout                                                                   | ไฟล์เดียวหรือการส่งต่อไปยังเครื่องมืออื่น        |

> **เคล็ดลับ**: ใช้ `directory-flat` เป็นค่าเริ่มต้น เว้นแต่คุณจะทำงานกับไฟล์เดียว หลีกเลี่ยง `inline` สำหรับอินพุตขนาดใหญ่ เนื่องจากอาจทำให้หน้าต่างบริบทล้นได้

## แฟล็ก

| แฟล็ก             | ค่าเริ่มต้น        | คำอธิบาย                                                                                     |
| ----------------- | ------------------ | -------------------------------------------------------------------------------------------- |
| `--output-mode`   | `directory-flat`   | โหมดเอาต์พุต: `inline`, `directory-flat` หรือ `directory-tree`                               |
| `--format`        | `skt`              | รูปแบบเอาต์พุต: `skt` หรือ `json`                                                            |
| `--max-lines`     | `500`              | จำนวนบรรทัดสูงสุดต่อไฟล์เอาต์พุตในโหมด flat/tree                                             |
| `--collect-test`  | `false`            | รวมไฟล์ทดสอบในการวิเคราะห์                                                                   |
| `--minify`        | `false`            | เปิดใช้การบีบอัดด้วยพจนานุกรมเพื่อลดการใช้โทเค็น                                             |
| `--edges`         | `false`            | รวมส่วน `[edges]` พร้อมข้อมูลความสัมพันธ์ (การเรียกใช้, การบรรจุ ฯลฯ)                          |
| `--clean`         | `false`            | ลบไฟล์ `.skt` ที่มีอยู่ในไดเรกทอรีเอาต์พุตก่อนเขียน                                            |
| `--workers`       | `NumCPU`           | จำนวนสูงสุดของ goroutines การแยกวิเคราะห์ที่ทำงานพร้อมกัน (0 = ใช้ทุกคอร์ CPU)                |
| `--verbose`       | `false`            | แสดงข้อมูลความคืบหน้าและเวลาในระหว่างการประมวลผล                                            |

## รูปแบบการใช้งานทั่วไป

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

ตัวอย่างเอาต์พุต JSON:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## การป้องกันเอาต์พุตที่ล้าสมัย

หากไดเรกทอรีเอาต์พุตมีไฟล์ `.skt` จากการรันก่อนหน้าแล้ว `codeknit` จะปฏิเสธการเขียนเอาต์พุตใหม่เพื่อป้องกันการผสมข้อมูลที่ล้าสมัยและใหม่

หากต้องการแทนที่พฤติกรรมนี้และล้างไดเรกทอรีเอาต์พุตก่อนเขียน ให้ใช้แฟล็ก `--clean`:

```bash
codeknit parse ./src --clean
```

การดำเนินการนี้จะทำให้มั่นใจได้ว่าเอาต์พุตจะสดใหม่และสอดคล้องกัน

## เคล็ดลับ

- ✅ **ใช้ `directory-flat` เป็นค่าเริ่มต้น** สำหรับโปรเจ็กต์ส่วนใหญ่ ช่วยสร้างสมดุลระหว่างความสามารถในการอ่านและการจัดการ
- 🔍 ใช้ `--minify` กับโค้ดเบสขนาดใหญ่เพื่อลดการใช้โทเค็นผ่านพจนานุกรมที่ใช้ร่วมกัน (`dict.skt`)
- 🔗 ส่วน `[edges]` **ถูกยกเว้นโดยค่าเริ่มต้น** เพื่อประหยัดโทเค็น ใช้ `--edges` เมื่อคุณต้องการข้อมูลความสัมพันธ์ เช่น `calls`, `contains` หรือ `inherits`
- 🧾 ใช้ `--format json` เมื่อสคริปต์หรือการผสานรวมต้องการข้อมูลที่มีโครงสร้างแทน `.skt`
- 🧹 ใช้ `--clean` เสมอเมื่อรันใหม่ในไดเรกทอรีเอาต์พุตเดียวกัน
- 📁 ใช้ `directory-tree` หากคุณต้องการเชื่อมโยงไฟล์ `.skt` กับไฟล์ซอร์สโดยตรงในตัวแก้ไขโค้ดของคุณ