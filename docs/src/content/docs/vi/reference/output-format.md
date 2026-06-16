---
title: Tài liệu tham khảo định dạng đầu ra
description: Tài liệu tham khảo đầy đủ về định dạng đầu ra .skt được sử dụng bởi codeknit.
---

Định dạng `.skt` (skeleton) là một định dạng văn bản nhỏ gọn, dễ đọc được `codeknit` sử dụng để biểu diễn cấu trúc mã đã trích xuất. Nó chứa các ký hiệu, mối quan hệ và siêu dữ liệu ở dạng tối giản phù hợp cho việc tiêu thụ bởi LLM và phân tích cấu trúc.

Tệp `.skt` được chia thành các phần. Mỗi phần bắt đầu bằng một tiêu đề trong dấu ngoặc vuông. Các phần có thể xuất hiện theo bất kỳ thứ tự nào, mặc dù `[symbols]` thường xuất hiện đầu tiên.

## [symbols]

Phần `[symbols]` liệt kê tất cả các ký hiệu đã trích xuất được nhóm theo tệp nguồn của chúng. Mỗi tệp được giới thiệu bằng tiêu đề `##` theo sau là đường dẫn tệp.

### Định dạng dòng

Mỗi ký hiệu được biểu diễn trên một dòng duy nhất với cấu trúc sau:

```
ShortID category/kind Lstart-Lend signature {properties}
```

### Các trường

- **ShortID**: Một định danh tuần tự được gán cho mỗi ký hiệu (ví dụ: `S1`, `S2`, `S3`). Được sử dụng làm tham chiếu trong các cạnh và các phần khác.
- **category/kind**: Một cặp phân cách bằng dấu gạch chéo cho biết danh mục và loại cụ thể của ký hiệu.
- **Lstart-Lend**: phạm vi dòng trong tệp nguồn nơi ký hiệu được định nghĩa (ví dụ: `L10-L15`).
- **signature**: Tên và thông tin kiểu của ký hiệu. Định dạng phụ thuộc vào ký hiệu:
  - `name` — cho các loại, giá trị, mô-đun
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

| Danh mục   | Loại                           | Ví dụ                                  |
| ---------- | ------------------------------ | -------------------------------------- |
| `callable` | function, method, constructor  | `callable/function`, `callable/method` |
| `type`     | class, interface, struct, enum | `type/class`, `type/interface`         |
| `value`    | variable, constant, field      | `value/variable`, `value/constant`     |
| `module`   | package, namespace             | `module/package`                       |
| `meta`     | type parameters, metadata      | `meta/type_parameter`                  |

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

Phần `[edges]` định nghĩa các mối quan hệ giữa các ký hiệu bằng cách sử dụng ShortID của chúng.

### Định dạng dòng

```
FromID --kind--> ToID1, ToID2
```

Nhiều ID đích được phân cách bằng dấu phẩy. Mỗi dòng biểu diễn một mối quan hệ có hướng.

### Các loại cạnh

| Loại         | Ý nghĩa                                    |
| ------------ | ------------------------------------------ |
| `calls`      | lời gọi hàm/phương thức                    |
| `contains`   | lớp chứa phương thức, mô-đun chứa hàm      |
| `inherits`   | lớp kế thừa từ lớp khác                    |
| `implements` | lớp triển khai giao diện                   |
| `overrides`  | phương thức ghi đè phương thức của lớp cha |
| `references` | ký hiệu tham chiếu đến ký hiệu khác        |
| `imports`    | mô-đun nhập mô-đun khác                    |
| `decorates`  | decorator được áp dụng cho một ký hiệu     |

### Ví dụ

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
S10 --inherits--> S2
S24 --implements--> S19
```

## [errors]

Phần `[errors]` liệt kê các tệp không thể được phân tích cú pháp hoàn toàn.

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

Phần `[dict]` chỉ xuất hiện khi sử dụng cờ `--minify`. Nó ánh xạ các mã từ điển ngắn tới các token chuỗi lặp lại để giảm kích thước đầu ra.

### Định dạng

Mỗi dòng ánh xạ một mã từ điển (`d0`, `d1`, v.v.) tới giá trị mở rộng của nó:

```
- d0: async=false
- d1: callable/method
- d2: exported
```

Trong phần còn lại của tệp, các mã này thay thế các giá trị đầy đủ của chúng.

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
S3 --returns--> S2

[errors]
- utils/broken.go: syntax error at line 5
```
