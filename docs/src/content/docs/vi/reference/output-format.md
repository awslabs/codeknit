---
title: Tài liệu tham khảo định dạng đầu ra
description: Tài liệu tham khảo đầy đủ về định dạng đầu ra .skt được sử dụng bởi codeknit.
---

Định dạng `.skt` (skeleton) là một định dạng văn bản nhỏ gọn, dễ đọc được sử dụng bởi `codeknit` để biểu diễn cấu trúc mã đã trích xuất. Nó chứa các ký hiệu, mối quan hệ và siêu dữ liệu ở dạng tối giản phù hợp cho việc tiêu thụ bởi LLM và phân tích cấu trúc.

Tệp `.skt` được chia thành các phần. Mỗi phần bắt đầu bằng một tiêu đề trong dấu ngoặc vuông. Các phần có thể xuất hiện theo bất kỳ thứ tự nào, mặc dù `[symbols]` thường xuất hiện đầu tiên.

## [symbols]

Phần `[symbols]` liệt kê tất cả các ký hiệu đã trích xuất được nhóm theo tệp nguồn của chúng. Mỗi tệp được giới thiệu bằng tiêu đề `##` theo sau là đường dẫn tệp.

### Định dạng dòng

Mỗi ký hiệu được biểu diễn trên một dòng với cấu trúc sau:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Trường

- **ShortID**: Một định danh tuần tự được gán cho mỗi ký hiệu (ví dụ: `S1`, `S2`, `S3`). Được sử dụng làm tham chiếu trong các cạnh và các phần khác.
- **category/kind**: Một cặp phân cách bằng dấu gạch chéo cho biết danh mục và loại cụ thể của ký hiệu.
- **Lstart-Lend**: Phạm vi dòng trong tệp nguồn nơi ký hiệu được định nghĩa (ví dụ: `L10-L15`).
- **signature**: Tên và thông tin kiểu của ký hiệu. Định dạng phụ thuộc vào ký hiệu:
  - `name` — cho các kiểu, giá trị, mô-đun
  - `name(params)` — cho các callable không có kiểu trả về
  - `name(params) -> returnType` — cho các callable có kiểu trả về
- **{properties}**: Siêu dữ liệu tùy chọn được đặt trong dấu ngoặc nhọn. Nhiều thuộc tính được phân cách bằng dấu phẩy.

### Tham số

- Trong các ngôn ngữ không có kiểu: `paramName`
- Trong các ngôn ngữ có kiểu: `paramName: type`
- Các tham chiếu kiểu khớp với các ký hiệu đã biết được thay thế bằng ShortID của chúng (ví dụ: `config: S5` thay vì `config: Config`).

### Thuộc tính

Các thuộc tính phổ biến bao gồm:

- `async`: `true` hoặc `false`
- `exported`: `true` hoặc `false`
- `static`: có mặt nếu ký hiệu là static
- `visibility=public|private|protected`
- `receiver=*TypeName`: cho các phương thức, chỉ ra kiểu nhận

### Danh mục và loại ký hiệu

| Danh mục    | Loại                                      | Ví dụ                                           |
| ----------- | ----------------------------------------- | ----------------------------------------------- |
| `callable`  | function, method, constructor             | `callable/function`, `callable/method`         |
| `type`      | class, interface, struct, enum            | `type/class`, `type/interface`                  |
| `value`     | variable, constant, field                 | `value/variable`, `value/constant`              |
| `module`    | package, namespace                        | `module/package`                                |
| `meta`      | type parameters, files, includes, metadata | `meta/type_parameter`, `meta/file`, `meta/include` |

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

Phần `[edges]` ghi lại các mối quan hệ khi `--edges` được bật. Các điểm cuối đã biết sử dụng ShortID của chúng; các mục tiêu không thể giải quyết thành ký hiệu có thể xuất hiện dưới dạng tên.

### Định dạng dòng

```
FromID --kind--> ToID1, ToID2
FromID --kind[unresolved]--> Target1, Target2
FromID --kind[ambiguous]--> Target [candidates=CandidateID1, CandidateID2]
```

Việc bỏ qua trạng thái có nghĩa là trình phân tích đã giải quyết được mối quan hệ. `unresolved` có nghĩa là nó không thể thiết lập được phụ thuộc; `ambiguous` có nghĩa là vẫn còn các cách giải thích cạnh tranh. ID ứng viên xác định các mục tiêu có thể, không phải là các kết nối đã xác nhận. Việc giải quyết dựa trên cú pháp và phạm vi, không phải là đảm bảo của trình biên dịch.

