---
title: Các lệnh Đồ thị
description: Trực quan hóa và phân tích cấu trúc codebase của bạn bằng các thuật toán đồ thị.
---

codeknit cung cấp hai lệnh đồ thị mạnh mẽ giúp bạn hiểu và cải thiện cấu trúc codebase: `graph show` để trực quan hóa tương tác và `graph analyze` để phân tích cấu trúc tự động.

## graph show

Tạo biểu đồ trực quan HTML tương tác cho codebase của bạn.

```bash
codeknit graph show <input-path>
```

Lệnh này phân tích codebase của bạn và tạo ra một tệp HTML độc lập với biểu đồ trực quan tương tác. Các **ký hiệu** (hàm, lớp, kiểu) xuất hiện dưới dạng nút, và các mối quan hệ của chúng (gọi, chứa, triển khai) dưới dạng **cạnh**. Biểu đồ trực quan sẽ tự động mở trong trình duyệt mặc định của bạn.

### Cờ

| Cờ               | Giá trị mặc định                 | Mô tả                                               |
| ---------------- | -------------------------------- | --------------------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Đường dẫn tệp HTML đầu ra                           |
| `--collect-test` | `false`                          | Bao gồm các tệp kiểm thử trong phân tích            |
| `--workers`      | `NumCPU`                         | Số lượng goroutine phân tích đồng thời tối đa       |
| `--verbose`      | `false`                          | Hiển thị thông tin tiến trình trong quá trình xử lý |

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

Chạy các thuật toán đồ thị cấu trúc trên codebase của bạn và xuất báo cáo `.skt` có thể đọc được bởi LLM chứa các thông tin chi tiết về chất lượng mã.

```bash
codeknit graph analyze <input-path>
```

Lệnh này phát hiện các vấn đề chất lượng mã phổ biến như **phụ thuộc vòng**, các **ký hiệu** hub, **mã chết**, **god class**, và các **nút thắt cổ chai** kiến trúc.

### Thuật toán

Phân tích bao gồm 17 thuật toán đồ thị cấu trúc:

- **Phụ thuộc vòng** (Tarjan's SCC)
- Phát hiện hub (độ kết nối **fan-in**/**fan-out** cao)
- Phát hiện mã mồ côi (ứng viên **mã chết**)
- Phát hiện **god class**/hàm (số lượng con quá mức)
- Chỉ số bất ổn (Robert C. Martin's Ce/(Ca+Ce))
- Chuỗi kế thừa sâu
- Độ trung tâm trung gian (**nút thắt cổ chai**)
- Điểm khớp (điểm lỗi đơn)
- PageRank (tầm quan trọng đệ quy)
- **Fan-in** bắc cầu (**phạm vi ảnh hưởng**)
- Mô phỏng lan truyền thay đổi
- Phụ thuộc gói vòng
- Phát hiện vi phạm lớp
- Khả năng tiếp cận từ các điểm vào
- Thành phần liên thông yếu
- Trọng số phụ thuộc (độ kết nối gói)
- Khoảng cách từ Chuỗi Chính (cân bằng A+I)

### Cờ

| Cờ                        | Giá trị mặc định                | Mô tả                                                                  |
| ------------------------- | ------------------------------- | ---------------------------------------------------------------------- |
| `-o`, `--output`          | `./skeleton/graph_analysis.skt` | Đường dẫn tệp `.skt` đầu ra                                            |
| `--collect-test`          | `false`                         | Bao gồm các tệp kiểm thử trong phân tích                               |
| `--workers`               | `NumCPU`                        | Số lượng goroutine phân tích đồng thời tối đa                          |
| `--verbose`               | `false`                         | Hiển thị thông tin tiến trình trong quá trình xử lý                    |
| `--fan-threshold`         | `10`                            | Ngưỡng **fan-in** hoặc **fan-out** tối thiểu để gắn cờ **ký hiệu** hub |
| `--god-threshold`         | `15`                            | Số lượng **cạnh** chứa tối thiểu để gắn cờ **god class**/hàm           |
| `--max-inheritance-depth` | `5`                             | Gắn cờ các chuỗi kế thừa sâu hơn mức này                               |
| `--top-n`                 | `30`                            | Giới hạn các phần đầu ra được xếp hạng; 0 = không giới hạn             |
| `--betweenness-threshold` | `0.001`                         | Giá trị độ trung tâm trung gian tối thiểu để báo cáo                   |
| `--propagation-cutoff`    | `0.05`                          | Xác suất tối thiểu để tiếp tục lan truyền thay đổi                     |

### Ví dụ

```skt
# Chạy phân tích cấu trúc với giá trị mặc định
codeknit graph analyze ./myproject

# Đầu ra và ngưỡng tùy chỉnh
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Hiển thị nhiều kết quả hơn cho mỗi phần
codeknit graph analyze ./myproject --top-n 50

# Bao gồm các tệp kiểm thử
codeknit graph analyze ./src --collect-test
```
