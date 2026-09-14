---
title: Dukungan Bahasa dan Fidelity API
description: Kemampuan ekstraksi yang telah diuji, pemilihan parser, dan keterbatasan bahasa yang diketahui.
---

<!-- Dirawat di docs/reference/language-support.md; disalin verbatim oleh docs/gen. -->

Codeknit mengekstrak sintaks tanpa menjalankan kompiler, mengevaluasi makro, atau memuat dependensi proyek. Matriks di bawah ini menjelaskan konstruk yang telah diuji, bukan kesesuaian bahasa secara lengkap atau jaminan resolusi nama lintas file.

Setiap bahasa memiliki pernyataan dasar parameter/tanda tangan kembalian yang tepat dalam `internal/plugin/api_fidelity_test.go`. Cakupan deklarasi juga diperiksa dalam `internal/plugin/coverage_test.go` dan paket pengujian bahasa individual.

| Bahasa      | Ekstensi                                    | Baseline tanda tangan                                                                                         | Cakupan deklarasi dan kontainer                                                                             |
| ----------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| C           | `.c`, `.h`                                  | Parameter bernama dan tipe kembalian dalam definisi dan prototipe                                             | Struct, union, enum, typedef, makro; deklarasi di dalam header guard                                       |
| C++         | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Parameter bernama dan tipe kembalian dalam fungsi, prototipe, dan deklarasi metode                            | Kelas, struct, metode, namespace, template; header guard dan wrapper `extern "C"`                          |
| C#          | `.cs`                                       | Parameter bertipe bernama dan tipe kembalian eksplisit                                                        | Kelas, struct, antarmuka, metode, field, delegasi, namespace                                                |
| Go          | `.go`                                       | Parameter terkelompok, tanpa nama, variadik, bertipe fungsi dan generik; hasil bernama/berganda; penerima generik | Field struct dengan tipe dan tag, field tersemat, tanda tangan metode antarmuka, parameter tipe, paket     |
| Java        | `.java`                                     | Parameter bertipe bernama dan tipe kembalian eksplisit                                                        | Kelas, antarmuka, konstruktor, metode, field, kelas bersarang, paket                                        |
| JavaScript  | `.js`, `.jsx`, `.mjs`, `.cjs`               | Parameter bernama; tanpa tipe kembalian yang disimpulkan                                                      | Fungsi, fungsi panah, kelas, field, metode, deklarasi yang mengandung JSX                                   |
| PHP         | `.php`                                      | Parameter bertipe bernama dan tipe kembalian eksplisit                                                        | Kelas, antarmuka, trait, enum, metode, properti, namespace                                                  |
| Python      | `.py`, `.pyi`                               | Anotasi, default, `*args`, `**kwargs`, pemisah posisi/kata kunci, anotasi kembalian                           | Fungsi, kelas, kelas bersarang, metode berdekorasi; penerima terikat konvensional diabaikan                 |
| Ruby        | `.rb`                                       | Parameter bernama; tanpa tipe kembalian yang disimpulkan                                                      | Kelas, modul, metode instance/singleton, konstanta                                                          |
| Rust        | `.rs`                                       | Pola bertipe, `self`/`&self`/`&mut self`, binding mutable, tipe kembalian eksplisit                           | Modul inline dan bersarang, field struct bernama, metode trait dan implementasi                             |
| Scala       | `.scala`, `.sc`                             | Parameter bertipe bernama dan tipe kembalian eksplisit                                                        | Kelas, trait, objek, metode, enum, paket                                                                    |
| TypeScript  | `.ts`, `.tsx`, `.mts`, `.cts`               | Parameter bertipe bernama dan anotasi kembalian eksplisit                                                     | Fungsi, fungsi panah, antarmuka, kelas, field, namespace; TSX menggunakan tata bahasa sendiri               |

## Pemilihan bahasa header

Semua perintah CLI pemrosesan sumber menerima `--header-language c|cpp`. Pilihan ini hanya berlaku untuk file `.h`. Ekstensi lainnya mempertahankan parser biasanya.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

