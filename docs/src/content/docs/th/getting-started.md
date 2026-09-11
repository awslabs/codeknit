---
title: เริ่มต้นใช้งาน
description: เริ่มต้นใช้งาน codeknit ภายใน 5 นาที
---

# เริ่มต้นใช้งาน

เริ่มต้นใช้งาน codeknit ภายใน 5 นาที

## 1. ข้อกำหนดเบื้องต้น

คุณจะต้องมี:

- Go 1.26+
- C compiler (CGo จำเป็นสำหรับ tree-sitter)

## 2. การติดตั้งจากซอร์สโค้ด

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# Binary is at ./bin/codeknit
```

## 3. เพิ่มใน PATH

เพิ่มไบนารีลงใน PATH ของ shell ของคุณ:

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

รีโหลด shell ของคุณหรือรัน `source ~/.bashrc` (หรือ `~/.zshrc`) เพื่อให้การเปลี่ยนแปลงมีผล

## 4. ตรวจสอบการติดตั้ง

ตรวจสอบว่า codeknit ทำงานได้:

```bash
codeknit --version
```

## 5. การแปลงครั้งแรก

รันการแปลงครั้งแรกบน codebase:

```bash
codeknit parse ./myproject
```

คำสั่งนี้:

- แปลงไฟล์ต้นฉบับทั้งหมดใน `./myproject`
- ดึงข้อมูลโครงสร้าง (ฟังก์ชัน คลาส ความสัมพันธ์)
- เขียนไฟล์ `.skt` ที่แบ่งเป็นชิ้นไปยัง `./skeleton/` (ไดเรกทอรีเอาต์พุตเริ่มต้น)

หากคุณรันคำสั่งนี้อีกครั้ง ให้ใช้ `--clean` เพื่อลบเอาต์พุตก่อนหน้า:

```bash
codeknit parse ./myproject --clean
```

## 6. การอ่านเอาต์พุต

ไฟล์ `.skt` มีข้อมูลโค้ดที่มีโครงสร้าง นี่คือตัวอย่างเล็กน้อย:

```skt
[symbols]
## src/main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {exported}
S3 callable/function L10-L12 NewServer(addr: string) -> *S2 {exported}
S4 callable/method L14-L19 Start() {receiver=*Server}

[edges]
S2 --contains--> S4
S3 --returns--> S2
```

ส่วนสำคัญ:

- `[symbols]`: นิยามที่จัดกลุ่มตามไฟล์ แสดงชื่อ ช่วงบรรทัด และข้อมูลเมตา
- `[edges]`: ความสัมพันธ์ เช่น `contains`, `calls`, `inherits` หรือ `returns`

## 7. ขั้นตอนถัดไป

เมื่อคุณได้รันการแปลงครั้งแรกแล้ว:

- เรียนรู้เพิ่มเติมเกี่ยวกับคำสั่ง parse: [คู่มือคำสั่ง Parse](/codeknit/th/guides/parse-command/)
- สำรวจการวิเคราะห์โครงสร้าง: [คู่มือคำสั่ง Graph](/codeknit/th/guides/graph-commands/)
- ทำความเข้าใจการตรวจจับรายการซ้ำ: [คู่มือคำสั่ง Fingerprint](/codeknit/th/guides/fingerprint-command/)
- อ่านรูปแบบเอาต์พุตทั้งหมด: [ข้อมูลอ้างอิงรูปแบบเอาต์พุต](/codeknit/th/reference/output-format/)
- ดูฟลैกที่มีทั้งหมด: [ข้อมูลอ้างอิง CLI flags](/codeknit/th/reference/cli-flags/)