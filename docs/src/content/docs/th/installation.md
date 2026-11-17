---
title: การติดตั้ง
description: วิธีการติดตั้ง codeknit บนระบบของคุณ
---

คู่มือนี้จะติดตั้ง codeknit จากซอร์ส เพิ่มลงใน `PATH` ของคุณ เปิดใช้งานการเติมคำในเชลล์ และตรวจสอบว่าโปรแกรมทำงานได้

:::tip[ใช้งานผู้ช่วยเขียนโค้ด AI หรือไม่?]
ผู้ใช้ Claude Code, Codex และ Kiro สามารถติดตั้งปลั๊กอิน codeknit แทนได้ ปลั๊กอินจะดาวน์โหลดไบนารีในครั้งแรกที่รัน ดูรายละเอียดที่ [ใช้งาน codeknit กับผู้ช่วย AI](../guides/ai-assistants/)
:::

## ติดตั้งจากซอร์ส

**ข้อกำหนดเบื้องต้น**

- Go 1.26 หรือใหกว่า
- คอมไพเลอร์ C codeknit ใช้ CGo ในการสร้าง tree-sitter parsers

**ขั้นตอน**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

ไบนารีจะถูกเขียนไปที่ `./bin/codeknit`

### ใช้ไบนารีที่สร้างไว้ล่วงหน้าแทน

ไบนารีที่สร้างไว้ล่วงหน้าสำหรับ Linux, macOS และ Windows (amd64 และ arm64) มีอยู่ใน [หน้ารุ่น](https://github.com/awslabs/codeknit/releases) ดาวน์โหลดไฟล์เก็บถาวรสำหรับแพลตฟอร์มของคุณ แตกไฟล์ และย้ายไบนารีไปยัง `PATH` ของคุณ แต่ละรุ่นมีไฟล์ `checksums.txt` ที่คุณสามารถตรวจสอบการดาวน์โหลดได้

## เพิ่ม codeknit ลงใน PATH ของคุณ

ข้ามส่วนนี้หากคุณย้ายไบนารีไปยังไดเรกทอรีที่อยู่ใน `PATH` อยู่แล้ว เช่น `/usr/local/bin`

แทนที่ `/path/to/codeknit` ด้วยไดเรกทอรีที่มีไบนารี หากคุณสร้างจากซอร์ส ไดเรกทอรีดังกล่าวคือไดเรกทอรี `bin` ภายใน repository ที่โคลนมา

เพิ่มบรรทัดสำหรับเชลล์ของคุณลงในไฟล์การกำหนดค่า:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

รีโหลดเชลล์ของคุณ หรือโหลดไฟล์ที่คุณแก้ไขใหม่:

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

สำหรับ fish ให้เปิดเทอร์มินัลใหม่

## เปิดใช้งานการเติมคำในเชลล์

codeknit สร้างสคริปต์การเติมคำสำหรับ bash, zsh, fish และ PowerShell รันคำสั่งสำหรับเชลล์ของคุณครั้งเดียว:

```bash
# bash
codeknit completion bash >> ~/.bashrc

# zsh
codeknit completion zsh >> ~/.zshrc

# fish
codeknit completion fish > ~/.config/fish/completions/codeknit.fish

# PowerShell
codeknit completion powershell >> $PROFILE
```

คำสั่งสำหรับ bash, zsh และ PowerShell จะเพิ่มไปยังโปรไฟล์ของคุณ การรันคำสั่งซ้ำจะเพิ่มสคริปต์ครั้งที่สอง ดังนั้นให้รันแต่ละคำสั่งเพียงครั้งเดียว

เปิดเชลล์ใหม่ หรือโหลดโปรไฟล์ของคุณใหม่เพื่อเปิดใช้งานการเติมคำ

## ตรวจสอบการติดตั้ง

```bash
codeknit --version
```

หากคำสั่งแสดงเวอร์ชัน แสดงว่า codeknit ได้รับการติดตั้งแล้ว หากเชลล์ของคุณรายงาน `command not found` ให้ตรวจสอบว่าไดเรกทอรีในบรรทัด `PATH` ของคุณมีไบนารี `codeknit` หรือไม่ จากนั้นรีโหลดเชลล์ของคุณ

## ตั้งค่าสภาพแวดล้อมการพัฒนา

ดำเนินการนี้เฉพาะเมื่อคุณวางแผนที่จะมีส่วนร่วมใน codeknit รันคำสั่งเหล่านี้จากรากของ repository ที่โคลนมา:

```bash
make deps    # ติดตั้ง dependencies
make setup   # กำหนดค่า git hooks
make test    # รันชุดทดสอบ
```

`make lint` จะรัน linter และ formatter

## ขั้นตอนถัดไป

- รันคำสั่งแรกของคุณใน [เริ่มต้นใช้งาน](../getting-started/)
- ดูรายละเอียดของแฟล็กทั้งหมดใน [CLI flags reference](../reference/cli-flags/)