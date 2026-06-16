---
title: Lệnh Fingerprint
description: Phát hiện mã trùng lặp và gần trùng lặp trên các tệp và ngôn ngữ bằng cách sử dụng fuzzy hashing.
---

Lệnh `codeknit fingerprint` phát hiện mã trùng lặp và gần trùng lặp trong codebase của bạn bằng cách sử dụng **Context-Triggered Piecewise Hashing (CTPH)**. Nó hoạt động trên các tệp và thậm chí trên các ngôn ngữ lập trình khác nhau bằng cách chuẩn hóa tên biến, chuỗi ký tự và chú thích kiểu trước khi tính toán **fingerprint cấu trúc**.

## Chức năng

`codeknit fingerprint` phân tích mọi hàm, phương thức, biến và kiểu trong codebase của bạn và tính toán **fingerprint cấu trúc đã chuẩn hóa** dựa trên:

- Luồng điều khiển (`if`, `for`, `while`, `switch`)
- Các phép toán (`=`, `+`, `==`, `&&`, `||`)
- Lời gọi, trả về, gán và tạo đối tượng
- Các cấu trúc ngôn ngữ như `try/catch`, `yield`, `await`, `defer`

Việc chuẩn hóa này có nghĩa là **sao chép đổi tên**, **tái cấu trúc đơn giản** và **logic tương đương trong các ngôn ngữ khác nhau** vẫn có thể được phát hiện là trùng lặp.

Thuật toán sử dụng **CTPH** (một biến thể của rolling hash) để tìm kiếm gần trùng lặp một cách hiệu quả. Mã tương tự tạo ra fingerprint tương tự, cho phép khớp mờ ngay cả khi mã đã được chỉnh sửa nhẹ.

## Cách sử dụng cơ bản

```bash
codeknit fingerprint ./src
```

Lệnh này sẽ:

- Phân tích tất cả các tệp nguồn trong `./src`
- Tính toán fingerprint cấu trúc
- Xuất kết quả ra `./skeleton/fingerprints.skt`
- Báo cáo các kết quả trùng khớp với **độ tương đồng** từ **65% đến 95%** (khoảng mặc định)

## Các cờ lệnh

| Cờ lệnh            | Giá trị mặc định              | Mô tả                                                                                                                                                                     |
| ------------------ | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`   | `./skeleton/fingerprints.skt` | Đường dẫn tệp `.skt` đầu ra                                                                                                                                               |
| `--min-similarity` | `65`                          | **Độ tương đồng** tối thiểu để báo cáo (0–100)                                                                                                                            |
| `--max-similarity` | `95`                          | **Độ tương đồng** tối đa để báo cáo (0–100)                                                                                                                               |
| `--show-all`       | `false`                       | Bao gồm phần `[fingerprints]` với dữ liệu token thô                                                                                                                       |
| `--rerank`         | `false`                       | Sắp xếp lại các ứng viên CTPH bằng cách sử dụng embeddings ngữ nghĩa qua Ollama để loại bỏ dương tính giả (yêu cầu: `ollama serve` và `ollama pull qwen3-embedding:0.6b`) |
| `--model`          | `qwen3-embedding:0.6b`        | Mô hình embedding Ollama để sử dụng với `--rerank`                                                                                                                        |
| `--collect-test`   | `false`                       | Bao gồm các tệp kiểm thử trong phân tích                                                                                                                                  |
| `--workers`        | `NumCPU`                      | Số lượng goroutine phân tích đồng thời tối đa (0 = sử dụng tất cả lõi CPU)                                                                                                |
| `--verbose`        | `false`                       | Hiển thị thông tin tiến trình trong quá trình xử lý                                                                                                                       |

## Định dạng đầu ra

Đầu ra là một tệp `.skt` với các phần sau:

### `[duplicates]` (luôn có mặt)

Liệt kê các cặp ký hiệu có **độ tương đồng** vượt ngưỡng:

```skt
[duplicates]
similarity:96%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  utils/str.go::TrimSpaces <-> lib/text.go::CleanString
```

Mỗi dòng hiển thị:

- Phần trăm **độ tương đồng**
- Ký hiệu bên trái (đường dẫn tệp, phạm vi, tên)
- Ký hiệu bên phải (đường dẫn tệp, phạm vi, tên)

### `[fingerprints]` (chỉ có với `--show-all`)

Chứa dữ liệu fingerprint thô cho mỗi ký hiệu:

```skt
[fingerprints]
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Các trường:

