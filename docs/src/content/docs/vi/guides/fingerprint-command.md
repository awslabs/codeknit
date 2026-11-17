---
title: Lệnh Fingerprint
description: Phát hiện mã trùng lặp và gần trùng lặp trên các tệp và ngôn ngữ bằng fuzzy hashing.
---

`codeknit fingerprint` tính toán một Context-Triggered Piecewise Hash (CTPH) trên các token phần thân đã được chuẩn hóa của mọi hàm, phương thức, biến và kiểu. Tên biến, chuỗi ký tự và chú thích kiểu được chuẩn hóa trước khi băm. Do đó, các bản sao-dán đã đổi tên và các bản chuyển đổi sang ngôn ngữ khác vẫn có thể khớp.

Hướng dẫn này chỉ ra cách tìm các bản trùng lặp, điều chỉnh phạm vi độ tương đồng, giảm dương tính giả và quyết định những gì cần tái cấu trúc.

## Trước khi bắt đầu

- `codeknit` đã được cài đặt. Xem [Cài đặt](../../installation/).
- Bạn có một thư mục nguồn để quét, chẳng hạn như `./src`.

## Tìm các bản trùng lặp trong dự án

Chạy lệnh trên thư mục bạn muốn kiểm tra:

```bash
codeknit fingerprint ./src
```

Lệnh ghi `./skeleton/fingerprints.skt` và báo cáo các cặp có độ tương đồng từ 65% đến 95%.

Mở tệp. Phần `[duplicates]` luôn hiện diện, với các kết quả khớp mạnh nhất được liệt kê đầu tiên:

```skt
[duplicates]
# similarity range: 65%-95%
similarity:93%  pkg/user.go::GetUser <-> pkg/admin.go::GetAdmin
similarity:88%  src/auth.ts::validateToken <-> src/session.ts::validateSession
similarity:81%  lib/math.py::compute_tax <-> lib/math.py::compute_vat
```

Mỗi dòng có dạng `similarity:<N>%  <file>::<symbol> <-> <file>::<symbol>`. Nếu không có gì nằm trong phạm vi, phần này chỉ chứa `# no duplicates found`.

Để biết cú pháp đầu ra đầy đủ, xem [tài liệu tham khảo định dạng đầu ra](../../reference/output-format/).

## Tìm các bản trùng lặp chính xác

Phạm vi mặc định giới hạn ở 95%, vì vậy các kết quả khớp 100% bị loại trừ. Để liệt kê các kết quả khớp chính xác, đặt cả hai giới hạn là 100:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

"Chính xác" có nghĩa là các luồng token đã chuẩn hóa hoàn chỉnh giống nhau trong cùng một danh mục ký hiệu. Điều này không có nghĩa là văn bản nguồn giống nhau, hoặc hai ký hiệu được chứng minh là tương đương.

## Mở rộng phạm vi để tìm các kết quả khớp lỏng lẻo hơn

Giảm giới hạn dưới để bao gồm mã có cùng hình dạng thuật toán nhưng khác biệt về chi tiết:

```bash
codeknit fingerprint ./src --min-similarity 50 --max-similarity 80
```

Các kết quả dưới 50% thường là nhiễu.

## Giảm dương tính giả bằng cách xếp hạng lại ngữ nghĩa

Nếu kết quả cấu trúc có nhiều nhiễu, xếp hạng lại các ứng viên bằng cách sử dụng embeddings từ máy chủ Ollama cục bộ.

1. Khởi động Ollama và tải mô hình embedding mặc định:

   ```bash
   ollama serve
   ollama pull qwen3-embedding:0.6b
   ```

2. Chạy lệnh với `--rerank`:

   ```bash
   codeknit fingerprint ./src --rerank
   ```

Để sử dụng mô hình embedding khác, truyền `--model`:

```bash
codeknit fingerprint ./src --rerank --model qwen3-embedding:4b
```

