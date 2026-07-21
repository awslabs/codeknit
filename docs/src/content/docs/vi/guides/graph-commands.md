---
title: Các lệnh Đồ thị
description: Trực quan hóa và phân tích cấu trúc codebase của bạn bằng các thuật toán đồ thị.
---

codeknit cung cấp các lệnh đồ thị để trực quan hóa cấu trúc, chạy phân tích tự động và kết hợp đồ thị phụ thuộc hiện tại với lịch sử thay đổi Git.

## graph show

Tạo ra một đồ thị HTML tương tác trực quan hóa codebase của bạn.

```bash
codeknit graph show <input-path>
```

Lệnh này phân tích codebase của bạn và tạo ra một tệp HTML độc lập với trực quan hóa đồ thị tương tác. Các ký hiệu (hàm, lớp, kiểu) xuất hiện dưới dạng các nút, và các mối quan hệ của chúng (gọi, chứa, triển khai) dưới dạng các cạnh. Trực quan hóa sẽ tự động mở trong trình duyệt mặc định của bạn.

### Cờ

| Cờ               | Mặc định                          | Mô tả                                  |
| ---------------- | -------------------------------- | -------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Đường dẫn tệp HTML đầu ra                        |
| `--collect-test` | `false`                          | Bao gồm các tệp kiểm thử trong phân tích               |
| `--workers`      | `NumCPU`                         | Số lượng goroutine phân tích đồng thời tối đa            |
| `--verbose`      | `false`                          | In thông tin tiến trình trong quá trình xử lý |

### Ví dụ

```skt
# Tạo trực quan hóa mặc định
codeknit graph show ./myproject

# Tệp đầu ra tùy chỉnh
codeknit graph show ./myproject -o graph.html

# Bao gồm các tệp kiểm thử
codeknit graph show ./src --collect-test
```

## graph analyze

Chạy các thuật toán đồ thị cấu trúc trên codebase của bạn và xuất ra báo cáo `.skt` có thể đọc được bởi LLM chứa các thông tin chi tiết về chất lượng mã.

```bash
codeknit graph analyze <input-path>
```

Lệnh này phát hiện các vấn đề chất lượng mã phổ biến như phụ thuộc vòng, các ký hiệu trung tâm, mã chết, god class, và các nút thắt cổ chai kiến trúc.

### Các thuật toán

Phân tích bao gồm 22 thuật toán đồ thị cấu trúc:

- Phụ thuộc vòng (Tarjan's SCC)
- Phát hiện trung tâm (kết nối fan-in/fan-out cao)
- Phát hiện mã mồ côi (ứng cử viên mã chết)
- Phát hiện god class/hàm (quá nhiều con)
- Chỉ số bất ổn định (Robert C. Martin's Ce/(Ca+Ce))
- Chuỗi kế thừa sâu
- Tính trung tâm giữa (phát hiện nút thắt cổ chai)
- Các điểm khớp (điểm lỗi đơn lẻ)
- PageRank (tầm quan trọng đệ quy)
- Transitive fan-in (phạm vi ảnh hưởng)
- Mô phỏng lan truyền thay đổi
- Phụ thuộc vòng giữa các gói
- Phát hiện vi phạm lớp
- Khả năng tiếp cận từ các điểm vào
- Các thành phần kết nối yếu
- Trọng số phụ thuộc (độ kết nối giữa các gói)
- Khoảng cách từ Chuỗi Chính (cân bằng A+I)
- Phát hiện phẫu thuật shotgun
- Phát hiện ghen tị tính năng
- Vi phạm phụ thuộc ổn định
- Vi phạm phân tách giao diện
- Độ sâu chứa

### Cờ

| Cờ                      | Mặc định                         | Mô tả                                              |
| ------------------------- | ------------------------------- | -------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Đường dẫn tệp `.skt` đầu ra                                  |
| `--collect-test`          | `false`                         | Bao gồm các tệp kiểm thử trong phân tích                           |
| `--workers`               | `NumCPU`                        | Số lượng goroutine phân tích đồng thời tối đa                        |
| `--verbose`               | `false`                         | In thông tin tiến trình trong quá trình xử lý             |
| `--fan-threshold`         | `10`                            | Số lượng fan-in hoặc fan-out tối thiểu để gắn cờ ký hiệu trung tâm           |
| `--god-threshold`         | `15`                            | Số lượng cạnh chứa tối thiểu để gắn cờ god class/hàm |
| `--max-inheritance-depth` | `5`                             | Gắn cờ các chuỗi kế thừa sâu hơn mức này                 |
| `--top-n`                 | `30`                            | Giới hạn các phần đầu ra được xếp hạng; 0 = không giới hạn                 |
| `--betweenness-threshold` | `0.001`                         | Giá trị tính trung tâm giữa tối thiểu để báo cáo           |
| `--propagation-cutoff`    | `0.05`                          | Xác suất tối thiểu để tiếp tục lan truyền thay đổi       |

### Ví dụ

```skt
# Chạy phân tích cấu trúc với mặc định
codeknit graph analyze ./myproject

# Đầu ra và ngưỡng tùy chỉnh
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Hiển thị nhiều kết quả hơn cho mỗi phần
codeknit graph analyze ./myproject --top-n 50

# Bao gồm các tệp kiểm thử
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Xếp hạng các tệp vừa thường xuyên thay đổi vừa quan trọng về mặt cấu trúc:

```bash
codeknit graph hotspots <input-path>
```

Điểm số kết hợp tần suất commit, mức độ thay đổi dòng và tính mới nhất với PageRank cấp tệp, transitive fan-in và tính trung tâm giữa. Báo cáo cũng xác định sự kết hợp thời gian giữa các tệp thường xuyên thay đổi trong cùng các commit.

Các commit merge được loại trừ theo mặc định. Các commit thay đổi hơn 50 tệp cũng bị loại trừ để các thay đổi hàng loạt được tạo ra, được bán hoặc cơ học không làm sai lệch kết quả.

### Cờ

| Cờ                     | Mặc định                   | Mô tả                                      |
| ------------------------ | ------------------------- | ------------------------------------------------ |
| `-o`, `--output`         | `./skeleton/hotspots.skt` | Đường dẫn tệp đầu ra                                 |
| `--format`               | `skt`                     | Định dạng đầu ra: `skt` hoặc `json`                   |
| `--since`                | `12mo`                    | Cửa sổ lịch sử, chẳng hạn như `180d`, `12mo`, hoặc `2y`  |
| `--max-commits`          | `2000`                    | Số lượng commit tối đa để kiểm tra                       |
| `--max-files-per-commit` | `50`                      | Loại trừ các commit thay đổi nhiều tệp hơn              |
| `--min-cochanges`        | `3`                       | Số lượng commit chia sẻ tối thiểu cho sự kết hợp thời gian     |
| `--top-n`                | `30`                      | Số lượng kết quả tối đa cho mỗi phần báo cáo               |
| `--include-merges`       | `false`                   | Bao gồm các commit merge                            |
| `--collect-test`         | `false`                   | Bao gồm các tệp kiểm thử                               |
| `--workers`              | `NumCPU`                  | Số lượng goroutine phân tích đồng thời tối đa            |
| `--verbose`              | `false`                   | In thông tin tiến trình                       |

### Ví dụ

```bash
# Phân tích 12 tháng gần nhất
codeknit graph hotspots ./myproject

# Phân tích hai năm và xuất ra JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Bao gồm các commit lớn hơn và yêu cầu kết hợp mạnh hơn
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```