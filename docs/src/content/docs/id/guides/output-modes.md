---
title: Mode Keluaran
description: Pilih mode keluaran yang tepat untuk ukuran proyek dan alur kerja Anda.
---

codeknit memisahkan dua keputusan: ke mana keluaran akan disimpan (**mode keluaran**) dan bagaimana keluaran tersebut dikodekan (**format keluaran**). Halaman ini menjelaskan mengapa setiap opsi ada, apa biayanya, dan kapan opsi tersebut sesuai. Untuk tata bahasa `.skt`, lihat [referensi format keluaran](../../reference/output-format/). Untuk perintah, lihat [panduan perintah parse](../parse-command/).

## Dua pilihan independen

Flag `--output-mode` memilih salah satu dari tiga mode: `directory-flat`, `directory-tree`, atau `inline`. Flag `--format` memilih `skt` (default) atau `json`. Setiap mode dapat dikombinasikan dengan salah satu format.

Kedua pilihan tersebut terpisah karena menjawab pertanyaan yang berbeda. Mode menentukan bagaimana keluaran dibagi dan ditempatkan di disk, yang penting bagi seseorang atau LLM yang membacanya. Format menentukan siapa yang mengonsumsinya: LLM atau pembaca manusia, atau skrip.

## directory-flat

`directory-flat` adalah mode default. Mode ini menulis file potongan bernomor, `map_001.skt`, `map_002.skt`, dan seterusnya, ke dalam `./skeleton/` kecuali Anda menentukan direktori keluaran lain.

**Mengapa mode ini ada.** Basis kode besar menghasilkan lebih banyak struktur daripada yang muat dalam satu jendela konteks. Pembagian potongan memungkinkan pembaca hanya membuka file yang dibutuhkan alih-alih memuat semuanya.

**Cara pembagian bekerja.** codeknit menjaga blok setiap file sumber tetap utuh dan memulai potongan baru di antara blok setelah potongan mencapai sekitar `--max-lines` (default 500). Batasan ini adalah target, bukan batasan keras: satu file sumber dengan blok yang sangat besar dapat menghasilkan potongan yang lebih panjang dari 500 baris. Menjaga blok tetap utuh berarti simbol tidak pernah dipisahkan dari saudaranya dalam file yang sama.

**Pertimbangan.** Potongan tidak diatur berdasarkan jalur sumber. Untuk menemukan struktur satu file, Anda mencari potongan untuk header `## source-path`-nya. Jika Anda sering mencari file tertentu, `directory-tree` lebih sesuai.

