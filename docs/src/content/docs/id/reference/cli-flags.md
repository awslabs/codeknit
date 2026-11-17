---
title: Referensi CLI
description: Referensi lengkap untuk semua perintah dan flag codeknit.
---

Halaman ini mencantumkan setiap perintah codeknit, argumennya, dan flag-nya. Untuk instruksi berbasis tugas, lihat [panduan perintah parse](../../guides/parse-command/), [panduan perintah graf](../../guides/graph-commands/), dan [panduan perintah fingerprint](../../guides/fingerprint-command/).

## Ringkasan perintah

| Perintah                                       | Tujuan                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | Meluncurkan antarmuka terminal interaktif.                     |
| `codeknit parse <input-path> [output-dir]`    | Mengekstrak simbol dan hubungan ke dalam keluaran `.skt` atau JSON. |
| `codeknit graph show <input-path>`            | Membuat graf HTML interaktif.                                  |
| `codeknit graph analyze <input-path>`         | Menjalankan algoritma graf struktural dan menulis laporan `.skt`. |
| `codeknit graph hotspots <input-path>`        | Memberi peringkat hotspot perubahan menggunakan riwayat Git dan struktur graf. |
| `codeknit fingerprint <input-path>`           | Mendeteksi kode duplikat dan hampir duplikat.                  |
| `codeknit completion <shell>`                 | Membuat skrip penyelesaian shell.                              |

