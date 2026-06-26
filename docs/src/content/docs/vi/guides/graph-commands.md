---
title: Các lệnh đồ thị
description: Trực quan hóa và phân tích cấu trúc codebase của bạn bằng các thuật toán đồ thị.
---

`codeknit` cung cấp hai lệnh đồ thị mạnh mẽ giúp bạn hiểu và cải thiện cấu trúc codebase của mình: `graph show` để trực quan hóa tương tác và `graph analyze` để phân tích cấu trúc tự động.

## graph show

Tạo ra một biểu đồ HTML tương tác về codebase của bạn.

```bash
codeknit graph show <input-path>
```

Lệnh này phân tích codebase của bạn và tạo ra một tệp HTML độc lập với biểu đồ trực quan tương tác. Các ký hiệu (hàm, lớp, kiểu) xuất hiện dưới dạng nút, và các mối quan hệ của chúng (gọi, chứa, triển khai) dưới dạng cạnh. Biểu đồ trực quan sẽ tự động mở trong trình duyệt mặc định của bạn.

### Cờ

| Cờ               | Mặc định                          | Mô tả                                  |
| ---------------- | -------------------------------- | -------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Đường dẫn tệp HTML đầu ra              |
| `--collect-test` | `false`                          | Bao gồm các tệp kiểm thử trong phân tích |
| `--workers`      | `NumCPU`                         | Số lượng goroutine phân tích đồng thời tối đa |
| `--verbose`      | `false`                          | In thông tin tiến trình trong quá trình xử lý |

### Ví dụ

```skt
# Tạo biểu đồ trực quan mặc định
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

Lệnh này phát hiện các vấn đề chất lượng mã phổ biến như phụ thuộc vòng, các ký hiệu trung tâm, mã chết, các lớp hàm khổng lồ và các nút thắt cổ chai kiến trúc.

### Các thuật toán

Phân tích bao gồm 17 thuật toán đồ thị cấu trúc:

- Phụ thuộc vòng (Tarjan's SCC)
- Phát hiện trung tâm (kết nối fan-in/fan-out cao)
- Phát hiện mã mồ côi (ứng viên mã chết)
- Phát hiện lớp/hàm khổng lồ (số lượng con quá mức)
- Chỉ số bất ổn (Ce/(Ca+Ce) của Robert C. Martin)
- Chuỗi kế thừa sâu
- Trung tâm giữa (phát hiện nút thắt cổ chai)
- Điểm khớp (điểm lỗi đơn lẻ)
- PageRank (tầm quan trọng đệ quy)
- Fan-in bắc cầu (phạm vi ảnh hưởng)
- Mô phỏng lan truyền thay đổi
- Phụ thuộc vòng gói
- Phát hiện vi phạm lớp
- Khả năng tiếp cận từ các điểm vào
- Thành phần kết nối yếu
- Trọng số phụ thuộc (độ kết nối gói)
- Khoảng cách từ Chuỗi Chính (cân bằng A+I)

### Cờ

| Cờ                      | Mặc định                         | Mô tả                                              |
| ----------------------- | ------------------------------- | -------------------------------------------------- |
| `-o`, `--output`        | `./skeleton/graph_analysis.skt` | Đường dẫn tệp `.skt` đầu ra                        |
| `--collect-test`        | `false`                         | Bao gồm các tệp kiểm thử trong phân tích           |
| `--workers`             | `NumCPU`                        | Số lượng goroutine phân tích đồng thời tối đa      |
| `--verbose`             | `false`                         | In thông tin tiến trình trong quá trình xử lý      |
| `--fan-threshold`       | `10`                            | Số lượng fan-in hoặc fan-out tối thiểu để gắn cờ ký hiệu trung tâm |
| `--god-threshold`       | `15`                            | Số lượng cạnh chứa tối thiểu để gắn cờ lớp/hàm khổng lồ |
| `--max-inheritance-depth` | `5`                           | Gắn cờ các chuỗi kế thừa sâu hơn mức này           |
| `--top-n`               | `30`                            | Giới hạn các phần đầu ra được xếp hạng; 0 = không giới hạn |
| `--betweenness-threshold` | `0.001`                       | Giá trị trung tâm giữa tối thiểu để báo cáo        |
| `--propagation-cutoff`  | `0.05`                          | Xác suất tối thiểu để tiếp tục lan truyền thay đổi |

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