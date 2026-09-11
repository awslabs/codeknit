---
title: Perintah Parse
description: Ekstrak informasi struktural dari kode sumber ke dalam file .skt atau JSON.
---

Perintah `codeknit parse` mengekstrak informasi struktural dari basis kode Anda — seperti fungsi, kelas, metode, variabel, dan hubungan di antara mereka — dan menampilkannya dalam format `.skt` yang ringkas secara default. Gunakan JSON saat Anda memerlukan keluaran yang dapat dibaca mesin untuk skrip, integrasi, atau alat downstream.

## Penggunaan Dasar

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Path ke direktori atau file yang ingin di-parse.
- **`[output-dir]`**: Direktori keluaran opsional. Jika tidak diberikan, secara default akan menggunakan `./skeleton`.

### Contoh

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## Mode Keluaran

Gunakan `--output-mode` untuk mengontrol bagaimana keluaran disusun. Tiga mode tersedia:

| Mode             | Deskripsi                                                                              | Paling Cocok Untuk                                    |
| ---------------- | ---------------------------------------------------------------------------------------- | --------------------------------------------------- |
| `directory-flat` | Menulis file `.skt` yang dipecah (misal `map_001.skt`, `map_002.skt`) ke direktori keluaran. | ✅ **Sebagian besar proyek** — mode default dan direkomendasikan |
| `directory-tree` | Mencerminkan struktur direktori sumber, membuat satu file `.skt` per file sumber.        | Menavigasi keluaran bersama dengan kode sumber             |
| `inline`         | Menampilkan semua keluaran ke stdout.                                                              | File tunggal atau piping ke alat lain               |

> **Tip**: Gunakan `directory-flat` secara default kecuali Anda bekerja dengan satu file. Hindari `inline` untuk input besar karena dapat membanjiri jendela konteks.

## Flags

| Flag             | Default          | Deskripsi                                                                  |
| ---------------- | ---------------- | ---------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Mode keluaran: `inline`, `directory-flat`, atau `directory-tree`                 |
| `--format`       | `skt`            | Format keluaran: `skt` atau `json`                                               |
| `--max-lines`    | `500`            | Jumlah baris maksimum per file keluaran dalam mode flat/tree                             |
| `--collect-test` | `false`          | Sertakan file pengujian dalam analisis                                               |
| `--minify`       | `false`          | Aktifkan kompresi berbasis kamus untuk mengurangi penggunaan token                    |
| `--edges`        | `false`          | Sertakan bagian `[edges]` dengan data hubungan (panggilan, berisi, dll.) |
| `--clean`        | `false`          | Hapus file `.skt` yang ada di direktori keluaran sebelum menulis          |
| `--workers`      | `NumCPU`         | Jumlah maksimum goroutine parsing bersamaan (0 = gunakan semua inti CPU)      |
| `--verbose`      | `false`          | Tampilkan informasi progres dan waktu selama pemrosesan                      |

## Pola Umum

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

Contoh keluaran JSON:

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

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## Perlindungan Keluaran Kedaluwarsa

Jika direktori keluaran sudah berisi file `.skt` dari eksekusi sebelumnya, `codeknit` akan menolak untuk menulis keluaran baru untuk mencegah pencampuran data kedaluwarsa dan baru.

Untuk mengatasi perilaku ini dan membersihkan direktori keluaran sebelum menulis, gunakan flag `--clean`:

```bash
codeknit parse ./src --clean
```

Hal ini memastikan kumpulan keluaran yang segar dan konsisten.

## Tips

- ✅ **Gunakan `directory-flat` secara default** untuk sebagian besar proyek. Ini menyeimbangkan keterbacaan dan pengelolaan.
- 🔍 Gunakan `--minify` pada basis kode besar untuk mengurangi penggunaan token melalui kamus bersama (`dict.skt`).
- 🔗 Bagian `[edges]` **dikecualikan secara default** untuk menghemat token. Gunakan `--edges` saat Anda memerlukan data hubungan seperti `calls`, `contains`, atau `inherits`.
- 🧾 Gunakan `--format json` saat skrip atau integrasi memerlukan data terstruktur alih-alih `.skt`.
- 🧹 Selalu gunakan `--clean` saat menjalankan ulang pada direktori keluaran yang sama.
- 📁 Gunakan `directory-tree` jika Anda ingin mengkorelasikan file `.skt` langsung dengan file sumber di editor Anda.