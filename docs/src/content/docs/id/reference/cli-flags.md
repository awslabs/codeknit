---
title: Referensi CLI
description: Referensi lengkap untuk semua perintah dan flag codeknit.
---

Semua perintah pemrosesan sumber menerima `--header-language c|cpp` untuk file `.h`. Default-nya adalah `c`; pilih `cpp` untuk header C++ yang diberi nama `.h`. Ekstensi lainnya tidak terpengaruh. Lihat [dukungan bahasa](../id/language-support/) untuk kebijakan pemilihan dan kemampuan ekstraksi yang telah diuji.

## codeknit

Meluncurkan antarmuka terminal interaktif (TUI), yang memandu Anda melalui perintah dan opsi yang tersedia.

```bash
codeknit
```

## codeknit parse

Ekstrak informasi struktural dari kode sumber ke dalam file `.skt` atau JSON.

```bash
codeknit parse <input-path> [output-dir]
```

| Flag             | Tipe   | Default          | Deskripsi                                                                                     |
| ---------------- | ------ | ---------------- | --------------------------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat` | Mode keluaran: `inline`, `directory-flat`, atau `directory-tree`                              |
| `--format`       | string | `skt`            | Format keluaran: `skt` atau `json`                                                            |
| `--max-lines`    | int    | `500`            | Jumlah baris maksimum per file keluaran (berlaku untuk mode `directory-flat` dan `directory-tree`) |
| `--collect-test` | bool   | `false`          | Sertakan file pengujian dalam analisis                                                        |
| `--minify`       | bool   | `false`          | Aktifkan minimisasi keluaran berbasis kamus                                                   |
| `--edges`        | bool   | `false`          | Sertakan bagian `[edges]` dalam keluaran (nonaktif secara default untuk menghemat token)      |
| `--clean`        | bool   | `false`          | Hapus file `.skt` yang usang dari direktori keluaran sebelum menulis                          |
| `--workers`      | int    | `0` (NumCPU)     | Jumlah maksimum goroutine parsing bersamaan                                                   |
| `--verbose`      | bool   | `false`          | Tampilkan informasi progres selama pemrosesan                                                 |

Direktori keluaran default ke `./skeleton` jika tidak ditentukan. Dalam mode `inline`, keluaran ditulis ke stdout dan tidak ada direktori yang digunakan. Dengan `--format json`, keluaran direktori ditulis sebagai `codeknit.json`.

## codeknit graph show

Hasilkan visualisasi graf HTML interaktif dari struktur basis kode.

```bash
codeknit graph show <input-path>
```

| Flag             | Tipe   | Default                          | Deskripsi                                  |
| ---------------- | ------ | -------------------------------- | ------------------------------------------ |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | Jalur file HTML keluaran                   |
| `--collect-test` | bool   | `false`                          | Sertakan file pengujian dalam analisis     |
| `--workers`      | int    | `0` (NumCPU)                     | Jumlah maksimum goroutine parsing bersamaan|
| `--verbose`      | bool   | `false`                          | Tampilkan informasi progres selama pemrosesan |

File HTML yang dihasilkan bersifat mandiri dan terbuka secara otomatis di browser default Anda.

## codeknit graph analyze

Jalankan algoritma analisis struktural dan hasilkan laporan `.skt` yang dapat dibaca oleh LLM.

```bash
codeknit graph analyze <input-path>
```

| Flag                      | Tipe    | Default                         | Deskripsi                                                   |
| ------------------------- | ------- | ------------------------------- | ----------------------------------------------------------- |
| `-o`, `--output`          | string  | `./skeleton/graph_analysis.skt` | Jalur file `.skt` keluaran                                  |
| `--collect-test`          | bool    | `false`                         | Sertakan file pengujian dalam analisis                      |
| `--workers`               | int     | `0` (NumCPU)                    | Jumlah maksimum goroutine parsing bersamaan                 |
| `--verbose`               | bool    | `false`                         | Tampilkan informasi progres selama pemrosesan               |
| `--fan-threshold`         | int     | `10`                            | Jumlah minimum fan-in atau fan-out untuk menandai simbol hub|
| `--god-threshold`         | int     | `15`                            | Jumlah minimum sisi contains untuk menandai god class/function |
| `--max-inheritance-depth` | int     | `5`                             | Tandai rantai pewarisan yang lebih dalam dari nilai ini     |
| `--top-n`                 | int     | `30`                            | Batasi bagian keluaran yang diberi peringkat; `0` berarti tanpa batas |
| `--betweenness-threshold` | float64 | `0.001`                         | Nilai minimum betweenness centrality untuk dilaporkan       |
| `--propagation-cutoff`    | float64 | `0.05`                          | Probabilitas minimum untuk melanjutkan simulasi propagasi perubahan |

## codeknit graph hotspots

Urutkan file menggunakan riwayat Git dan kepentingan struktural, serta laporkan
keterkaitan temporal antara file yang berulang kali berubah bersama.

```bash
codeknit graph hotspots <input-path>
```

| Flag                     | Tipe   | Default                   | Deskripsi                                     |
| ------------------------ | ------ | ------------------------- | --------------------------------------------- |
| `-o`, `--output`         | string | `./skeleton/hotspots.skt` | Jalur file keluaran                           |
| `--format`               | string | `skt`                     | Format keluaran: `skt` atau `json`            |
| `--since`                | string | `12mo`                    | Jendela riwayat, seperti `180d`, `12mo`, atau `2y` |
| `--max-commits`          | int    | `2000`                    | Jumlah maksimum commit yang diperiksa        |
| `--max-files-per-commit` | int    | `50`                      | Kecualikan commit yang mengubah lebih banyak file |
| `--min-cochanges`        | int    | `3`                       | Jumlah minimum commit bersama untuk keterkaitan temporal |
| `--top-n`                | int    | `30`                      | Jumlah maksimum hasil per bagian laporan      |
| `--include-merges`       | bool   | `false`                   | Sertakan merge commit                         |
| `--collect-test`         | bool   | `false`                   | Sertakan file pengujian                       |
| `--workers`              | int    | `0` (NumCPU)              | Jumlah maksimum goroutine parsing bersamaan   |
| `--verbose`              | bool   | `false`                   | Tampilkan informasi progres                   |

## codeknit fingerprint

Deteksi duplikat dan hampir duplikat kode menggunakan fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Flag               | Tipe   | Default                       | Deskripsi                                                                                                                            |
| ------------------ | ------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `-o`, `--output`   | string | `./skeleton/fingerprints.skt` | Jalur file `.skt` keluaran                                                                                                           |
| `--min-similarity` | int    | `65`                          | Persentase kemiripan minimum untuk dilaporkan (0–100)                                                                                |
| `--max-similarity` | int    | `95`                          | Persentase kemiripan maksimum untuk dilaporkan (0–100)                                                                              |
| `--show-all`       | bool   | `false`                       | Sertakan bagian `[fingerprints]` dengan data token mentah                                                                            |
| `--rerank`         | bool   | `false`                       | Temukan tetangga semantik dan urut ulang kandidat menggunakan embedding Ollama (memerlukan `ollama serve` dan `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | string | `qwen3-embedding:0.6b`        | Model embedding Ollama yang digunakan dengan `--rerank`                                                                             |
| `--collect-test`   | bool   | `false`                       | Sertakan file pengujian dalam analisis                                                                                               |
| `--workers`        | int    | `0` (NumCPU)                  | Jumlah maksimum goroutine parsing bersamaan                                                                                          |
| `--verbose`        | bool   | `false`                       | Tampilkan informasi progres selama pemrosesan                                                                                        |

## codeknit completion

Hasilkan skrip penyelesaian shell untuk shell yang didukung.

```bash
codeknit completion <shell>
```

Shell yang didukung: `bash`, `zsh`, `fish`, `powershell`.

## Flag global

| Flag           | Deskripsi                       |
| -------------- | ------------------------------- |
| `--version`    | Tampilkan informasi versi       |
| `--help`, `-h` | Tampilkan bantuan untuk perintah saat ini |