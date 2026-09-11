---
title: Mode Keluaran
description: Pilih mode keluaran yang tepat untuk ukuran proyek dan alur kerja Anda.
---

`codeknit` mendukung tiga mode keluaran, yang dikontrol oleh flag `--output-mode`. Setiap mode menentukan bagaimana struktur kode yang diekstrak ditulis ke disk (atau stdout).

Mode keluaran terpisah dari format keluaran. Format default adalah `.skt`; gunakan `--format json` untuk mengeluarkan hasil parse yang sama dalam bentuk JSON yang dapat dibaca mesin. Dalam mode direktori, JSON ditulis ke `codeknit.json`. Dalam mode `inline`, JSON ditulis ke stdout.

### directory-flat (default, direkomendasikan)

- **Perilaku**: Menulis file `.skt` dalam potongan seperti `map_001.skt`, `map_002.skt`, dll.
- **Direktori keluaran**: `./skeleton/` secara default
- **Pembagian**: File dibagi ketika melebihi batas `--max-lines` (default: 500 baris)
- **Kasus penggunaan**: Terbaik untuk sebagian besar proyek. Menjaga keluaran tetap terorganisir dan mudah dibaca dengan membatasi ukuran file. Anda dapat membaca hanya potongan yang relevan dengan tugas Anda.
- **Minifikasi**: Ketika `--minify` diaktifkan, file `dict.skt` juga dibuat di direktori keluaran, berisi pemetaan token untuk nilai yang dikompresi.

Contoh:

```bash
codeknit parse ./src
# Output: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **Perilaku**: Mencerminkan struktur direktori sumber secara persis.
- **Direktori keluaran**: `./skeleton/` secara default
- **Pemetaan**: Satu file `.skt` dibuat per file sumber, pada path yang sesuai.
- **Kasus penggunaan**: Ideal ketika Anda ingin dengan cepat mencari struktur file tertentu. Berguna untuk navigasi bersama dengan basis kode asli.

Contoh:

```bash
codeknit parse ./src --output-mode directory-tree
# Output: ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt, etc.
```

### inline

- **Perilaku**: Mencetak semua keluaran ke stdout.
- **Direktori keluaran**: Tidak ada yang dibuat
- **Kasus penggunaan**: Hanya direkomendasikan untuk file tunggal atau proyek yang sangat kecil (kurang dari 5 file). Berguna saat mengalirkan keluaran ke alat lain atau memeriksa file tunggal secara interaktif.

Contoh:

```bash
codeknit parse ./src/main.go --output-mode inline
# Output: printed directly to terminal
```

### Format JSON

- **Perilaku**: Mengeluarkan satu dokumen JSON yang berisi `files`, `symbols`, `edges` opsional, dan `errors` opsional.
- **Lokasi keluaran**: `codeknit.json` dalam mode direktori, atau stdout dalam mode `inline`.
- **Kasus penggunaan**: Terbaik untuk skrip, integrasi editor, pemeriksaan CI, dan alat yang membutuhkan data terstruktur.

Contoh:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Contoh keluaran:

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

### Tabel Keputusan

| Mode             | Terbaik untuk                                | Lokasi keluaran                                     |
| ---------------- | ------------------------------------------- | --------------------------------------------------- |
| `directory-flat` | Sebagian besar proyek (default, direkomendasikan)    | `./skeleton/map_001.skt`, `map_002.skt`, ...        |
| `directory-tree` | Navigasi keluaran bersama kode sumber       | `./skeleton/<path tercermin>.skt`                    |
| `inline`         | File tunggal, mengalirkan ke alat lain      | stdout — hanya gunakan untuk file tunggal atau proyek sangat kecil |

| Format | Terbaik untuk                           | Keluaran                                                   |
| ------ | -------------------------------------- | -------------------------------------------------------- |
| `skt`  | Konteks LLM dan inspeksi manusia       | File `.skt` atau stdout                                   |
| `json` | Skrip dan integrasi terstruktur        | `codeknit.json` dalam mode direktori, atau stdout dalam `inline` |

### Aturan Praktis

- **Jika ragu** → gunakan `directory-flat` (default)
- **Inspeksi file tunggal** → `inline` dapat diterima
- **Lebih dari beberapa file** → utamakan `directory-flat` atau `directory-tree`
- **Basis kode besar** → tambahkan `--minify` untuk mengurangi penggunaan token
- **Menjalankan ulang pada keluaran yang sama** → gunakan `--clean` untuk menghapus file `.skt` yang usang

### Minifikasi

Flag `--minify` mengaktifkan kompresi berbasis kamus untuk token yang berulang (misalnya, kunci properti seperti `exported`, `async`, atau nama tipe umum). Ketika diaktifkan:

- Nilai yang berulang diganti dengan kode pendek (`d0`, `d1`, `d2`, ...)
- File `dict.skt` ditulis ke direktori keluaran, memetakan kode ke nilai asli
- Secara signifikan mengurangi ukuran keluaran untuk basis kode besar
- Berfungsi dalam mode `directory-flat` dan `directory-tree`

Contoh keluaran yang diminifikasi:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```

Format ini mempertahankan informasi lengkap sambil meminimalkan jejak token, menjadikannya ideal untuk analisis berbasis LLM.