Việc xếp hạng lại bổ sung các ứng viên ngữ nghĩa và lọc dựa trên điểm số kết hợp. Nó giảm dương tính giả nhưng không loại bỏ tất cả. Nó cũng có thể loại bỏ một ứng viên có cấu trúc chính xác, và phạm vi độ tương đồng vẫn được áp dụng. Embeddings được tính toán từ cùng một ảnh chụp nguồn đã phân tích được sử dụng cho việc trích xuất.

## Kiểm tra các fingerprint thô

Để xem fingerprint của mọi ký hiệu, thêm `--show-all`:

```bash
codeknit fingerprint ./src --show-all
```

Thao tác này bổ sung phần `[fingerprints]`, được nhóm theo tệp nguồn, một dòng cho mỗi ký hiệu:

```skt
[fingerprints]
## src/auth.ts
validateToken  FP:3:a1b2c3...:d4e5f6...  tokens:8e0f1a2b...
```

Định dạng dòng là `<symbol-name>  FP:<blocksize>:<hash1>:<hash2>  tokens:<hex>`. `tokens` là luồng token đã chuẩn hóa được mã hóa hex.

## Ghi báo cáo vào nơi khác

```bash
codeknit fingerprint ./src -o duplicates.skt
```

## Kiểm soát các tệp được quét

Các tệp kiểm thử được bỏ qua theo mặc định. Bao gồm chúng với `--collect-test`:

```bash
codeknit fingerprint ./src --collect-test
```

Để thu hẹp phạm vi quét, sử dụng các cờ lựa chọn nguồn. Trích dẫn các mẫu để shell không mở rộng chúng:

```bash
codeknit fingerprint ./src --include '**/*.ts' --exclude '**/generated/'
```

Các điều khiển lựa chọn khác:

