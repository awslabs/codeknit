---
title: Menggunakan dengan Asisten AI
description: Pasang plugin codeknit untuk Claude Code, Codex, dan Kiro.
---

Pasang plugin codeknit di Claude Code, Codex, atau Kiro, tetap perbarui, dan ajukan pertanyaan tentang struktur dan kode duplikat kepada asisten. Plugin ini berisi satu keterampilan, `codeknit`. Keterampilan ini memberi tahu asisten kapan harus menjalankan `codeknit parse`, `codeknit graph analyze`, dan `codeknit fingerprint`, serta cara membaca keluaran `.skt`.

Anda tidak perlu memasang biner `codeknit` sendiri. Pertama kali keterampilan dijalankan, ia akan menjalankan `scripts/install-codeknit.sh` yang terikat, yang melakukan hal berikut:

1. Keluar segera jika `codeknit` sudah ada di `PATH` atau di `~/.local/bin`.
2. Jika tidak, mengunduh rilis GitHub terbaru untuk OS dan arsitektur Anda.
3. Memverifikasi unduhan terhadap `checksums.txt` rilis.
4. Memasang biner ke `~/.local/bin` dan mencetak jalur absolutnya pada baris terakhir keluaran.

Jika `~/.local/bin` tidak ada di `PATH`, asisten akan memanggil codeknit dengan jalur lengkap tersebut selama sisa sesi.

Untuk mengontrol pemasangan, atur variabel lingkungan ini:

| Variabel | Efek |
| --- | --- |
| `CODEKNIT_VERSION` | Memperbaiki rilis, misalnya `0.6.0`. |
| `CODEKNIT_INSTALL_DIR` | Mengubah lokasi pemasangan. |

Untuk memasang biner sendiri, lihat [Pemasangan](../../installation/).

## Pasang plugin

Pilih bagian sesuai dengan asisten Anda.

### Claude Code

Jalankan perintah berikut di shell Anda. Di dalam sesi Claude Code, jalankan perintah yang sama sebagai `/plugin ...`.

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

1. Buka panel Powers.
2. Pilih **Add Custom Power**, lalu **Import power from GitHub**.
3. Masukkan `https://github.com/awslabs/codeknit`.

### Kiro CLI

Kiro CLI memasang dari jalur lokal, jadi klon repositori terlebih dahulu:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### Tentang `--sparse`

Opsi `--sparse` bersifat opsional. Opsi ini membatasi klon hanya pada manifes plugin dan `skills/` alih-alih seluruh pohon sumber.

### Pasang keterampilan secara manual

Jika asisten Anda membaca Agent Skills tetapi tidak mendukung plugin, salin `skills/codeknit/` ke direktori keterampilan untuk asisten Anda:

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## Perbarui plugin

| Asisten | Perintah atau tindakan |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Pilih **Check for updates** di panel Powers. |
| Kiro CLI | Jalankan `git -C ~/.codeknit-power pull`, lalu jalankan `kiro-cli powers install ~/.codeknit-power` lagi. |

Memperbarui plugin tidak akan memperbarui biner yang sudah terpasang. Untuk mengubah versi biner, atur `CODEKNIT_VERSION` seperti dijelaskan di atas.

Jika Anda meningkatkan codeknit dan ekstraksi berubah, ID simbol juga dapat berubah. Minta asisten untuk menghasilkan ulang semua keluaran alih-alih mencampur potongan dari berbagai eksekusi.

## Apa yang diajarkan keterampilan kepada asisten

Keterampilan ini memberikan aturan berikut kepada asisten, sehingga Anda tidak perlu mengulanginya di setiap permintaan:

- **Pilih perintah berdasarkan pertanyaan.**
  - `parse` menjawab apa yang ada, di mana lokasinya, dan apa yang memanggil, mewarisi, atau mengandung apa.
  - `graph analyze` menemukan siklus, hub, kode mati, god class, dan hambatan.
  - `fingerprint` menemukan kandidat salin-tempel dan refaktor.
