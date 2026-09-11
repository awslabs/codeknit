---
title: Instalasi
description: Cara menginstal codeknit pada sistem Anda.
---

codeknit dapat diinstal dari sumber. Langkah-langkah berikut akan memandu Anda dalam menyiapkan codeknit pada sistem Anda.

## Dari sumber

Metode instalasi utama adalah dengan membangun dari sumber. Anda akan memerlukan:

- Go 1.26+
- Compiler C (diperlukan untuk tree-sitter via CGo)

Klon repositori dan bangun biner:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Biner yang telah dikompilasi akan tersedia di `./bin/codeknit`.

## Tambahkan ke PATH

Untuk menjalankan `codeknit` dari direktori mana pun, tambahkan lokasi biner ke PATH sistem Anda.

Untuk **bash** (`~/.bashrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

Untuk **zsh** (`~/.zshrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

Untuk **fish** (`~/.config/fish/config.fish`):

```fish
fish_add_path /path/to/codeknit
```

Setelah memperbarui konfigurasi shell Anda, muat ulang dengan menjalankan `source ~/.bashrc` (atau `~/.zshrc`) atau mulai ulang terminal Anda.

## Penyelesaian shell

codeknit mendukung penyelesaian otomatis untuk shell populer. Instal penyelesaian menggunakan perintah berikut:

Untuk **bash**:

```bash
codeknit completion bash >> ~/.bashrc
```

Untuk **zsh**:

```bash
codeknit completion zsh >> ~/.zshrc
```

Untuk **fish**:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

Untuk **PowerShell**:

```powershell
codeknit completion powershell >> $PROFILE
```

## Verifikasi instalasi

Setelah instalasi, verifikasi bahwa codeknit telah diatur dengan benar:

```bash
codeknit --version
```

## Pengaturan pengembangan

Jika Anda berkontribusi pada codeknit, jalankan perintah tambahan berikut:

Instal dependensi pengembangan:

```bash
make deps
```

Siapkan git hooks:

```bash
make setup
```

Jalankan rangkaian pengujian:

```bash
make test
```