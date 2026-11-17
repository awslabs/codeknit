---
title: Tài liệu tham chiếu về định dạng đầu ra
description: Tài liệu tham chiếu đầy đủ về định dạng đầu ra .skt được sử dụng bởi codeknit.
---

Trang này mô tả ngữ pháp của đầu ra `.skt` và đầu ra JSON tương đương. Để biết các quy tắc quyết định cách thức giải quyết mối quan hệ, hãy xem [Hỗ trợ ngôn ngữ và độ trung thực của API](../language-support/).

## Tổng quan

Tệp `.skt` là một tệp văn bản nhỏ gọn bao gồm các phần. Mỗi phần bắt đầu bằng một dòng tiêu đề trong dấu ngoặc vuông.

| Phần          | Xuất hiện khi                                      | Nội dung                                          |
| ------------- | -------------------------------------------------- | ------------------------------------------------- |
| `[symbols]`   | Luôn luôn                                          | Định nghĩa ký hiệu được nhóm theo tệp nguồn       |
| `[edges]`     | `--edges` được thiết lập                           | Mối quan hệ giữa các ký hiệu                      |
| `[errors]`    | Có lỗi quét hoặc phân tích cú pháp xảy ra          | Các tệp không thể xử lý hoàn toàn                 |
| `[dict]`      | `--minify` được thiết lập                          | Từ điển ánh xạ mã ngắn tới các token lặp lại      |

Trong các **chế độ đầu ra** thư mục, `[errors]` và `[dict]` có thể được ghi vào các tệp phụ trợ `warnings.skt` và `dict.skt`. Trong chế độ `inline`, chúng xuất hiện trong cùng một đầu ra với các phần khác.

