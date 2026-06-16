---
title: Cài đặt
description: Cách cài đặt codeknit trên hệ thống của bạn.
---

codeknit có thể được cài đặt từ mã nguồn. Các bước sau sẽ hướng dẫn bạn thiết lập codeknit trên hệ thống của mình.

## Từ mã nguồn

Phương pháp cài đặt chính là xây dựng từ mã nguồn. Bạn sẽ cần:

- Go 1.26+
- Trình biên dịch C (yêu cầu cho tree-sitter thông qua CGo)

Sao chép kho lưu trữ và xây dựng tệp nhị phân:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Tệp nhị phân đã biên dịch sẽ có sẵn tại `./bin/codeknit`.

## Thêm vào PATH

Để chạy `codeknit` từ bất kỳ thư mục nào, hãy thêm vị trí tệp nhị phân vào PATH của hệ thống.

Đối với **bash** (`~/.bashrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

Đối với **zsh** (`~/.zshrc`):

```bash
export PATH="$PATH:/path/to/codeknit"
```

Đối với **fish** (`~/.config/fish/config.fish`):

```fish
fish_add_path /path/to/codeknit
```

Sau khi cập nhật cấu hình shell, hãy tải lại bằng cách chạy `source ~/.bashrc` (hoặc `~/.zshrc`) hoặc khởi động lại terminal.

## Hoàn thành shell

codeknit hỗ trợ tự động hoàn thành cho các shell phổ biến. Cài đặt hoàn thành bằng các lệnh sau:

Đối với **bash**:

```bash
codeknit completion bash >> ~/.bashrc
```

Đối với **zsh**:

```bash
codeknit completion zsh >> ~/.zshrc
```

Đối với **fish**:

```bash
codeknit completion fish > ~/.config/fish/completions/codeknit.fish
```

Đối với **PowerShell**:

```powershell
codeknit completion powershell >> $PROFILE
```

## Xác minh cài đặt

Sau khi cài đặt, hãy xác minh codeknit đã được thiết lập chính xác:

```bash
codeknit --version
```

## Thiết lập phát triển

Nếu bạn đóng góp cho codeknit, hãy chạy các lệnh bổ sung sau:

Cài đặt các phụ thuộc phát triển:

```bash
make deps
```

Thiết lập git hooks:

```bash
make setup
```

Chạy bộ kiểm tra:

```bash
make test
```
