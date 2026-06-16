---
title: Bắt đầu nhanh
description: Bắt đầu sử dụng codeknit trong vòng chưa đầy 5 phút.
---

# Bắt đầu nhanh

Bắt đầu sử dụng **codeknit** trong vòng chưa đầy 5 phút.

## 1. Điều kiện tiên quyết

Bạn cần:

- Go 1.26+
- Trình biên dịch C (CGo là bắt buộc cho **tree-sitter**)

## 2. Cài đặt từ mã nguồn

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
# Binary nằm tại ./bin/codeknit
```

## 3. Thêm vào PATH

Thêm binary vào PATH của shell:

```bash
# bash (~/.bashrc)
export PATH="$PATH:$(pwd)/bin"

# zsh (~/.zshrc)
export PATH="$PATH:$(pwd)/bin"

# fish (~/.config/fish/config.fish)
fish_add_path $(pwd)/bin
```

Tải lại shell hoặc chạy `source ~/.bashrc` (hoặc `~/.zshrc`) để thay đổi có hiệu lực.

## 4. Xác minh cài đặt

Kiểm tra **codeknit** hoạt động:

```bash
codeknit --version
```

## 5. Phân tích đầu tiên

Chạy phân tích đầu tiên trên một codebase:

```bash
codeknit parse ./myproject
```

Lệnh này sẽ:

- Phân tích tất cả các tệp nguồn trong `./myproject`
- Trích xuất thông tin cấu trúc (hàm, lớp, mối quan hệ)
- Ghi các tệp `.skt` đã phân mảnh vào `./skeleton/` (thư mục đầu ra mặc định)

Nếu bạn chạy lại lệnh này, hãy sử dụng `--clean` để xóa đầu ra trước đó:

```bash
codeknit parse ./myproject --clean
```

## 6. Đọc đầu ra

Các tệp `.skt` chứa thông tin mã có cấu trúc. Dưới đây là một ví dụ nhỏ:

```skt
[symbols]
## src/main.go
S1 module/package L1-L1 main {}
S2 type/struct L5-L8 Server {exported}
S3 callable/function L10-L12 NewServer(addr: string) -> *S2 {exported}
S4 callable/method L14-L19 Start() {receiver=*Server}

[edges]
S2 --contains--> S4
S3 --returns--> S2
```

Các phần chính:

- `[symbols]`: Các định nghĩa được nhóm theo tệp, hiển thị tên, **phạm vi dòng** và siêu dữ liệu
- `[edges]`: Các mối quan hệ như `contains`, `calls`, `inherits` hoặc `returns`

## 7. Các bước tiếp theo

Bây giờ bạn đã chạy phân tích đầu tiên:

- Tìm hiểu thêm về lệnh parse: [Hướng dẫn lệnh Parse](/codeknit/vi/guides/parse-command/)
- Khám phá phân tích cấu trúc: [Hướng dẫn lệnh Graph](/codeknit/vi/guides/graph-commands/)
- Hiểu về phát hiện trùng lặp: [Hướng dẫn lệnh Fingerprint](/codeknit/vi/guides/fingerprint-command/)
- Đọc định dạng đầu ra đầy đủ: [Tài liệu tham khảo định dạng đầu ra](/codeknit/vi/reference/output-format/)
- Xem tất cả các cờ có sẵn: [Tài liệu tham khảo cờ CLI](/codeknit/vi/reference/cli-flags/)
