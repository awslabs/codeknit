---
title: Perintah Parse
description: Mengekstrak informasi struktural dari kode sumber ke dalam file .skt atau JSON.
---

`codeknit parse` membaca file sumber dan menulis simbol-simbolnya (dan, secara opsional, hubungan antar mereka) sebagai file `.skt` atau JSON. Panduan ini mencakup situasi yang paling mungkin Anda hadapi. Untuk daftar lengkap flag yang digunakan bersama semua perintah, lihat [referensi flag CLI](../../reference/cli-flags/). Untuk tata bahasa keluaran, lihat [referensi format keluaran](../../reference/output-format/).

## Parse proyek untuk pertama kali

```bash
codeknit parse ./src
```

codeknit menulis file `.skt` yang dipecah ke dalam `./skeleton/`: `map_001.skt`, `map_002.skt`, dan seterusnya. Buka pecahan-pecahan tersebut secara selektif alih-alih memuat semuanya sekaligus.

Untuk menulis ke lokasi lain, berikan direktori keluaran sebagai argumen kedua:

```bash
codeknit parse ./src ./output
```

Jangan gunakan direktori sumber, atau salah satu direktori induknya, sebagai direktori keluaran.

## Pilih mode keluaran

Pilih mode dengan `--output-mode`.

| Mode             | Gunakan saat                                    | Hasil                                                                         |
| ---------------- | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `directory-flat` | Melakukan parse sebagian besar proyek (default) | `map_001.skt`, `map_002.skt`, ...                                              |
| `directory-tree` | Anda ingin melihat struktur satu file          | Satu file `.skt` per file sumber, mencerminkan pohon sumber                    |
| `inline`         | Melakukan parse satu file, atau piping ke alat lain | Keluaran di stdout; tidak ada direktori keluaran yang digunakan                |

Dalam mode pohon, nama file tetap mempertahankan ekstensi sumber: `src/foo.c` menjadi `src/foo.c.skt`, dan `src/foo.h` menjadi `src/foo.h.skt`. File besar dipecah menjadi `foo.c_part1.skt`, `foo.c_part2.skt`, dan seterusnya.

```bash
codeknit parse ./src --output-mode directory-tree
```

Hindari penggunaan `inline` pada seluruh proyek. Ini akan menulis semuanya ke stdout dan dapat membanjiri terminal atau jendela konteks LLM Anda.

Untuk informasi lebih lanjut tentang pemilihan mode, lihat [Mode keluaran](../output-modes/).

## Jalankan ulang pada proyek yang sama

codeknit menolak untuk menulis ke direktori yang sudah berisi keluaran yang dihasilkan untuk format yang dipilih. Ini mencegah pencampuran pecahan yang usang dan baru. Tambahkan `--clean` untuk mengganti keluaran sebelumnya:

```bash
codeknit parse ./src --clean
```

`--clean` tidak menghapus apa pun di awal. codeknit menghasilkan keluaran baru di direktori staging dan mengganti file yang dihasilkan sebelumnya hanya setelah pembuatan berhasil. Jika pembuatan gagal, atau proses terputus, keluaran sebelumnya tetap dipertahankan atau dikembalikan pada saat berikutnya. File di direktori keluaran yang tidak dihasilkan oleh codeknit akan tetap dipertahankan. Jika salah satu dari mereka bertabrakan dengan jalur yang dihasilkan, proses akan gagal.

Dalam mode JSON, aturan yang sama berlaku untuk `codeknit.json`.

**Jebakan: pecahan campuran.** ID simbol (`S1`, `S2`, ...) dan hasil hubungan dapat berubah antar proses dan versi. Setelah meningkatkan ke 0.5.0, hasilkan ulang seluruh set keluaran dengan `--clean`. Jangan pernah menggabungkan pecahan dari proses yang berbeda.

## Parse satu file

Arahkan `parse` ke file dan gunakan mode `inline`:

```bash
codeknit parse ./src/main.go --output-mode inline
```

