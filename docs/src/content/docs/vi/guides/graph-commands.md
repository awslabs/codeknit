---
title: Các lệnh Đồ thị
description: Trực quan hóa và phân tích cấu trúc codebase của bạn bằng các thuật toán đồ thị.
---

Sử dụng các lệnh con `graph` để xem cách code của bạn kết nối như thế nào, tìm ra các vấn đề về cấu trúc và quyết định tệp nào cần refactor trước. Hướng dẫn này giả định bạn đã [cài đặt codeknit](../../installation/) và có thể chạy [`codeknit parse`](../parse-command/).

| Tác vụ | Lệnh |
| --- | --- |
| Khám phá đồ thị một cách trực quan | `codeknit graph show` |
| Tìm các vấn đề về cấu trúc | `codeknit graph analyze` |
| Tìm các tệp quan trọng, dễ thay đổi | `codeknit graph hotspots` |

## Trực quan hóa đồ thị code

`graph show` phân tích đầu vào và ghi một tệp HTML độc lập. Các ký hiệu (hàm, lớp, kiểu) là các nút. Các mối quan hệ (gọi, chứa, triển khai) là các cạnh.

```bash
codeknit graph show ./myproject
```

Tệp được ghi vào `./skeleton/codeknit-graph.html` và mở trong trình duyệt mặc định của bạn.

Để chọn đường dẫn đầu ra hoặc bao gồm các tệp kiểm thử:

```bash
codeknit graph show ./myproject -o graph.html
codeknit graph show ./src --collect-test
```

| Cờ | Mặc định | Mục đích |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/codeknit-graph.html` | Đường dẫn tệp HTML đầu ra |
| `--collect-test` | `false` | Bao gồm các tệp kiểm thử |
| `--workers` | `0` (`NumCPU`) | Số lượng goroutine phân tích đồng thời tối đa |
| `--verbose` | `false` | In thông tin tiến trình |

### Kiểm tra các mối quan hệ chưa được giải quyết

Trình xem chỉ vẽ các mối quan hệ đã được giải quyết. Nó hiển thị số lượng các mối quan hệ không chắc chắn (`unresolved` hoặc `ambiguous`) nhưng không vẽ chúng hoặc các liên kết ứng viên của chúng.

Để xem các cạnh không chắc chắn riêng lẻ, chạy `parse` với `--edges`:

```bash
codeknit parse ./src --output-mode inline --edges
```

Thêm `--format json` để xuất JSON. Xem [tài liệu tham khảo định dạng đầu ra](../../reference/output-format/) để biết cách các cạnh và ứng viên được ghi.

## Phân tích cấu trúc và tìm các vấn đề

`graph analyze` chạy 22 thuật toán đồ thị và ghi báo cáo `.skt` có thể đọc được bởi LLM.

```bash
codeknit graph analyze ./myproject
```

Báo cáo được ghi vào `./skeleton/graph_analysis.skt`. Đọc trực tiếp hoặc cung cấp cho LLM cùng với mã nguồn liên quan.

### Điều chỉnh báo cáo

Tăng ngưỡng để thấy ít phát hiện hơn, hoặc giảm nó để thấy nhiều hơn:

```bash
codeknit graph analyze ./myproject -o analysis.skt --fan-threshold 15
codeknit graph analyze ./myproject --top-n 50
codeknit graph analyze ./src --collect-test
```

`--top-n 0` loại bỏ giới hạn trên các phần được xếp hạng.

| Cờ | Mặc định | Mục đích |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/graph_analysis.skt` | Đường dẫn tệp `.skt` đầu ra |
| `--collect-test` | `false` | Bao gồm các tệp kiểm thử |
| `--workers` | `0` (`NumCPU`) | Số lượng goroutine phân tích đồng thời tối đa |
| `--verbose` | `false` | In thông tin tiến trình |
| `--fan-threshold` | `10` | Số lượng fan-in hoặc fan-out tối thiểu để gắn cờ một ký hiệu hub |
| `--god-threshold` | `15` | Số lượng cạnh chứa tối thiểu để gắn cờ một god class hoặc hàm |
| `--max-inheritance-depth` | `5` | Gắn cờ các chuỗi kế thừa sâu hơn mức này |
| `--top-n` | `30` | Giới hạn các phần được xếp hạng; `0` có nghĩa là không giới hạn |
| `--betweenness-threshold` | `0.001` | Giá trị trung tâm betweenness tối thiểu để báo cáo |
| `--propagation-cutoff` | `0.05` | Điểm heuristic tối thiểu để tiếp tục lan truyền thay đổi |

