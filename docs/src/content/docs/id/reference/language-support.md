---
title: Dukungan Bahasa dan Fidelity API
description: Kemampuan ekstraksi yang telah diuji, pemilihan parser, dan keterbatasan bahasa yang diketahui.
---

<!-- Dirawat di docs/reference/language-support.md; disalin secara verbatim oleh docs/gen. -->

Codeknit mengekstrak sintaks tanpa menjalankan kompiler, mengevaluasi makro, atau memuat dependensi proyek. Matriks di bawah ini menjelaskan konstruk yang telah diuji, bukan kesesuaian bahasa secara lengkap atau jaminan resolusi nama lintas file.

Setiap bahasa memiliki pernyataan parameter/pengembalian dasar yang tepat dalam `internal/plugin/api_fidelity_test.go`. Cakupan deklarasi juga diperiksa dalam `internal/plugin/coverage_test.go` dan paket pengujian bahasa individual.

| Bahasa     | Ekstensi                                    | Baseline tanda tangan                                                                                         | Cakupan deklarasi dan kontainer                                                                             |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Parameter bernama dan tipe pengembalian dalam definisi dan prototipe                                          | Struct, union, enum, typedef, makro; deklarasi di dalam header guard                                       |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Parameter bernama dan tipe pengembalian dalam fungsi, prototipe, dan deklarasi metode                         | Kelas, struct, metode, namespace, template; header guard dan wrapper `extern "C"`                          |
| C#         | `.cs`                                       | Parameter dan tipe pengembalian yang diberi nama dan tipe eksplisit                                           | Kelas, struct, antarmuka, metode, field, delegasi, namespace                                               |
| Go         | `.go`                                       | Parameter yang dikelompokkan, tidak bernama, variadik, bertipe fungsi, dan generik; hasil bernama/ganda; penerima generik | Field struct dengan tipe dan tag, field tersemat, tanda tangan metode antarmuka, parameter tipe, paket     |
| Java       | `.java`                                     | Parameter dan tipe pengembalian yang diberi nama dan tipe eksplisit                                           | Kelas, antarmuka, konstruktor, metode, field, kelas bersarang, paket                                       |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Parameter bernama; tanpa tipe pengembalian yang disimpulkan                                                   | Fungsi, arrow function, kelas, field, metode, deklarasi yang mengandung JSX                                |
| PHP        | `.php`                                      | Parameter dan tipe pengembalian yang diberi nama dan tipe eksplisit                                           | Kelas, antarmuka, trait, enum, metode, properti, namespace                                                 |
| Python     | `.py`, `.pyi`                               | Anotasi, default, `*args`, `**kwargs`, pemisah posisi/kata kunci, anotasi pengembalian                       | Fungsi, kelas, kelas bersarang, metode yang didekorasi; penerima terikat konvensional dihilangkan          |
| Ruby       | `.rb`                                       | Parameter bernama; tanpa tipe pengembalian yang disimpulkan                                                   | Kelas, modul, metode instance/singleton, konstanta                                                         |
| Rust       | `.rs`                                       | Pola bertipe, `self`/`&self`/`&mut self`, binding mutable, tipe pengembalian eksplisit                       | Modul inline dan bersarang, field struct bernama, metode trait dan metode implementasi                     |
| Scala      | `.scala`, `.sc`                             | Parameter dan tipe pengembalian yang diberi nama dan tipe eksplisit                                           | Kelas, trait, objek, metode, enum, paket                                                                   |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Parameter bertipe bernama dan anotasi pengembalian eksplisit                                                 | Fungsi, arrow function, antarmuka, kelas, field, namespace; TSX menggunakan tata bahasanya sendiri         |

## Pemilihan bahasa header

Semua perintah CLI pemrosesan sumber menerima `--header-language c|cpp`. Pilihan ini hanya berlaku untuk file `.h`. Ekstensi lainnya mempertahankan parser biasanya.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

Defaultnya adalah `c`, mempertahankan ekstraksi makro C, typedef, dan union. Gunakan `cpp` untuk header C++ yang diberi nama `.h`; file yang diberi nama `.hpp` dan `.hxx` sudah menggunakan C++. Pilihan ini eksplisit karena kedua tata bahasa dapat menerima sebagian sintaks bahasa lain tanpa mengekstrak struktur lengkapnya.

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

Metode antarmuka Go dan field bernama Go/Rust muncul sebagai simbol dalam mode normal dan mode fingerprint. Metode dan field mempertahankan tipe pemiliknya. Modul Rust bersarang mempertahankan nama yang memenuhi syarat seperti `service.storage.read`, termasuk ketika modul lain mendeklarasikan fungsi dengan nama yang sama.

Rust mengizinkan field dan metode dengan nama yang sama. Nama scoped field membawa sufiks `#field` secara internal sehingga ID dan sisi containment-nya tetap berbeda; nama tampilan dan tanda tangan mempertahankan nama field asli.

Tanda tangan Python mempertahankan anotasi, default, splat, dan pemisah konvensi pemanggilan. Parameter pertama konvensional `self` atau `cls` dihilangkan dari metode terikat, tetapi dipertahankan dalam fungsi bebas atau metode statis.

Tanda tangan multibaris mempertahankan jeda baris sumber dalam JSON. SKT meng-escape jeda baris sebagai `\n` dan `\r` sehingga setiap simbol tetap berada pada satu baris fisik keluaran.

## Keterbatasan yang diketahui

- Deklarator pointer, referensi, array, dan fungsi-pointer C/C++ yang kompleks masih dapat memiliki tanda tangan yang tidak lengkap. Deklarasi yang dihasilkan makro tidak diekspansi.
- Penanganan konstruktor TypeScript/JavaScript dan beberapa bentuk parameter opsional, rest, atau destructured belum lengkap.
- Kendala generik lanjutan dan bentuk tipe di luar kasus yang diuji bervariasi menurut bahasa. Tipe pengembalian tidak disimpulkan.
- Makro Rust tidak diekspansi. Deklarasi modul out-of-line dicatat; file-file tersebut harus disertakan dalam pemindaian. Field tuple, detail varian enum, dan item terkait memiliki cakupan parsial.
- Ekstraksi sintaks tidak menyelesaikan pengiriman dinamis, overload, dependensi yang diimpor, atau semantik konfigurasi build dengan presisi kompiler.
- Kesalahan sintaks parsial masih dapat menghilangkan deklarasi. Periksa peringatan yang dilaporkan dan verifikasi sumber sebelum menerapkan refaktor.

Perubahan ekstraksi yang lebih lengkap mengubah jumlah simbol dan ShortID. Hasilkan ulang keluaran skeleton yang ada saat melakukan upgrade alih-alih mencampurnya dengan keluaran yang lebih lama.