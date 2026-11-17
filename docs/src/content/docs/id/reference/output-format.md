---
title: Referensi Format Keluaran
description: Referensi lengkap untuk format keluaran .skt yang digunakan oleh codeknit.
---

Halaman ini menjelaskan tata bahasa dari keluaran `.skt` dan keluaran JSON yang setara. Untuk aturan yang menentukan bagaimana hubungan diselesaikan, lihat [Dukungan Bahasa dan Fidelity API](../language-support/).

## Gambaran Umum

File `.skt` adalah file teks ringkas yang terdiri dari beberapa bagian. Setiap bagian dimulai dengan baris header dalam tanda kurung siku.

| Bagian       | Hadir ketika                                      | Isi                                             |
| ------------ | ------------------------------------------------- | ----------------------------------------------- |
| `[symbols]`  | Selalu                                            | Definisi simbol dikelompokkan berdasarkan file sumber |
| `[edges]`    | `--edges` diaktifkan                              | Hubungan antar simbol                           |
| `[errors]`   | Terjadi kesalahan pemindaian atau penguraian      | File yang tidak dapat diproses sepenuhnya       |
| `[dict]`     | `--minify` diaktifkan                             | Kamus yang memetakan kode pendek ke token berulang |

Dalam mode keluaran direktori, `[errors]` dan `[dict]` dapat ditulis ke file tambahan `warnings.skt` dan `dict.skt`. Dalam mode `inline`, mereka muncul dalam keluaran yang sama dengan bagian lainnya.