Các mục tiêu chỉ được nhóm khi nguồn, loại, độ phân giải và tập ứng viên của chúng khớp nhau. Tên không có ký hiệu là tương đối với tệp nguồn. Các định danh phức tạp sử dụng ID đầy đủ được trích dẫn với các dấu ngoặc kép, dấu gạch chéo ngược và ngắt dòng được thoát. Một tên theo nghĩa đen giống như ShortID phải được trích dẫn như một ID đầy đủ, chẳng hạn như `"app.go::S2"`. Định nghĩa ký hiệu có thể xuất hiện trong các tệp hoặc khối đầu ra sau này.

```skt
S4 --references[unresolved]--> string, bool
S5 --calls[ambiguous]--> Helper [candidates=S8, S9]
S6 --references[unresolved]--> "app.go::S2"
```

### Điểm cuối tệp và include

Các include của C/C++ có các ký hiệu siêu dữ liệu rõ ràng. Các cách viết tiêu đề xuất hiện một lần trong danh sách ký hiệu, và các cạnh sử dụng ShortID mặc dù việc giải quyết tiêu đề vẫn chưa được giải quyết:

```skt
[symbols]
## src/main.c
S1 meta/file L1-L3 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3
```

Một ký hiệu include ghi lại chỉ thị đã viết, không phải là khai báo tiêu đề đã giải quyết. Các include lặp lại trong một tệp chia sẻ một điểm cuối. Siêu dữ liệu tệp/include bị loại khỏi xếp hạng phụ thuộc và phát hiện mã chết.

Các mối quan hệ không chắc chắn được bảo toàn trong đầu ra phân tích cú pháp nhưng bị loại khỏi các số liệu phụ thuộc. Các báo cáo đồ thị hiển thị số lượng mối quan hệ bị loại trừ; trình xem HTML hiển thị số lượng của chúng mà không vẽ các cạnh không chắc chắn riêng lẻ.

Khi nâng cấp lên `0.5.0`, hãy tái tạo tất cả các khối đầu ra cùng nhau. ID ký hiệu và kết quả mối quan hệ có thể thay đổi, vì vậy các khối từ các lần chạy khác nhau không được trộn lẫn.

### Loại cạnh

| Loại         | Ý nghĩa                                         |
| ------------ | ----------------------------------------------- |
| `calls`      | lời gọi hàm/phương thức                         |
| `contains`   | lớp chứa phương thức, mô-đun chứa hàm          |
| `inherits`   | lớp kế thừa lớp khác                            |
| `implements` | lớp triển khai giao diện                        |
| `overrides`  | phương thức ghi đè phương thức cha             |
| `references` | ký hiệu tham chiếu đến ký hiệu khác             |
| `imports`    | mô-đun nhập mô-đun khác                         |
| `decorates`  | decorator được áp dụng cho một ký hiệu          |

### Ví dụ

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

Phần `[errors]` liệt kê các tệp không thể phân tích cú pháp hoàn toàn.

### Định dạng

Mỗi dòng bắt đầu bằng `-` theo sau là đường dẫn tệp và thông báo lỗi:

```
- path/to/file.go: syntax error at line 42
```

### Ví dụ

```skt
[errors]
- src/broken.go: unexpected token at line 10
- tests/corner_case.py: unterminated string literal
```

## [dict]

Phần `[dict]` chỉ xuất hiện khi cờ `--minify` được sử dụng. Nó ánh xạ các mã từ điển ngắn tới các chuỗi được lặp lại để giảm kích thước đầu ra.

### Định dạng

Mỗi dòng ánh xạ một mã từ điển (`d0`, `d1`, v.v.) tới giá trị mở rộng của nó:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Trong phần còn lại của tệp, các mã này thay thế các giá trị đầy đủ của chúng.

Các loại mối quan hệ cũng có thể sử dụng mã từ điển trong khi các nhãn trạng thái vẫn dễ đọc: `S1 --d2[unresolved]--> S3`. Mở rộng `d2` bằng từ điển và giữ nguyên trạng thái unresolved.

### Ví dụ

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

## Ví dụ đầy đủ

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

## Mối quan hệ JSON

Sử dụng JSON khi một tập lệnh hoặc tích hợp cần đầu ra có cấu trúc:

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Mỗi cạnh bao gồm `from`, `to`, và `kind`. ID nguồn đã biết xuất hiện trong `from_short`; ID mục tiêu đã giải quyết xuất hiện trong `to_short`. Các cạnh không chắc chắn thêm `resolution` và, khi có sẵn, `candidates` chứa ID ký hiệu đầy đủ. Chúng bỏ qua `to_short`, ngay cả khi mục tiêu có ký hiệu siêu dữ liệu include.

Ví dụ, một lời gọi mơ hồ có thể được biểu diễn như sau:

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

JSON và SKT bảo toàn cùng một mức độ không chắc chắn với các mã hóa khác nhau. Xem [Hỗ trợ ngôn ngữ và độ trung thực API](../language-support/) để biết các quy tắc giải quyết và giới hạn đã biết.