- Tên ký hiệu
- `FP:<version>:<hash1>:<hash2>` — fingerprint CTPH
- `tokens:<hex>` — luồng token đã chuẩn hóa của nội dung

Phần này hữu ích cho việc gỡ lỗi hoặc xây dựng các công cụ hạ nguồn.

## Các mẫu lệnh phổ biến

```bash
# Quét mặc định
codeknit fingerprint ./src
```

```bash
# Tìm các bản sao chính xác
codeknit fingerprint ./src --min-similarity 100
```

```bash
# Tìm mã có độ tương đồng vừa phải (ví dụ: cùng thuật toán, tên khác nhau)
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

```bash
# Sử dụng sắp xếp lại ngữ nghĩa để giảm dương tính giả
# Yêu cầu: ollama serve && ollama pull qwen3-embedding:0.6b
codeknit fingerprint ./src --rerank
```

```bash
# Sử dụng mô hình embedding khác để sắp xếp lại
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

```bash
# Xuất danh sách fingerprint đầy đủ (cho các công cụ phân tích)
codeknit fingerprint ./src --show-all
```

```bash
# Tệp đầu ra tùy chỉnh
codeknit fingerprint ./src -o duplicates.skt
```

## Lựa chọn khoảng độ tương đồng

| Khoảng  | Hướng dẫn                                                                                             |
| ------- | ----------------------------------------------------------------------------------------------------- |
| 96–100% | Các bản sao cấu trúc chính xác hoặc gần chính xác. Gần như chắc chắn là sao chép-dán.                 |
| 85–95%  | Gần trùng lặp. Thường là sao chép-dán với chỉnh sửa nhỏ (ví dụ: đổi tên biến, thêm ghi log).          |
| 65–84%  | Khoảng mặc định. **Độ tương đồng** cấu trúc cao. Ứng viên tốt để tái cấu trúc.                        |
| 50–64%  | **Độ tương đồng** vừa phải. Cùng hình dạng thuật toán nhưng chi tiết khác nhau. Cần xem xét thủ công. |
| < 50%   | Thường là nhiễu. Không phải trùng lặp có ý nghĩa.                                                     |

## Mẹo

- **Fingerprint đo cấu trúc, không đo ý nghĩa**: Điểm **độ tương đồng** cao có nghĩa là mã _trông_ giống nhau, không phải là mã _thực hiện_ cùng một việc. Luôn xem xét cả hai ký hiệu.
- **Sử dụng `--rerank` cho kết quả nhiễu**: Nếu bạn nhận được nhiều dương tính giả, hãy bật sắp xếp lại ngữ nghĩa để lọc các kết quả trùng khớp bằng cách sử dụng embeddings.
- **Bỏ qua các nội dung ngắn**: Các ký hiệu có ít hơn 4 token đã chuẩn hóa (ví dụ: các getter đơn giản) sẽ bị bỏ qua để tránh nhiễu.
- **Khớp chéo ngôn ngữ hoạt động**: Các cấu trúc tương đương (ví dụ: một hàm Python và một hàm Go có cùng logic) có thể khớp, nhưng các mẫu đặc thù ngôn ngữ có thể tạo ra các kết quả trùng khớp có **độ tương đồng** thấp không chính xác.
- **Kết quả trùng khớp là tín hiệu, không phải phán quyết**: Hãy coi mỗi kết quả trùng khớp là một gợi ý để điều tra — không phải bằng chứng tự động về trùng lặp.