- `--max-file-size BYTES` bỏ qua các tệp lớn hơn. Mặc định là `10485760` (10 MiB). `0` vô hiệu hóa giới hạn.
- `--strict` thất bại trước khi ghi đầu ra nếu quá trình quét hoặc phân tích không hoàn chỉnh.
- `--header-language c|cpp` chọn trình phân tích cú pháp cho các tệp `.h` mơ hồ. Xem [Hỗ trợ ngôn ngữ](../../reference/language-support/#header-language-selection).

Các liên kết tượng trưng bị bỏ qua, các tệp quá kích thước và cảnh báo cú pháp xuất hiện trên stderr. Bao gồm không ghi đè `.gitignore` hoặc lọc kiểm thử, và loại trừ được ưu tiên hơn bao gồm. Xem [tài liệu tham khảo độ tin cậy phân tích](../../reference/analysis-reliability/) để biết các đảm bảo chính xác.

## Chọn phạm vi độ tương đồng

| Phạm vi     | Những gì thường xuất hiện                                      |
| ----------- | -------------------------------------------------------------- |
| `96-100%`   | Các bản trùng lặp cấu trúc chính xác hoặc gần chính xác; sao chép rõ ràng |
| `85-95%`    | Các bản gần trùng lặp; thường là sao chép với chỉnh sửa nhỏ    |
| `65-84%`    | Phạm vi mặc định; độ tương đồng cấu trúc mạnh, thường là ứng viên tái cấu trúc |
| `50-64%`    | Cùng hình dạng thuật toán với chi tiết khác nhau               |
| Dưới 50%    | Thường là nhiễu                                                |

Để xem hai hàng đầu tiên, tăng `--max-similarity` lên `100`.

## Quyết định những gì cần tái cấu trúc

1. Bắt đầu với phạm vi mặc định và đọc các cặp đầu tiên.
2. Mở cả hai ký hiệu của mỗi cặp và so sánh chúng.
3. Xem một kết quả khớp là tín hiệu, không phải bằng chứng. Fingerprints đo lường hình dạng cấu trúc, không phải tên định danh hoặc ý định. Cùng một logic được viết theo cách lặp và đệ quy sẽ cho điểm thấp, và mã không liên quan có cùng hình dạng có thể cho điểm cao.
4. Các cặp trong phạm vi 85–95% thường là ứng viên an toàn nhất để hợp nhất. Các cặp trong phạm vi 65–84% thường cần xem xét kỹ hơn.
5. Nếu danh sách có nhiều nhiễu, thu hẹp phạm vi hoặc thêm `--rerank`.

Các kết quả khớp đa ngôn ngữ hoạt động cho các cấu trúc tương đương. Các tính năng đặc thù ngôn ngữ có thể tạo ra các kết quả khớp độ tương đồng thấp giả mạo.

Phần trăm độ tương đồng là một xếp hạng heuristic, không phải khoảng cách chỉnh sửa. Bộ hiệu chuẩn đã gắn nhãn của dự án báo cáo độ chính xác và khả năng thu hồi riêng biệt với điểm số độ tương đồng. Đây là một kiểm tra hồi quy, không phải ước lượng độ chính xác trên mã của bạn.

## Dự án lớn

- Dưới 500 ký hiệu đủ điều kiện, codeknit chấm điểm mọi cặp cùng danh mục.
- Từ 500 trở lên, các nhóm token đã chuẩn hóa chính xác hoàn chỉnh luôn được giữ lại. Các cặp khác sử dụng khoảng 50 lân cận hàng đầu cho mỗi ký hiệu, vì vậy kết quả gần trùng lặp có thể không đầy đủ.
- Đối với các bản trùng lặp chính xác, sử dụng lệnh 100/100 ở trên. Nó không bị ảnh hưởng bởi phép xấp xỉ.
- Thứ tự là xác định cho các đầu vào và điểm số ổn định.

## Khắc phục sự cố

| Triệu chứng | Nguyên nhân và cách khắc phục |
| ----------- | ----------------------------- |
| `# no duplicates found` | Không có cặp nào nằm trong phạm vi. Giảm `--min-similarity` hoặc tăng `--max-similarity`. |
| Một hàm ngắn đã biết bị thiếu | Các ký hiệu có ít hơn 4 token đã chuẩn hóa bị bỏ qua. |
| Các bản sao chính xác bị thiếu | Giới hạn tối đa mặc định là 95%. Sử dụng `--min-similarity 100 --max-similarity 100`. Với `--rerank`, bộ lọc ngữ nghĩa cũng có thể loại bỏ chúng. |
| `--rerank` thất bại | Kiểm tra rằng `ollama serve` đang chạy và mô hình đã được tải. |
| Kết quả một phần với cảnh báo trên stderr | Sửa các tệp được báo cáo, hoặc thêm `--strict` để thất bại thay vì xuất bản đầu ra không hoàn chỉnh. |

## Các cờ được sử dụng trong hướng dẫn này

| Cờ                | Mặc định                       | Mô tả                                                                 |
| ----------------- | ------------------------------ | --------------------------------------------------------------------- |
| `-o`, `--output`  | `./skeleton/fingerprints.skt`  | Đường dẫn tệp `.skt` đầu ra.                                         |
| `--min-similarity` | `65`                          | Phần trăm độ tương đồng tối thiểu để báo cáo (0–100).                 |
| `--max-similarity` | `95`                          | Phần trăm độ tương đồng tối đa để báo cáo (0–100).                    |
| `--show-all`      | `false`                        | Cũng xuất phần `[fingerprints]` với dữ liệu token thô.                |
| `--rerank`        | `false`                        | Xếp hạng lại các ứng viên với embeddings ngữ nghĩa qua Ollama. Yêu cầu `ollama serve` và `ollama pull qwen3-embedding:0.6b`. |
| `--model`         | `qwen3-embedding:0.6b`         | Mô hình embedding Ollama để sử dụng với `--rerank`.                   |
| `--collect-test`  | `false`                        | Bao gồm các tệp kiểm thử trong phân tích.                             |
| `--workers`       | `NumCPU`                       | Số lượng goroutine phân tích cú pháp đồng thời tối đa.                |
| `--verbose`       | `false`                        | In thông tin tiến trình.                                              |

Để biết danh sách đầy đủ, bao gồm `--include`, `--exclude`, `--max-file-size`, `--strict`, và `--header-language`, xem [tài liệu tham khảo cờ CLI](../../reference/cli-flags/).

## Liên quan

- [Lệnh Parse](../parse-command/)
- [Lệnh Đồ thị](../graph-commands/)