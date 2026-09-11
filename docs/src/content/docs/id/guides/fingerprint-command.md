---
title: Perintah Fingerprint
description: Deteksi kode duplikat dan hampir duplikat di seluruh file dan bahasa menggunakan fuzzy hashing.
---

Perintah `codeknit fingerprint` mendeteksi kode duplikat dan **hampir duplikat** di seluruh basis kode Anda menggunakan **Context-Triggered Piecewise Hashing (CTPH)**. Perintah ini bekerja lintas file dan bahkan lintas bahasa pemrograman dengan menormalisasi nama variabel, literal string, dan anotasi tipe sebelum menghitung **fingerprint struktural**.

## Apa yang dilakukan

`codeknit fingerprint` menganalisis setiap fungsi, metode, variabel, dan tipe dalam basis kode Anda dan menghitung **fingerprint struktural yang dinormalisasi** berdasarkan:

- Alur kontrol (`if`, `for`, `while`, `switch`)
- Operasi (`=`, `+`, `==`, `&&`, `||`)
- Panggilan, pengembalian, penugasan, dan pembuatan objek
- Konstruksi bahasa seperti `try/catch`, `yield`, `await`, `defer`

Normalisasi ini berarti bahwa **salin-tempel yang diubah namanya**, **refaktor trivial**, dan **logika yang setara dalam bahasa berbeda** masih dapat dideteksi sebagai duplikat.

Algoritma ini menggunakan **CTPH** (variasi rolling hash) untuk menemukan **hampir duplikat** secara efisien. Kode yang mirip menghasilkan fingerprint yang mirip, memungkinkan pencocokan fuzzy bahkan ketika kode telah sedikit dimodifikasi.

## Penggunaan dasar

```bash
codeknit fingerprint ./src
```

Perintah ini:

- Mengurai semua file sumber di `./src`
- Menghitung fingerprint struktural
- Mengeluarkan hasil ke `./skeleton/fingerprints.skt`
- Melaporkan kecocokan dengan **kemiripan** antara **65% dan 95%** (rentang default)

## Flag

| Flag               | Default                       | Deskripsi                                                                                                                                                |
| ------------------ | ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Jalur file `.skt` keluaran                                                                                                                                |
| `--min-similarity` | `65`                          | Persentase **kemiripan** minimum untuk dilaporkan (0–100)                                                                                                |
| `--max-similarity` | `95`                          | Persentase **kemiripan** maksimum untuk dilaporkan (0–100)                                                                                               |
| `--show-all`       | `false`                       | Sertakan bagian `[fingerprints]` dengan data token mentah                                                                                                |
| `--rerank`         | `false`                       | Temukan tetangga semantik dan urutkan ulang kandidat menggunakan embedding Ollama (memerlukan: `ollama serve` dan `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Model embedding Ollama yang digunakan dengan `--rerank`                                                                                                  |
| `--collect-test`   | `false`                       | Sertakan file pengujian dalam analisis                                                                                                                   |
| `--workers`        | `NumCPU`                      | Jumlah maksimum goroutine penguraian konkuren (0 = gunakan semua inti CPU)                                                                              |
| `--verbose`        | `false`                       | Cetak informasi progres selama pemrosesan                                                                                                                |

## Format keluaran

Keluaran berupa file `.skt` dengan bagian-bagian berikut:

### `[duplicates]` (selalu ada)

Mencantumkan pasangan simbol dengan **kemiripan** di atas ambang batas:

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

Setiap baris menampilkan:

- Persentase **kemiripan**
- Simbol kiri (jalur file, cakupan, nama)
- Simbol kanan (jalur file, cakupan, nama)

### `[fingerprints]` (hanya dengan `--show-all`)

Berisi data fingerprint mentah untuk setiap simbol:

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Field:

- Nama simbol
- `FP:<version>:<hash1>:<hash2>` — fingerprint CTPH
- `tokens:<hex>` — aliran token tubuh yang dinormalisasi

Bagian ini berguna untuk debugging atau membangun alat downstream.

## Pola umum

```bash
# Default scan
codeknit fingerprint ./src
```

```bash
# Find only exact duplicates
codeknit fingerprint ./src --min-similarity 100
```

```bash
# Find moderately similar code (e.g. same algorithm, different names)
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

```bash
# Use semantic matching to find additional candidates and reduce false positives
# Requires: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank
```

```bash
# Use a different embedding model for semantic matching
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

```bash
# Output full fingerprint listing (for analysis tools)
codeknit fingerprint ./src --show-all
```

```bash
# Custom output file
codeknit fingerprint ./src -o duplicates.skt
```

## Memilih rentang kemiripan

| Rentang  | Panduan                                                                                     |
| -------- | ------------------------------------------------------------------------------------------- |
| 96–100%  | Duplikat struktural eksak atau hampir eksak. Hampir pasti salin-tempel.                     |
| 85–95%   | **Hampir duplikat**. Biasanya salin-tempel dengan perubahan kecil (misal variabel diganti nama, log ditambahkan). |
| 65–84%   | Rentang default. **Kemiripan** struktural yang kuat. Kandidat bagus untuk refaktor.         |
| 50–64%   | **Kemiripan** sedang. Bentuk algoritmik sama tetapi detail berbeda. Tinjau secara manual.   |
| < 50%    | Biasanya noise. Bukan duplikasi yang bermakna.                                              |

## Tips

- **Fingerprint mengukur struktur, bukan makna**: Skor **kemiripan** tinggi berarti kode _terlihat_ mirip, bukan berarti _melakukan_ hal yang sama. Selalu tinjau kedua simbol.
- **Gunakan `--rerank` untuk pencocokan semantik**: Embedding menambahkan tetangga semantik yang mungkin terlewatkan oleh pengambilan struktural dan memfilter kandidat yang tidak setuju secara semantik.
- **Tubuh pendek dilewati**: Simbol dengan kurang dari 4 token yang dinormalisasi (misal getter sederhana) diabaikan untuk menghindari noise.
- **Pencocokan lintas bahasa berfungsi**: Konstruksi yang setara (misal, fungsi Python dan fungsi Go dengan logika yang sama) dapat cocok, tetapi pola spesifik bahasa dapat menghasilkan kecocokan **kemiripan** rendah yang palsu.
- **Kecocokan adalah sinyal, bukan keputusan**: Anggap setiap kecocokan sebagai dorongan untuk menyelidiki — bukan bukti otomatis duplikasi.