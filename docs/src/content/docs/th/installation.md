---
title: การติดตั้ง
description: วิธีการติดตั้ง codeknit บนระบบของคุณ
---

codeknit สามารถติดตั้งจากซอร์สโค้ดได้ ขั้นตอนต่อไปนี้จะแนะนำคุณในการตั้งค่า codeknit บนระบบของคุณ

## จากซอร์สโค้ด

วิธีการติดตั้งหลักคือการสร้างจากซอร์สโค้ด คุณจะต้องมี:

- Go 1.26+
- ตัวแปลภาษา C (จำเป็นสำหรับ tree-sitter ผ่าน CGo)

โคลน repository และสร้างไบนารี:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

ไบนารีที่คอมไพล์แล้วจะอยู่ที่ `./bin/codeknit`

## เพิ่มใน PATH

หากต้องการรัน `codeknit` จากไดเรกทอรีใด ๆ ให้เพิ่มตำแหน่งของไบนารีลงใน PATH ของระบบของคุณ

สำหรับ **bash** (`~/.bashrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

สำหรับ **zsh** (`~/.zshrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

สำหรับ **fish** (`~/.config/fish/config.fish`):

```fish
fish_add_path /path/to/codeknit
```

หลังจากอัปเดตการกำหนดค่าเชลล์ของคุณแล้ว ให้โหลดใหม่โดยรัน `source ~/.bashrc` (หรือ `~/.zshrc`) หรือรีสตาร์ทเทอร์มินัลของคุณ

## การเติมคำอัตโนมัติในเชลล์

codeknit รองรับการเติมคำอัตโนมัติสำหรับเชลล์ยอดนิยม ติดตั้งการเติมคำอัตโนมัติโดยใช้คำสั่งเหล่านี้:

สำหรับ **bash**:

```bash
codeknit completion bash >> ~/.bashrc
```

สำหรับ **zsh**:

```bash
codeknit completion zsh >> ~/.zshrc
```

สำหรับ **fish**:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

สำหรับ **PowerShell**:

```powershell
codeknit completion powershell >> $PROFILE
```

## ตรวจสอบการติดตั้ง

หลังจากการติดตั้งแล้ว ให้ตรวจสอบว่า codeknit ถูกตั้งค่าอย่างถูกต้อง:

```bash
codeknit --version
```

## การตั้งค่าสำหรับการพัฒนา

หากคุณกำลังมีส่วนร่วมในการพัฒนา codeknit ให้รันคำสั่งเพิ่มเติมเหล่านี้:

ติดตั้ง dependencies สำหรับการพัฒนา:

```bash
make deps
```

ตั้งค่า git hooks:

```bash
make setup
```

รันชุดทดสอบ:

```bash
make test
```