**Minifikasi.** Dengan `--minify`, kamus ditulis ke `dict.skt` di samping potongan. Lihat [Minifikasi](#minifikasi).

## directory-tree

`directory-tree` mencerminkan pohon sumber Anda di bawah direktori keluaran. Setiap file sumber mendapatkan file `.skt` pada jalur yang sesuai.

**Penamaan.** codeknit menambahkan `.skt` ke nama file sumber lengkap alih-alih mengganti ekstensi. `src/foo.c` menjadi `src/foo.c.skt`, dan `src/foo.h` menjadi `src/foo.h.skt`. Tanpa aturan ini, keduanya akan dipetakan ke `foo.skt` dan bertabrakan. Ketika keluaran file dibagi, bagian-bagiannya dinamai `foo.c_part1.skt`, `foo.c_part2.skt`, dan seterusnya. Jalur keluaran yang bertabrakan, termasuk tabrakan hanya huruf besar-kecil, gagal secara terlihat alih-alih menimpa satu sama lain.

**Pembagian berbeda dari mode flat.** Dalam mode pohon, `--max-lines` adalah batasan keras pada setiap potongan sumber, dengan minimum 3. Setiap potongan simbol mengulangi header `[symbols]` dan `## source-path` sehingga dapat dibaca sendiri. Potongan hanya sisi membawa `[edges]`. Batasan berlaku untuk potongan sumber, bukan untuk file tambahan seperti `dict.skt` dan `warnings.skt`.

**Pertimbangan.** Keluaran pohon mudah dinavigasi bersama sumber, tetapi menghasilkan banyak file kecil. Proyek besar dapat menghasilkan ribuan file, dan pembaca yang ingin gambaran luas harus membuka banyak file.

## inline

`inline` menulis semuanya ke stdout. Mode ini tidak membuat direktori keluaran atau area staging.

**Mengapa mode ini ada.** Mode ini cocok untuk dua kasus: memeriksa satu file dengan cepat, dan mengalirkan keluaran ke alat lain. Ini adalah satu-satunya mode di mana tidak ada yang ditulis ke disk.

**Pertimbangan.** Tidak ada yang dibagi. Seluruh hasil masuk ke terminal atau konteks konsumen sekaligus, sehingga membebani pembaca pada apa pun di luar beberapa file. Anggap mode ini sesuai untuk satu file atau proyek yang sangat kecil, kira-kira kurang dari lima file. Karena stdout tidak dapat di-rollback, konsumen yang gagal dapat melihat keluaran parsial. Mode direktori tidak memiliki masalah ini.

## Format keluaran

| Format | Paling cocok untuk | Keluaran |
| --- | --- | --- |
| `skt` | Konteks LLM dan inspeksi manusia | File `.skt` atau stdout |
| `json` | Skrip dan integrasi terstruktur | `codeknit.json` dalam mode direktori, atau stdout dalam mode inline |

**skt** ringkas dan berorientasi baris. Setiap simbol berada pada satu baris fisik, dengan jeda baris dalam tanda tangan di-escape, sehingga pembaca dapat memuat sebagian keluaran tanpa merusak strukturnya. Itulah mengapa pembagian potongan bekerja untuk `.skt`.

**json** ada untuk konsumen yang membutuhkan bidang bertipe alih-alih teks untuk dibaca. Ini adalah satu dokumen dengan `files`, `symbols`, dan, ketika sisi disertakan, `edges`. Karena ini adalah satu dokumen, tidak dibagi berdasarkan jumlah baris. Dalam mode direktori `--max-lines` oleh karena itu tidak berpengaruh pada JSON, dan proyek besar menghasilkan satu `codeknit.json` besar.

Contoh ringkas:

```json
{
  "files": ["app.go"],
  "symbols": [
    {"id": "app.go::User", "short_id": "S1", "name": "User", "file": "app.go",
     "category": "type", "kind": "struct", "signature": "type User struct", "span": [3, 3]},
    {"id": "app.go::Save", "short_id": "S2", "name": "Save", "file": "app.go",
     "category": "callable", "kind": "function", "signature": "Save(u: S1)", "span": [5, 5]}
  ],
  "edges": [
    {"from": "app.go::Save", "from_short": "S2", "to": "app.go::User",
     "to_short": "S1", "kind": "references"}
  ]
}
```

## Memilih mode

| Mode | Paling cocok untuk | Keluaran |
| --- | --- | --- |
| `directory-flat` | Sebagian besar proyek (default) | `./skeleton/map_001.skt`, `map_002.skt`, ... |
| `directory-tree` | Navigasi bersama sumber | `./skeleton/<jalur tercermin>.skt` |
| `inline` | Satu file, pipa | stdout |

Aturan praktis:

- **Jika ragu**, gunakan `directory-flat`, mode default.
- **Satu file**: `inline` sudah cukup.
- **Lebih dari beberapa file**: gunakan `directory-flat` atau `directory-tree`.
- **Basis kode besar**: tambahkan `--minify`.
- **Menjalankan ulang ke direktori yang sama**: tambahkan `--clean`.

## Staging, menjalankan ulang, dan `--clean`

Mode direktori tidak menulis ke direktori keluaran saat proses berlangsung. codeknit menghasilkan keluaran di direktori staging saudara dan mempublikasikannya hanya setelah pembuatan berhasil. Jalankan yang gagal oleh karena itu meninggalkan keluaran sebelumnya tetap utuh.

Secara default, codeknit menolak untuk menulis ketika keluaran yang dihasilkan untuk format yang dipilih sudah ada. Alasannya adalah ID simbol (`S1`, `S2`, ...) dapat berubah setiap kali ekstraksi berubah. Mencampur potongan dari berbagai jalankan akan memberikan sisi yang menunjuk ke simbol yang salah. `--clean` memilih untuk penggantian: setelah pembuatan berhasil, file `.skt` usang (dalam mode SKT) atau `codeknit.json` (dalam mode JSON) diganti. Entri yang tidak terkait dalam direktori dipertahankan. Jika salah satunya bertabrakan dengan jalur yang dihasilkan, jalankan gagal alih-alih menimpanya.

Publikasi mengambil kunci khusus direktori, mencadangkan keluaran sebelumnya, dan memulihkannya jika penggantian gagal. Jika proses terputus di antara dua penggantian direktori, jalankan berikutnya memulihkan cadangan sebelum menghasilkan apa pun. Ini memulihkan dari kegagalan proses. Ini tidak membuat pembaruan direktori menjadi instan: ada interval singkat antara penggantian, sehingga pembaca yang membutuhkan akses tanpa gangguan harus berkoordinasi dengan pembuatan. Ini juga tidak melindungi dari kegagalan penyimpanan atau kehilangan daya.

Jangan pilih direktori sumber, atau leluhur darinya, sebagai direktori keluaran. Untuk jaminan publikasi penuh, lihat [referensi keandalan analisis](../../reference/analysis-reliability/).

## Minifikasi

`--minify` menerapkan kompresi berbasis kamus. codeknit mengganti token berulang, seperti `callable/method` atau `exported`, dengan kode pendek (`d0`, `d1`, ...). Pemetaan disimpan dalam bagian `[dict]`, yang dalam mode direktori adalah file `dict.skt`.

Penghematan berasal dari pengulangan. Sebagian besar baris simbol berbagi kategori, jenis, dan properti yang sama, sehingga basis kode besar mengulangi string ini ribuan kali. Biayanya adalah indireksi: pembaca harus memperluas kode menggunakan kamus, dan satu potongan tidak lagi mandiri. Kamus adalah keluaran tambahan terpisah, sehingga tidak dihitung terhadap `--max-lines`.

Keluaran yang diminifikasi mengkodekan jenis hubungan tetapi menjaga penanda resolusi tetap dapat dibaca, sehingga sisi `unresolved` dan `ambiguous` tetap dapat dibedakan. Untuk proyek kecil, penghematan kecil dan langkah tambahan jarang sepadan. Untuk proyek besar, penghematan dalam token bisa signifikan.

## Halaman terkait

- [Panduan perintah parse](../parse-command/): perintah untuk setiap mode.
- [Referensi format keluaran](../../reference/output-format/): tata bahasa `.skt` dan JSON.
- [Referensi flag CLI](../../reference/cli-flags/): setiap flag dan default.
- [Referensi keandalan analisis](../../reference/analysis-reliability/): penamaan keluaran dan jaminan publikasi.