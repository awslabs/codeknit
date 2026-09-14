---
title: Lệnh Parse
description: Trích xuất thông tin cấu trúc từ mã nguồn vào các tệp .skt hoặc JSON.
---

Lệnh `codeknit parse` trích xuất thông tin cấu trúc từ codebase của bạn — như hàm, lớp, phương thức, biến và các mối quan hệ của chúng — và xuất ra dưới định dạng `.skt` gọn nhẹ theo mặc định. Sử dụng JSON khi bạn cần đầu ra có thể đọc được bằng máy cho các tập lệnh, tích hợp hoặc công cụ hạ nguồn.

## Cách sử dụng cơ bản

```bash
codeknit parse <input-path> [output-dir]
```

- **`<input-path>`**: Đường dẫn đến thư mục hoặc tệp bạn muốn phân tích.
- **`[output-dir]`**: Thư mục đầu ra tùy chọn. Nếu không được cung cấp, mặc định là `./skeleton`.

### Ví dụ

```bash
# Parse a project, output to default directory ./skeleton
codeknit parse ./src

# Parse and write to a custom output directory
codeknit parse ./src ./output

# Parse a single file and output to stdout
codeknit parse ./src/main.go --output-mode inline

# Emit machine-readable JSON to stdout
codeknit parse ./src --output-mode inline --format json
```

## Chế độ đầu ra

Sử dụng `--output-mode` để kiểm soát cách cấu trúc đầu ra. Có ba chế độ khả dụng:

| Chế độ            | Mô tả                                                                                     | Phù hợp nhất cho                                      |
| ----------------- | ----------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| `directory-flat`  | Ghi các tệp `.skt` phân mảnh (ví dụ: `map_001.skt`, `map_002.skt`) vào thư mục đầu ra.    | ✅ **Hầu hết các dự án** — chế độ mặc định và được khuyến nghị |
| `directory-tree`  | Phản chiếu cấu trúc thư mục nguồn, tạo một tệp `.skt` cho mỗi tệp nguồn.                  | Duyệt đầu ra cùng với mã nguồn                       |
| `inline`          | Xuất tất cả đầu ra ra stdout.                                                             | Tệp đơn hoặc chuyển tiếp đến các công cụ khác        |

> **Mẹo**: Sử dụng `directory-flat` theo mặc định trừ khi bạn làm việc với một tệp đơn. Tránh `inline` đối với đầu vào lớn vì nó có thể làm quá tải cửa sổ ngữ cảnh.

## Cờ

| Cờ               | Mặc định         | Mô tả                                                                                     |
| ---------------- | ---------------- | ----------------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Chế độ đầu ra: `inline`, `directory-flat`, hoặc `directory-tree`                          |
| `--format`       | `skt`            | Định dạng đầu ra: `skt` hoặc `json`                                                       |
| `--max-lines`    | `500`            | Số dòng tối đa trên mỗi tệp đầu ra trong chế độ flat/tree                                 |
| `--collect-test` | `false`          | Bao gồm các tệp kiểm thử trong phân tích                                                  |
| `--minify`       | `false`          | Bật nén dựa trên từ điển để giảm sử dụng token                                            |
| `--edges`        | `false`          | Bao gồm phần `[edges]` với dữ liệu mối quan hệ (gọi, chứa, v.v.)                          |
| `--clean`        | `false`          | Xóa các tệp `.skt` hiện có trong thư mục đầu ra trước khi ghi                             |
| `--workers`      | `NumCPU`         | Số lượng goroutine phân tích đồng thời tối đa (0 = sử dụng tất cả lõi CPU)               |
| `--verbose`      | `false`          | In thông tin tiến trình và thời gian trong quá trình xử lý                                |

## Các mẫu phổ biến

```bash
# First run on a project
codeknit parse ./src
```

```bash
# Re-run and clean previous output
codeknit parse ./src --clean
```

