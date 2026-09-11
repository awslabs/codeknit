---
title: Referensi Format Keluaran
description: Referensi lengkap untuk format keluaran .skt yang digunakan oleh codeknit.
---

Format `.skt` (skeleton) adalah format teks ringkas dan mudah dibaca yang digunakan oleh `codeknit` untuk merepresentasikan struktur kode yang diekstrak. Format ini berisi simbol, hubungan, dan metadata dalam bentuk minimal yang cocok untuk konsumsi LLM dan analisis struktural.

File `.skt` dibagi menjadi beberapa bagian. Setiap bagian dimulai dengan header dalam tanda kurung siku. Bagian-bagian dapat muncul dalam urutan apa pun, meskipun `[symbols]` biasanya muncul pertama.

## [symbols]

Bagian `[symbols]` mencantumkan semua simbol yang diekstrak, dikelompokkan berdasarkan file sumbernya. Setiap file diperkenalkan dengan header `##` diikuti oleh path file.

### Format baris

Setiap simbol direpresentasikan dalam satu baris dengan struktur berikut:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Field

- **ShortID**: Pengidentifikasi berurutan yang diberikan untuk setiap simbol (misalnya, `S1`, `S2`, `S3`). Digunakan sebagai referensi dalam sisi dan bagian lainnya.
- **kategori/jenis**: Pasangan dipisahkan oleh garis miring yang menunjukkan kategori dan jenis spesifik simbol.
- **Lstart-Lend**: Rentang baris dalam file sumber tempat simbol didefinisikan (misalnya, `L10-L15`).
- **tanda tangan**: Nama dan informasi tipe simbol. Format tergantung pada simbol:
  - `nama` — untuk tipe, nilai, modul
  - `nama(params)` — untuk callable tanpa tipe kembalian
  - `nama(params) -> returnType` — untuk callable dengan tipe kembalian
- **{properti}**: Metadata opsional yang diapit oleh kurung kurawal. Beberapa properti dipisahkan oleh koma.

### Parameter

- Dalam bahasa tanpa tipe: `namaParam`
- Dalam bahasa dengan tipe: `namaParam: tipe`
- Referensi tipe yang cocok dengan simbol yang diketahui diganti dengan ShortID mereka (misalnya, `config: S5` alih-alih `config: Config`).

### Properti

Properti umum meliputi:

- `async`: `true` atau `false`
- `exported`: `true` atau `false`
- `static`: ada jika simbol bersifat statis
- `visibility=public|private|protected`
- `receiver=*TypeName`: untuk metode, menunjukkan tipe penerima

### Kategori dan jenis simbol

| Kategori   | Jenis                          | Contoh                               |
| ---------- | ------------------------------ | ------------------------------------ |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

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

Bagian `[edges]` mendefinisikan hubungan antara simbol menggunakan ShortID mereka.

### Format baris

```
FromID --kind--> ToID1, ToID2
```

Beberapa ID target dipisahkan oleh koma. Setiap baris merepresentasikan satu hubungan searah.

### Jenis sisi

| Jenis         | Arti                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | pemanggilan fungsi/metode                      |
| `contains`   | kelas berisi metode, modul berisi fungsi       |
| `inherits`   | kelas mewarisi kelas lain                      |
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

Bagian `[errors]` mencantumkan file yang tidak dapat diurai sepenuhnya.

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

Bagian `[dict]` hanya muncul ketika flag `--minify` digunakan. Bagian ini memetakan kode kamus pendek ke token string berulang untuk mengurangi ukuran keluaran.

### Format

Setiap baris memetakan kode kamus (`d0`, `d1`, dll.) ke nilai yang diperluas:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Di bagian lain file, kode-kode ini menggantikan nilai lengkap mereka.

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
S3 --returns--> S2

[errors]
- utils/broken.go: syntax error at line 5
```