Lima perintah memproses kode sumber: `parse`, `graph show`, `graph analyze`, `graph hotspots`, dan `fingerprint`. Mereka berbagi flag yang dijelaskan dalam [Flag pemrosesan sumber bersama](#flag-pemrosesan-sumber-bersama).

## Flag global

| Flag              | Tipe | Default | Deskripsi                  |
| ----------------- | ---- | ------- | ---------------------------- |
| `--version`       | bool | `false` | Menampilkan informasi versi. |
| `-h`, `--help`    | bool | `false` | Menampilkan bantuan untuk perintah. |

## Flag pemrosesan sumber bersama

Flag ini muncul dalam tabel setiap perintah pemrosesan sumber.

| Flag                    | Tipe                | Default               | Nilai yang Diizinkan       | Deskripsi                                                                          |
| ----------------------- | ------------------- | --------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                   | `c`, `cpp`           | Parser yang digunakan untuk file `.h`.                                              |
| `--include GLOB`        | string, dapat diulang | Semua sumber yang didukung | Pola glob gaya gitignore | Memasukkan jalur sumber yang cocok. Beberapa pola membentuk gabungan.               |
| `--exclude GLOB`        | string, dapat diulang | Tidak ada             | Pola glob gaya gitignore | Mengecualikan jalur atau direktori yang cocok.                                      |
| `--max-file-size BYTES` | int                 | `10485760` (10 MiB)   | `0` atau lebih besar | Melewati file sumber yang lebih besar dari ukuran ini. `0` menonaktifkan batas.     |
| `--strict`              | bool                | `false`               | —                    | Gagal sebelum keluaran ditulis jika terjadi diagnosa pemindaian atau parsing.        |

### `--header-language`

Hanya berlaku untuk file `.h`. Ekstensi lain tetap menggunakan parser biasanya (`.hpp` dan `.hxx` sudah menggunakan C++).

- `c` (default) mempertahankan ekstraksi makro C, typedef, dan union.
- `cpp` memilih parser C++. Gunakan untuk proyek C++ yang header-nya dinamai `.h`.

Antarmuka pengguna terminal (TUI) menggunakan pilihan default `c` untuk file `.h`.

```bash
codeknit parse ./src --header-language cpp
```

Untuk kemampuan ekstraksi yang diuji dan batasan yang diketahui per bahasa, lihat [Dukungan Bahasa dan Fidelity API](../language-support/). Halaman tersebut menjelaskan konstruksi yang diuji, bukan cakupan bahasa atau semantik kompiler yang lengkap.

### Pemilihan sumber

- Pola menggunakan glob gaya gitignore yang dipisahkan dengan garis miring, relatif terhadap direktori masukan. Untuk masukan file tunggal, pola cocok dengan nama dasar file.
- Kutip pola agar shell tidak mengembangkannya.
- Pengecualian lebih diutamakan daripada pemasukan.
- Pemasukan tidak mengesampingkan aturan `.gitignore`, penyaringan file uji, atau pengecualian bawaan.
- Pola negasi (`!pattern`) tidak diterima dalam `--include` atau `--exclude`.
- Pemindaian direktori menghormati file `.gitignore` leluhur dan bersarang secara independen dari direktori kerja saat ini.
- Memilih secara eksplisit file sumber reguler melewati penyaringan abaikan dan uji. Pemasukan, pengecualian, dan batas ukuran masih berlaku.
- File uji dikecualikan kecuali `--collect-test` disetel.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### Diagnostik dan `--strict`

- Simbol tautan, simbol tautan direktori, dan file non-reguler dilewati dan dilaporkan.
- Jalur yang tidak dapat dibaca, pembacaan file abaikan yang gagal, sumber yang terlalu besar, dan peringatan sintaks membuat analisis tidak lengkap.
- Secara default, hasil parsial yang dapat digunakan tetap disimpan. Diagnostik muncul di stderr. `parse` juga mencatatnya dalam `errors` JSON atau dalam `[errors]` SKT.
- Dengan `--strict`, analisis yang tidak lengkap ditolak sebelum keluaran ditulis.
- Eksekusi tanpa sumber yang didukung yang dipilih, atau di mana setiap pembacaan sumber gagal, selalu keluar dengan tidak berhasil.
- TUI menerapkan kebijakan simbol tautan yang sama dan default 10 MiB. Formulirnya tidak menampilkan flag pemilihan dan strict.

Lihat [Keandalan Analisis](../analysis-reliability/) untuk jaminan yang tepat.

## `codeknit`

```bash
codeknit
```

Menjalankan `codeknit` tanpa subperintah akan meluncurkan antarmuka terminal interaktif. Antarmuka ini menawarkan perintah `parse`, `graph show`, `graph analyze`, `graph hotspots`, dan `fingerprint` beserta opsi-opsi mereka.

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| Argumen       | Wajib | Default      | Deskripsi                                                              |
| -------------- | -------- | ------------ | ------------------------------------------------------------------------ |
| `input-path`   | Ya       | —            | Direktori sumber atau file sumber tunggal.                              |
| `output-dir`   | Tidak    | `./skeleton` | Direktori keluaran. Tidak digunakan saat `--output-mode` adalah `inline`. |

| Flag                    | Tipe                | Default            | Nilai yang Diizinkan                                   | Deskripsi                                                                                         |
| ----------------------- | ------------------- | ------------------ | ------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`   | `inline`, `directory-flat`, `directory-tree`     | Tempat dan cara keluaran ditulis.                                                                  |
| `--format`              | string              | `skt`              | `skt`, `json`                                    | Format keluaran.                                                                                   |
| `--max-lines`           | int                 | `500`              | Untuk `directory-tree`, 3 atau lebih besar       | Target baris untuk keluaran datar. Batas keras potongan sumber untuk keluaran pohon.               |
| `--collect-test`        | bool                | `false`            | —                                                | Memasukkan file uji.                                                                              |
| `--minify`              | bool                | `false`            | —                                                | Mengaktifkan kompresi keluaran berbasis kamus.                                                     |
| `--edges`               | bool                | `false`            | —                                                | Memasukkan bagian `[edges]`. Nonaktif secara default untuk menghemat token.                        |
| `--clean`               | bool                | `false`            | —                                                | Mengganti keluaran yang dihasilkan sebelumnya, hanya setelah pembuatan berhasil.                   |
| `--workers`             | int                 | `0` (`NumCPU`)     | `0` atau lebih besar                             | Jumlah maksimum goroutine parsing konkuren.                                                        |
| `--verbose`             | bool                | `false`            | —                                                | Menampilkan informasi progres.                                                                     |
| `--header-language`     | string              | `c`                | `c`, `cpp`                                       | Parser yang digunakan untuk file `.h`.                                                             |
| `--include GLOB`        | string, dapat diulang | Semua sumber yang didukung | —                                             | Memasukkan jalur sumber yang cocok.                                                                |
| `--exclude GLOB`        | string, dapat diulang | Tidak ada          | —                                                | Mengecualikan jalur atau direktori yang cocok.                                                     |
| `--max-file-size BYTES` | int                 | `10485760`         | `0` menonaktifkan batas                           | Melewati file sumber yang lebih besar.                                                             |
| `--strict`              | bool                | `false`            | —                                                | Gagal sebelum keluaran jika terjadi diagnosa pemindaian atau parsing.                              |

### Mode keluaran

| Nilai             | Perilaku                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| `inline`          | Menulis ke stdout. Tidak ada direktori keluaran yang digunakan.                                  |
| `directory-flat`  | Menulis file terpotong ke direktori keluaran. Blok sumber tetap utuh, sehingga `--max-lines` adalah target pemisahan daripada batas keras. |
| `directory-tree`  | Mencerminkan pohon sumber. Nama mempertahankan ekstensi sumber (`foo.c.skt`, `foo.h.skt`). File terpisah menggunakan `foo.c_part1.skt`. |

### Perilaku keluaran

- Tanpa `--clean`, keluaran yang dihasilkan yang ada untuk format yang dipilih di direktori keluaran menyebabkan kesalahan.
- Dengan `--clean`, file `.skt` yang usang (format SKT) atau `codeknit.json` (format JSON) diganti hanya setelah pembuatan berhasil. Entri yang tidak terkait tetap dipertahankan.
- Dalam format JSON, mode direktori menulis `codeknit.json`. JSON adalah dokumen tunggal dan tidak dipisah berdasarkan jumlah baris.
- `dict.skt` dan `warnings.skt` adalah nama file bantu yang dicadangkan.
- Jangan gunakan direktori sumber atau salah satu leluhurnya sebagai direktori keluaran.

Lihat [Mode keluaran](../../guides/output-modes/) dan [referensi format keluaran](../output-format/) untuk detailnya.

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| Argumen       | Wajib | Deskripsi                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Ya       | Direktori sumber atau file sumber tunggal. |

| Flag                    | Tipe                | Default                            | Nilai yang Diizinkan          | Deskripsi                                                       |
| ----------------------- | ------------------- | ---------------------------------- | ----------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`   | Jalur file              | File HTML keluaran.                                               |
| `--collect-test`        | bool                | `false`                            | —                       | Memasukkan file uji.                                              |
| `--workers`             | int                 | `0` (`NumCPU`)                     | `0` atau lebih besar    | Jumlah maksimum goroutine parsing konkuren.                       |
| `--verbose`             | bool                | `false`                            | —                       | Menampilkan informasi progres.                                    |
| `--header-language`     | string              | `c`                                | `c`, `cpp`              | Parser yang digunakan untuk file `.h`.                            |
| `--include GLOB`        | string, dapat diulang | Semua sumber yang didukung         | —                       | Memasukkan jalur sumber yang cocok.                               |
| `--exclude GLOB`        | string, dapat diulang | Tidak ada                          | —                       | Mengecualikan jalur atau direktori yang cocok.                    |
| `--max-file-size BYTES` | int                 | `10485760`                         | `0` menonaktifkan batas | Melewati file sumber yang lebih besar.                            |
| `--strict`              | bool                | `false`                            | —                       | Gagal sebelum keluaran jika terjadi diagnosa pemindaian atau parsing. |

Perintah ini menulis file HTML mandiri dan membukanya di peramban default. Penampil menggambar hubungan yang terselesaikan dan menampilkan jumlah hubungan yang tidak pasti.

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| Argumen       | Wajib | Deskripsi                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Ya       | Direktori sumber atau file sumber tunggal. |

| Flag                       | Tipe                | Default                            | Nilai yang Diizinkan          | Deskripsi                                                  |
| -------------------------- | ------------------- | ---------------------------------- | ----------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`    | Jalur file              | File `.skt` keluaran.                                        |
| `--collect-test`           | bool                | `false`                            | —                       | Memasukkan file uji.                                         |
| `--workers`                | int                 | `0` (`NumCPU`)                     | `0` atau lebih besar    | Jumlah maksimum goroutine parsing konkuren.                  |
| `--verbose`                | bool                | `false`                            | —                       | Menampilkan informasi progres.                               |
| `--fan-threshold`          | int                 | `10`                               | —                       | Minimum fan-in atau fan-out untuk menandai simbol hub.       |
| `--god-threshold`          | int                 | `15`                               | —                       | Jumlah sisi `contains` minimum untuk menandai god class atau fungsi. |
| `--max-inheritance-depth`  | int                 | `5`                                | —                       | Menandai rantai pewarisan yang lebih dalam dari nilai ini.   |
| `--top-n`                  | int                 | `30`                               | `0` = tanpa batas       | Membatasi bagian keluaran yang diberi peringkat.             |
| `--betweenness-threshold`  | float               | `0.001`                            | —                       | Nilai sentralitas betweenness minimum untuk dilaporkan.      |
| `--propagation-cutoff`     | float               | `0.05`                             | —                       | Probabilitas minimum untuk melanjutkan propagasi perubahan.  |
| `--header-language`        | string              | `c`                                | `c`, `cpp`              | Parser yang digunakan untuk file `.h`.                       |
| `--include GLOB`           | string, dapat diulang | Semua sumber yang didukung         | —                       | Memasukkan jalur sumber yang cocok.                          |
| `--exclude GLOB`           | string, dapat diulang | Tidak ada                          | —                       | Mengecualikan jalur atau direktori yang cocok.               |
| `--max-file-size BYTES`    | int                 | `10485760`                         | `0` menonaktifkan batas | Melewati file sumber yang lebih besar.                       |
| `--strict`                 | bool                | `false`                            | —                       | Gagal sebelum keluaran jika terjadi diagnosa pemindaian atau parsing. |

Metrik dependensi mengecualikan hubungan yang tidak pasti dan metadata file/sertakan. Laporan mencatat jumlah hubungan yang dikecualikan.

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| Argumen       | Wajib | Deskripsi                                                  |
| -------------- | -------- | ------------------------------------------------------------ |
| `input-path`   | Ya       | Direktori sumber di dalam repositori Git.                    |

| Flag                       | Tipe                | Default                    | Nilai yang Diizinkan                              | Deskripsi                                                  |
| -------------------------- | ------------------- | -------------------------- | ------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`  | Jalur file                                   | File keluaran.                                               |
| `--format`                 | string              | `skt`                      | `skt`, `json`                               | Format keluaran.                                             |
| `--since`                  | string              | `12mo`                     | Jendela seperti `180d`, `12mo`, atau `2y`   | Jendela riwayat.                                             |
| `--max-commits`            | int                 | `2000`                     | —                                           | Jumlah maksimum commit yang diperiksa.                       |
| `--max-files-per-commit`   | int                 | `50`                       | —                                           | Mengecualikan commit yang menyentuh lebih banyak file daripada nilai ini. |
| `--min-cochanges`          | int                 | `3`                        | —                                           | Jumlah minimum commit bersama untuk melaporkan kopling temporal. |
| `--top-n`                  | int                 | `30`                       | —                                           | Hasil maksimum per bagian.                                   |
| `--include-merges`         | bool                | `false`                    | —                                           | Memasukkan merge commit dalam metrik riwayat.                |
| `--collect-test`           | bool                | `false`                    | —                                           | Memasukkan file uji.                                         |
| `--workers`                | int                 | `0` (`NumCPU`)             | `0` atau lebih besar                         | Jumlah maksimum goroutine parsing konkuren.                  |
| `--verbose`                | bool                | `false`                    | —                                           | Menampilkan informasi progres.                               |
| `--header-language`        | string              | `c`                        | `c`, `cpp`                                  | Parser yang digunakan untuk file `.h`.                       |
| `--include GLOB`           | string, dapat diulang | Semua sumber yang didukung | —                                           | Memasukkan jalur sumber yang cocok.                          |
| `--exclude GLOB`           | string, dapat diulang | Tidak ada                  | —                                           | Mengecualikan jalur atau direktori yang cocok.               |
| `--max-file-size BYTES`    | int                 | `10485760`                 | `0` menonaktifkan batas                     | Melewati file sumber yang lebih besar.                       |
| `--strict`                 | bool                | `false`                    | —                                           | Gagal sebelum keluaran jika terjadi diagnosa pemindaian atau parsing. |

Perintah ini menggabungkan frekuensi commit, churn baris, dan kebaruan dengan PageRank tingkat file, fan-in transitif, dan sentralitas betweenness. Perintah ini juga melaporkan kopling temporal antara file yang berubah bersama. Merge commit dikecualikan secara default.

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| Argumen       | Wajib | Deskripsi                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Ya       | Direktori sumber atau file sumber tunggal. |

| Flag                    | Tipe                | Default                          | Nilai yang Diizinkan          | Deskripsi                                                                                                      |
| ----------------------- | ------------------- | -------------------------------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`    | Jalur file              | File `.skt` keluaran.                                                                                           |
| `--min-similarity`      | int                 | `65`                             | `0`–`100`               | Persentase kemiripan minimum untuk dilaporkan.                                                                  |
| `--max-similarity`      | int                 | `95`                             | `0`–`100`               | Persentase kemiripan maksimum untuk dilaporkan.                                                                 |
| `--show-all`            | bool                | `false`                          | —                       | Juga mengeluarkan bagian `[fingerprints]` dengan data token mentah.                                            |
| `--rerank`              | bool                | `false`                          | —                       | Menemukan tetangga semantik dan mengurutkan ulang kandidat dengan embedding Ollama. Membutuhkan `ollama serve` dan `ollama pull qwen3-embedding:0.6b`. |
| `--model`               | string              | `qwen3-embedding:0.6b`           | Nama model Ollama       | Model embedding Ollama yang digunakan dengan `--rerank`.                                                        |
| `--collect-test`        | bool                | `false`                          | —                       | Memasukkan file uji.                                                                                            |
| `--workers`             | int                 | `0` (`NumCPU`)                   | `0` atau lebih besar    | Jumlah maksimum goroutine parsing konkuren.                                                                     |
| `--verbose`             | bool                | `false`                          | —                       | Menampilkan informasi progres.                                                                                  |
| `--header-language`     | string              | `c`                              | `c`, `cpp`              | Parser yang digunakan untuk file `.h`.                                                                          |
| `--include GLOB`        | string, dapat diulang | Semua sumber yang didukung      | —                       | Memasukkan jalur sumber yang cocok.                                                                             |
| `--exclude GLOB`        | string, dapat diulang | Tidak ada                       | —                       | Mengecualikan jalur atau direktori yang cocok.                                                                  |
| `--max-file-size BYTES` | int                 | `10485760`                       | `0` menonaktifkan batas | Melewati file sumber yang lebih besar.                                                                          |
| `--strict`              | bool                | `false`                          | —                       | Gagal sebelum keluaran jika terjadi diagnosa pemindaian atau parsing.                                           |

### Rentang kemiripan

Rentang default, 65–95%, mengecualikan kecocokan 100%. Untuk melaporkan kecocokan token ternormalisasi yang tepat, atur kedua batas ke `100`:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Kandidat token ternormalisasi yang tepat bersifat lengkap. Dengan 500 atau lebih simbol yang memenuhi syarat, kandidat hampir duplikat lainnya menggunakan pengambilan tetangga top-50 yang perkiraan. `--rerank` dapat menghapus kandidat yang secara struktural tepat. Lihat [Keandalan Analisis](../analysis-reliability/#jaminan-deteksi-duplikat) untuk detailnya.

## `codeknit completion`

```bash
codeknit completion <shell>
```

| Argumen  | Wajib | Nilai yang Diizinkan                         | Deskripsi                                   |
| --------- | -------- | -------------------------------------- | --------------------------------------------- |
| `shell`   | Ya       | `bash`, `zsh`, `fish`, `powershell`    | Shell tempat skrip penyelesaian akan dibuat. |

Skrip ditulis ke stdout.

```bash
codeknit completion zsh >> ~/.zshrc
```

## Halaman terkait

- [Referensi format keluaran](../output-format/)
- [Dukungan Bahasa dan Fidelity API](../language-support/)
- [Keandalan Analisis](../analysis-reliability/)