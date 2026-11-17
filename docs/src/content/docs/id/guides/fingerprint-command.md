---
title: Perintah Fingerprint
description: Deteksi kode duplikat dan hampir duplikat di berbagai file dan bahasa menggunakan fuzzy hashing.
---

`codeknit fingerprint` menghitung Context-Triggered Piecewise Hash (CTPH) dari token tubuh yang dinormalisasi untuk setiap fungsi, metode, variabel, dan tipe. Nama variabel, literal string, dan anotasi tipe dinormalisasi sebelum proses hashing. Salinan yang diubah namanya dan porting ke bahasa lain tetap dapat cocok.

Panduan ini menunjukkan cara menemukan duplikat, menyesuaikan rentang kemiripan, mengurangi positif palsu, dan menentukan apa yang perlu direfaktor.

## Sebelum memulai

- `codeknit` telah terpasang. Lihat [Instalasi](../../installation/).
- Anda memiliki direktori sumber untuk dipindai, seperti `./src`.

## Temukan duplikat dalam proyek

Jalankan perintah pada direktori yang ingin diperiksa:

```bash
codeknit fingerprint ./src
```

Perintah akan menulis `./skeleton/fingerprints.skt` dan melaporkan pasangan dengan kemiripan antara 65% hingga 95%.

Buka file tersebut. Bagian `[duplicates]` selalu ada, dengan kecocokan terkuat ditampilkan pertama:

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

Setiap baris memiliki format `similarity:<N>%  <file>::<symbol> <-> <file>::<symbol>`. Jika tidak ada yang masuk dalam rentang, bagian tersebut hanya berisi `# no duplicates found`.

Untuk tata bahasa keluaran lengkap, lihat [referensi format keluaran](../../reference/output-format/).

## Temukan duplikat eksak saja

Rentang default memiliki batas atas 95%, sehingga kecocokan 100% dikecualikan. Untuk menampilkan kecocokan eksak, atur kedua batas menjadi 100:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

"Eksak" berarti aliran token yang dinormalisasi sepenuhnya identik dalam kategori simbol yang sama. Ini tidak berarti teks sumber identik, atau bahwa kedua simbol terbukti setara.

## Perluas rentang untuk menemukan kecocokan yang lebih longgar

Turunkan batas minimum untuk mencakup kode yang memiliki bentuk algoritmik serupa tetapi berbeda dalam detail:

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

Hasil di bawah 50% biasanya hanya noise.

## Kurangi positif palsu dengan reranking semantik

Jika hasil struktural terlalu bising, urutkan ulang kandidat dengan embedding dari server Ollama lokal.

1. Jalankan Ollama dan unduh model embedding default:

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. Jalankan perintah dengan `--rerank`:

   ```bash
   codeknit fingerprint ./src --rerank
   ```

Untuk menggunakan model embedding lain, berikan `--model`:

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

Reranking menambahkan kandidat semantik dan memfilter berdasarkan skor gabungan. Ini mengurangi positif palsu tetapi tidak menghilangkan semuanya. Reranking juga dapat menghapus kandidat yang secara struktural eksak, dan rentang kemiripan tetap berlaku. Embedding dihitung dari snapshot sumber yang sama yang digunakan untuk ekstraksi.

## Periksa fingerprint mentah

Untuk melihat fingerprint setiap simbol, tambahkan `--show-all`:

```bash
codeknit fingerprint ./src --show-all
```

Ini menambahkan bagian `[fingerprints]`, dikelompokkan berdasarkan file sumber, satu baris per simbol:

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Format baris adalah `<symbol-name>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`. `tokens` adalah aliran token yang dinormalisasi dalam bentuk hex.

## Tulis laporan di lokasi lain

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## Kendalikan file yang dipindai

File pengujian dilewati secara default. Sertakan dengan `--collect-test`:

```bash
codeknit fingerprint ./src --collect-test
```

Untuk mempersempit pemindaian, gunakan flag pemilihan sumber. Gunakan tanda kutip agar shell tidak mengembangkan pola:

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

Kontrol pemilihan lainnya:

