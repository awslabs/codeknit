---
title: Các lệnh Đồ thị
description: Trực quan hóa và phân tích cấu trúc codebase của bạn bằng các thuật toán đồ thị.
---

codeknit cung cấp các lệnh đồ thị để trực quan hóa cấu trúc, chạy phân tích tự động và kết hợp đồ thị phụ thuộc hiện tại với lịch sử thay đổi Git.

## graph show

Tạo ra một biểu đồ HTML tương tác để trực quan hóa codebase của bạn.

```bash
codeknit graph show <input-path>
```

Lệnh này phân tích codebase của bạn và tạo ra một tệp HTML độc lập với biểu đồ trực quan tương tác. Các ký hiệu (hàm, lớp, kiểu) xuất hiện dưới dạng các nút, và các mối quan hệ của chúng (gọi, chứa, triển khai) dưới dạng các cạnh. Trực quan hóa sẽ tự động mở trong trình duyệt mặc định của bạn.

Trình xem hiển thị các mối quan hệ đã được giải quyết. Nó hiển thị số lượng không chắc chắn nhưng không vẽ các mối quan hệ chưa được giải quyết hoặc mơ hồ riêng lẻ cũng như các liên kết ứng viên của chúng. Hãy kiểm tra các mối quan hệ đó bằng đầu ra phân tích:

```bash
codeknit parse ./src --output-mode inline --edges
codeknit parse ./src --output-mode inline --format json --edges
```

### Cờ

| Cờ               | Mặc định                          | Mô tả                                  |
| ---------------- | --------------------------------- | -------------------------------------- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html`  | Đường dẫn tệp HTML đầu ra              |
| `--collect-test` | `false`                           | Bao gồm các tệp kiểm thử trong phân tích |
| `--workers`      | `NumCPU`                          | Số lượng goroutine phân tích đồng thời tối đa |
| `--verbose`      | `false`                           | In thông tin tiến trình trong quá trình xử lý |

### Ví dụ

```bash
# Generate default visualization
codeknit graph show ./myproject

# Custom output file
codeknit graph show ./myproject -o graph.html

# Include test files
codeknit graph show ./src --collect-test
```

## graph analyze

Chạy các thuật toán đồ thị cấu trúc trên codebase của bạn và xuất ra báo cáo `.skt` có thể đọc được bởi LLM chứa các thông tin chi tiết về chất lượng mã.

```bash
codeknit graph analyze <input-path>
```

Lệnh này báo cáo các ứng viên cần xem xét, bao gồm phụ thuộc vòng, các ký hiệu trung tâm, mã chết, god class, và các nút thắt cổ chai kiến trúc.

### Các thuật toán

Phân tích bao gồm 22 thuật toán đồ thị cấu trúc:

- Phụ thuộc vòng (Tarjan's SCC)
- Phát hiện trung tâm (kết nối fan-in/fan-out cao)
- Phát hiện mã mồ côi (ứng viên mã chết)
- Phát hiện god class/hàm (quá nhiều con)
- Chỉ số bất ổn định (Robert C. Martin's Ce/(Ca+Ce))
- Chuỗi kế thừa sâu
- Trung tâm giữa (phát hiện nút thắt cổ chai)
- Điểm khớp (điểm lỗi đơn)
- PageRank (tầm quan trọng đệ quy)
- Fan-in bắc cầu (phạm vi ảnh hưởng)
- Mô phỏng lan truyền thay đổi
- Phụ thuộc vòng gói
- Phát hiện vi phạm lớp
- Khả năng tiếp cận từ các điểm vào
- Các thành phần kết nối yếu
- Trọng số phụ thuộc (độ kết nối gói)
- Khoảng cách từ Chuỗi Chính (cân bằng A+I)
- Phát hiện phẫu thuật shotgun
- Phát hiện ghen tị tính năng
- Vi phạm phụ thuộc ổn định
- Vi phạm phân tách giao diện
- Độ sâu chứa

### Giải thích báo cáo

Các chỉ số phụ thuộc sử dụng các mối quan hệ đã được giải quyết. Báo cáo đếm các mối quan hệ chưa được giải quyết và mơ hồ bị loại trừ. Siêu dữ liệu tệp/bao gồm bị loại khỏi xếp hạng phụ thuộc và phát hiện mã chết.

- Fan-in và fan-out đếm các láng giềng riêng biệt. Các mối quan hệ lặp lại không làm tăng các số liệu này, và chứa không phải là một phụ thuộc.
- Độ sâu kế thừa xem xét tất cả các cơ sở và giao diện. Các hệ thống phân cấp vòng không được gán độ sâu.
- Khả năng tiếp cận bắt đầu từ các callable có tên `main`, `Main`, hoặc `init` và theo các cạnh phụ thuộc. Việc tiếp cận một thành viên làm cho vùng chứa của nó trở nên liên quan mà không đánh dấu mọi anh chị em là có thể tiếp cận. Nếu không có điểm vào được nhận diện, các phát hiện về khả năng tiếp cận sẽ bị bỏ qua.
- Các đối số callback và các callable được trả về là các tham chiếu, không phải bằng chứng cho thấy callback được thực thi.
- Lan truyền thay đổi sử dụng đường phụ thuộc mạnh nhất. Điểm số và quy tắc lớp dựa trên thư mục của nó là các heuristic, không phải xác suất đo lường.

Các lời gọi từ bên ngoài, điều phối động, và các phụ thuộc bên ngoài phạm vi quét có thể thay đổi kết luận. Xem [Hỗ trợ Ngôn ngữ và Độ trung thực API](../../reference/language-support/) để biết các quy tắc và giới hạn giải quyết.

### Cờ

| Cờ                         | Mặc định                         | Mô tả                                              |
| -------------------------- | -------------------------------- | -------------------------------------------------- |
| `-o`, `--output`           | `./skeleton/graph_analysis.skt`  | Đường dẫn tệp `.skt` đầu ra                        |
| `--collect-test`           | `false`                          | Bao gồm các tệp kiểm thử trong phân tích           |
| `--workers`                | `NumCPU`                         | Số lượng goroutine phân tích đồng thời tối đa      |
| `--verbose`                | `false`                          | In thông tin tiến trình trong quá trình xử lý      |
| `--fan-threshold`          | `10`                             | Ngưỡng fan-in hoặc fan-out tối thiểu để gắn cờ ký hiệu trung tâm |
| `--god-threshold`          | `15`                             | Số lượng cạnh chứa tối thiểu để gắn cờ god class/hàm |
| `--max-inheritance-depth`  | `5`                              | Gắn cờ các chuỗi kế thừa sâu hơn mức này           |
| `--top-n`                  | `30`                             | Giới hạn các phần đầu ra được xếp hạng; 0 = không giới hạn |
| `--betweenness-threshold`  | `0.001`                          | Giá trị trung tâm giữa tối thiểu để báo cáo        |
| `--propagation-cutoff`     | `0.05`                           | Điểm heuristic tối thiểu để tiếp tục lan truyền thay đổi |

### Ví dụ

```bash
# Run structural analysis with defaults
codeknit graph analyze ./myproject