Defaultnya adalah `c`, mempertahankan ekstraksi makro C, typedef, dan union. Gunakan `cpp` untuk header C++ yang diberi nama `.h`; file bernama `.hpp` dan `.hxx` sudah menggunakan C++. Pilihan ini eksplisit karena kedua tata bahasa dapat menerima sebagian sintaks bahasa lain tanpa mengekstrak struktur lengkapnya.

Header guard dan cabang kondisional dilalui tanpa mengevaluasi kondisinya. Hasilnya dapat mencakup deklarasi dari cabang yang saling eksklusif. Ekspansi makro, resolusi include, dan flag build tidak diterapkan. TUI menggunakan pilihan C default untuk file `.h`.

## Mempertahankan bentuk API

Sebagai contoh, Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

menghasilkan:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Metode antarmuka Go dan field bernama Go/Rust muncul sebagai simbol dalam mode normal dan fingerprint. Metode dan field mempertahankan tipe pemiliknya. Modul Rust bersarang mempertahankan nama yang memenuhi syarat seperti `service.storage.read`, termasuk ketika modul lain mendeklarasikan fungsi dengan nama yang sama.

Rust mengizinkan field dan metode dengan nama yang sama. Nama scoped field membawa sufiks `#field` secara internal sehingga ID dan sisi kontainernya tetap berbeda; nama tampilan dan tanda tangan mempertahankan nama field asli.

Tanda tangan Python mempertahankan anotasi, default, splat, dan pemisah konvensi pemanggilan. Parameter pertama konvensional `self` atau `cls` diabaikan dari metode terikat, tetapi dipertahankan dalam fungsi bebas atau metode statis.

Tanda tangan multibaris mempertahankan jeda baris sumber dalam JSON. SKT meng-escape jeda baris sebagai `\n` dan `\r` sehingga setiap simbol tetap berada pada satu baris keluaran fisik.

## Resolusi hubungan

Hubungan mempertahankan kualifikasi penerima dan kepemilikan leksikal. Impor relatif dan impor JavaScript/TypeScript serta Python yang diganti nama memberikan konteks target. Deklarasi paket Go dan Java dapat diselesaikan lintas file dalam direktori yang sama. Impor yang tidak cocok tidak pernah kembali ke deklarasi yang tidak terkait, dan kandidat yang bersaing tetap ambigu.

Binding penerima lokal lebih diutamakan daripada impor namespace, termasuk ketika parameter membayangi namespace yang diimpor. Impor paket Python mempertimbangkan deklarasi dalam modul `__init__` paket, bukan modul turunan arbitrer. Re-ekspor yang tidak dapat ditetapkan dari modul tersebut tetap tidak terselesaikan.

Ekstraksi callable Java, C++, C#, Scala, dan TypeScript mempertahankan rentang byte deklarasi sehingga badan yang kelebihan beban dapat memiliki pemanggil dan cakupan alias yang berbeda, termasuk deklarasi pada baris yang sama. Memilih target yang kelebihan beban tetap konservatif. Hubungan override eksplisit dalam Java, C++, C#, Scala, dan TypeScript mencari deklarasi yang diwarisi terdekat; basis yang hilang, siklus, atau kandidat yang bersaing mempertahankan ketidakpastian alih-alih menghasilkan tautan mandiri.

Ini adalah analisis berbasis sintaks, bukan kompiler atau grafik panggilan runtime. Spesifikasi paket JavaScript polos, re-ekspor paket, konfigurasi build, tipe penerima kompleks, dan pengiriman dinamis mungkin tetap tidak terselesaikan. Nama unik di tempat lain dalam repositori tidak cukup sebagai bukti dependensi.

JSON mempertahankan sisi yang tidak pasti dengan `resolution: "unresolved"` atau `"ambiguous"` dan menyertakan ID kandidat jika tersedia. Sisi ini tidak memiliki target ShortID.

