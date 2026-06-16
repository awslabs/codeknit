---
title: Lệnh Parse
description: Trích xuất thông tin cấu trúc từ mã nguồn vào các tệp .skt.
---

Lệnh `codeknit parse` trích xuất thông tin cấu trúc từ codebase của bạn — như hàm, lớp, phương thức, biến và các mối quan hệ của chúng — và xuất ra dưới định dạng `.skt` nhỏ gọn được thiết kế để tiêu thụ hiệu quả bởi các mô hình ngôn ngữ lớn (LLM) và công cụ phân tích.

## Cách sử dụng cơ bản

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Đường dẫn đến thư mục hoặc tệp bạn muốn phân tích cú pháp.
- **`[output-dir]`**: Thư mục đầu ra tùy chọn. Nếu không được cung cấp, mặc định là `./skeleton`.

### Ví dụ

```bash
# Phân tích cú pháp một dự án, xuất ra thư mục mặc định ./skeleton
codeknit parse ./src

# Phân tích cú pháp và ghi vào thư mục đầu ra tùy chỉnh
codeknit parse ./src ./output

# Phân tích cú pháp một tệp đơn và xuất ra stdout
codeknit parse ./src/main.go --output-mode inline
```

## Chế độ đầu ra

Sử dụng `--output-mode` để kiểm soát cách cấu trúc đầu ra. Ba chế độ có sẵn:

| Chế độ           | Mô tả                                                                                  | Phù hợp nhất cho                                               |
| ---------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------- |
| `directory-flat` | Ghi các tệp `.skt` phân mảnh (ví dụ: `map_001.skt`, `map_002.skt`) vào thư mục đầu ra. | ✅ **Hầu hết các dự án** — chế độ mặc định và được khuyến nghị |
| `directory-tree` | Phản ánh cấu trúc thư mục nguồn, tạo một tệp `.skt` cho mỗi tệp nguồn.                 | Duyệt đầu ra cùng với mã nguồn                                 |
| `inline`         | Xuất tất cả đầu ra ra stdout.                                                          | Tệp đơn hoặc chuyển tiếp đến các công cụ khác                  |

> **Mẹo**: Sử dụng `directory-flat` theo mặc định trừ khi bạn làm việc với một tệp đơn. Tránh sử dụng `inline` cho các đầu vào lớn vì nó có thể làm quá tải cửa sổ ngữ cảnh.

## Cờ

| Cờ               | Giá trị mặc định | Mô tả                                                                              |
| ---------------- | ---------------- | ---------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Định dạng đầu ra: `inline`, `directory-flat`, hoặc `directory-tree`                |
| `--max-lines`    | `500`            | Số dòng tối đa cho mỗi tệp đầu ra trong chế độ flat/tree                           |
| `--collect-test` | `false`          | Bao gồm các tệp kiểm thử trong phân tích                                           |
| `--minify`       | `false`          | Bật nén dựa trên từ điển để giảm sử dụng token                                     |
| `--edges`        | `false`          | Bao gồm phần `[edges]` với dữ liệu mối quan hệ (gọi, chứa, v.v.)                   |
| `--clean`        | `false`          | Xóa các tệp `.skt` hiện có trong thư mục đầu ra trước khi ghi                      |
| `--workers`      | `NumCPU`         | Số lượng goroutine phân tích cú pháp đồng thời tối đa (0 = sử dụng tất cả lõi CPU) |
| `--verbose`      | `false`          | In thông tin tiến trình và thời gian trong quá trình xử lý                         |

## Mẫu thường dùng

```bash
# Chạy lần đầu trên một dự án
codeknit parse ./src
```

```bash
# Chạy lại và dọn dẹp đầu ra trước đó
codeknit parse ./src --clean
```

```bash
# Phân tích cú pháp một tệp đơn ra stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Nén đầu ra cho các codebase lớn
codeknit parse ./src --minify
```

```bash
# Bao gồm các cạnh mối quan hệ (ví dụ: để phân tích phụ thuộc)
codeknit parse ./src --edges
```

```bash
# Phản ánh cấu trúc cây thư mục nguồn trong đầu ra
codeknit parse ./src --output-mode directory-tree
```

## Bảo vệ đầu ra cũ

Nếu thư mục đầu ra đã chứa các tệp `.skt` từ lần chạy trước, `codeknit` sẽ từ chối ghi đầu ra mới để ngăn việc trộn lẫn dữ liệu cũ và mới.

Để ghi đè hành vi này và dọn dẹp thư mục đầu ra trước khi ghi, hãy sử dụng cờ `--clean`:

```bash
codeknit parse ./src --clean
```

Điều này đảm bảo một bộ đầu ra mới và nhất quán.

## Mẹo

- ✅ **Sử dụng `directory-flat` theo mặc định** cho hầu hết các dự án. Nó cân bằng giữa khả năng đọc và quản lý.
- 🔍 Sử dụng `--minify` trên các codebase lớn để giảm sử dụng token thông qua từ điển chia sẻ (`dict.skt`).
- 🔗 Phần `[edges]` **được loại trừ theo mặc định** để tiết kiệm token. Sử dụng `--edges` khi bạn cần dữ liệu mối quan hệ như `calls`, `contains`, hoặc `inherits`.
- 🧹 Luôn sử dụng `--clean` khi chạy lại trên cùng một thư mục đầu ra.
- 📁 Sử dụng `directory-tree` nếu bạn muốn tương quan các tệp `.skt` trực tiếp với các tệp nguồn trong trình soạn thảo của mình.