# Custom output and thresholds
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15

# Show more results per section
codeknit graph analyze ./myproject --top-n 50

# Remove the cap on ranked sections
codeknit graph analyze ./myproject --top-n 0

# Include test files
codeknit graph analyze ./src --collect-test
```

## graph hotspots

Xếp hạng các tệp vừa thường xuyên thay đổi vừa quan trọng về mặt cấu trúc:

```bash
codeknit graph hotspots <input-path>
```

Điểm số kết hợp tần suất commit, sự thay đổi dòng, và tính mới nhất với PageRank cấp tệp, fan-in bắc cầu, và trung tâm giữa. Báo cáo cũng xác định sự kết hợp thời gian giữa các tệp thường xuyên thay đổi trong cùng các commit.

Các commit merge bị loại trừ theo mặc định. Các commit thay đổi hơn 50 tệp cũng bị loại trừ để các thay đổi hàng loạt được tạo ra, được bán hoặc cơ học không làm méo mó kết quả.

### Cờ

| Cờ                      | Mặc định                   | Mô tả                                     |
| ----------------------- | -------------------------- | ----------------------------------------- |
| `-o`, `--output`        | `./skeleton/hotspots.skt`  | Đường dẫn tệp đầu ra                      |
| `--format`              | `skt`                      | Định dạng đầu ra: `skt` hoặc `json`       |
| `--since`               | `12mo`                     | Cửa sổ lịch sử, chẳng hạn `180d`, `12mo`, hoặc `2y` |
| `--max-commits`         | `2000`                     | Số lượng commit tối đa để kiểm tra        |
| `--max-files-per-commit`| `50`                       | Loại trừ các commit thay đổi nhiều tệp hơn |
| `--min-cochanges`       | `3`                        | Số lượng commit chung tối thiểu cho sự kết hợp thời gian |
| `--top-n`               | `30`                       | Số lượng kết quả tối đa cho mỗi phần báo cáo |
| `--include-merges`      | `false`                    | Bao gồm các commit merge                   |
| `--collect-test`        | `false`                    | Bao gồm các tệp kiểm thử                  |
| `--workers`             | `NumCPU`                   | Số lượng goroutine phân tích đồng thời tối đa |
| `--verbose`             | `false`                    | In thông tin tiến trình                   |

### Ví dụ

```bash
# Analyze the last 12 months
codeknit graph hotspots ./myproject

# Analyze two years and emit JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Include larger commits and require stronger coupling
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```