SKT menggunakan notasi sisi yang sama untuk semua hubungan, dengan penanda resolusi opsional:

```text
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

Titik akhir dan kandidat yang diketahui menggunakan ShortID, termasuk titik akhir hubungan yang tidak terselesaikan. Target yang tidak pasti tanpa simbol adalah nama relatif terhadap file sumber. Nama literal yang terlihat seperti ShortID, dan identitas kompleks lainnya, menggunakan ID lengkap yang dikutip untuk menghindari ambiguitas. Sisi dengan sumber, jenis, resolusi, dan kumpulan kandidat yang sama berbagi satu daftar target. Pembaca SKT mempertahankan resolusi dan kandidat, termasuk referensi ke simbol dalam potongan keluaran berikutnya. Keluaran yang diminifikasi mengkodekan jenis hubungan sambil menjaga penanda resolusi tetap terbaca.

Include C/C++ memiliki simbol eksplisit `meta/file` dan `meta/include`. Ejaan header muncul sekali dalam daftar simbol; sisi menggunakan ShortID mereka:

```text
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

Simbol include menjelaskan direktif tertulis, bukan deklarasi header yang terselesaikan. Include berulang dalam file berbagi titik akhir. Metadata file/include dikecualikan dari peringkat dependensi dan temuan kode mati; menambahkan ID tidak mengklaim bahwa header telah ditemukan melalui konfigurasi pencarian include kompiler.

Metrik dependensi dan tautan grafik HTML hanya menggunakan sisi yang terselesaikan. Laporan menunjukkan jumlah hubungan yang dikecualikan, dan grafik HTML menampilkan jumlah ketidakpastian. Fan-in dan fan-out menghitung tetangga yang berbeda. Analisis pewarisan mempertimbangkan semua basis dan antarmuka; pewarisan siklik tidak memiliki kedalaman yang dilaporkan. Keterjangkauan dimulai dari callable bernama `main`, `Main`, atau `init` dan mengikuti sisi dependensi. Mencapai anggota membuat kontainernya relevan tanpa membuat setiap saudara dapat dijangkau. Tanpa titik masuk yang dikenali, temuan keterjangkauan dihilangkan.

Argumen callback dan callable yang dikembalikan adalah referensi, bukan bukti invokasi. Resolusi alias mempertahankan identitas file, fungsi, dan objek; penugasan yang bertentangan tetap tidak pasti. Temuan grafik adalah kandidat tinjauan: pemanggil eksternal dan perilaku dinamis dapat mengubah hasil. Aturan lapisan berbasis direktori dan bobot propagasi perubahan adalah heuristik, bukan probabilitas yang diukur.

## Keterbatasan yang diketahui

- Deklarator pointer, referensi, array, dan fungsi-pointer C/C++ yang kompleks masih dapat memiliki tanda tangan yang tidak lengkap. Deklarasi yang dihasilkan makro tidak diperluas.
- Penanganan konstruktor TypeScript/JavaScript dan beberapa bentuk parameter opsional, rest, atau destructured tidak lengkap.
- Kendala generik lanjutan dan bentuk tipe di luar kasus yang diuji bervariasi menurut bahasa. Tipe kembalian tidak disimpulkan.
- Makro Rust tidak diperluas. Deklarasi modul out-of-line dicatat; file mereka harus disertakan dalam pemindaian. Field tuple, detail varian enum, dan item terkait memiliki cakupan parsial.
- Ekstraksi sintaks tidak menyelesaikan pengiriman dinamis, kelebihan beban, dependensi yang diimpor, atau semantik konfigurasi build dengan presisi kompiler.
- Kesalahan sintaks parsial masih dapat menghilangkan deklarasi. Periksa peringatan yang dilaporkan dan verifikasi sumber sebelum menerapkan refaktor.

Ekstraksi yang lebih lengkap mengubah jumlah simbol dan ShortID. Hasilkan ulang keluaran skeleton yang ada saat meningkatkan alih-alih mencampurnya dengan keluaran yang lebih lama.