---
title: Perintah Grafik
description: Visualisasikan dan analisis struktur basis kode Anda dengan algoritma graf.
---

codeknit menyediakan perintah grafik untuk memvisualisasikan struktur, menjalankan analisis otomatis, dan menggabungkan grafik dependensi saat ini dengan riwayat perubahan Git.

## graph show

Menghasilkan visualisasi grafik HTML interaktif dari basis kode Anda.

```bash
codeknit graph show <input-path>
```

Perintah ini mem-parsing basis kode Anda dan menghasilkan file HTML mandiri dengan visualisasi grafik interaktif. Simbol (fungsi, kelas, tipe) muncul sebagai node, dan hubungan mereka (panggilan, berisi, mengimplementasikan) sebagai sisi. Visualisasi akan terbuka secara otomatis di peramban default Anda.

### Flags

| Flag             | Default                          | Deskripsi                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Jalur file HTML keluaran                     |
| `--collect-test` | `false`                          | Sertakan file pengujian dalam analisis       |
| `--workers`      | `NumCPU`                         | Jumlah maksimum goroutine parsing bersamaan  |
| `--verbose`      | `false`                          | Cetak informasi progres selama pemrosesan    |

### Contoh

```skt
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

Menjalankan algoritma grafik struktural pada basis kode Anda dan menghasilkan laporan `.skt` yang dapat dibaca oleh LLM yang berisi wawasan kualitas kode.

```bash
codeknit graph analyze <input-path>
```

Perintah ini mendeteksi masalah kualitas kode umum seperti dependensi siklik, simbol hub, kode mati, god class, dan hambatan arsitektural.

### Algoritma

Analisis mencakup 22 algoritma grafik struktural:

- Dependensi siklik (Tarjan's SCC)
- Deteksi hub (tinggi fan-in/fan-out coupling)
- Deteksi yatim piatu (kandidat dead code)
- Deteksi god class/function (anak berlebihan)
- Metrik ketidakstabilan (Robert C. Martin's Ce/(Ca+Ce))
- Rantai pewarisan yang dalam
- Betweenness centrality (deteksi hambatan)
- Titik artikulasi (titik kegagalan tunggal)
- PageRank (kepentingan rekursif)
- Transitive fan-in (cakupan dampak)
- Simulasi propagasi perubahan
- Dependensi paket sirkular
- Deteksi pelanggaran lapisan
- Keterjangkauan dari titik masuk
- Komponen terhubung lemah
- Bobot dependensi (kekuatan coupling paket)
- Jarak dari Main Sequence (keseimbangan A+I)
- Deteksi shotgun surgery
- Deteksi feature envy
- Pelanggaran stable dependency
- Pelanggaran interface segregation
- Kedalaman containment

### Flags

| Flag                      | Default                         | Deskripsi                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Jalur file `.skt` keluaran                               |
| `--collect-test`          | `false`                         | Sertakan file pengujian dalam analisis                   |
| `--workers`               | `NumCPU`                        | Jumlah maksimum goroutine parsing bersamaan              |
| `--verbose`               | `false`                         | Cetak informasi progres selama pemrosesan                |
| `--fan-threshold`         | `10`                            | Minimum fan-in atau fan-out untuk menandai simbol hub    |
| `--god-threshold`         | `15`                            | Jumlah minimum sisi berisi untuk menandai god class/function |
| `--max-inheritance-depth` | `5`                             | Tandai rantai pewarisan yang lebih dalam dari ini        |
| `--top-n`                 | `30`                            | Batasi bagian keluaran yang diberi peringkat; 0 = tanpa batas |
| `--betweenness-threshold` | `0.001`                         | Nilai betweenness centrality minimum untuk dilaporkan    |
| `--propagation-cutoff`    | `0.05`                          | Probabilitas minimum untuk melanjutkan propagasi perubahan |

### Contoh

```skt
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Memberi peringkat file yang sering diubah dan penting secara struktural:

```bash
codeknit graph hotspots <input-path>
```

Skor ini menggabungkan frekuensi commit, churn baris, dan kebaruan dengan PageRank tingkat file, transitive fan-in, dan betweenness centrality. Laporan juga mengidentifikasi coupling temporal antara file yang berulang kali diubah dalam commit yang sama.

Commit merge dikecualikan secara default. Commit yang mengubah lebih dari 50 file juga dikecualikan sehingga perubahan massal yang dihasilkan, vendored, atau mekanis tidak merusak hasil.

### Flags

| Flag                     | Default                   | Deskripsi                                      |
| ------------------------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Jalur file keluaran                             |
| `--format`               | `skt`                     | Format keluaran: `skt` atau `json`              |
| `--since`                | `12mo`                    | Jendela riwayat, seperti `180d`, `12mo`, atau `2y` |
| `--max-commits`          | `2000`                    | Jumlah maksimum commit yang diperiksa           |
| `--max-files-per-commit` | `50`                      | Kecualikan commit yang mengubah lebih banyak file |
| `--min-cochanges`        | `3`                       | Jumlah minimum commit bersama untuk coupling temporal |
| `--top-n`                | `30`                      | Jumlah maksimum hasil per bagian laporan        |
| `--include-merges`       | `false`                   | Sertakan commit merge                           |
| `--collect-test`         | `false`                   | Sertakan file pengujian                         |
| `--workers`              | `NumCPU`                  | Jumlah maksimum goroutine parsing bersamaan     |
| `--verbose`              | `false`                   | Cetak informasi progres                         |

### Contoh

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```