`--format json` tạo ra một tài liệu JSON với cùng thông tin. Xem [đầu ra JSON](#đầu-ra-json).

## [symbols]

Các ký hiệu được nhóm theo tệp nguồn. Mỗi nhóm bắt đầu bằng dòng tiêu đề `## <đường-dẫn>`, theo sau là một dòng cho mỗi ký hiệu.

```skt
[symbols]
## pkg/services/auth.go
S1 module/package L1-L1 services {}
S2 type/struct L5-L8 AuthService {exported}
```

### Dòng ký hiệu

```text
ShortID category/kind Lstart-Lend signature {properties}
```

| Trường          | Mô tả                                                                                               |
| --------------- | --------------------------------------------------------------------------------------------------- |
| `ShortID`       | Định danh tuần tự: `S1`, `S2`, `S3`, v.v. Các cạnh tham chiếu đến các ký hiệu bằng giá trị này.     |
| `category/kind` | Phân loại của ký hiệu. Xem [Danh mục và loại](#danh-mục-và-loại).                                   |
| `Lstart-Lend`   | Dòng đầu tiên và dòng cuối cùng của ký hiệu trong tệp nguồn.                                        |
| `signature`     | Tên ký hiệu, với tham số và kiểu trả về cho các callable. Xem [Định dạng chữ ký](#định-dạng-chữ-ký). |
| `{properties}`  | Siêu dữ liệu trong dấu ngoặc nhọn. Xem [Thuộc tính](#thuộc-tính).                                   |

Một dòng ký hiệu nằm trên một dòng vật lý. Ngắt dòng bên trong chữ ký được thoát dưới dạng `\n` và `\r`.

ID ký hiệu có thể thay đổi khi trích xuất thay đổi. Hãy tái tạo tất cả các phần đầu ra cùng nhau thay vì trộn các phần từ các lần chạy khác nhau. Điều này đặc biệt áp dụng khi nâng cấp lên 0.5.0, vì ID ký hiệu và kết quả giải quyết mối quan hệ có thể thay đổi.

### Danh mục và loại

| Danh mục    | Loại                                      | Ví dụ                                           |
| ----------- | ----------------------------------------- | ----------------------------------------------- |
| `callable`  | function, method, constructor             | `callable/function`, `callable/method`          |
| `type`      | class, interface, struct, enum            | `type/class`, `type/interface`                  |
| `value`     | variable, constant, field                 | `value/variable`, `value/constant`              |
| `module`    | package, namespace                        | `module/package`                                |
| `meta`      | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

### Định dạng chữ ký

| Định dạng                     | Được sử dụng cho                              |
| ----------------------------- | --------------------------------------------- |
| `name`                        | Các loại, giá trị và mô-đun                   |
| `name(params)`                | Callable không có kiểu trả về                 |
| `name(params) -> returnType`  | Callable có kiểu trả về                       |

Các tham số được viết dưới dạng:

| Định dạng       | Được sử dụng cho              |
| --------------- | ----------------------------- |
| `name`          | Ngôn ngữ không có kiểu        |
| `name: type`    | Ngôn ngữ có kiểu              |

Khi một tham chiếu kiểu khớp với một ký hiệu đã biết, tên kiểu được thay thế bằng ID ngắn của ký hiệu đó. Ví dụ, `-> *S2` tham chiếu đến một con trỏ tới ký hiệu `S2`.

### Thuộc tính

Các thuộc tính xuất hiện trong dấu ngoặc nhọn sau chữ ký, phân tách bằng dấu phẩy. Dấu ngoặc nhọn rỗng (`{}`) có nghĩa là ký hiệu không có thuộc tính.

| Định dạng thuộc tính | Ví dụ                 |
| -------------------- | --------------------- |
| Cờ                   | `exported`, `async`, `static` |
| `key=value`          | `visibility=public`, `receiver=*AuthService`, `exported=false`, `resolution=unresolved` |

### Ví dụ

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

Phần `[edges]` liệt kê các mối quan hệ giữa các ký hiệu. Nó chỉ được viết khi các cạnh được yêu cầu với `--edges`.

### Dòng cạnh

```text
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

| Phần                     | Mô tả                                                                           |
| ------------------------ | ------------------------------------------------------------------------------- |
| `FromID`                 | Điểm cuối nguồn.                                                                |
| `kind`                   | Loại mối quan hệ. Xem [Loại cạnh](#loại-cạnh).                                 |
| `[unresolved]`, `[ambiguous]` | Đánh dấu giải quyết tùy chọn. Xem [Đánh dấu giải quyết](#đánh-dấu-giải-quyết). |
| `ToID1, ToID2`           | Một hoặc nhiều mục tiêu, phân tách bằng dấu phẩy.                               |
| `[candidates=...]`       | Các mục tiêu có thể cho một cạnh mơ hồ, dưới dạng danh sách phân tách bằng dấu phẩy trong dấu ngoặc. |

Các mục tiêu chia sẻ một dòng chỉ khi nguồn, loại, đánh dấu giải quyết và tập ứng viên giống nhau. Một định nghĩa được tham chiếu bởi một cạnh có thể xuất hiện trong một tệp hoặc phần đầu ra sau đó. Hãy tải tất cả các phần trước khi diễn giải các cạnh.

### Đánh dấu giải quyết

| Đánh dấu         | Ý nghĩa                                                                              |
| ---------------- | ------------------------------------------------------------------------------------ |
| Không có         | Trình phân tích đã giải quyết được mối quan hệ.                                      |
| `[unresolved]`   | Trình phân tích không thể thiết lập được phụ thuộc.                                   |
| `[ambiguous]`    | Có nhiều hơn một cách giải thích khả dĩ. Các ID ứng viên liệt kê các mục tiêu có thể. |

Việc thiếu đánh dấu không phải là một đảm bảo từ trình biên dịch. Các ID ứng viên xác định các mục tiêu có thể, không phải là các kết nối đã xác nhận. Sự không chắc chắn luôn được viết trong ngữ pháp SKT này, không bao giờ dưới dạng JSON nhúng.

### Định dạng điểm cuối

| Định dạng                      | Khi được sử dụng                                                                                         | Ví dụ                     |
| ------------------------------ | -------------------------------------------------------------------------------------------------------- | ------------------------- |
| Short ID                       | Điểm cuối hoặc ứng viên là một ký hiệu đã biết. Điều này bao gồm các điểm cuối của các cạnh chưa giải quyết và các ký hiệu include. | `S4`, `S8`       |
| Tên đơn                        | Mục tiêu không có ký hiệu. Tên này là tương đối với tệp nguồn.                                          | `Refresh`, `string`       |
| ID đầy đủ có dấu ngoặc kép     | Định danh phức tạp, hoặc một tên theo nghĩa đen giống như một short ID.                                 | `"app.go::S2"`            |

Các ID đầy đủ có dấu ngoặc kép thoát các dấu ngoặc kép, dấu gạch chéo ngược và ngắt dòng. Một tên theo nghĩa đen như `S2` phải được viết dưới dạng ID đầy đủ có dấu ngoặc kép để không bị đọc nhầm là short ID.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Loại cạnh

| Loại         | Ý nghĩa                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | Lời gọi hàm/phương thức                         |
| `contains`   | Lớp chứa phương thức, mô-đun chứa hàm           |
| `inherits`   | Lớp kế thừa từ lớp khác                         |
| `implements` | Lớp triển khai giao diện                        |
| `overrides`  | Phương thức ghi đè phương thức cha              |
| `references` | Ký hiệu tham chiếu đến ký hiệu khác             |
| `imports`    | Mô-đun nhập mô-đun khác                         |
| `decorates`  | Bộ trang trí được áp dụng cho ký hiệu           |

Các callable được trả về và đối số callback được ghi lại là `references`, không phải là một loại riêng biệt.

### Ký hiệu tệp và include của C và C++

Các nguồn C và C++ tạo ra các ký hiệu `meta/file` và `meta/include`, vì vậy các mối quan hệ include có short ID. Mỗi cách viết tiêu đề xuất hiện một lần trong danh sách ký hiệu, và các cạnh sử dụng short ID. Các include lặp lại trong một tệp chia sẻ một điểm cuối.

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

Một ký hiệu include ghi lại chỉ thị như đã viết. Nó không đại diện cho một khai báo tiêu đề đã giải quyết.

### Ảnh hưởng đến phân tích và trực quan hóa

- Các cạnh không chắc chắn và siêu dữ liệu tệp/include không đóng góp vào các chỉ số phụ thuộc. Các báo cáo hiển thị số lượng các mối quan hệ bị loại trừ.
- Đồ thị HTML vẽ các liên kết đã giải quyết và hiển thị số lượng các mối quan hệ không chắc chắn. Nó không vẽ các cạnh không chắc chắn riêng lẻ.

### Ví dụ

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

Phần `[errors]` liệt kê các tệp không được xử lý hoàn toàn. Nó chỉ xuất hiện khi có chẩn đoán.

```text
- <path>: <message>
```

```skt
[errors]
- path/to/broken.go: syntax error at line 42
```

## [dict]

Phần `[dict]` chỉ xuất hiện trong đầu ra đã được rút gọn (`--minify`). Nó ánh xạ các mã ngắn tới các token lặp lại.

| Phần       | Mô tả                                                                 |
| ---------- | --------------------------------------------------------------------- |
| `d0`, `d1`, … | Mã từ điển.                                                        |
| Dòng mục   | `- <code>: <token>`                                                   |

Để đọc đầu ra đã được rút gọn, hãy thay thế mỗi mã bằng giá trị từ điển của nó. Các mã có thể thay thế các cặp danh mục/loại và thuộc tính trong `[symbols]`. Chúng cũng có thể thay thế các loại mối quan hệ trong `[edges]`. Các đánh dấu không chắc chắn vẫn có thể đọc được. Ví dụ, `S1 --d2[unresolved]--> S3` là một cạnh chưa giải quyết có loại là giá trị của `d2`.

### Ví dụ đã rút gọn

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

Ở đây `S2` là một `callable/method` với thuộc tính `async=false`.

## Ví dụ đầy đủ

Đầu ra sau đây chứa `[symbols]`, `[edges]`, và `[errors]`. Nó không được rút gọn, vì vậy không có phần `[dict]`; xem [ví dụ đã rút gọn](#ví-dụ-đã-rút-gọn).

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

## Đầu ra JSON

`--format json`