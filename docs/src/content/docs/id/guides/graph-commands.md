---
title: Perintah Graf
description: Visualisasikan dan analisis struktur basis kode Anda dengan algoritma graf.
---

codeknit menyediakan perintah graf untuk memvisualisasikan struktur, menjalankan analisis otomatis, dan menggabungkan graf dependensi saat ini dengan riwayat perubahan Git.

## graph show

Menghasilkan visualisasi graf HTML interaktif dari basis kode Anda.

```bash
codeknit graph show <input-path>
```

Perintah ini mengurai basis kode Anda dan menghasilkan file HTML mandiri dengan visualisasi graf interaktif. Simbol (fungsi, kelas, tipe) muncul sebagai node, dan hubungan mereka (panggilan, berisi, mengimplementasikan) sebagai sisi. Visualisasi akan terbuka secara otomatis di peramban default Anda.

Pemirsa menampilkan hubungan yang telah diselesaikan. Ia menampilkan jumlah ketidakpastian tetapi tidak menggambar hubungan yang tidak terselesaikan atau ambigu secara individual atau tautan kandidat mereka. Periksa hubungan tersebut dengan keluaran parse:

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### Flag

| Flag             | Default                          | Deskripsi                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Jalur file HTML keluaran                    |
| `--collect-test` | `false`                          | Sertakan file pengujian dalam analisis      |
| `--workers`      | `NumCPU`                         | Jumlah maksimum goroutine pengurai konkuren |
| `--verbose`      | `false`                          | Cetak informasi progres selama pemrosesan   |

### Contoh

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

Menjalankan algoritma graf struktural pada basis kode Anda dan menghasilkan laporan `.skt` yang dapat dibaca oleh LLM yang berisi wawasan kualitas kode.

```bash
codeknit graph analyze <input-path>
```

Perintah ini melaporkan kandidat untuk ditinjau, termasuk dependensi siklik, simbol hub, kode mati, god class, dan hambatan arsitektural.

### Algoritma

Analisis mencakup 22 algoritma graf struktural:

- Dependensi siklik (Tarjan's SCC)
- Deteksi hub (kopling fan-in/fan-out tinggi)
- Deteksi yatim piatu (kandidat kode mati)
- Deteksi god class/fungsi (anak berlebihan)
- Metrik ketidakstabilan (Robert C. Martin's Ce/(Ca+Ce))
- Rantai pewarisan yang dalam
- Betweenness centrality (deteksi hambatan)
- Titik artikulasi (titik kegagalan tunggal)
- PageRank (kepentingan rekursif)
- Fan-in transitif (cakupan dampak)
- Simulasi propagasi perubahan
- Dependensi paket sirkular
- Deteksi pelanggaran lapisan
- Keterjangkauan dari titik masuk
- Komponen terhubung lemah
- Bobot dependensi (kekuatan kopling paket)
- Jarak dari Main Sequence (keseimbangan A+I)
- Deteksi operasi shotgun
- Deteksi iri fitur
- Pelanggaran dependensi stabil
- Pelanggaran segregasi antarmuka
- Kedalaman penampungan

### Menafsirkan laporan

Metrik dependensi menggunakan hubungan yang telah diselesaikan. Laporan menghitung hubungan yang tidak terselesaikan dan ambigu yang dikecualikan. Metadata file/sertakan dikecualikan dari peringkat dependensi dan temuan kode mati.

- Fan-in dan fan-out menghitung tetangga yang berbeda. Hubungan berulang tidak meningkatkan jumlah ini, dan penampungan bukanlah dependensi.
- Kedalaman pewarisan mempertimbangkan semua basis dan antarmuka. Hierarki siklik tidak diberi kedalaman.
- Keterjangkauan dimulai dari callable bernama `main`, `Main`, atau `init` dan mengikuti sisi dependensi. Mencapai anggota membuat kontainernya relevan tanpa menandai setiap saudara sebagai terjangkau. Tanpa titik masuk yang dikenali, temuan keterjangkauan dihilangkan.
- Argumen callback dan callable yang dikembalikan adalah referensi, bukan bukti bahwa callback dieksekusi.
- Propagasi perubahan menggunakan jalur dependensi terkuat. Skor dan aturan lapisan berbasis direktori adalah heuristik, bukan probabilitas yang diukur.

Pemanggil eksternal, pengiriman dinamis, dan dependensi di luar pemindaian dapat mengubah kesimpulan. Lihat [Dukungan Bahasa dan Fidelity API](../../reference/language-support/) untuk aturan dan batasan resolusi.

### Flag

| Flag                      | Default                         | Deskripsi                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Jalur file `.skt` keluaran                              |
| `--collect-test`          | `false`                         | Sertakan file pengujian dalam analisis                  |
| `--workers`               | `NumCPU`                        | Jumlah maksimum goroutine pengurai konkuren             |
| `--verbose`               | `false`                         | Cetak informasi progres selama pemrosesan               |
| `--fan-threshold`         | `10`                            | Minimum fan-in atau fan-out untuk menandai simbol hub   |
| `--god-threshold`         | `15`                            | Jumlah minimum sisi berisi untuk menandai god class/fungsi |
| `--max-inheritance-depth` | `5`                             | Tandai rantai pewarisan yang lebih dalam dari ini       |
| `--top-n`                 | `30`                            | Batasi bagian keluaran yang diberi peringkat; 0 = tanpa batas |
| `--betweenness-threshold` | `0.001`                         | Nilai betweenness centrality minimum untuk dilaporkan   |
| `--propagation-cutoff`    | `0.05`                          | Skor heuristik minimum untuk melanjutkan propagasi perubahan |

### Contoh

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Memberi peringkat file yang sering diubah dan penting secara struktural:

```bash
codeknit graph hotspots <input-path>
```

Skor menggabungkan frekuensi commit, churn baris, dan kebaruan dengan PageRank tingkat file, fan-in transitif, dan betweenness centrality. Laporan juga mengidentifikasi kopling temporal antara file yang berulang kali diubah dalam commit yang sama.

Commit penggabungan dikecualikan secara default. Commit yang mengubah lebih dari 50 file juga dikecualikan sehingga perubahan massal yang dihasilkan, vendored, atau mekanis tidak mendistorsi hasil.

### Flag

| Flag                     | Default                   | Deskripsi                                     |
| ------------------------ | ------------------------- | ----------------------------------------------- |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Jalur file keluaran                            |
| `--format`               | `skt`                     | Format keluaran: `skt` atau `json`             |
| `--since`                | `12mo`                    | Jendela riwayat, seperti `180d`, `12mo`, atau `2y` |
| `--max-commits`          | `2000`                    | Jumlah maksimum commit yang diperiksa         |
| `--max-files-per-commit` | `50`                      | Kecualikan commit yang mengubah lebih banyak file |
| `--min-cochanges`        | `3`                       | Jumlah minimum commit bersama untuk kopling temporal |
| `--top-n`                | `30`                      | Jumlah maksimum hasil per bagian laporan       |
| `--include-merges`       | `false`                   | Sertakan commit penggabungan                   |
| `--collect-test`         | `false`                   | Sertakan file pengujian                        |
| `--workers`              | `NumCPU`                  | Jumlah maksimum goroutine pengurai konkuren    |
| `--verbose`              | `false`                   | Cetak informasi progres                        |

### Contoh

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```