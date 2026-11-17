---
title: Perintah Graf
description: Visualisasikan dan analisis struktur basis kode Anda dengan algoritma graf.
---

Gunakan subperintah `graph` untuk melihat bagaimana kode Anda terhubung, menemukan masalah struktural, dan menentukan file mana yang harus direfaktor terlebih dahulu. Panduan ini mengasumsikan Anda telah [menginstal codeknit](../../installation/) dan dapat menjalankan [`codeknit parse`](../parse-command/).

| Tugas | Perintah |
| --- | --- |
| Jelajahi graf secara visual | `codeknit graph show` |
| Temukan masalah struktural | `codeknit graph analyze` |
| Temukan file penting yang mudah berubah | `codeknit graph hotspots` |

## Visualisasikan graf kode

`graph show` mem-parsing masukan dan menulis file HTML mandiri. Simbol (fungsi, kelas, tipe) adalah node. Hubungan (panggilan, berisi, mengimplementasikan) adalah sisi.

```bash
codeknit graph show ./myproject
```

File ditulis ke `./skeleton/codeknit-graph.html` dan terbuka di peramban bawaan Anda.

Untuk memilih jalur keluaran atau menyertakan file pengujian:

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| Flag | Default | Tujuan |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Jalur file HTML keluaran |
| `--collect-test` | `false` | Sertakan file pengujian |
| `--workers` | `0` (`NumCPU`) | Jumlah maksimum goroutine parsing konkuren |
| `--verbose` | `false` | Cetak informasi progres |

### Periksa hubungan yang tidak terselesaikan

Pemirsa hanya menggambar hubungan yang terselesaikan. Ia menampilkan jumlah hubungan yang tidak pasti (`unresolved` atau `ambiguous`) tetapi tidak menggambarnya atau tautan kandidatnya.

Untuk melihat sisi yang tidak pasti secara individual, jalankan `parse` dengan `--edges`:

```bash
codeknit parse ./src --output-mode inline --edges
```

Tambahkan `--format json` untuk keluaran JSON. Lihat [referensi format keluaran](../../reference/output-format/) untuk cara sisi dan kandidat ditulis.

## Analisis struktur dan temukan masalah

`graph analyze` menjalankan 22 algoritma graf dan menulis laporan `.skt` yang dapat dibaca LLM.

```bash
codeknit graph analyze ./myproject
```

Laporan ditulis ke `./skeleton/graph_analysis.skt`. Baca langsung, atau berikan ke LLM bersama dengan sumber yang relevan.

### Sesuaikan laporan