Ketika Anda memilih file sumber secara eksplisit, codeknit melewati penyaringan file pengabaian dan file pengujian untuk file tersebut. Pola inklusi/eksklusi dan batas ukuran masih berlaku.

## Dapatkan JSON untuk skrip

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Dokumen JSON memiliki field tingkat atas `files`, `symbols`, `edges` (dihilangkan ketika tidak ada), dan `errors` (dihilangkan ketika tidak ada). Contoh kecil:

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

Untuk menulis JSON ke disk, hilangkan `--output-mode inline`. Dalam mode direktori, JSON masuk ke `codeknit.json`. JSON adalah dokumen tunggal dan tidak dipecah berdasarkan jumlah baris.

## Sertakan hubungan

Secara default, keluaran hanya memiliki bagian `[symbols]`, yang menghemat token. Tambahkan `--edges` ketika Anda perlu mengetahui apa yang memanggil, mengandung, mewarisi, atau mereferensikan apa:

```bash
codeknit parse ./src --edges
```

### Baca hubungan yang tidak pasti

Codeknit menganalisis sintaks tanpa kompiler, sehingga beberapa hubungan tidak dapat ditetapkan. Setiap sisi menunjukkan seberapa pasti hubungan tersebut:

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->` adalah hubungan yang terselesaikan.
- `--kind[unresolved]-->` berarti codeknit tidak dapat menetapkan target.
- `--kind[ambiguous]-->` berarti lebih dari satu target mungkin. ID kandidat dicantumkan jika tersedia.

Kandidat adalah target yang mungkin, bukan dependensi yang dikonfirmasi. Titik akhir yang diketahui menggunakan ShortID. Target yang tidak terselesaikan tanpa simbol yang cocok dapat muncul sebagai nama, relatif terhadap file sumber.

Dalam JSON, sisi yang tidak pasti membawa `resolution` (`"unresolved"` atau `"ambiguous"`) dan, jika tersedia, `candidates` sebagai ID simbol lengkap. Mereka menghilangkan `to_short`.

Untuk C dan C++, direktif `#include` menjadi simbol `meta/file` dan `meta/include`, sehingga sisi mereka memiliki ShortID. Ini mencatat direktif seperti yang tertulis. Ini tidak berarti codeknit menemukan header.

Metrik graf mengecualikan sisi yang tidak pasti dan melaporkan berapa banyak yang dikecualikan. Penampil graf HTML menunjukkan jumlah hubungan yang tidak pasti dan tidak menggambarnya. Untuk memeriksa sisi yang tidak pasti secara individual, gunakan `--edges` pada `parse`. Lihat [Perintah graf](../graph-commands/) untuk laporan tersebut.

