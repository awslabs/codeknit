---
title: Cài đặt
description: Cách cài đặt codeknit trên hệ thống của bạn.
---

Hướng dẫn này cài đặt codeknit từ nguồn, thêm nó vào `PATH`, kích hoạt tính năng hoàn thiện shell và kiểm tra xem nó có chạy được không.

:::tip[Đang sử dụng trợ lý mã hóa AI?]
Người dùng Claude Code, Codex và Kiro có thể cài đặt plugin codeknit thay thế. Plugin sẽ tải xuống tệp nhị phân lần đầu tiên chạy. Xem [Sử dụng codeknit với trợ lý AI](../guides/ai-assistants/).
:::

## Cài đặt từ nguồn

**Điều kiện tiên quyết**

- Go 1.26 hoặc mới hơn
- Trình biên dịch C. codeknit sử dụng CGo để xây dựng các trình phân tích cú pháp tree-sitter.

**Các bước**

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Tệp nhị phân được ghi vào `./bin/codeknit`.

### Sử dụng tệp nhị phân dựng sẵn thay thế

Các tệp nhị phân dựng sẵn cho Linux, macOS và Windows (amd64 và arm64) có trên [trang releases](https://github.com/awslabs/codeknit/releases). Tải xuống kho lưu trữ cho nền tảng của bạn, giải nén và di chuyển tệp nhị phân vào `PATH`. Mỗi bản phát hành bao gồm tệp `checksums.txt` để bạn có thể xác minh tệp tải xuống.

## Thêm codeknit vào PATH

Bỏ qua phần này nếu bạn đã di chuyển tệp nhị phân vào thư mục đã có trong `PATH`, chẳng hạn như `/usr/local/bin`.

Thay thế `/path/to/codeknit` bằng thư mục chứa tệp nhị phân. Nếu bạn xây dựng từ nguồn, đó là thư mục `bin` bên trong kho lưu trữ đã nhân bản.

Thêm dòng cho shell của bạn vào tệp cấu hình của nó:

```bash
# bash (~/.bashrc)
export PATH="$PATH:/path/to/codeknit"

# zsh (~/.zshrc)
export PATH="$PATH:/path/to/codeknit"

# fish (~/.config/fish/config.fish)
fish_add_path /path/to/codeknit
```

Tải lại shell hoặc nguồn tệp bạn đã chỉnh sửa:

```bash
source ~/.bashrc   # bash
source ~/.zshrc    # zsh
```

Đối với fish, hãy mở một terminal mới.

## Kích hoạt tính năng hoàn thiện shell

codeknit tạo các tập lệnh hoàn thiện cho bash, zsh, fish và PowerShell. Chạy lệnh cho shell của bạn một lần:

```bash
# bash
codeknit completion bash >> ~/.bashrc

# zsh
codeknit completion zsh >> ~/.zshrc

# fish
codeknit completion fish > ~/.config/fish/completions/codeknit.fish

# PowerShell
codeknit completion powershell >> $PROFILE
```

Các lệnh bash, zsh và PowerShell sẽ thêm vào tệp cấu hình của bạn. Chạy lại chúng sẽ thêm tập lệnh lần thứ hai, vì vậy chỉ chạy mỗi lệnh một lần.

Mở một shell mới hoặc nguồn tệp cấu hình của bạn để kích hoạt tính năng hoàn thiện.

## Xác minh cài đặt

```bash
codeknit --version
```

Nếu lệnh in ra một phiên bản, codeknit đã được cài đặt. Nếu shell báo `command not found`, hãy kiểm tra xem thư mục trong dòng `PATH` của bạn có chứa tệp nhị phân `codeknit` không, sau đó tải lại shell.

## Thiết lập môi trường phát triển

Chỉ thực hiện việc này nếu bạn dự định đóng góp cho codeknit. Chạy các lệnh này từ thư mục gốc của kho lưu trữ đã nhân bản:

```bash
make deps    # cài đặt các phụ thuộc
make setup   # cấu hình git hooks
make test    # chạy bộ kiểm tra
```

`make lint` chạy trình kiểm tra và định dạng.

## Các bước tiếp theo

- Chạy lệnh đầu tiên của bạn trong [Bắt đầu](../getting-started/).
- Duyệt mọi cờ trong [tài liệu tham khảo cờ CLI](../reference/cli-flags/).