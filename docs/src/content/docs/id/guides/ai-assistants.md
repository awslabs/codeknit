---
title: Menggunakan dengan Asisten AI
description: Siapkan codeknit sebagai skill untuk Kiro, Claude Code, dan asisten pengodean AI lainnya.
---

codeknit dilengkapi dengan skill siap pakai yang mengajarkan asisten pengodean AI cara menggunakannya secara efektif. Skill ini memungkinkan asisten untuk mengekstrak struktur kode, mendeteksi duplikat, dan melakukan analisis struktural tanpa perlu prompting manual.

## Gambaran skill

codeknit menyediakan dua skill:

- **`codeknit-parse`**: Mengajarkan asisten untuk mengekstrak struktur kode (fungsi, kelas, metode, variabel) dan hubungan (panggilan, pewarisan, penampungan) ke dalam file `.skt`.
- **`codeknit-fingerprint`**: Mengajarkan asisten untuk mendeteksi kode duplikat dan hampir duplikat menggunakan fuzzy hashing.

Setiap skill mencakup dokumentasi yang dibaca oleh asisten sesuai kebutuhan untuk memahami penggunaan, flag, format keluaran, dan alur kerja.

## Instalasi

Gunakan helper installer untuk menyalin direktori skill ke folder skill asisten Anda. Installer hanya mengunduh file skill yang sudah dibundel, sehingga Anda tidak perlu mengkloning repositori.

Instal untuk **Codex**, **Kiro**, dan **Claude Code**:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash
```

Instal untuk satu asisten:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant codex
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant kiro
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant claude
```

Dari checkout lokal, Anda dapat menggunakan helper Makefile:

```bash
make skills-install-dry-run
make skills-install
```

Installer melewati direktori skill yang sudah ada secara default. Untuk menggantinya, tambahkan `--force`:

```bash
curl -fsSL https://raw.githubusercontent.com/awslabs/codeknit/main/scripts/install-skills.sh | bash -s -- --assistant all --force
```

Setelah instalasi, asisten secara otomatis mengetahui cara menjalankan perintah codeknit, memilih flag yang sesuai, dan menginterpretasikan keluaran `.skt`.

## Apa yang diajarkan oleh setiap skill

### codeknit-parse

Skill `codeknit-parse` mengajarkan asisten untuk:

- Menjalankan `codeknit parse` dengan flag yang sesuai untuk berbagai skenario
- Memilih mode keluaran yang tepat:
  - `directory-flat` (default) untuk sebagian besar proyek
  - `inline` untuk file tunggal atau input kecil
  - `directory-tree` untuk mencerminkan struktur sumber
- Membaca dan menginterpretasikan file keluaran `.skt`, termasuk bagian `[symbols]`, `[edges]`, dan bagian opsional `[dict]`
- Menggunakan data struktural untuk refactoring, pemetaan dependensi, dan tinjauan kode
- Menjalankan `codeknit graph analyze` untuk wawasan kualitas kode yang lebih mendalam (dependensi siklik, simbol hub, god class, dll.)

### codeknit-fingerprint

Skill `codeknit-fingerprint` mengajarkan asisten untuk:

- Menggunakan `codeknit fingerprint` untuk deteksi duplikat, audit DRY, dan identifikasi refactor
- Memilih rentang kemiripan yang sesuai (`--min-similarity`, `--max-similarity`)
- Membaca bagian `[duplicates]` untuk mengidentifikasi kode hampir duplikat
- Memahami bahwa fingerprint mengukur bentuk struktural, bukan maksud semantik
- Menggunakan `--rerank` dengan embeddings Ollama untuk mengurangi false positive jika diperlukan

## Contoh alur kerja

### Analisis struktural

1. Minta asisten untuk menganalisis struktur basis kode Anda
2. Asisten menjalankan `codeknit parse ./src` dan membaca file `.skt` yang dihasilkan
3. Asisten menjawab pertanyaan struktural: dependensi, rantai panggilan, dead code
4. Untuk wawasan yang lebih mendalam, asisten menjalankan `codeknit graph analyze ./src` dan menginterpretasikan laporan

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### Deteksi duplikat

1. Minta asisten untuk menemukan kode yang diduplikasi
2. Asisten menjalankan `codeknit fingerprint ./src`
3. Asisten membaca bagian `[duplicates]` dalam keluaran
4. Asisten menyelidiki pasangan yang ditandai dan mengusulkan konsolidasi

```skt
[duplicates]
S1, S2: 87% similarity
S3, S4: 76% similarity
```

## Tips

- **Selalu baca file `.skt`, bukan sumber mentah, untuk pertanyaan struktural** — file tersebut berisi struktur yang diekstrak dalam format yang ringkas dan dapat diandalkan
- Gunakan `codeknit graph analyze` untuk mengungkap masalah kualitas kode seperti dependensi siklik, simbol hub, dan rantai pewarisan yang dalam
- Jalankan `codeknit fingerprint` sebelum refactor besar untuk mengidentifikasi kode yang disalin-tempel yang harus dikonsolidasikan
- Format `.skt` dirancang untuk efisien dalam penggunaan token, menjadikannya ideal untuk jendela konteks LLM
- Gunakan `--minify` untuk lebih mengurangi penggunaan token saat memproses basis kode besar