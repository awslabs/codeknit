---
title: Referensi Format Keluaran
description: Referensi lengkap untuk format keluaran .skt yang digunakan oleh codeknit.
---

Format `.skt` (skeleton) adalah format teks yang ringkas dan mudah dibaca manusia yang digunakan oleh `codeknit` untuk merepresentasikan struktur kode yang diekstrak. Format ini berisi simbol, hubungan, dan metadata dalam bentuk minimal yang cocok untuk konsumsi LLM dan analisis struktural.

File `.skt` dibagi menjadi beberapa bagian. Setiap bagian dimulai dengan header dalam tanda kurung siku. Bagian-bagian dapat muncul dalam urutan apa pun, meskipun `[symbols]` biasanya muncul pertama.

## [symbols]

Bagian `[symbols]` mencantumkan semua simbol yang diekstrak, dikelompokkan berdasarkan file sumbernya. Setiap file diperkenalkan dengan header `##` diikuti oleh path file.

### Format baris

Setiap simbol direpresentasikan dalam satu baris dengan struktur berikut:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Kolom

- **ShortID**: Pengidentifikasi berurutan yang diberikan untuk setiap simbol (mis., `S1`, `S2`, `S3`). Digunakan sebagai referensi dalam sisi dan bagian lainnya.
- **kategori/jenis**: Pasangan yang dipisahkan oleh garis miring yang menunjukkan kategori dan jenis spesifik simbol.
- **Lstart-Lend**: Rentang baris dalam file sumber tempat simbol didefinisikan (mis., `L10-L15`).
- **tanda_tangan**: Nama dan informasi tipe simbol. Format tergantung pada simbol:
  - `name` — untuk tipe, nilai, modul
  - `name(params)` — untuk callable tanpa tipe kembalian
  - `name(params) -> returnType` — untuk callable dengan tipe kembalian
- **{properti}**: Metadata opsional yang diapit kurung kurawal. Beberapa properti dipisahkan dengan koma.

### Parameter

- Dalam bahasa tanpa tipe: `paramName`
- Dalam bahasa dengan tipe: `paramName: type`
- Referensi tipe yang cocok dengan simbol yang diketahui diganti dengan ShortID mereka (mis., `config: S5` alih-alih `config: Config`).

### Properti

Properti umum termasuk:

- `async`: `true` atau `false`
- `exported`: `true` atau `false`
- `static`: ada jika simbol adalah statis
- `visibility=public|private|protected`
- `receiver=*TypeName`: untuk metode, menunjukkan tipe penerima

### Kategori dan jenis simbol

| Kategori   | Jenis                                      | Contoh                                           |
| ---------- | ------------------------------------------ | -------------------------------------------------- |
| `callable` | function, method, constructor              | `callable/function`, `callable/method`             |
| `type`     | class, interface, struct, enum             | `type/class`, `type/interface`                     |
| `value`    | variable, constant, field                  | `value/variable`, `value/constant`                 |
| `module`   | package, namespace                         | `module/package`                                   |
| `meta`     | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

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

Bagian `[edges]` mencatat hubungan ketika `--edges` diaktifkan. Titik akhir yang diketahui menggunakan ShortID mereka; target yang tidak dapat diselesaikan ke simbol mungkin muncul sebagai nama.

### Format baris

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

Status yang dihilangkan berarti penganalisis berhasil menyelesaikan hubungan. `unresolved` berarti tidak dapat menetapkan dependensi; `ambiguous` berarti ada beberapa interpretasi yang mungkin. ID Kandidat mengidentifikasi target yang mungkin, bukan koneksi yang dikonfirmasi. Penyelesaian didasarkan pada sintaks dan cakupan, bukan jaminan kompiler.

Target dikelompokkan hanya ketika sumber, jenis, resolusi, dan kumpulan kandidat mereka cocok. Nama tanpa simbol relatif terhadap file sumber. Identitas kompleks menggunakan ID lengkap yang dikutip dengan tanda kutip yang di-escape, garis miring terbalik, dan jeda baris. Nama literal yang menyerupai ShortID harus dikutip sebagai ID lengkap, seperti `"app.go::S2"`. Definisi simbol dapat muncul di file atau potongan keluaran selanjutnya.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Titik akhir file dan include