- **Gunakan `directory-flat` sebagai mode keluaran default.** Asisten menulis file terpotong ke `./skeleton/` dan hanya membaca potongan yang diperlukan. Asisten menggunakan mode keluaran `inline` hanya untuk satu file atau beberapa file, karena `inline` membuang semuanya ke jendela konteks.
- **Baca keluaran `.skt`, bukan sumber mentah, untuk pertanyaan struktural.** Asisten membuka file sumber hanya untuk memeriksa atau mengubah detail implementasi, menggunakan rentang baris dari keluaran. `references/output-format.md` keterampilan berisi format lengkap. Lihat juga [referensi format keluaran](../../reference/output-format/).
- **Pilih rentang kemiripan dan perlakukan kecocokan sebagai sinyal.** Fingerprint membandingkan bentuk struktural, bukan maksud. Asisten membaca kedua simbol dari setiap pasangan sebelum mengusulkan refaktor.

## Minta analisis struktural

1. Minta asisten untuk menganalisis struktur basis kode Anda. Misalnya: "Petakan struktur `./src` dan beri tahu saya modul mana yang bergantung pada paket auth."
2. Asisten memastikan biner terpasang, lalu menjalankan `codeknit parse` dengan sisi diaktifkan dan membaca keluaran `.skt`.
3. Ajukan pertanyaan lanjutan tentang dependensi, rantai panggilan, atau kode mati. Asisten menjawab dari keluaran.
4. Untuk analisis yang lebih mendalam, minta tinjauan struktural. Misalnya: "Temukan dependensi siklik dan god class di `./src`." Asisten menjalankan `codeknit graph analyze` dan membaca laporan.

Hubungan yang ditandai `[unresolved]` atau `[ambiguous]` dalam keluaran tidak pasti, bukan dependensi yang dikonfirmasi. Perlakukan jawaban yang bergantung padanya sebagai kandidat untuk diverifikasi.

Untuk perintah di balik permintaan ini, lihat [panduan perintah parse](../parse-command/) dan [panduan perintah graf](../graph-commands/).

## Minta deteksi duplikat

1. Minta asisten untuk menemukan kode yang diduplikasi. Misalnya: "Temukan fungsi salin-tempel di `./src`."
2. Asisten menjalankan `codeknit fingerprint` dan membaca bagian `[duplicates]` dari `./skeleton/fingerprints.skt`.
3. Asisten membuka kedua simbol dari setiap pasangan yang ditandai, lalu mengusulkan refaktor.
4. Jika hasilnya bising, minta asisten untuk menyesuaikan rentang kemiripan atau menambahkan pemeringkatan semantik ulang. Pemeringkatan ulang memerlukan server Ollama lokal dengan model embedding yang diunduh.

Untuk opsi di balik permintaan ini, lihat [panduan perintah fingerprint](../fingerprint-command/).

## Tips

- Ingatkan asisten untuk membaca file `.skt`, bukan sumber mentah, untuk pertanyaan struktural. Hal ini menjaga konteks tetap kecil, karena formatnya ringkas dan menyimpan simbol, hubungan, dan rentang baris.
- Gunakan `graph analyze` untuk mendapatkan wawasan kualitas kode. Temuan adalah kandidat tinjauan, karena pemanggil eksternal dan pengiriman dinamis dapat mengubah hasil.
- Jalankan fingerprint sebelum refaktor besar untuk menemukan salin-tempel yang harus dikonsolidasikan.
- Jika asisten menjalankan ulang `parse` ke direktori yang sudah berisi keluaran `.skt`, codeknit akan menolak untuk menulis. Minta asisten untuk menambahkan `--clean` untuk mengganti keluaran sebelumnya.
- Jika unduhan biner gagal (offline atau platform tidak didukung), dapatkan rilis secara manual dari [halaman rilis](https://github.com/awslabs/codeknit/releases), atau lihat [Pemasangan](../../installation/).
- Jika suatu eksekusi melaporkan hasil parsial, periksa stderr dan bagian `[errors]`. Lihat [keandalan analisis](../../reference/analysis-reliability/) untuk detailnya.