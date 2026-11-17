---
title: Keandalan Analisis
description: Pemilihan sumber, analisis tidak lengkap, publikasi keluaran, dan jaminan deteksi duplikat.
---

<!-- Dipertahankan di docs/reference/analysis-reliability.md; disalin verbatim oleh docs/gen. -->

Perilaku ini menjelaskan perubahan Go/CLI yang belum dirilis setelah versi 0.5.0.

## Pemilihan sumber dan diagnostik

Kelima perintah pemrosesan sumber (`parse`, `fingerprint`, `graph show`, `graph analyze`, dan `graph hotspots`) menerima:

| Flag                    | Default               | Arti                                                             |
| ----------------------- | --------------------- | ---------------------------------------------------------------- |
| `--include GLOB`        | Semua sumber yang didukung | Sertakan jalur sumber yang cocok; ulangi untuk memilih beberapa pola. |
| `--exclude GLOB`        | Tidak ada             | Kecualikan jalur atau direktori yang cocok; ulangi untuk beberapa pola. |
| `--max-file-size BYTES` | `10485760` (10 MiB)   | Lewati file sumber yang lebih besar; `0` menonaktifkan batas ini. |
| `--strict`              | `false`               | Gagal sebelum mengeluarkan hasil jika terjadi diagnostik pemindaian atau parsing. |

Pola menggunakan glob gaya gitignore yang dipisahkan oleh garis miring, relatif terhadap direktori masukan yang dipilih. Untuk masukan berkas tunggal, pola cocok dengan basename-nya. Kutip pola agar shell tidak mengembangkannya. Include membentuk gabungan; exclude lebih diutamakan. Include tidak mengesampingkan `.gitignore`, penyaringan tes, atau pengecualian bawaan. Pola negasi (`!pattern`) tidak diterima dalam flag ini; gunakan kontrol include dan exclude yang terpisah.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
codeknit fingerprint ./src --max-file-size 2097152 --include '**/*.ts'
codeknit parse ./src/app.go --output-mode inline --format json
```

Pemindaian direktori menghormati berkas `.gitignore` leluhur dan bersarang relatif terhadap lokasi proyek mereka, terlepas dari direktori kerja proses. Direktori `.git` terdekat atau penanda worktree menetapkan batas repositori. Di luar repositori, berkas ignore leluhur tetap berlaku. Subdirektori yang dipilih di bawah direktori yang diabaikan tetap diabaikan. Memilih secara eksplisit berkas sumber reguler melewati ignore dan penyaringan tes, tetapi tetap menerapkan batas include/exclude dan ukuran.

Symlink sumber, symlink direktori, dan sumber non-reguler dilewati dan dilaporkan. Pembacaan dibatasi pada root pemindaian, dan batas ukuran diberlakukan kembali saat membaca sehingga berkas yang tumbuh setelah ditemukan tidak dapat melewatinya. Pengecualian direktori bawaan mencakup `.git`, `.hg`, `.svn`, `node_modules`, cache Python, dan lingkungan virtual Python yang umum. Berkas tes tetap dikecualikan kecuali `--collect-test` disetel.

Direktori/berkas yang tidak dapat dibaca, pembacaan berkas ignore yang gagal, symlink yang dilewati, sumber yang terlalu besar, dan peringatan sintaks membuat analisis tidak lengkap. Secara default, hasil parsial yang dapat digunakan dipertahankan dan diagnostik muncul di stderr. JSON parse menyertakan catatan jalur/alasan dalam `errors`; SKT menyertakan `[errors]` secara inline atau di `warnings.skt`. Perintah graph/fingerprint juga melaporkan diagnostik di stderr. Berkas yang sengaja diabaikan atau difilter bukanlah peringatan. `--strict` menolak hasil tidak lengkap sebelum publikasi keluaran. Masukan tanpa sumber yang didukung yang dipilih, atau eksekusi di mana setiap pembacaan sumber gagal, selalu keluar dengan tidak berhasil.

TUI menggunakan kebijakan symlink yang sama dan default 10 MiB. Formulirnya belum menampilkan flag pemilihan dan strict yang baru.

## Nama keluaran dan publikasi

Keluaran tree menambahkan `.skt` ke nama sumber lengkap: `src/foo.c` menjadi `src/foo.c.skt`, sementara `src/foo.h` menjadi `src/foo.h.skt`. Keluaran split menggunakan `src/foo.c_part1.skt`, `src/foo.c_part2.skt`, dan seterusnya. Berkas kamus dan diagnostik mempertahankan nama yang dipesan `dict.skt` dan `warnings.skt`. Jalur keluaran yang bertentangan, termasuk tabrakan hanya huruf besar/kecil, gagal secara terlihat.

Tree SKT memerlukan `--max-lines` minimal 3. Setiap potongan simbol mengulangi header `[symbols]` dan `## source-path`. Potongan hanya sisi membawa `[edges]`. Setiap potongan sumber menghormati batas; berkas kamus dan diagnostik adalah keluaran bantu terpisah. Flat SKT menjaga blok sumber tetap utuh, sehingga batas barisnya adalah target pemisahan daripada batas keras untuk blok sumber besar. JSON adalah dokumen tunggal dan tidak dipisah berdasarkan jumlah baris.