Naikkan ambang batas untuk melihat lebih sedikit temuan, atau turunkan untuk melihat lebih banyak:

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0` menghapus batasan pada bagian yang diberi peringkat.

| Flag | Default | Tujuan |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | Jalur file `.skt` keluaran |
| `--collect-test` | `false` | Sertakan file pengujian |
| `--workers` | `0` (`NumCPU`) | Jumlah maksimum goroutine parsing konkuren |
| `--verbose` | `false` | Cetak informasi progres |
| `--fan-threshold` | `10` | Minimum fan-in atau fan-out untuk menandai simbol hub |
| `--god-threshold` | `15` | Jumlah minimum sisi berisi untuk menandai god class atau fungsi |
| `--max-inheritance-depth` | `5` | Tandai rantai pewarisan yang lebih dalam dari ini |
| `--top-n` | `30` | Batasi bagian yang diberi peringkat; `0` berarti tanpa batas |
| `--betweenness-threshold` | `0.001` | Minimum betweenness centrality untuk dilaporkan |
| `--propagation-cutoff` | `0.05` | Skor heuristik minimum untuk melanjutkan propagasi perubahan |

### Ketahui temuan mana yang menjawab pertanyaan mana

| Algoritma | Gunakan untuk menemukan |
| --- | --- |
| Dependensi siklik (Tarjan's SCC) | Kelompok simbol yang saling bergantung dalam satu loop |
| Deteksi hub | Simbol dengan fan-in atau fan-out coupling yang tinggi |
| Deteksi yatim | Kandidat dead code |
| Deteksi god class/fungsi | Simbol yang berisi jumlah anak yang berlebihan |
| Metrik ketidakstabilan | Simbol dan paket yang rapuh atau stabil, menggunakan Ce/(Ca+Ce) |
| Rantai pewarisan yang dalam | Hierarki yang lebih dalam dari `--max-inheritance-depth` |
| Betweenness centrality | Simbol hambatan yang dilalui banyak jalur dependensi |
| Titik artikulasi | Titik kegagalan tunggal yang penghapusannya membagi graf |
| PageRank | Simbol yang penting secara rekursif |
| Transitive fan-in | Cakupan dampak: segala sesuatu yang bergantung pada simbol, langsung atau tidak langsung |
| Simulasi propagasi perubahan | Seberapa jauh perubahan pada simbol kemungkinan menyebar |
| Dependensi paket siklik | Loop dependensi tingkat paket |
| Deteksi pelanggaran lapisan | Dependensi yang melintasi lapisan ke arah yang salah |
| Keterjangkauan dari titik masuk | Kode yang tidak dijangkau oleh titik masuk mana pun |
| Komponen terhubung lemah | Kluster kode yang terisolasi |
| Bobot dependensi | Seberapa kuat paket-paket terhubung |
| Jarak dari Main Sequence | Paket yang abstrak dan ketidakstabilannya tidak seimbang |
| Deteksi operasi shotgun | Perubahan yang akan menyentuh banyak tempat dalam sebuah paket |
| Deteksi iri fitur | Simbol yang menggunakan paket lain lebih dari paketnya sendiri |
| Pelanggaran dependensi stabil | Paket stabil yang bergantung pada paket yang kurang stabil |
| Pelanggaran segregasi antarmuka | Antarmuka yang anggotanya sebagian besar tidak digunakan oleh penggunanya |
| Kedalaman penampung | Simbol yang bersarang dalam |

### Baca hasil dengan benar

Aturan-aturan ini mempengaruhi seberapa jauh Anda dapat mempercayai temuan:

- **Hubungan yang tidak pasti dikecualikan.** Metrik dependensi hanya menggunakan sisi yang terselesaikan. Laporan menghitung hubungan `unresolved` dan `ambiguous` yang dikecualikan.
- **Metadata file dan include diabaikan.** Simbol `meta/file` dan `meta/include` tidak mempengaruhi peringkat dependensi atau temuan dead code.
- **Fan-in dan fan-out menghitung tetangga yang berbeda.** Penampungan tidak dihitung sebagai dependensi.
- **Kedalaman pewarisan mencakup semua basis dan antarmuka.** Hierarki siklik tidak diberi kedalaman.
- **Keterjangkauan dimulai dari callable `main`, `Main`, atau `init`** dan mengikuti sisi dependensi. Anggota yang dapat dijangkau membuat penampungnya relevan, tetapi bukan anggota saudaranya. Jika pemindaian tidak menemukan titik masuk yang dikenali, temuan keterjangkauan dihilangkan.
- **Callback adalah referensi, bukan panggilan.** Argumen callback atau callable yang dikembalikan tidak membuktikan callback berjalan.
- **Skor adalah heuristik.** Propagasi perubahan menggunakan jalur dependensi terkuat. Skor dan aturan lapisannya bukan probabilitas terukur atau jaminan kompiler.

### Bertindak berdasarkan temuan dengan aman

Anggap setiap temuan sebagai kandidat untuk ditinjau. Sebelum mengubah kode, buka simbol yang disebutkan laporan dan periksa:

- Pemanggil eksternal di luar direktori yang dipindai.
- Dispatch dinamis.
- Dependensi yang tidak Anda pindai.

Analisis berbasis sintaksis, sehingga salah satu dari ini dapat mengubah kesimpulan. Lihat [resolusi hubungan dalam referensi dukungan bahasa](../../reference/language-support/#relationship-resolution) untuk aturan dan batasan resolusi.

## Temukan hotspot yang sadar sejarah

`graph hotspots` memberi peringkat file yang sering berubah dan berada di posisi struktural yang penting. Ia juga melaporkan coupling temporal: file yang berulang kali berubah dalam commit yang sama. Jalankan pada direktori di dalam repositori Git.

```bash
codeknit graph hotspots ./myproject
```

Laporan ditulis ke `./skeleton/hotspots.skt`.

Ia menggabungkan dua kelompok metrik:

- **Sejarah:** frekuensi commit, churn baris yang ditambahkan dan dihapus, dan kebaruan.
- **Struktur:** PageRank tingkat file, transitive fan-in, dan betweenness.

Laporan mencakup tingkat kepercayaan berdasarkan jumlah commit yang relevan yang dianalisis. Anggap peringkat dari laporan dengan kepercayaan rendah sebagai tentatif.

### Ubah jendela sejarah dan filter

```bash
# Analisis dua tahun dan tulis JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Izinkan commit yang lebih besar dan memerlukan coupling temporal yang lebih kuat
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

