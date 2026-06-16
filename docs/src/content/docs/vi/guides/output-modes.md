---
title: Chế độ đầu ra
description: Chọn chế độ đầu ra phù hợp cho kích thước dự án và quy trình làm việc của bạn.
---

codeknit hỗ trợ ba chế độ đầu ra, được điều khiển bởi cờ `--output-mode`. Mỗi chế độ xác định cách cấu trúc mã được trích xuất ghi vào đĩa (hoặc stdout).

### directory-flat (mặc định, được khuyến nghị)

- **Hành vi**: Ghi các tệp `.skt` phân mảnh như `map_001.skt`, `map_002.skt`, v.v.
- **Thư mục đầu ra**: `./skeleton/` theo mặc định
- **Phân mảnh**: Các tệp được phân mảnh khi vượt quá giới hạn `--max-lines` (mặc định: 500 dòng)
- **Trường hợp sử dụng**: Tốt nhất cho hầu hết các dự án. Giữ đầu ra có tổ chức và dễ đọc bằng cách giới hạn kích thước tệp. Bạn có thể chỉ đọc các phân mảnh liên quan đến nhiệm vụ của mình.
- **Nén**: Khi `--minify` được bật, một tệp `dict.skt` cũng được tạo trong thư mục đầu ra, chứa ánh xạ token cho các giá trị đã nén.

Ví dụ:

```bash
codeknit parse ./src
# Đầu ra: ./skeleton/map_001.skt, map_002.skt, ...
```

### directory-tree

- **Hành vi**: Phản ánh chính xác cấu trúc thư mục nguồn.
- **Thư mục đầu ra**: `./skeleton/` theo mặc định
- **Ánh xạ**: Một tệp `.skt` được tạo cho mỗi tệp nguồn, tại đường dẫn tương ứng.
- **Trường hợp sử dụng**: Lý tưởng khi bạn muốn tra cứu nhanh cấu trúc của một tệp cụ thể. Hữu ích cho việc điều hướng cùng với mã nguồn gốc.

Ví dụ:

```bash
codeknit parse ./src --output-mode directory-tree
# Đầu ra: ./skeleton/src/handler.skt, ./skeleton/pkg/db.skt, v.v.
```

### inline

- **Hành vi**: Xuất tất cả đầu ra ra stdout.
- **Thư mục đầu ra**: Không tạo thư mục
- **Trường hợp sử dụng**: Chỉ được khuyến nghị cho các tệp đơn hoặc dự án rất nhỏ (ít hơn 5 tệp). Hữu ích khi chuyển đầu ra đến một công cụ khác hoặc kiểm tra một tệp đơn lẻ tương tác.

Ví dụ:

```bash
codeknit parse ./src/main.go --output-mode inline
# Đầu ra: in trực tiếp ra terminal
```

### Bảng quyết định

| Chế độ           | Phù hợp nhất cho                               | Vị trí đầu ra                                       |
| ---------------- | ---------------------------------------------- | --------------------------------------------------- |
| `directory-flat` | Hầu hết các dự án (mặc định, được khuyến nghị) | `./skeleton/map_001.skt`, `map_002.skt`, ...        |
| `directory-tree` | Điều hướng đầu ra cùng với mã nguồn            | `./skeleton/<đường dẫn phản ánh>.skt`               |
| `inline`         | Tệp đơn, chuyển đến công cụ khác               | stdout — chỉ sử dụng cho tệp đơn hoặc dự án rất nhỏ |

### Nguyên tắc chung

- **Khi không chắc chắn** → sử dụng `directory-flat` (mặc định)
- **Kiểm tra tệp đơn** → `inline` có thể chấp nhận được
- **Nhiều hơn một vài tệp** → ưu tiên `directory-flat` hoặc `directory-tree`
- **Dự án mã nguồn lớn** → thêm `--minify` để giảm sử dụng token
- **Chạy lại trên cùng đầu ra** → sử dụng `--clean` để xóa các tệp `.skt` cũ

### Nén

Cờ `--minify` kích hoạt nén dựa trên từ điển cho các token lặp lại (ví dụ: các khóa thuộc tính như `exported`, `async`, hoặc các tên kiểu phổ biến). Khi được bật:

- Các giá trị lặp lại được thay thế bằng các mã ngắn (`d0`, `d1`, `d2`, ...)
- Một tệp `dict.skt` được ghi vào thư mục đầu ra, ánh xạ các mã với các giá trị gốc
- Giảm đáng kể kích thước đầu ra cho các dự án mã nguồn lớn
- Hoạt động trong cả hai chế độ `directory-flat` và `directory-tree`

Ví dụ đầu ra đã nén:

```skt
[dict]
- d0: exported
- d1: callable/function

[symbols]
## src/main.py
S1 d1 L1-L5 main() {d0}
```