### Biết thuật toán nào trả lời câu hỏi nào

| Thuật toán | Sử dụng để tìm |
| --- | --- |
| Phụ thuộc vòng (Tarjan's SCC) | Các nhóm ký hiệu phụ thuộc lẫn nhau theo vòng lặp |
| Phát hiện hub | Các ký hiệu có tính kết nối fan-in hoặc fan-out cao |
| Phát hiện orphan | Các ứng viên mã chết |
| Phát hiện god class/hàm | Các ký hiệu chứa quá nhiều phần tử con |
| Chỉ số bất ổn | Các ký hiệu và gói dễ vỡ hoặc ổn định, sử dụng Ce/(Ca+Ce) |
| Chuỗi kế thừa sâu | Các hệ thống phân cấp sâu hơn `--max-inheritance-depth` |
| Betweenness centrality | Các nút thắt cổ chai mà nhiều đường phụ thuộc đi qua |
| Điểm khớp | Các điểm lỗi đơn lẻ mà khi loại bỏ sẽ chia cắt đồ thị |
| PageRank | Các ký hiệu quan trọng một cách đệ quy |
| Fan-in bắc cầu | Phạm vi ảnh hưởng: mọi thứ phụ thuộc vào một ký hiệu, trực tiếp hoặc gián tiếp |
| Mô phỏng lan truyền thay đổi | Mức độ lan rộng của một thay đổi đối với một ký hiệu |
| Phụ thuộc vòng giữa các gói | Các vòng lặp phụ thuộc ở cấp độ gói |
| Phát hiện vi phạm lớp | Các phụ thuộc vượt qua các lớp theo hướng sai |
| Khả năng tiếp cận từ các điểm vào | Mã mà không điểm vào nào tiếp cận được |
| Thành phần kết nối yếu | Các cụm mã bị cô lập |
| Trọng số phụ thuộc | Mức độ kết nối mạnh giữa các gói |
| Khoảng cách từ Chuỗi Chính | Các gói có tính trừu tượng và bất ổn không cân bằng |
| Phát hiện shotgun surgery | Các thay đổi sẽ ảnh hưởng đến nhiều nơi trong một gói |
| Phát hiện feature envy | Các ký hiệu sử dụng gói khác nhiều hơn gói của chính nó |
| Vi phạm phụ thuộc ổn định | Các gói ổn định phụ thuộc vào các gói kém ổn định hơn |
| Vi phạm phân tách giao diện | Các giao diện mà hầu hết các thành viên không được sử dụng bởi người dùng của chúng |
| Độ sâu chứa | Các ký hiệu lồng nhau sâu |

### Đọc kết quả một cách chính xác

Những quy tắc này ảnh hưởng đến mức độ bạn có thể tin tưởng một phát hiện:

- **Các mối quan hệ không chắc chắn bị loại trừ.** Các chỉ số phụ thuộc chỉ sử dụng các cạnh đã được giải quyết. Báo cáo đếm số lượng các mối quan hệ `unresolved` và `ambiguous` mà nó đã loại bỏ.
- **Siêu dữ liệu tệp và bao gồm bị bỏ qua.** Các ký hiệu `meta/file` và `meta/include` không ảnh hưởng đến xếp hạng phụ thuộc hoặc phát hiện mã chết.
- **Fan-in và fan-out đếm các láng giềng riêng biệt.** Chứa không được tính là một phụ thuộc.
- **Độ sâu kế thừa bao gồm tất cả các base và interface.** Các hệ thống phân cấp vòng không được gán độ sâu.
- **Khả năng tiếp cận bắt đầu từ các callable `main`, `Main`, hoặc `init`** và theo các cạnh phụ thuộc. Một thành viên có thể tiếp cận làm cho container của nó có liên quan, nhưng không làm cho các thành viên anh em của nó có liên quan. Nếu quá trình quét không tìm thấy điểm vào được nhận diện, các phát hiện về khả năng tiếp cận bị bỏ qua.
- **Callback là các tham chiếu, không phải lời gọi.** Một đối số callback hoặc callable được trả về không chứng minh rằng callback được chạy.
- **Điểm số là heuristic.** Lan truyền thay đổi sử dụng đường phụ thuộc mạnh nhất. Điểm số của nó và các quy tắc lớp không phải là xác suất đo lường hoặc đảm bảo của trình biên dịch.

### Hành động dựa trên các phát hiện một cách an toàn

Xem mỗi phát hiện như một ứng viên để xem xét. Trước khi thay đổi mã, hãy mở các ký hiệu mà báo cáo đề cập và kiểm tra:

- Các caller bên ngoài thư mục được quét.
- Dispatch động.
- Các phụ thuộc mà bạn không quét.

Phân tích dựa trên cú pháp, vì vậy bất kỳ điều nào trong số này đều có thể thay đổi kết luận. Xem [giải quyết mối quan hệ trong tài liệu tham khảo hỗ trợ ngôn ngữ](../../reference/language-support/#relationship-resolution) để biết các quy tắc và giới hạn giải quyết.

## Tìm các điểm nóng theo lịch sử

`graph hotspots` xếp hạng các tệp thay đổi thường xuyên và nằm ở các vị trí quan trọng về mặt cấu trúc. Nó cũng báo cáo tính kết hợp theo thời gian: các tệp thay đổi cùng nhau trong các commit lặp đi lặp lại. Chạy nó trên một thư mục bên trong kho Git.

```bash
codeknit graph hotspots ./myproject
```

Báo cáo được ghi vào `./skeleton/hotspots.skt`.

Nó kết hợp hai nhóm chỉ số:

- **Lịch sử:** tần suất commit, churn dòng thêm và xóa, và tính mới.
- **Cấu trúc:** PageRank cấp tệp, fan-in bắc cầu, và betweenness.

Báo cáo bao gồm mức độ tin cậy dựa trên số lượng commit liên quan được phân tích. Xem xếp hạng từ báo cáo có mức độ tin cậy thấp như tạm thời.

### Thay đổi cửa sổ lịch sử và bộ lọc

```bash
# Phân tích hai năm và ghi JSON
codeknit graph hotspots ./myproject --since 2y --format json -o hotspots.json

# Cho phép các commit lớn hơn và yêu cầu tính kết hợp theo thời gian mạnh hơn
codeknit graph hotspots . --max-files-per-commit 100 --min-cochanges 5
```

Các commit merge bị loại trừ trừ khi bạn truyền `--include-merges`. Các commit chạm đến nhiều tệp hơn `--max-files-per-commit` bị bỏ qua như các commit hàng loạt, vì vậy các thay đổi được tạo hoặc cơ học không làm méo mó xếp hạng.

| Cờ | Mặc định | Mục đích |
| --- | --- | --- |
| `-o`, `--output` | `./skeleton/hotspots.skt` | Đường dẫn tệp đầu ra |
| `--format` | `skt` | Định dạng đầu ra: `skt` hoặc `json` |
| `--since` | `12mo` | Cửa sổ lịch sử, chẳng hạn như `180d`, `12mo`, hoặc `2y` |
| `--max-commits` | `2000` | Số lượng commit tối đa để kiểm tra |
| `--max-files-per-commit` | `50` | Loại trừ các commit thay đổi nhiều tệp hơn mức này |
| `--min-cochanges` | `3` | Số lượng commit chung tối thiểu để báo cáo tính kết hợp theo thời gian |
| `--top-n` | `30` | Số lượng kết quả tối đa mỗi phần |
| `--include-merges` | `false` | Bao gồm các commit merge trong chỉ số lịch sử |
| `--collect-test` | `false` | Bao gồm các tệp kiểm thử |
| `--workers` | `0` (`NumCPU`) | Số lượng goroutine phân tích đồng thời tối đa |
| `--verbose` | `false` | In thông tin tiến trình |

### Đọc đầu ra JSON

Với `--format json`, mỗi điểm nóng có `score`, `history_score`, `structure_score`, `recency_score`, `pagerank`, `betweenness`, `commits`, `churn`, và `transitive_fan_in`. Mỗi mục tính kết hợp theo thời gian có `left`, `right`, `strength`, và `co_changes`. Cấp cao nhất cũng ghi lại `commits_visited`, `commits_analyzed`, `skipped_merges`, và `skipped_bulk_commits`. Sử dụng các số liệu này để kiểm tra lượng lịch sử mà xếp hạng dựa trên.

## Kiểm soát các nguồn được quét

Tất cả năm lệnh xử lý nguồn (`parse`, `fingerprint`, `graph show`, `graph analyze`, và `graph hotspots`) chấp nhận các cờ lựa chọn giống nhau:

| Cờ | Mặc định | Mục đích |
| --- | --- | --- |
| `--include GLOB` | Tất cả các nguồn được hỗ trợ | Bao gồm các đường dẫn khớp; có thể lặp lại |
| `--exclude GLOB` | Không có | Loại trừ các đường dẫn hoặc thư mục khớp; có thể lặp lại. Loại trừ thắng bao gồm |
| `--max-file-size BYTES` | `10485760` (10 MiB) | Bỏ qua các tệp lớn hơn; `0` vô hiệu hóa giới hạn |
| `--strict` | `false` | Thất bại trước khi ghi đầu ra nếu quá trình quét hoặc phân tích không hoàn chỉnh |

Các mẫu là các glob kiểu gitignore tương đối với thư mục đầu vào. Trích dẫn chúng để shell của bạn không mở rộng chúng. Bao gồm không ghi đè các quy tắc `.gitignore` hoặc lọc kiểm thử.

```bash
codeknit graph analyze ./src --include '**/*.go' --exclude '**/generated/' --strict
```

Các liên kết tượng trưng và các tệp không phải tệp thông thường bị bỏ qua và được báo cáo. Các vấn đề quét và phân tích xuất hiện trên stderr. Nếu không có `--strict`, codeknit giữ lại các kết quả một phần có thể sử dụng, vì vậy hãy kiểm tra stderr trước khi tin tưởng một báo cáo. Với `--strict`, một phân tích không hoàn chỉnh dừng trước khi bất kỳ đầu ra nào được ghi. Một lần chạy mà mọi nguồn đọc đều thất bại luôn thất bại.

Để biết các đảm bảo chính xác, hãy xem [tài liệu tham khảo độ tin cậy phân tích](../../reference/analysis-reliability/).

## Các trang liên quan

- [Tài liệu tham khảo cờ CLI](../../reference/cli-flags/): mọi cờ cho mọi lệnh.
- [Tài liệu tham khảo định dạng đầu ra](../../reference/output-format/): cú pháp `.skt` và JSON.
- [Tài liệu tham khảo hỗ trợ ngôn ngữ](../../reference/language-support/): giới hạn trích xuất và giải quyết.
- [Hướng dẫn lệnh parse](../parse-command/): trích xuất các ký hiệu và cạnh cho các tệp riêng lẻ.