`--format json` menghasilkan dokumen JSON dengan informasi yang sama. Lihat [keluaran JSON](#keluaran-json).

## [symbols]

Simbol dikelompokkan berdasarkan file sumber. Setiap kelompok dimulai dengan baris header `## <path>`, diikuti oleh satu baris per simbol.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### Baris simbol

```text
ShortID kategori/jenis Lstart-Lend tanda_tangan {properti}
```

| Field          | Deskripsi                                                                                               |
| -------------- | ------------------------------------------------------------------------------------------------------- |
| `ShortID`      | Pengidentifikasi berurutan: `S1`, `S2`, `S3`, dan seterusnya. Sisi merujuk ke simbol dengan nilai ini. |
| `kategori/jenis` | Klasifikasi simbol. Lihat [Kategori dan jenis](#kategori-dan-jenis).                                   |
| `Lstart-Lend`  | Baris pertama dan terakhir dari simbol dalam file sumber.                                              |
| `tanda_tangan` | Nama simbol, dengan parameter dan tipe kembalian untuk *callable*. Lihat [Format tanda tangan](#format-tanda-tangan). |
| `{properti}`   | Metadata dalam kurung kurawal. Lihat [Properti](#properti).                                            |

Baris simbol tetap pada satu baris fisik. Pemisah baris di dalam tanda tangan di-escape sebagai `\n` dan `\r`.

ID simbol dapat berubah ketika ekstraksi berubah. Hasilkan ulang semua potongan keluaran bersama-sama alih-alih mencampur potongan dari berbagai proses. Hal ini berlaku terutama saat meningkatkan ke versi 0.5.0, karena ID simbol dan hasil penyelesaian hubungan dapat berubah.

### Kategori dan jenis

| Kategori   | Jenis                                      | Contoh                                           |
| ---------- | ------------------------------------------ | ------------------------------------------------ |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`           |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                   |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`               |
| `module`   | package, namespace                         | `module/package`                                 |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### Format tanda tangan

| Bentuk                        | Digunakan untuk                                  |
| ----------------------------- | ------------------------------------------------ |
| `nama`                        | Tipe, nilai, dan modul                           |
| `nama(params)`                | *Callable* tanpa tipe kembalian                  |
| `nama(params) -> tipeKembali` | *Callable* dengan tipe kembalian                 |

Parameter ditulis sebagai:

| Bentuk          | Digunakan untuk                  |
| --------------- | -------------------------------- |
| `nama`          | Bahasa tanpa tipe                |
| `nama: tipe`    | Bahasa dengan tipe               |

Ketika referensi tipe cocok dengan simbol yang diketahui, nama tipe diganti dengan ID pendek simbol tersebut. Misalnya, `-> *S2` merujuk ke pointer ke simbol `S2`.

### Properti

Properti muncul dalam kurung kurawal setelah tanda tangan, dipisahkan oleh koma. Kurung kurawal kosong (`{}`) berarti simbol tidak memiliki properti.

| Bentuk properti        | Contoh                 |
| ---------------------- | ---------------------- |
| *Flag*                 | `exported`, `async`, `static` |
| `kunci=nilai`          | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

### Contoh

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
```

## [edges]

Bagian `[edges]` mencantumkan hubungan antar simbol. Bagian ini hanya ditulis ketika sisi diminta dengan `--edges`.

### Baris sisi

```text
FromID --jenis--> ToID1, ToID2
FromID --jenis[unresolved]--> Target1, Target2
FromID --jenis[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

| Bagian                     | Deskripsi                                                                           |
| -------------------------- | ----------------------------------------------------------------------------------- |
| `FromID`                   | Titik akhir sumber.                                                                 |
| `jenis`                    | Jenis hubungan. Lihat [Jenis sisi](#jenis-sisi).                                   |
| `[unresolved]`, `[ambiguous]` | Penanda resolusi opsional. Lihat [Penanda resolusi](#penanda-resolusi).         |
| `ToID1, ToID2`             | Satu atau lebih target, dipisahkan oleh koma.                                       |
| `[candidates=...]`         | Kemungkinan target untuk sisi ambigu, sebagai daftar terpisah koma dalam kurung.   |

Target berbagi satu baris hanya ketika sumber, jenis, penanda resolusi, dan kumpulan kandidat identik. Definisi yang dirujuk oleh sisi dapat muncul dalam file atau potongan keluaran yang lebih baru. Muat semua potongan sebelum menafsirkan sisi.

### Penanda resolusi

| Penanda         | Arti                                                                              |
| --------------- | --------------------------------------------------------------------------------- |
| Tidak ada       | Analyzer berhasil menyelesaikan hubungan.                                         |
| `[unresolved]`  | Analyzer tidak dapat menetapkan dependensi.                                       |
| `[ambiguous]`   | Lebih dari satu interpretasi mungkin tetap ada. ID kandidat mencantumkan target yang mungkin. |

Penanda yang hilang bukanlah jaminan kompiler. ID kandidat mengidentifikasi target yang mungkin, bukan koneksi yang dikonfirmasi. Ketidakpastian selalu ditulis dalam tata bahasa SKT ini, tidak pernah sebagai JSON yang disematkan.

### Bentuk titik akhir

| Bentuk                      | Digunakan ketika                                                                                         | Contoh                     |
| --------------------------- | --------------------------------------------------------------------------------------------------------- | -------------------------- |
| Short ID                    | Titik akhir atau kandidat adalah simbol yang diketahui. Ini termasuk titik akhir dari sisi yang tidak terselesaikan dan simbol *include*. | `S4`, `S8`       |
| Nama polos                  | Target tidak memiliki simbol. Nama relatif terhadap file sumber.                                        | `Refresh`, `string`        |
| ID lengkap dalam tanda kutip | Identitas kompleks, atau nama literal menyerupai Short ID.                                              | `"app.go::S2"`             |

ID lengkap dalam tanda kutip meng-escape tanda kutip, garis miring terbalik, dan pemisah baris. Nama literal seperti `S2` harus ditulis sebagai ID lengkap dalam tanda kutip agar tidak dibaca sebagai Short ID.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Jenis sisi

| Jenis         | Arti                                         |
| ------------- | -------------------------------------------- |
| `calls`       | pemanggilan fungsi/metode                    |
| `contains`    | kelas berisi metode, modul berisi fungsi     |
| `inherits`    | kelas mewarisi kelas lain                    |
| `implements`  | kelas mengimplementasikan antarmuka          |
| `overrides`   | metode menimpa metode induk                  |
| `references`  | simbol merujuk ke simbol lain                |
| `imports`     | modul mengimpor modul lain                   |
| `decorates`   | dekorator diterapkan pada simbol             |

*Callable* yang dikembalikan dan argumen *callback* dicatat sebagai `references`, bukan sebagai jenis terpisah.

### Simbol file dan include C dan C++

Sumber C dan C++ menghasilkan simbol `meta/file` dan `meta/include`, sehingga hubungan *include* memiliki Short ID. Setiap ejaan header muncul sekali dalam daftar simbol, dan sisi menggunakan Short ID. *Include* berulang dalam satu file berbagi satu titik akhir.

```skt
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

Simbol *include* mencatat direktif seperti yang tertulis. Simbol ini tidak mewakili deklarasi header yang terselesaikan.

### Dampak pada analisis dan visualisasi

- Sisi yang tidak pasti dan metadata file/*include* tidak berkontribusi pada metrik dependensi. Laporan menunjukkan jumlah hubungan yang dikecualikan.
- Grafik HTML menggambar tautan yang terselesaikan dan menampilkan jumlah hubungan yang tidak pasti. Grafik ini tidak menggambar sisi yang tidak pasti secara individual.

### Contoh

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

Bagian `[errors]` mencantumkan file yang tidak sepenuhnya diproses. Bagian ini hanya ada ketika terdapat diagnostik.

```text
- <path>: <pesan>
```

```skt
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict]

Bagian `[dict]` hanya muncul dalam keluaran yang diminifikasi (`--minify`). Bagian ini memetakan kode pendek ke token berulang.

| Bagian       | Deskripsi                                                                 |
| ------------ | ------------------------------------------------------------------------- |
| `d0`, `d1`, … | Kode kamus.                                                               |
| Baris entri  | `- <kode>: <token>`                                                       |

Untuk membaca keluaran yang diminifikasi, ganti setiap kode dengan nilai kamusnya. Kode dapat menggantikan pasangan kategori/jenis dan properti dalam `[symbols]`. Kode juga dapat menggantikan jenis hubungan dalam `[edges]`. Penanda ketidakpastian tetap dapat dibaca. Misalnya, `S1 --d2[unresolved]--> S3` adalah sisi yang tidak terselesaikan yang jenisnya adalah nilai dari `d2`.

### Contoh diminifikasi

```skt
[dict]
- d0: async=false
- d1: callable/method
- d2: exported

[symbols]
## src/handler.py
S1 type/class L1-L6 Handler {}
S2 d1 L2-L3 __init__(name) {d0}
S3 d1 L5-L6 handle(request) {d0}

[edges]
S1 --contains--> S2, S3
```

Di sini `S2` adalah `callable/method` dengan properti `async=false`.

## Contoh lengkap

Keluaran berikut berisi `[symbols]`, `[edges]`, dan `[errors]`. Keluaran ini tidak diminifikasi, sehingga tidak memiliki bagian `[dict]`; lihat [contoh diminifikasi](#contoh-diminifikasi).

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
S3 callable/function L10-L12 NewAuthService(secret: string, ttl: int) -> *S2 {exported}
S4 callable/method L14-L19 Authenticate(token: string) {exported, receiver=*AuthService}
S5 callable/function L29-L31 verifyToken(token: string) -> bool {exported=false}
## pkg/util/hash.go
S6 callable/function L3-L5 Hash(data: string) -> string {exported}
## pkg/util/hash_alt.go
S7 callable/function L3-L5 Hash(data: string) -> string {exported}

[edges]
S1 --contains--> S2, S3, S4, S5
S2 --contains--> S4
S3 --references--> S2
S4 --calls--> S5
S4 --calls[unresolved]--> Refresh
S5 --calls[ambiguous]--> Hash [candidates=S6, S7]

[errors]
- pkg/legacy/broken.go: syntax error at line 42
```

## Keluaran JSON

`--format json`