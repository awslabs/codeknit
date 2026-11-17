---
title: Tài liệu tham khảo CLI
description: Tài liệu tham khảo đầy đủ cho tất cả các lệnh và cờ của codeknit.
---

Trang này liệt kê mọi lệnh của codeknit, các đối số và cờ của chúng. Để xem hướng dẫn theo tác vụ, hãy tham khảo [hướng dẫn lệnh parse](../../guides/parse-command/), [hướng dẫn lệnh đồ thị](../../guides/graph-commands/) và [hướng dẫn lệnh fingerprint](../../guides/fingerprint-command/).

## Tóm tắt lệnh

| Lệnh                                          | Mục đích                                                        |
| --------------------------------------------- | -------------------------------------------------------------- |
| `codeknit`                                    | Khởi chạy giao diện terminal tương tác.                        |
| `codeknit parse <input-path> [output-dir]`    | Trích xuất ký hiệu và mối quan hệ vào đầu ra `.skt` hoặc JSON. |
| `codeknit graph show <input-path>`            | Tạo đồ thị HTML tương tác.                                     |
| `codeknit graph analyze <input-path>`         | Chạy các thuật toán đồ thị cấu trúc và ghi báo cáo `.skt`.     |
| `codeknit graph hotspots <input-path>`        | Xếp hạng các điểm nóng thay đổi bằng lịch sử Git và cấu trúc đồ thị. |
| `codeknit fingerprint <input-path>`           | Phát hiện mã trùng lặp và gần trùng lặp.                       |
| `codeknit completion <shell>`                 | Tạo script hoàn thành shell.                                   |

