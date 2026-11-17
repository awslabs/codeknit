---
title: Bắt đầu nhanh
description: Bắt đầu sử dụng codeknit trong vòng chưa đầy 5 phút.
---

Trong hướng dẫn này, bạn sẽ xây dựng codeknit từ mã nguồn, chạy phân tích đầu tiên và đọc tệp `.skt` mà nó xuất ra. Bạn sẽ phân tích một gói nhỏ từ kho lưu trữ codeknit, vì vậy bạn không cần dự án riêng của mình.

## Điều kiện tiên quyết

- **Go 1.26 hoặc mới hơn.** Chạy `go version` để kiểm tra.
- **Trình biên dịch C.** codeknit sử dụng tree-sitter, cần CGo.
- **Git** và **Make**.

## 1. Xây dựng codeknit

Sao chép kho lưu trữ và xây dựng:

```bash
git clone https://github.com/awslabs/codeknit.git
cd codeknit
make build
```

Bạn sẽ thấy quá trình xây dựng hoàn tất mà không có lỗi, và một tệp nhị phân tại `./bin/codeknit`.

## 2. Thêm codeknit vào PATH

Từ thư mục `codeknit`, thêm `bin` vào `PATH` bằng dòng lệnh cho shell của bạn.

Đối với bash hoặc zsh:

```bash
export PATH="$PATH:$(pwd)/bin"
```

Đối với fish:

```fish
fish_add_path (pwd)/bin
```

Thao tác này áp dụng cho phiên bash hoặc zsh hiện tại. Để giữ lại cho các phiên sau, hãy thêm cùng dòng lệnh đó vào `~/.bashrc` hoặc `~/.zshrc`, với đường dẫn tuyệt đối đến `bin` thay vì `$(pwd)/bin`.

## 3. Xác minh cài đặt

```bash
codeknit --version
```

Bạn sẽ thấy phiên bản codeknit được in ra.

## 4. Chạy phân tích đầu tiên

Ở lại trong thư mục `codeknit` và phân tích gói `internal/console`:

```bash
codeknit parse ./internal/console --edges
```

codeknit quét thư mục, phân tích từng tệp nguồn được hỗ trợ và ghi tệp `.skt`. *Tệp `.skt`* là tệp văn bản nhỏ gọn liệt kê các **ký hiệu** trong mã của bạn và các mối quan hệ giữa chúng. Cờ `--edges` thêm các mối quan hệ. Chúng được bỏ qua theo mặc định để tiết kiệm không gian.

Bạn sẽ thấy một bản tóm tắt ngắn khi lệnh hoàn tất và một thư mục `./skeleton` mới trong thư mục hiện tại.

Liệt kê đầu ra:

```bash
ls ./skeleton
```

Bạn sẽ thấy một tệp như `map_001.skt`.

## 5. Đọc đầu ra

Mở tệp:

```bash
cat ./skeleton/map_001.skt
```

Bạn sẽ thấy phần `[symbols]` đầu tiên. *Ký hiệu* là một mục có tên trong mã của bạn, chẳng hạn như hàm, kiểu hoặc trường. Các ký hiệu được nhóm dưới tiêu đề `## <đường dẫn đến>/<tệp nguồn>`. Đầu ra của bạn bao gồm các dòng như sau:

```skt
[symbols]
## <path to>/console.go
S1 module/package L5-L5 console
S2 type/struct L17-L25 Console {exported}
...
S10 callable/function L28-L41 New() -> *S2 {exported}
...
S12 callable/method L49-L49 Error(msg: string) {exported, receiver=*Console}
```

Mỗi dòng có các phần giống nhau:

| Phần | Ví dụ | Ý nghĩa |
| --- | --- | --- |
| Short ID | `S10` | Nhãn ngắn cho ký hiệu, được sử dụng ở nơi khác trong đầu ra |
| Danh mục và loại | `callable/function` | Loại ký hiệu đó là gì |
| Phạm vi dòng | `L28-L41` | Các dòng mà ký hiệu bao phủ trong tệp nguồn |
| Chữ ký | `New() -> *S2` | Tên, tham số và kiểu trả về |
| Thuộc tính | `{exported}` | Chi tiết bổ sung trong dấu ngoặc nhọn |

Trong `New() -> *S2`, `S2` là Short ID của kiểu `Console`. codeknit thay thế tên kiểu bằng Short ID khi biết ký hiệu. Số dòng có thể khác một chút nếu bản sao của bạn mới hơn.

Cuộn xuống phần `[edges]`. *Cạnh* là mối quan hệ giữa hai ký hiệu. Mỗi dòng cạnh ghi tên ký hiệu nguồn, loại mối quan hệ và ký hiệu hoặc các ký hiệu đích:

```skt
[edges]
S2 --contains--> S4
S4 --calls--> S5
```

Ví dụ này cho thấy hình dạng của một dòng cạnh. Nó có nghĩa là "`S2` chứa `S4`" và "`S4` gọi `S5`". Bạn có thể tra cứu từng Short ID trong `[symbols]` để xem nó đề cập đến gì.

## Những gì bạn đã đạt được

Bạn đã xây dựng codeknit, đưa nó vào `PATH`, chạy `codeknit parse` và đọc các phần `[symbols]` và `[edges]` của đầu ra `.skt`.

## Các bước tiếp theo

- Phân tích dự án của riêng bạn, chọn bố cục đầu ra khác và chạy lại một cách an toàn: xem [hướng dẫn lệnh parse](../guides/parse-command/).
- Trực quan hóa và phân tích **cấu trúc mã** của bạn: xem [hướng dẫn lệnh đồ thị](../guides/graph-commands/).
- Tìm mã **trùng lặp** và **gần trùng lặp**: xem [hướng dẫn lệnh fingerprint](../guides/fingerprint-command/).
- Tra cứu mọi phần của định dạng `.skt`: xem [tài liệu tham khảo định dạng đầu ra](../reference/output-format/).
- Tra cứu mọi cờ dòng lệnh: xem [tài liệu tham khảo cờ CLI](../reference/cli-flags/).