- `--max-file-size BYTES` melewati file yang lebih besar. Default adalah `10485760` (10 MiB). `0` menonaktifkan batas.
- `--strict` gagal sebelum menulis keluaran jika pemindaian atau parsing tidak lengkap.
- `--header-language c|cpp` memilih parser untuk file `.h` yang ambigu. Lihat [Dukungan Bahasa](../../reference/language-support/#header-language-selection).

Symlink yang dilewati, file yang terlalu besar, dan peringatan sintaks muncul di stderr. Include tidak mengesampingkan `.gitignore` atau penyaringan pengujian, dan exclude lebih diutamakan daripada include. Lihat [referensi keandalan analisis](../../reference/analysis-reliability/) untuk jaminan yang tepat.

## Pilih rentang kemiripan

| Rentang     | Apa yang biasanya ditemukan                                      |
| ----------- | ---------------------------------------------------------------- |
| `96-100%`   | Duplikat struktural eksak atau hampir eksak; salinan jelas       |
| `85-95%`    | Hampir duplikat; biasanya salinan dengan perubahan sepele        |
| `65-84%`    | Rentang default; kemiripan struktural kuat, sering kandidat refaktor |
| `50-64%`    | Bentuk algoritmik yang sama dengan detail berbeda               |
| Di bawah 50%| Biasanya noise                                                   |

Untuk melihat dua baris teratas, naikkan `--max-similarity` menjadi `100`.

## Tentukan apa yang perlu direfaktor

1. Mulailah dengan rentang default dan baca pasangan teratas.
2. Buka kedua simbol dari setiap pasangan dan bandingkan.
3. Anggap kecocokan sebagai sinyal, bukan bukti. Fingerprint mengukur bentuk struktural, bukan nama pengenal atau maksud. Logika yang sama yang ditulis secara iteratif dan rekursif mendapat skor rendah, dan kode yang tidak terkait dengan bentuk yang sama dapat mendapat skor tinggi.
4. Pasangan dalam rentang 85–95% biasanya merupakan kandidat teraman untuk digabungkan. Pasangan dalam rentang 65–84% sering memerlukan pemeriksaan lebih lanjut.
5. Jika daftar terlalu bising, persempit rentang atau tambahkan `--rerank`.

Kecocokan lintas bahasa berlaku untuk konstruksi yang setara. Fitur khusus bahasa dapat menghasilkan kecocokan kemiripan rendah yang palsu.

Persentase kemiripan adalah heuristik peringkat, bukan jarak edit. Fixture kalibrasi berlabel proyek melaporkan presisi dan recall secara terpisah dari skor kemiripan. Ini adalah pemeriksaan regresi, bukan estimasi akurasi pada kode Anda.

## Proyek besar

- Di bawah 500 simbol yang memenuhi syarat, `codeknit` memberi skor setiap pasangan dalam kategori yang sama.
- Pada 500 atau lebih, grup token yang dinormalisasi secara eksak selalu dipertahankan. Pasangan lain menggunakan tetangga top-50 per simbol, sehingga hasil hampir duplikat dapat tidak lengkap.
- Untuk duplikat eksak, gunakan perintah 100/100 di atas. Ini tidak terpengaruh oleh aproksimasi.
- Pengurutan bersifat deterministik untuk input dan skor yang stabil.

## Pemecahan masalah

| Gejala | Penyebab dan solusi |
| ------ | ------------------- |
| `# no duplicates found` | Tidak ada pasangan yang masuk dalam rentang. Turunkan `--min-similarity` atau naikkan `--max-similarity`. |
| Fungsi pendek yang diketahui hilang | Simbol dengan kurang dari 4 token yang dinormalisasi dilewati. |
| Salinan eksak hilang | Maksimum default adalah 95%. Gunakan `--min-similarity 100 --max-similarity 100`. Dengan `--rerank`, filter semantik juga dapat menghapusnya. |
| `--rerank` gagal | Pastikan `ollama serve` berjalan dan model telah diunduh. |
| Hasil parsial dengan peringatan di stderr | Perbaiki file yang dilaporkan, atau tambahkan `--strict` untuk gagal daripada menerbitkan keluaran yang tidak lengkap. |

## Flag yang digunakan dalam panduan ini

| Flag               | Default                       | Deskripsi                                                                 |
| ------------------ | ----------------------------- | ------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Jalur file `.skt` keluaran.                                               |
| `--min-similarity` | `65`                          | Persentase kemiripan minimum untuk dilaporkan (0–100).                    |
| `--max-similarity` | `95`                          | Persentase kemiripan maksimum untuk dilaporkan (0–100).                   |
| `--show-all`       | `false`                       | Juga keluarkan bagian `[fingerprints]` dengan data token mentah.          |
| `--rerank`         | `false`                       | Urutkan ulang kandidat dengan embedding semantik melalui Ollama. Memerlukan `ollama serve` dan `ollama pull qwen3-embedding:0.6b`. |
| `--model`          | `qwen3-embedding:0.6b`        | Model embedding Ollama yang digunakan dengan `--rerank`.                  |
| `--collect-test`   | `false`                       | Sertakan file pengujian dalam analisis.                                   |
| `--workers`        | `NumCPU`                      | Jumlah maksimum goroutine parsing konkuren.                               |
| `--verbose`        | `false`                       | Cetak informasi progres.                                                  |

Untuk daftar lengkap, termasuk `--include`, `--exclude`, `--max-file-size`, `--strict`, dan `--header-language`, lihat [referensi flag CLI](../../reference/cli-flags/).

## Terkait

- [Perintah Parse](../parse-command/)
- [Perintah Graf](../graph-commands/)