Năm lệnh xử lý mã nguồn: `parse`, `graph show`, `graph analyze`, `graph hotspots` và `fingerprint`. Chúng chia sẻ các cờ được mô tả trong [Cờ xử lý nguồn chung](#các-cờ-xử-lý-nguồn-chung).

## Cờ toàn cục

| Cờ              | Kiểu | Mặc định | Mô tả                  |
| ----------------- | ---- | ------- | ---------------------------- |
| `--version`       | bool | `false` | Hiển thị thông tin phiên bản. |
| `-h`, `--help`    | bool | `false` | Hiển thị trợ giúp cho lệnh.   |

## Các cờ xử lý nguồn chung

Các cờ này xuất hiện trong bảng của mỗi lệnh xử lý nguồn.

| Cờ                    | Kiểu                | Mặc định               | Giá trị cho phép       | Mô tả                                                                          |
| ----------------------- | ------------------- | --------------------- | -------------------- | ------------------------------------------------------------------------------------ |
| `--header-language`     | string              | `c`                   | `c`, `cpp`           | Bộ phân tích cú pháp được sử dụng cho tệp `.h`.                                                          |
| `--include GLOB`        | string, có thể lặp lại | Tất cả các nguồn được hỗ trợ | Glob kiểu gitignore | Bao gồm các đường dẫn nguồn phù hợp. Nhiều mẫu tạo thành một hợp.                       |
| `--exclude GLOB`        | string, có thể lặp lại | Không có              | Glob kiểu gitignore | Loại trừ các đường dẫn hoặc thư mục phù hợp.                                               |
| `--max-file-size BYTES` | int                 | `10485760` (10 MiB)   | `0` hoặc lớn hơn       | Bỏ qua các tệp nguồn lớn hơn kích thước này. `0` vô hiệu hóa giới hạn.                     |
| `--strict`              | bool                | `false`               | —                    | Thất bại trước khi ghi đầu ra nếu xảy ra chẩn đoán quét hoặc phân tích cú pháp.                    |

### `--header-language`

Chỉ áp dụng cho tệp `.h`. Các phần mở rộng khác giữ bộ phân tích cú pháp thông thường (`.hpp` và `.hxx` đã sử dụng C++).

- `c` (mặc định) giữ lại việc trích xuất macro, typedef và union của C.
- `cpp` chọn bộ phân tích cú pháp C++. Sử dụng cho các dự án C++ có tiêu đề được đặt tên `.h`.

Giao diện người dùng terminal sử dụng lựa chọn `c` mặc định cho tệp `.h`.

```bash
codeknit parse ./src --header-language cpp
```

Để biết khả năng trích xuất đã được kiểm tra và giới hạn đã biết cho từng ngôn ngữ, hãy xem [Hỗ trợ ngôn ngữ và độ trung thực API](../language-support/). Trang đó mô tả các cấu trúc đã được kiểm tra, không phải phạm vi bao phủ ngôn ngữ hoặc ngữ nghĩa trình biên dịch hoàn chỉnh.

### Lựa chọn nguồn

- Các mẫu sử dụng glob kiểu gitignore phân tách bằng dấu gạch chéo tương đối với thư mục đầu vào. Đối với đầu vào tệp đơn, các mẫu khớp với tên cơ sở của tệp.
- Trích dẫn các mẫu để shell không mở rộng chúng.
- Loại trừ có ưu tiên cao hơn bao gồm.
- Bao gồm không ghi đè các quy tắc `.gitignore`, lọc tệp kiểm thử hoặc loại trừ tích hợp sẵn.
- Mẫu phủ định (`!pattern`) không được chấp nhận trong `--include` hoặc `--exclude`.
- Quét thư mục tuân theo các tệp `.gitignore` tổ tiên và lồng nhau độc lập với thư mục làm việc hiện tại.
- Việc chọn rõ ràng một tệp nguồn thông thường sẽ bỏ qua việc bỏ qua và lọc kiểm thử. Giới hạn bao gồm, loại trừ và kích thước vẫn được áp dụng.
- Tệp kiểm thử bị loại trừ trừ khi `--collect-test` được đặt.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

### Chẩn đoán và `--strict`

- Liên kết tượng trưng, liên kết tượng trưng thư mục và tệp không thông thường được bỏ qua và báo cáo.
- Đường dẫn không thể đọc được, lỗi đọc tệp bỏ qua, tệp nguồn quá lớn và cảnh báo cú pháp làm cho phân tích không hoàn chỉnh.
- Theo mặc định, các kết quả một phần có thể sử dụng được giữ lại. Chẩn đoán xuất hiện trên stderr. `parse` cũng ghi lại chúng trong `errors` JSON hoặc trong SKT `[errors]`.
- Với `--strict`, một phân tích không hoàn chỉnh bị từ chối trước khi đầu ra được ghi.
- Một lần chạy không có nguồn được hỗ trợ đã chọn, hoặc trong đó mọi lần đọc nguồn đều thất bại, luôn thoát không thành công.
- Giao diện người dùng terminal áp dụng cùng chính sách liên kết tượng trưng và mặc định 10 MiB. Các biểu mẫu của nó không hiển thị cờ lựa chọn và strict.

Xem [Độ tin cậy phân tích](../analysis-reliability/) để biết các đảm bảo chính xác.

## `codeknit`

```bash
codeknit
```

Chạy `codeknit` mà không có lệnh con sẽ khởi chạy giao diện terminal tương tác. Giao diện này cung cấp các lệnh `parse`, `graph show`, `graph analyze`, `graph hotspots` và `fingerprint` cùng với các tùy chọn của chúng.

## `codeknit parse`

```bash
codeknit parse <input-path> [output-dir]
```

| Đối số       | Bắt buộc | Mặc định      | Mô tả                                                              |
| -------------- | -------- | ------------ | ------------------------------------------------------------------------ |
| `input-path`   | Có       | —            | Thư mục nguồn hoặc tệp nguồn đơn.                                  |
| `output-dir`   | Không    | `./skeleton` | Thư mục đầu ra. Không được sử dụng khi `--output-mode` là `inline`.             |

| Cờ                    | Kiểu                | Mặc định            | Giá trị cho phép                                   | Mô tả                                                                                         |
| ----------------------- | ------------------- | ------------------ | ------------------------------------------------ | --------------------------------------------------------------------------------------------------- |
| `--output-mode`         | string              | `directory-flat`   | `inline`, `directory-flat`, `directory-tree`     | Nơi và cách thức ghi đầu ra.                                                                    |
| `--format`              | string              | `skt`              | `skt`, `json`                                    | Định dạng đầu ra.                                                                                      |
| `--max-lines`           | int                 | `500`              | Đối với `directory-tree`, 3 hoặc lớn hơn               | Mục tiêu dòng cho đầu ra phẳng. Giới hạn cứng cho phân đoạn nguồn trong đầu ra dạng cây.                               |
| `--collect-test`        | bool                | `false`            | —                                                | Bao gồm tệp kiểm thử.                                                                                 |
| `--minify`              | bool                | `false`            | —                                                | Bật nén đầu ra dựa trên từ điển.                                                         |
| `--edges`               | bool                | `false`            | —                                                | Bao gồm phần `[edges]`. Tắt theo mặc định để tiết kiệm token.                                       |
| `--clean`               | bool                | `false`            | —                                                | Thay thế đầu ra đã tạo trước đó, chỉ sau khi tạo thành công.                                   |
| `--workers`             | int                 | `0` (`NumCPU`)     | `0` hoặc lớn hơn                                   | Số lượng goroutine phân tích cú pháp đồng thời tối đa.                                                              |
| `--verbose`             | bool                | `false`            | —                                                | In thông tin tiến trình.                                                                         |
| `--header-language`     | string              | `c`                | `c`, `cpp`                                       | Bộ phân tích cú pháp được sử dụng cho tệp `.h`.                                                                         |
| `--include GLOB`        | string, có thể lặp lại  | Tất cả các nguồn được hỗ trợ | —                                             | Bao gồm các đường dẫn nguồn phù hợp.                                                                      |
| `--exclude GLOB`        | string, có thể lặp lại  | Không có               | —                                                | Loại trừ các đường dẫn hoặc thư mục phù hợp.                                                              |
| `--max-file-size BYTES` | int                 | `10485760`         | `0` vô hiệu hóa giới hạn                           | Bỏ qua các tệp nguồn lớn hơn.                                                                           |
| `--strict`              | bool                | `false`            | —                                                | Thất bại trước khi ghi đầu ra nếu xảy ra chẩn đoán quét hoặc phân tích cú pháp.                                              |

### Chế độ đầu ra

| Giá trị             | Hành vi                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------ |
| `inline`          | Ghi vào stdout. Không sử dụng thư mục đầu ra.                                                    |
| `directory-flat`  | Ghi các tệp được phân đoạn vào thư mục đầu ra. Các khối nguồn được giữ nguyên, vì vậy `--max-lines` là mục tiêu phân tách chứ không phải giới hạn cứng. |
| `directory-tree`  | Phản chiếu cây nguồn. Tên giữ phần mở rộng nguồn (`foo.c.skt`, `foo.h.skt`). Các tệp được phân tách sử dụng `foo.c_part1.skt`. |

### Hành vi đầu ra

- Không có `--clean`, đầu ra đã tạo trước đó cho định dạng đã chọn trong thư mục đầu ra sẽ gây ra lỗi.
- Với `--clean`, các tệp `.skt` lỗi thời (định dạng SKT) hoặc `codeknit.json` (định dạng JSON) được thay thế chỉ sau khi tạo thành công. Các mục không liên quan được giữ nguyên.
- Ở định dạng JSON, các chế độ thư mục ghi `codeknit.json`. JSON là một tài liệu đơn và không được phân tách theo số dòng.
- `dict.skt` và `warnings.skt` là tên tệp phụ trợ đã được đặt trước.
- Không sử dụng thư mục nguồn hoặc một trong các thư mục tổ tiên của nó làm thư mục đầu ra.

Xem [Chế độ đầu ra](../../guides/output-modes/) và [tài liệu tham khảo định dạng đầu ra](../output-format/) để biết chi tiết.

```bash
codeknit parse ./src
codeknit parse ./src ./out --output-mode directory-tree
codeknit parse ./src --output-mode inline --format json --edges
```

## `codeknit graph show`

```bash
codeknit graph show <input-path>
```

| Đối số       | Bắt buộc | Mô tả                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Có       | Thư mục nguồn hoặc tệp nguồn đơn.  |

| Cờ                    | Kiểu                | Mặc định                            | Giá trị cho phép          | Mô tả                                                       |
| ----------------------- | ------------------- | ---------------------------------- | ----------------------- | ----------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/codeknit-graph.html`   | Đường dẫn tệp               | Tệp HTML đầu ra.                                                 |
| `--collect-test`        | bool                | `false`                            | —                       | Bao gồm tệp kiểm thử.                                               |
| `--workers`             | int                 | `0` (`NumCPU`)                     | `0` hoặc lớn hơn          | Số lượng goroutine phân tích cú pháp đồng thời tối đa.                            |
| `--verbose`             | bool                | `false`                            | —                       | In thông tin tiến trình.                                       |
| `--header-language`     | string              | `c`                                | `c`, `cpp`              | Bộ phân tích cú pháp được sử dụng cho tệp `.h`.                                       |
| `--include GLOB`        | string, có thể lặp lại  | Tất cả các nguồn được hỗ trợ              | —                       | Bao gồm các đường dẫn nguồn phù hợp.                                    |
| `--exclude GLOB`        | string, có thể lặp lại  | Không có                               | —                       | Loại trừ các đường dẫn hoặc thư mục phù hợp.                            |
| `--max-file-size BYTES` | int                 | `10485760`                         | `0` vô hiệu hóa giới hạn  | Bỏ qua các tệp nguồn lớn hơn.                                         |
| `--strict`              | bool                | `false`                            | —                       | Thất bại trước khi ghi đầu ra nếu xảy ra chẩn đoán quét hoặc phân tích cú pháp.            |

Lệnh ghi một tệp HTML độc lập và mở nó trong trình duyệt mặc định. Trình xem vẽ các mối quan hệ đã được giải quyết và hiển thị số lượng các mối quan hệ không chắc chắn.

```bash
codeknit graph show ./src -o graph.html
```

## `codeknit graph analyze`

```bash
codeknit graph analyze <input-path>
```

| Đối số       | Bắt buộc | Mô tả                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Có       | Thư mục nguồn hoặc tệp nguồn đơn.  |

| Cờ                       | Kiểu                | Mặc định                            | Giá trị cho phép          | Mô tả                                                  |
| -------------------------- | ------------------- | ---------------------------------- | ----------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/graph_analysis.skt`    | Đường dẫn tệp               | Tệp `.skt` đầu ra.                                          |
| `--collect-test`           | bool                | `false`                            | —                       | Bao gồm tệp kiểm thử.                                          |
| `--workers`                | int                 | `0` (`NumCPU`)                     | `0` hoặc lớn hơn          | Số lượng goroutine phân tích cú pháp đồng thời tối đa.                       |
| `--verbose`                | bool                | `false`                            | —                       | In thông tin tiến trình.                                 |
| `--fan-threshold`          | int                 | `10`                               | —                       | Ngưỡng fan-in hoặc fan-out tối thiểu để gắn cờ một ký hiệu trung tâm.              |
| `--god-threshold`          | int                 | `15`                               | —                       | Số lượng cạnh `contains` tối thiểu để gắn cờ một god class hoặc hàm. |
| `--max-inheritance-depth`  | int                 | `5`                                | —                       | Gắn cờ các chuỗi kế thừa sâu hơn giá trị này.              |
| `--top-n`                  | int                 | `30`                               | `0` = không giới hạn          | Giới hạn các phần đầu ra được xếp hạng.                                  |
| `--betweenness-threshold`  | float               | `0.001`                            | —                       | Giá trị trung tâm betweenness tối thiểu để báo cáo.              |
| `--propagation-cutoff`     | float               | `0.05`                             | —                       | Xác suất tối thiểu để tiếp tục lan truyền thay đổi.          |
| `--header-language`        | string              | `c`                                | `c`, `cpp`              | Bộ phân tích cú pháp được sử dụng cho tệp `.h`.                                  |
| `--include GLOB`           | string, có thể lặp lại  | Tất cả các nguồn được hỗ trợ              | —                       | Bao gồm các đường dẫn nguồn phù hợp.                               |
| `--exclude GLOB`           | string, có thể lặp lại  | Không có                               | —                       | Loại trừ các đường dẫn hoặc thư mục phù hợp.                       |
| `--max-file-size BYTES`    | int                 | `10485760`                         | `0` vô hiệu hóa giới hạn  | Bỏ qua các tệp nguồn lớn hơn.                                    |
| `--strict`                 | bool                | `false`                            | —                       | Thất bại trước khi ghi đầu ra nếu xảy ra chẩn đoán quét hoặc phân tích cú pháp.       |

Các chỉ số phụ thuộc loại trừ các mối quan hệ không chắc chắn và siêu dữ liệu tệp/bao gồm. Báo cáo ghi lại số lượng mối quan hệ bị loại trừ.

```bash
codeknit graph analyze ./src -o analysis.skt --fan-threshold 15 --top-n 50
```

## `codeknit graph hotspots`

```bash
codeknit graph hotspots <input-path>
```

| Đối số       | Bắt buộc | Mô tả                                                  |
| -------------- | -------- | ------------------------------------------------------------ |
| `input-path`   | Có       | Thư mục nguồn bên trong kho lưu trữ Git.                    |

| Cờ                       | Kiểu                | Mặc định                    | Giá trị cho phép                              | Mô tả                                                  |
| -------------------------- | ------------------- | -------------------------- | ------------------------------------------- | ------------------------------------------------------------ |
| `-o`, `--output`           | string              | `./skeleton/hotspots.skt`  | Đường dẫn tệp                                   | Tệp đầu ra.                                                 |
| `--format`                 | string              | `skt`                      | `skt`, `json`                               | Định dạng đầu ra.                                               |
| `--since`                  | string              | `12mo`                     | Cửa sổ như `180d`, `12mo`, hoặc `2y`      | Cửa sổ lịch sử.                                              |
| `--max-commits`            | int                 | `2000`                     | —                                           | Số lượng commit tối đa để kiểm tra.                                  |
| `--max-files-per-commit`   | int                 | `50`                       | —                                           | Loại trừ các commit chạm đến nhiều tệp hơn giá trị này.       |
| `--min-cochanges`          | int                 | `3`                        | —                                           | Số lượng commit chia sẻ tối thiểu để báo cáo ghép nối thời gian.          |
| `--top-n`                  | int                 | `30`                       | —                                           | Số lượng kết quả tối đa cho mỗi phần.                                 |
| `--include-merges`         | bool                | `false`                    | —                                           | Bao gồm các commit merge trong chỉ số lịch sử.                    |
| `--collect-test`           | bool                | `false`                    | —                                           | Bao gồm tệp kiểm thử.                                          |
| `--workers`                | int                 | `0` (`NumCPU`)             | `0` hoặc lớn hơn                              | Số lượng goroutine phân tích cú pháp đồng thời tối đa.                       |
| `--verbose`                | bool                | `false`                    | —                                           | In thông tin tiến trình.                                 |
| `--header-language`        | string              | `c`                        | `c`, `cpp`                                  | Bộ phân tích cú pháp được sử dụng cho tệp `.h`.                                  |
| `--include GLOB`           | string, có thể lặp lại  | Tất cả các nguồn được hỗ trợ      | —                                           | Bao gồm các đường dẫn nguồn phù hợp.                               |
| `--exclude GLOB`           | string, có thể lặp lại  | Không có                       | —                                           | Loại trừ các đường dẫn hoặc thư mục phù hợp.                       |
| `--max-file-size BYTES`    | int                 | `10485760`                 | `0` vô hiệu hóa giới hạn                      | Bỏ qua các tệp nguồn lớn hơn.                                    |
| `--strict`                 | bool                | `false`                    | —                                           | Thất bại trước khi ghi đầu ra nếu xảy ra chẩn đoán quét hoặc phân tích cú pháp.       |

Lệnh kết hợp tần suất commit, mức độ thay đổi dòng và tính mới nhất với PageRank cấp tệp, fan-in bắc cầu và trung tâm betweenness. Nó cũng báo cáo ghép nối thời gian giữa các tệp thay đổi cùng nhau. Các commit merge bị loại trừ theo mặc định.

```bash
codeknit graph hotspots ./src --since 2y --format json -o hotspots.json
```

## `codeknit fingerprint`

```bash
codeknit fingerprint <input-path>
```

| Đối số       | Bắt buộc | Mô tả                              |
| -------------- | -------- | ---------------------------------------- |
| `input-path`   | Có       | Thư mục nguồn hoặc tệp nguồn đơn.  |

| Cờ                    | Kiểu                | Mặc định                          | Giá trị cho phép          | Mô tả                                                                                                      |
| ----------------------- | ------------------- | -------------------------------- | ----------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output`        | string              | `./skeleton/fingerprints.skt`    | Đường dẫn tệp               | Tệp `.skt` đầu ra.                                                                                              |
| `--min-similarity`      | int                 | `65`                             | `0`–`100`               | Phần trăm độ tương đồng tối thiểu để báo cáo.                                                                         |
| `--max-similarity`      | int                 | `95`                             | `0`–`100`               | Phần trăm độ tương đồng tối đa để báo cáo.                                                                         |
| `--show-all`            | bool                | `false`                          | —                       | Cũng phát ra phần `[fingerprints]` với dữ liệu token thô.                                                      |
| `--rerank`              | bool                | `false`                          | —                       | Tìm các láng giềng ngữ nghĩa và sắp xếp lại các ứng viên với các embedding Ollama. Yêu cầu `ollama serve` và `ollama pull qwen3-embedding:0.6b`. |
| `--model`               | string              | `qwen3-embedding:0.6b`           | Tên mô hình Ollama       | Mô hình embedding Ollama để sử dụng với `--rerank`.                                                                   |
| `--collect-test`        | bool                | `false`                          | —                       | Bao gồm tệp kiểm thử.                                                                                              |
| `--workers`             | int                 | `0` (`NumCPU`)                   | `0` hoặc lớn hơn          | Số lượng goroutine phân tích cú pháp đồng thời tối đa.                                                                           |
| `--verbose`             | bool                | `false`                          | —                       | In thông tin tiến trình.                                                                                      |
| `--header-language`     | string              | `c`                              | `c`, `cpp`              | Bộ phân tích cú pháp được sử dụng cho tệp `.h`.                                                                                      |
| `--include GLOB`        | string, có thể lặp lại  | Tất cả các nguồn được hỗ trợ            | —                       | Bao gồm các đường dẫn nguồn phù hợp.                                                                                   |
| `--exclude GLOB`        | string, có thể lặp lại  | Không có                             | —                       | Loại trừ các đường dẫn hoặc thư mục phù hợp.                                                                           |
| `--max-file-size BYTES` | int                 | `10485760`                       | `0` vô hiệu hóa giới hạn  | Bỏ qua các tệp nguồn lớn hơn.                                                                                        |
| `--strict`              | bool                | `false`                          | —                       | Thất bại trước khi ghi đầu ra nếu xảy ra chẩn đoán quét hoặc phân tích cú pháp.                                                           |

### Phạm vi độ tương đồng

Phạm vi mặc định, 65–95%, loại trừ các kết quả trùng khớp 100%. Để báo cáo các kết quả trùng khớp token đã chuẩn hóa chính xác, hãy đặt cả hai giới hạn thành `100`:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Các ứng viên token đã chuẩn hóa chính xác là hoàn chỉnh. Với 500 ký hiệu đủ điều kiện trở lên, các ứng viên gần trùng lặp khác sử dụng truy xuất láng giềng top-50 gần đúng. `--rerank` có thể loại bỏ một ứng viên chính xác về mặt cấu trúc. Xem [Độ tin cậy phân tích](../analysis-reliability/#đảm-bảo-phát-hiện-trùng-lặp) để biết chi tiết.

## `codeknit completion`

```bash
codeknit completion <shell>
```

| Đối số  | Bắt buộc | Giá trị cho phép                         | Mô tả                                   |
| --------- | -------- | -------------------------------------- | --------------------------------------------- |
| `shell`   | Có       | `bash`, `zsh`, `fish`, `powershell`    | Shell để tạo script hoàn thành. |

Script được ghi vào stdout.

```bash
codeknit completion zsh >> ~/.zshrc
```

## Các trang liên quan

- [Tài liệu tham khảo định dạng đầu ra](../output-format/)
- [Hỗ trợ ngôn ngữ và độ trung thực API](../language-support/)
- [Độ tin cậy phân tích](../analysis-reliability/)