Pembangkitan direktori selesai di direktori staging saudara. Tanpa `--clean`, keluaran yang dihasilkan yang ada untuk format yang dipilih menyebabkan kesalahan. Dengan `--clean`, berkas `.skt` usang (mode SKT) atau `codeknit.json` (mode JSON) diganti hanya setelah pembangkitan berhasil. Entri yang tidak terkait dipertahankan; tabrakan dengan satu menyebabkan kesalahan. Jangan gunakan direktori sumber atau leluhurnya sebagai direktori keluaran.

Publikasi memegang kunci khusus direktori, mencadangkan keluaran sebelumnya, dan memulihkannya jika penggantian gagal. Jika proses berhenti antara penggantian nama direktori, pemanggilan berikutnya memulihkan cadangan sebelum menghasilkan apa pun. Berkas `.NAME.codeknit.lock` saudara sengaja tetap ada untuk mengoordinasikan penerbit di masa depan; jangan hapus kunci aktif atau cadangan pemulihan. Ada interval singkat antara dua penggantian nama direktori, sehingga pembaca yang memerlukan akses tanpa gangguan harus berkoordinasi dengan pembangkitan. Ini menyediakan pemulihan kegagalan proses, bukan jaminan terhadap kegagalan penyimpanan atau kehilangan daya.

Laporan individual ditulis dan disinkronkan ke saudara sementara sebelum mengganti tujuan. Berkas yang ada tidak dipotong selama pembangkitan. Stdout inline tidak dapat dikembalikan jika konsumennya gagal.

## Jaminan deteksi duplikat

Pencocokan eksak berarti aliran token yang dinormalisasi lengkap identik dalam kategori simbol yang sama di antara deklarasi yang memenuhi syarat untuk fingerprinting. Ini tidak berarti teks sumber identik atau kesetaraan semantik yang terbukti. Grup kandidat eksak menggunakan aliran token itu sendiri, bukan kesetaraan fuzzy hash.

Di bawah 500 simbol yang memenuhi syarat, deteksi struktural membandingkan semua pasangan kategori yang sama. Pada 500 atau lebih, semua pasangan token eksak dipertahankan terlepas dari pengambilan kandidat, dan pasangan lainnya menggunakan sekitar 50 tetangga struktural teratas per simbol. Oleh karena itu, pengambilan hampir duplikat tidak lengkap pada masukan besar. Pengambilan tetap memindai semua vektor; ini tidak membuat algoritma keseluruhan subkuadratik.

Penyaringan rentang kemiripan masih berlaku. Rentang default adalah 65–95%, yang mengecualikan kecocokan 100%. Untuk menyertakan kecocokan eksak, gunakan:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Opsi `--rerank` menambahkan kandidat semantik, menerapkan kesepakatan kosinus, dan menyaring berdasarkan skor gabungan; ini dapat menghapus kandidat yang secara struktural eksak. Teks embedding berasal dari snapshot sumber yang sama yang digunakan untuk ekstraksi. Masukan stabil dan skor tetap menghasilkan pengurutan deterministik, termasuk seri; respons model itu sendiri adalah masukan eksternal.

`internal/emitter/calibration_test.go` mengevaluasi fixture sumber berlabel kecil pada ambang batas 70%, melaporkan positif/negatif benar dan salah, presisi, dan recall terpisah dari skor kemiripan. Fixture memiliki satu pasangan eksak, dua pasangan klon dekat, dan tiga pasangan tidak terkait, diuji di bawah dan melintasi batas 500 simbol. Ini adalah pemeriksaan regresi, bukan perkiraan representatif dari akurasi dunia nyata. Anggaran kalibrasi dan kinerja multibahasa yang lebih besar tetap menjadi pekerjaan roadmap.