Commit merge dikecualikan kecuali Anda memberikan `--include-merges`. Commit yang menyentuh lebih banyak file daripada `--max-files-per-commit` dilewati sebagai commit massal, sehingga perubahan yang dihasilkan atau mekanis tidak mengganggu peringkat.

| Flag | Default | Tujuan |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | Jalur file keluaran |
| `--format` | `skt` | Format keluaran: `skt` atau `json` |
| `--since` | `12mo` | Jendela sejarah, seperti `180d`, `12mo`, atau `2y` |
| `--max-commits` | `2000` | Jumlah maksimum commit yang akan diperiksa |
| `--max-files-per-commit` | `50` | Kecualikan commit yang mengubah lebih banyak file daripada ini |
| `--min-cochanges` | `3` | Jumlah minimum commit bersama untuk melaporkan coupling temporal |
| `--top-n` | `30` | Hasil maksimum per bagian |
| `--include-merges` | `false` | Sertakan commit merge dalam metrik sejarah |
| `--collect-test` | `false` | Sertakan file pengujian |
| `--workers` | `0` (`NumCPU`) | Jumlah maksimum goroutine parsing konkuren |
| `--verbose` | `false` | Cetak informasi progres |

### Baca keluaran JSON

Dengan `--format json`, setiap hotspot memiliki `score`, `history_score`, `structure_score`, `recency_score`, `pagerank`, `betweenness`, `commits`, `churn`, dan `transitive_fan_in`. Setiap entri coupling temporal memiliki `left`, `right`, `strength`, dan `co_changes`. Tingkat atas juga mencatat `commits_visited`, `commits_analyzed`, `skipped_merges`, dan `skipped_bulk_commits`. Gunakan hitungan ini untuk memeriksa seberapa banyak sejarah yang menjadi dasar peringkat.

## Kontrol sumber mana yang dipindai

Kelima perintah pemrosesan sumber (`parse`, `fingerprint`, `graph show`, `graph analyze`, dan `graph hotspots`) menerima flag pemilihan yang sama:

| Flag | Default | Tujuan |
| --- | --- | --- |
| `--include GLOB` | Semua sumber yang didukung | Sertakan jalur yang cocok; dapat diulang |
| `--exclude GLOB` | Tidak ada | Kecualikan jalur atau direktori yang cocok; dapat diulang. Pengecualian mengalahkan penyertaan |
| `--max-file-size BYTES` | `10485760` (10 MiB) | Lewati file yang lebih besar; `0` menonaktifkan batas |
| `--strict` | `false` | Gagal sebelum menulis keluaran jika pemindaian atau parsing tidak lengkap |

Pola adalah glob gaya gitignore relatif terhadap direktori masukan. Kutip pola tersebut agar shell Anda tidak mengembangkannya. Penyertaan tidak mengesampingkan aturan `.gitignore` atau penyaringan pengujian.

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

Symlink dan file non-reguler dilewati dan dilaporkan. Masalah pemindaian dan parsing muncul di stderr. Tanpa `--strict`, codeknit menyimpan hasil parsial yang dapat digunakan, jadi periksa stderr sebelum mempercayai laporan. Dengan `--strict`, analisis yang tidak lengkap berhenti sebelum keluaran apa pun ditulis. Jalankan di mana setiap pembacaan sumber gagal selalu gagal.

Untuk jaminan yang tepat, lihat [referensi keandalan analisis](../../reference/analysis-reliability/).

## Halaman terkait

- [Referensi flag CLI](../../reference/cli-flags/): setiap flag untuk setiap perintah.
- [Referensi format keluaran](../../reference/output-format/): sintaks `.skt` dan JSON.
- [Referensi dukungan bahasa](../../reference/language-support/): batasan ekstraksi dan resolusi.
- [Panduan perintah parse](../parse-command/): ekstrak simbol dan sisi untuk file individual.