Tata bahasa sisi ada dalam [referensi format keluaran](../../reference/output-format/). Aturan dan batasan penyelesaian ada dalam [referensi dukungan bahasa](../../reference/language-support/#relationship-resolution).

## Kurangi ukuran keluaran

Untuk mengompresi token yang berulang, tambahkan `--minify`:

```bash
codeknit parse ./src --minify
```

Ini menulis kamus ke `dict.skt` (atau bagian `[dict]` secara inline). Ganti kode seperti `d0` dan `d1` dengan nilai kamus saat membaca. Jenis hubungan juga dikodekan, tetapi label penyelesaian seperti `[unresolved]` tetap dapat dibaca.

Untuk mengontrol ukuran pecahan, gunakan `--max-lines` (default `500`):

- Dalam mode `directory-flat`, ini adalah target. Blok sumber tidak pernah dipecah, sehingga pecahan dapat melebihinya dengan satu blok.
- Dalam mode `directory-tree`, ini adalah batas keras untuk setiap pecahan sumber. Minimum adalah `3`. Setiap pecahan mengulangi header `[symbols]` dan `## source-path`.

## Parse header C++ bernama `.h`

Codeknit melakukan parse file `.h` sebagai C secara default. Untuk header C++ dengan ekstensi tersebut, atur `--header-language cpp`:

```bash
codeknit parse ./src --header-language cpp
```

Flag ini hanya memengaruhi file `.h`. File bernama `.hpp` dan `.hxx` sudah di-parse sebagai C++, dan ekstensi lain mempertahankan parser biasanya.

## Kontrol file mana yang dipindai

Kelima perintah pemrosesan sumber menerima flag ini. Flag tersebut tercantum secara lengkap dalam [referensi flag CLI](../../reference/cli-flags/).

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` dan `--exclude GLOB` dapat diulang, pola bergaya gitignore relatif terhadap direktori masukan. Kutip mereka agar shell Anda tidak mengembangkannya. Eksklusi menang. Inklusi tidak mengesampingkan `.gitignore`, penyaringan file pengujian, atau pengecualian bawaan.
- `--max-file-size BYTES` melewati file yang lebih besar. Defaultnya adalah `10485760` (10 MiB). `0` menonaktifkan batas.
- `--strict` gagal sebelum menulis keluaran apa pun jika pemindaian atau parse tidak lengkap. Defaultnya adalah `false`.

Codeknit melewati dan melaporkan symlink dan file non-reguler. Aturan `.gitignore` induk berlaku secara independen dari direktori kerja Anda.

**Jebakan: hasil parsial.** Jalur yang tidak dapat dibaca, file yang terlalu besar, symlink yang dilewati, dan peringatan sintaks membuat analisis tidak lengkap. Tanpa `--strict`, codeknit tetap menyimpan hasil yang dapat digunakan dan melaporkan masalah:

- pada stderr,
- dalam field JSON `errors`, atau
- dalam bagian SKT `[errors]` (inline, atau dalam `warnings.skt` dalam mode direktori).

Masukan tanpa sumber yang didukung yang dipilih, atau proses di mana setiap pembacaan sumber gagal, selalu keluar dengan tidak berhasil. Untuk jaminan yang tepat, lihat [Keandalan analisis](../../reference/analysis-reliability/).

## Flag parse

Flag ini khusus untuk `parse`. Flag bersama ada dalam [referensi flag CLI](../../reference/cli-flags/).

| Flag               | Default          | Deskripsi                                                                                       |
| ------------------ | ---------------- | ----------------------------------------------------------------------------------------------- |
| `--output-mode`    | `directory-flat` | Mode keluaran: `inline`, `directory-flat`, atau `directory-tree`.                               |
| `--format`         | `skt`            | Format keluaran: `skt` atau `json`.                                                             |
| `--header-language`| `c`              | Parser untuk file `.h`: `c` atau `cpp`. Ekstensi lain tidak berubah.                            |
| `--max-lines`      | `500`            | Target baris untuk keluaran flat; batas keras (minimum 3) untuk setiap pecahan sumber pohon.    |
| `--collect-test`   | `false`          | Sertakan file pengujian.                                                                        |
| `--minify`         | `false`          | Gunakan kompresi berbasis kamus.                                                                |
| `--edges`          | `false`          | Sertakan bagian `[edges]` (nonaktif secara default untuk menghemat token).                      |
| `--clean`          | `false`          | Ganti keluaran yang dihasilkan sebelumnya, hanya setelah pembuatan berhasil.                    |
| `--workers`        | `NumCPU`         | Jumlah maksimum goroutine parsing konkuren.                                                     |
| `--verbose`        | `false`          | Cetak informasi progres.                                                                        |

## Tips

- Gunakan `directory-flat` untuk sebagian besar proyek, dan baca hanya pecahan yang Anda butuhkan.
- Gunakan `--minify` pada basis kode besar untuk menghemat token.
- Berikan `--edges` hanya ketika Anda membutuhkan data hubungan.
- Periksa [matriks kemampuan yang diuji dan batasan yang diketahui](../../reference/language-support/) sebelum mengandalkan ekstraksi untuk bahasa tertentu.

## Lihat juga

- [Mode keluaran](../output-modes/)
- [Referensi format keluaran](../../reference/output-format/)
- [Keandalan analisis](../../reference/analysis-reliability/)
- [Perintah graf](../graph-commands/)