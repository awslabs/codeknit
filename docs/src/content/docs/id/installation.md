---
title: Instalasi
description: Cara menginstal codeknit di sistem Anda.
---

Panduan ini menginstal codeknit dari sumber, menambahkannya ke `PATH`, mengaktifkan penyelesaian shell, dan memeriksa apakah aplikasi berjalan.

:::tip[Menggunakan asisten pengkodean AI?]
Pengguna Claude Code, Codex, dan Kiro dapat menginstal plugin codeknit sebagai gantinya. Plugin akan mengunduh biner saat pertama kali dijalankan. Lihat [Gunakan codeknit dengan asisten AI](../guides/ai-assistants/).
:::

## Instal dari sumber

**Prasyarat**

- Go 1.26 atau lebih baru
- Kompiler C. codeknit menggunakan CGo untuk membangun parser tree-sitter.

**Langkah-langkah**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Biner akan disimpan di `./bin/codeknit`.

### Gunakan biner pra-bangun sebagai gantinya

Biner pra-bangun untuk Linux, macOS, dan Windows (amd64 dan arm64) tersedia di [halaman rilis](https://github.com/awslabs/codeknit/releases). Unduh arsip untuk platform Anda, ekstrak, dan pindahkan biner ke `PATH`. Setiap rilis menyertakan file `checksums.txt` yang dapat Anda gunakan untuk memverifikasi unduhan.

## Tambahkan codeknit ke PATH

Lewati bagian ini jika Anda telah memindahkan biner ke direktori yang sudah ada di `PATH`, seperti `/usr/local/bin`.

Ganti `/path/to/codeknit` dengan direktori yang berisi biner. Jika Anda membangun dari sumber, direktori tersebut adalah `bin` di dalam repositori yang telah dikloning.

Tambahkan baris berikut untuk shell Anda ke file konfigurasi:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

Muat ulang shell, atau muat ulang file yang telah diedit:

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

Untuk fish, buka terminal baru.

## Aktifkan penyelesaian shell

codeknit menghasilkan skrip penyelesaian untuk bash, zsh, fish, dan PowerShell. Jalankan perintah untuk shell Anda sekali:

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

Perintah untuk bash, zsh, dan PowerShell menambahkan skrip ke profil Anda. Menjalankannya lagi akan menambahkan skrip untuk kedua kalinya, jadi jalankan masing-masing hanya sekali.

Buka shell baru, atau muat ulang profil Anda, untuk mengaktifkan penyelesaian.

## Verifikasi instalasi

```bash
codeknit --version
```

Jika perintah menampilkan versi, codeknit telah terinstal. Jika shell melaporkan `command not found`, periksa apakah direktori dalam baris `PATH` Anda berisi biner `codeknit`, lalu muat ulang shell.

## Siapkan lingkungan pengembangan

Lakukan ini hanya jika Anda berencana untuk berkontribusi pada codeknit. Jalankan perintah berikut dari root repositori yang telah dikloning:

```bash
make deps    # instal dependensi
make setup   # konfigurasi git hooks
make test    # jalankan rangkaian pengujian
```

`make lint` menjalankan linter dan formatter.

## Langkah selanjutnya

- Jalankan perintah pertama Anda di [Memulai](../getting-started/).
- Jelajahi setiap flag di [referensi flag CLI](../reference/cli-flags/).