Include C/C++ memiliki simbol metadata eksplisit. Ejaan header muncul sekali dalam daftar simbol, dan sisi menggunakan ShortID meskipun resolusi header tetap unresolved:

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

Simbol include mencatat direktif yang ditulis, bukan deklarasi header yang diselesaikan. Include berulang dalam satu file berbagi titik akhir. Metadata file/include dikecualikan dari peringkat dependensi dan temuan kode mati.

Hubungan yang tidak pasti dipertahankan dalam keluaran parse tetapi dikecualikan dari metrik dependensi. Laporan graf menunjukkan jumlah hubungan yang dikecualikan; penampil HTML menunjukkan jumlah mereka tanpa menggambar sisi yang tidak pasti secara individual.

Saat meningkatkan ke `0.5.0`, hasilkan kembali semua potongan keluaran bersama-sama. ID Simbol dan hasil hubungan dapat berubah, sehingga potongan dari berbagai run tidak boleh dicampur.

### Jenis sisi

| Jenis         | Arti                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | pemanggilan fungsi/metode                      |
| `contains`   | kelas berisi metode, modul berisi fungsi       |
| `inherits`   | kelas memperluas kelas lain                    |
| `implements` | kelas mengimplementasikan antarmuka            |
| `overrides`  | metode menimpa metode induk                    |
| `references` | simbol mereferensikan simbol lain              |
| `imports`    | modul mengimpor modul lain                     |
| `decorates`  | dekorator diterapkan pada simbol               |

### Contoh

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

Bagian `[errors]` mencantumkan file yang tidak dapat di-parse sepenuhnya.

### Format

Setiap baris dimulai dengan `-` diikuti oleh path file dan pesan kesalahan:

```
- path/to/file.go: syntax error at line 42
```

### Contoh

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

Bagian `[dict]` hanya muncul ketika flag `--minify` digunakan. Ini memetakan kode kamus pendek ke token string berulang untuk mengurangi ukuran keluaran.

### Format

Setiap baris memetakan kode kamus (`d0`, `d1`, dll.) ke nilai yang diperluas:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Di bagian lain file, kode ini menggantikan nilai lengkap mereka.

Jenis hubungan juga dapat menggunakan kode kamus sementara label status tetap dapat dibaca: `S1 --d2[unresolved]--> S3`. Perluas `d2` menggunakan kamus dan pertahankan status unresolved.

### Contoh

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

## Contoh lengkap

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {d0}
S3 d1 L10-L12 NewServer(addr: string) -> *S2 {d0}
S4 callable/method L14-L20 Serve() {d0, receiver=*Server}
S5 callable/function L22-L25 handleError(err: error) -> bool {}

[edges]
S2 --contains--> S4
S4 --calls--> S5
S3 --references--> S2

[errors]
- utils/broken.go: syntax error at line 5
```

## Hubungan JSON

Gunakan JSON ketika skrip atau integrasi membutuhkan keluaran terstruktur:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Setiap sisi mencakup `from`, `to`, dan `kind`. ID sumber yang diketahui muncul dalam `from_short`; ID target yang diselesaikan muncul dalam `to_short`. Sisi yang tidak pasti menambahkan `resolution` dan, jika tersedia, `candidates` yang berisi ID simbol lengkap. Mereka menghilangkan `to_short`, bahkan ketika target memiliki simbol metadata include.

Misalnya, pemanggilan ambigu dapat direpresentasikan sebagai:

```json
{
  "from": "app.py::run",
  "from_short": "S1",
  "to": "app.py::Helper",
  "kind": "calls",
  "resolution": "ambiguous",
  "candidates": ["helpers/a.py::Helper", "helpers/b.py::Helper"]
}
```

JSON dan SKT mempertahankan ketidakpastian yang sama dengan enkode yang berbeda. Lihat [Dukungan Bahasa dan Fidelity API](../language-support/) untuk aturan resolusi dan batasan yang diketahui.