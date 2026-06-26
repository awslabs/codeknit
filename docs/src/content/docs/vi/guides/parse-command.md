---
title: Lệnh Parse
description: Trích xuất thông tin cấu trúc từ mã nguồn vào các tệp .skt hoặc JSON.
---

Lệnh `codeknit parse` trích xuất thông tin cấu trúc từ cơ sở mã của bạn — chẳng hạn như hàm, lớp, phương thức, biến và các mối quan hệ của chúng — và xuất ra dưới định dạng `.skt` gọn nhẹ theo mặc định. Sử dụng JSON khi bạn cần đầu ra có thể đọc được bằng máy cho các tập lệnh, tích hợp hoặc công cụ hạ nguồn.

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

# Xuất JSON có thể đọc được bằng máy ra stdout
codeknit parse ./src --output-mode inline --format json
```

## Chế độ đầu ra

Sử dụng `--output-mode` để kiểm soát cách cấu trúc đầu ra. Có ba chế độ khả dụng:

| Chế độ            | Mô tả                                                                                     | Phù hợp nhất cho                                      |
| ----------------- | ----------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| `directory-flat`  | Ghi các tệp `.skt` phân mảnh (ví dụ: `map_001.skt`, `map_002.skt`) vào thư mục đầu ra.    | ✅ **Hầu hết các dự án** — chế độ mặc định và được khuyến nghị |
| `directory-tree`  | Phản chiếu cấu trúc thư mục nguồn, tạo một tệp `.skt` cho mỗi tệp nguồn.                  | Duyệt đầu ra cùng với mã nguồn                       |
| `inline`          | Xuất tất cả đầu ra ra stdout.                                                             | Tệp đơn hoặc chuyển tiếp đến các công cụ khác        |

> **Mẹo**: Sử dụng `directory-flat` theo mặc định trừ khi bạn làm việc với một tệp đơn. Tránh sử dụng `inline` cho các đầu vào lớn vì nó có thể làm quá tải các cửa sổ ngữ cảnh.

## Cờ

| Cờ               | Mặc định         | Mô tả                                                                                     |
| ---------------- | ---------------- | ----------------------------------------------------------------------------------------- |
| `--output-mode`  | `directory-flat` | Chế độ đầu ra: `inline`, `directory-flat`, hoặc `directory-tree`                          |
| `--format`       | `skt`            | Định dạng đầu ra: `skt` hoặc `json`                                                       |
| `--max-lines`    | `500`            | Số dòng tối đa cho mỗi tệp đầu ra trong chế độ flat/tree                                  |
| `--collect-test` | `false`          | Bao gồm các tệp kiểm thử trong phân tích                                                  |
| `--minify`       | `false`          | Bật nén dựa trên từ điển để giảm sử dụng token                                             |
| `--edges`        | `false`          | Bao gồm phần `[edges]` với dữ liệu mối quan hệ (gọi, chứa, v.v.)                          |
| `--clean`        | `false`          | Xóa các tệp `.skt` hiện có trong thư mục đầu ra trước khi ghi                             |
| `--workers`      | `NumCPU`         | Số lượng goroutine phân tích cú pháp đồng thời tối đa (0 = sử dụng tất cả lõi CPU)       |
| `--verbose`      | `false`          | In thông tin tiến trình và thời gian xử lý                                                |

## Các mẫu phổ biến

```bash
# Chạy lần đầu trên một dự án
codeknit parse ./src
```

```bash
# Chạy lại và làm sạch đầu ra trước đó
codeknit parse ./src --clean
```

```bash
# Phân tích cú pháp một tệp đơn ra stdout
codeknit parse ./src/main.go --output-mode inline
```

```bash
# Nén đầu ra cho các cơ sở mã lớn
codeknit parse ./src --minify
```

```bash
# Bao gồm các cạnh mối quan hệ (ví dụ: để phân tích phụ thuộc)
codeknit parse ./src --edges
```

```bash
# Xuất JSON cho một công cụ khác
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
# Phản chiếu cấu trúc cây thư mục nguồn trong đầu ra
codeknit parse ./src --output-mode directory-tree
```

## Bảo vệ đầu ra cũ

Nếu thư mục đầu ra đã chứa các tệp `.skt` từ lần chạy trước, `codeknit` sẽ từ chối ghi đầu ra mới để ngăn chặn việc trộn lẫn dữ liệu cũ và mới.

Để ghi đè hành vi này và làm sạch thư mục đầu ra trước khi ghi, hãy sử dụng cờ `--clean`:

```bash
codeknit parse ./src --clean
```

Điều này đảm bảo một tập đầu ra mới và nhất quán.

## Mẹo

- ✅ **Sử dụng `directory-flat` theo mặc định** cho hầu hết các dự án. Nó cân bằng giữa tính dễ đọc và khả năng quản lý.
- 🔍 Sử dụng `--minify` trên các cơ sở mã lớn để giảm sử dụng token thông qua từ điển chia sẻ (`dict.skt`).
- 🔗 Phần `[edges]` **được loại trừ theo mặc định** để tiết kiệm token. Sử dụng `--edges` khi bạn cần dữ liệu mối quan hệ như `calls`, `contains`, hoặc `inherits`.
- 🧾 Sử dụng `--format json` khi một tập lệnh hoặc tích hợp cần dữ liệu có cấu trúc thay vì `.skt`.
- 🧹 Luôn sử dụng `--clean` khi chạy lại trên cùng một thư mục đầu ra.
- 📁 Sử dụng `directory-tree` nếu bạn muốn tương quan các tệp `.skt` trực tiếp với các tệp nguồn trong trình soạn thảo của mình.