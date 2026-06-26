---
title: Tài liệu tham khảo CLI
description: Tài liệu tham khảo đầy đủ cho tất cả các lệnh và cờ của codeknit.
---

## codeknit

Khởi chạy giao diện dòng lệnh tương tác (TUI), hướng dẫn bạn qua các lệnh và tùy chọn có sẵn.

```bash
codeknit
```

## codeknit parse

Trích xuất thông tin cấu trúc từ mã nguồn vào các tệp `.skt` hoặc JSON.

```bash
codeknit parse <input-path> [output-dir]
```

| Cờ               | Kiểu   | Mặc định          | Mô tả                                                                                     |
| ---------------- | ------ | ---------------- | ----------------------------------------------------------------------------------------- |
| `--output-mode`  | string | `directory-flat` | Chế độ đầu ra: `inline`, `directory-flat`, hoặc `directory-tree`                          |
| `--format`       | string | `skt`            | Định dạng đầu ra: `skt` hoặc `json`                                                      |
| `--max-lines`    | int    | `500`            | Số dòng tối đa cho mỗi tệp đầu ra (áp dụng cho chế độ `directory-flat` và `directory-tree`) |
| `--collect-test` | bool   | `false`          | Bao gồm các tệp kiểm thử trong phân tích                                                  |
| `--minify`       | bool   | `false`          | Kích hoạt nén đầu ra dựa trên từ điển                                                     |
| `--edges`        | bool   | `false`          | Bao gồm phần `[edges]` trong đầu ra (tắt theo mặc định để tiết kiệm token)                 |
| `--clean`        | bool   | `false`          | Xóa các tệp `.skt` cũ trong thư mục đầu ra trước khi ghi                                  |
| `--workers`      | int    | `0` (NumCPU)     | Số lượng goroutine phân tích đồng thời tối đa                                             |
| `--verbose`      | bool   | `false`          | Hiển thị thông tin tiến trình trong quá trình xử lý                                        |

Thư mục đầu ra mặc định là `./skeleton` khi không được chỉ định. Trong chế độ `inline`, đầu ra được ghi ra stdout và không sử dụng thư mục nào. Với `--format json`, đầu ra thư mục được ghi dưới dạng `codeknit.json`.

## codeknit graph show

Tạo biểu đồ HTML tương tác để trực quan hóa cấu trúc mã của codebase.

```bash
codeknit graph show <input-path>
```

| Cờ               | Kiểu   | Mặc định                          | Mô tả                                  |
| ---------------- | ------ | -------------------------------- | -------------------------------------- |
| `-o`, `--output` | string | `./skeleton/codeknit-graph.html` | Đường dẫn tệp HTML đầu ra              |
| `--collect-test` | bool   | `false`                          | Bao gồm các tệp kiểm thử trong phân tích |
| `--workers`      | int    | `0` (NumCPU)                     | Số lượng goroutine phân tích đồng thời tối đa |
| `--verbose`      | bool   | `false`                          | Hiển thị thông tin tiến trình trong quá trình xử lý |

Tệp HTML được tạo ra là tự chứa và tự động mở trong trình duyệt mặc định của bạn.

## codeknit graph analyze

Chạy các thuật toán phân tích cấu trúc và xuất báo cáo `.skt` có thể đọc được bởi LLM.

```bash
codeknit graph analyze <input-path>
```

| Cờ                      | Kiểu    | Mặc định                         | Mô tả                                                   |
| ----------------------- | ------- | ------------------------------- | ------------------------------------------------------- |
| `-o`, `--output`        | string  | `./skeleton/graph_analysis.skt` | Đường dẫn tệp `.skt` đầu ra                            |
| `--collect-test`        | bool    | `false`                         | Bao gồm các tệp kiểm thử trong phân tích                |
| `--workers`             | int     | `0` (NumCPU)                    | Số lượng goroutine phân tích đồng thời tối đa           |
| `--verbose`             | bool    | `false`                         | Hiển thị thông tin tiến trình trong quá trình xử lý      |
| `--fan-threshold`       | int     | `10`                            | Ngưỡng fan-in hoặc fan-out tối thiểu để gắn cờ ký hiệu trung tâm |
| `--god-threshold`       | int     | `15`                            | Số lượng cạnh chứa tối thiểu để gắn cờ cho god class/function |
| `--max-inheritance-depth` | int   | `5`                             | Gắn cờ cho các chuỗi kế thừa sâu hơn giá trị này         |
| `--top-n`               | int     | `30`                            | Giới hạn các phần đầu ra được xếp hạng; `0` nghĩa là không giới hạn |
| `--betweenness-threshold` | float64 | `0.001`                       | Giá trị trung tâm betweenness tối thiểu để báo cáo      |
| `--propagation-cutoff`  | float64 | `0.05`                          | Xác suất tối thiểu để tiếp tục mô phỏng lan truyền thay đổi |

## codeknit fingerprint

Phát hiện mã trùng lặp và gần trùng lặp bằng cách sử dụng fuzzy hashing.

```bash
codeknit fingerprint <input-path>
```

| Cờ               | Kiểu   | Mặc định                       | Mô tả                                                                                                                  |
| ---------------- | ------ | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `-o`, `--output` | string | `./skeleton/fingerprints.skt` | Đường dẫn tệp `.skt` đầu ra                                                                                           |
| `--min-similarity` | int  | `65`                          | Độ tương đồng tối thiểu để báo cáo (0–100)                                                                             |
| `--max-similarity` | int  | `95`                          | Độ tương đồng tối đa để báo cáo (0–100)                                                                               |
| `--show-all`     | bool   | `false`                       | Bao gồm phần `[fingerprints]` với dữ liệu token thô                                                                    |
| `--rerank`       | bool   | `false`                       | Sắp xếp lại các ứng viên CTPH bằng cách sử dụng embeddings ngữ nghĩa qua Ollama (yêu cầu `ollama serve` và `ollama pull qwen3-embedding:0.6b`) |
| `--model`        | string | `qwen3-embedding:0.6b`        | Mô hình embedding Ollama để sử dụng với `--rerank`                                                                     |
| `--collect-test` | bool   | `false`                       | Bao gồm các tệp kiểm thử trong phân tích                                                                              |
| `--workers`      | int    | `0` (NumCPU)                  | Số lượng goroutine phân tích đồng thời tối đa                                                                         |
| `--verbose`      | bool   | `false`                       | Hiển thị thông tin tiến trình trong quá trình xử lý                                                                    |

## codeknit completion

Tạo các tập lệnh hoàn thành shell cho các shell được hỗ trợ.

```bash
codeknit completion <shell>
```

Các shell được hỗ trợ: `bash`, `zsh`, `fish`, `powershell`.

## Cờ toàn cục

| Cờ              | Mô tả                       |
| --------------- | --------------------------- |
| `--version`     | Hiển thị thông tin phiên bản |
| `--help`, `-h`  | Hiển thị trợ giúp cho lệnh hiện tại |