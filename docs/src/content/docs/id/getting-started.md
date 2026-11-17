---
title: Memulai
description: Mulai menggunakan codeknit dalam waktu kurang dari 5 menit.
---

Dalam tutorial ini, Anda akan membangun codeknit dari sumber, menjalankan parse pertama, dan membaca keluaran `.skt` yang dihasilkan. Anda akan mem-parsing paket kecil dari repositori codeknit, sehingga Anda tidak memerlukan proyek sendiri.

## Prasyarat

- **Go 1.26 atau lebih baru.** Jalankan `go version` untuk memeriksa.
- **Kompiler C.** codeknit menggunakan tree-sitter, yang membutuhkan CGo.
- **Git** dan **Make**.

## 1. Bangun codeknit

Klona repositori dan bangun:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Anda akan melihat proses pembangunan selesai tanpa kesalahan, dan biner di `./bin/codeknit`.

## 2. Tambahkan codeknit ke PATH Anda

Dari direktori `codeknit`, tambahkan `bin` ke `PATH` Anda menggunakan baris untuk shell Anda.

Untuk bash atau zsh:

```bash
export PATH="$PATH:$(pwd)/bin"
```

Untuk fish:

```fish
fish_add_path (pwd)/bin
```

Ini berlaku untuk sesi bash atau zsh saat ini. Untuk menyimpannya untuk sesi berikutnya, tambahkan baris yang sama ke `~/.bashrc` atau `~/.zshrc`, dengan path absolut ke `bin` menggantikan `$(pwd)/bin`.

## 3. Verifikasi instalasi

```bash
codeknit --version
```

Anda akan melihat versi codeknit tercetak.

## 4. Jalankan parse pertama Anda

Tetap di direktori `codeknit` dan parse paket `internal/console`:

```bash
codeknit parse ./internal/console --edges
```

codeknit memindai direktori, mem-parsing setiap file sumber yang didukung, dan menulis file `.skt`. *File `.skt`* adalah file teks ringkas yang mencantumkan simbol dalam kode Anda dan hubungan di antara mereka. Bendera `--edges` menambahkan hubungan. Hubungan ini dihilangkan secara default untuk menghemat ruang.

Anda akan melihat ringkasan singkat saat perintah selesai, dan direktori `./skeleton` baru di direktori saat ini.

Daftar keluaran:

```bash
ls ./skeleton
```

Anda akan melihat file seperti `map_001.skt`.

## 5. Baca keluaran

Buka file:

```bash
cat ./skeleton/map_001.skt
```

Anda akan melihat bagian `[symbols]` terlebih dahulu. *Simbol* adalah item bernama dalam kode Anda, seperti fungsi, tipe, atau field. Simbol dikelompokkan di bawah heading `## <source file>`. Keluaran Anda mencakup baris seperti ini:

```skt
[symbols]
## <path to>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exported}
...
S10 callable/function L28-L41 New() -> *S2 {exported}
...
S12 callable/method L49-L49 Error(msg: string) {exported, receiver=*Console}
```

Setiap baris memiliki bagian yang sama:

| Bagian | Contoh | Arti |
| --- | --- | --- |
| ShortID | `S10` | Label pendek untuk simbol, digunakan di tempat lain dalam keluaran |
| Kategori dan jenis | `callable/function` | Jenis simbol apa itu |
| Rentang baris | `L28-L41` | Baris yang dicakup simbol dalam file sumber |
| Tanda tangan | `New() -> *S2` | Nama, parameter, dan tipe pengembalian |
| Properti | `{exported}` | Detail tambahan dalam kurung kurawal |

Dalam `New() -> *S2`, `S2` adalah ShortID dari tipe `Console`. codeknit mengganti nama tipe dengan ShortID ketika mengetahui simbolnya. Nomor baris dapat sedikit berbeda jika checkout Anda lebih baru.

Gulir ke bawah ke bagian `[edges]`. *Sisi* adalah hubungan antara dua simbol. Setiap baris sisi mencantumkan simbol sumber, jenis hubungan, dan simbol target atau simbol:

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

Contoh ini menunjukkan bentuk baris sisi. Ini berarti "`S2` contains `S4`" dan "`S4` calls `S5`". Anda dapat mencari setiap ShortID di `[symbols]` untuk melihat apa yang dirujuknya.

## Apa yang telah Anda capai

Anda telah membangun codeknit, menambahkannya ke `PATH`, menjalankan `codeknit parse`, dan membaca bagian `[symbols]` dan `[edges]` dari keluaran `.skt`.

## Langkah selanjutnya

- Parse proyek Anda sendiri, pilih tata letak keluaran yang berbeda, dan jalankan kembali dengan aman: lihat [panduan perintah parse](../guides/parse-command/).
- Visualisasikan dan analisis struktur kode Anda: lihat [panduan perintah graf](../guides/graph-commands/).
- Temukan kode duplikat dan hampir duplikat: lihat [panduan perintah fingerprint](../guides/fingerprint-command/).
- Cari setiap bagian dari format `.skt`: lihat [referensi format keluaran](../reference/output-format/).
- Cari setiap bendera baris perintah: lihat [referensi bendera CLI](../reference/cli-flags/).