```bash
# Parse a single file to stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Minify output for large codebases
codeknit parse ./src --minify
```

```bash
# Include relationship edges (e.g., for dependency analysis)
codeknit parse ./src --edges
```

```bash
# Emit JSON for another tool
codeknit parse ./src --output-mode inline --format json --edges
```

Ví dụ đầu ra JSON:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

```bash
# Mirror source tree structure in output
codeknit parse ./src --output-mode directory-tree
```

## Sự không chắc chắn trong mối quan hệ

Với `--edges`, Codeknit bảo toàn các mối quan hệ ngay cả khi cú pháp và phạm vi không xác định được mục tiêu duy nhất:

```skt
S1 --calls--> S2
S1 --references[unresolved]--> string, bool
S1 --calls[ambiguous]--> Helper [candidates=S3, S4]
```

Việc bỏ qua trạng thái có nghĩa là đã được phân giải bởi bộ phân tích. `unresolved` có nghĩa là không thể thiết lập được phụ thuộc; `ambiguous` có nghĩa là vẫn còn các cách giải thích cạnh tranh. Các ứng viên là các mục tiêu có thể, không phải là kết nối đã được xác nhận. Các điểm cuối đã biết sử dụng ShortID; các mục tiêu chưa được phân giải không có ký hiệu có thể xuất hiện dưới dạng tên.

JSON mang cùng thông tin trong `resolution` và `candidates`, và bỏ qua `to_short` đối với các mục tiêu không chắc chắn. C/C++ bao gồm các ký hiệu `meta/file` và `meta/include` nên các cạnh sử dụng ShortID; điều này ghi lại các chỉ thị đã viết mà không khẳng định việc phân giải tiêu đề ở cấp trình biên dịch.

Các chỉ số phụ thuộc loại trừ các mối quan hệ không chắc chắn. Đồ thị HTML hiển thị số lượng của chúng mà không vẽ các liên kết không chắc chắn riêng lẻ. Xem [tài liệu tham khảo định dạng đầu ra](../../reference/output-format/) để biết ngữ pháp và [tài liệu tham khảo hỗ trợ ngôn ngữ](../../reference/language-support/) để biết giới hạn phân giải.

## Bảo vệ đầu ra cũ

Nếu thư mục đầu ra đã chứa các tệp `.skt` từ lần chạy trước, `codeknit` sẽ từ chối ghi đầu ra mới để ngăn việc trộn dữ liệu cũ và mới.

Để ghi đè hành vi này và dọn dẹp thư mục đầu ra trước khi ghi, hãy sử dụng cờ `--clean`:

```bash
codeknit parse ./src --clean
```

Điều này đảm bảo một tập đầu ra mới, nhất quán.

Khi nâng cấp lên `0.5.0`, hãy tạo lại toàn bộ tập đầu ra. ID ký hiệu và kết quả mối quan hệ có thể thay đổi; không trộn các phần từ các lần chạy khác nhau.

## Mẹo

- ✅ **Sử dụng `directory-flat` theo mặc định** cho hầu hết các dự án. Nó cân bằng giữa khả năng đọc và quản lý.
- 🔍 Sử dụng `--minify` trên các codebase lớn để giảm sử dụng token thông qua từ điển chia sẻ (`dict.skt`).
- 🔗 Phần `[edges]` **bị loại trừ theo mặc định** để tiết kiệm token. Sử dụng `--edges` khi bạn cần dữ liệu mối quan hệ như `calls`, `contains`, hoặc `inherits`.
- 🧾 Sử dụng `--format json` khi một tập lệnh hoặc tích hợp cần dữ liệu có cấu trúc thay vì `.skt`.
- 🧹 Luôn sử dụng `--clean` khi chạy lại trên cùng thư mục đầu ra.
- 📁 Sử dụng `directory-tree` nếu bạn muốn tương quan các tệp `.skt` trực tiếp với các tệp nguồn trong trình soạn thảo của mình.