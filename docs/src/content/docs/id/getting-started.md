---
title: Memulai
description: Mulai menggunakan codeknit dalam waktu kurang dari 5 menit.
---

# Memulai

Mulai menggunakan codeknit dalam waktu kurang dari 5 menit.

## 1. Prasyarat

Anda memerlukan:

- Go 1.26+
- Kompilator C (CGo diperlukan untuk tree-sitter)

## 2. Instalasi dari sumber

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# Binary is at ./bin/codeknit
```

## 3. Tambahkan ke PATH

Tambahkan binary ke PATH shell Anda:

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

Muat ulang shell Anda atau jalankan `source ~/.bashrc` (atau `~/.zshrc`) agar perubahan berlaku.

## 4. Verifikasi instalasi

Pastikan codeknit berfungsi:

```bash
codeknit --version
```

## 5. Parse pertama

Jalankan parse pertama pada basis kode:

```bash
codeknit parse ./myproject
```

Perintah ini:

- Mem-parsing semua file sumber di `./myproject`
- Mengekstrak informasi struktur kode (fungsi, kelas, hubungan)
- Menulis file `.skt` yang telah di-chunk ke `./skeleton/` (direktori keluaran default)

Jika Anda menjalankan perintah ini lagi, gunakan `--clean` untuk menghapus keluaran sebelumnya:

```bash
codeknit parse ./myproject --clean
```

## 6. Membaca keluaran

File `.skt` berisi informasi kode terstruktur. Berikut contoh kecil:

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

Bagian utama:

- `[symbols]`: Definisi dikelompokkan berdasarkan file, menampilkan nama, rentang baris, dan metadata
- `[edges]`: Hubungan seperti `contains`, `calls`, `inherits`, atau `returns`

## 7. Langkah selanjutnya

Setelah menjalankan parse pertama:

- Pelajari lebih lanjut tentang perintah parse: [Panduan perintah parse](/codeknit/id/guides/parse-command/)
- Jelajahi analisis struktur: [Panduan perintah graf](/codeknit/id/guides/graph-commands/)
- Pahami deteksi duplikat: [Panduan perintah fingerprint](/codeknit/id/guides/fingerprint-command/)
- Baca format keluaran lengkap: [Referensi format keluaran](/codeknit/id/reference/output-format/)
- Lihat semua flag yang tersedia: [Referensi flag CLI](/codeknit/id/reference/cli-flags/)