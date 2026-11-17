---
title: Lệnh Parse
description: Trích xuất thông tin cấu trúc từ mã nguồn vào các tệp .skt hoặc JSON.
---

`codeknit parse` đọc các tệp nguồn và ghi các ký hiệu của chúng (và tùy chọn, các mối quan hệ giữa chúng) dưới dạng tệp `.skt` hoặc JSON. Hướng dẫn này bao gồm các tình huống bạn có khả năng gặp phải nhiều nhất. Để xem danh sách đầy đủ các cờ được chia sẻ bởi tất cả các lệnh, hãy xem [tài liệu tham khảo cờ CLI](../../reference/cli-flags/). Để xem ngữ pháp đầu ra, hãy xem [tài liệu tham khảo định dạng đầu ra](../../reference/output-format/).

## Phân tích một dự án lần đầu tiên

```bash
codeknit parse ./src
```

codeknit ghi các tệp `.skt` đã chia nhỏ vào `./skeleton/`: `map_001.skt`, `map_002.skt`, v.v. Mở các phần nhỏ một cách chọn lọc thay vì tải tất cả cùng một lúc.

Để ghi vào một vị trí khác, hãy truyền thư mục đầu ra làm đối số thứ hai:

```bash
codeknit parse ./src ./output
```

Không sử dụng thư mục nguồn, hoặc một trong các thư mục cha của nó, làm thư mục đầu ra.

## Chọn chế độ đầu ra

Chọn chế độ bằng `--output-mode`.

| Chế độ            | Sử dụng khi                                    | Kết quả                                                                         |
| ----------------- | ---------------------------------------------- | ------------------------------------------------------------------------------ |
| `directory-flat`  | Phân tích hầu hết các dự án (mặc định)         | `map_001.skt`, `map_002.skt`, ...                                              |
| `directory-tree`  | Bạn muốn tra cứu cấu trúc của một tệp          | Một tệp `.skt` cho mỗi tệp nguồn, phản ánh cây thư mục nguồn                   |
| `inline`          | Phân tích một tệp đơn, hoặc chuyển tiếp đến công cụ khác | Đầu ra trên stdout; không sử dụng thư mục đầu ra                              |

Trong chế độ cây, tên tệp giữ nguyên phần mở rộng nguồn: `src/foo.c` trở thành `src/foo.c.skt`, và `src/foo.h` trở thành `src/foo.h.skt`. Các tệp lớn được chia thành `foo.c_part1.skt`, `foo.c_part2.skt`, v.v.

```bash
codeknit parse ./src --output-mode directory-tree
```

Tránh sử dụng `inline` cho toàn bộ dự án. Nó ghi tất cả vào stdout và có thể làm tràn terminal hoặc cửa sổ ngữ cảnh LLM của bạn.

Để biết thêm về cách chọn chế độ, hãy xem [Chế độ đầu ra](../output-modes/).

## Chạy lại trên cùng một dự án

codeknit từ chối ghi vào thư mục đã chứa đầu ra được tạo cho định dạng đã chọn. Điều này ngăn việc trộn lẫn các phần nhỏ cũ và mới. Thêm `--clean` để thay thế đầu ra trước đó:

```bash
codeknit parse ./src --clean
```

`--clean` không xóa bất cứ thứ gì ngay từ đầu. codeknit tạo đầu ra mới trong thư mục tạm thời và chỉ thay thế các tệp được tạo cũ sau khi quá trình tạo thành công. Nếu quá trình tạo thất bại, hoặc bị gián đoạn, đầu ra trước đó được giữ lại hoặc khôi phục trong lần chạy tiếp theo. Các tệp trong thư mục đầu ra mà codeknit không tạo ra được bảo toàn. Nếu một trong số chúng trùng với đường dẫn được tạo, quá trình chạy sẽ thất bại.

Trong chế độ JSON, các quy tắc tương tự áp dụng cho `codeknit.json`.

**Cạm bẫy: phần nhỏ trộn lẫn.** ID ký hiệu (`S1`, `S2`, ...) và kết quả mối quan hệ có thể thay đổi giữa các lần chạy và giữa các phiên bản. Sau khi nâng cấp lên 0.5.0, hãy tạo lại toàn bộ tập đầu ra với `--clean`. Không bao giờ kết hợp các phần nhỏ từ các lần chạy khác nhau.

## Phân tích một tệp đơn

Chỉ định `parse` vào tệp và sử dụng chế độ `inline`:

```bash
codeknit parse ./src/main.go --output-mode inline
```

Khi bạn chọn một tệp nguồn thông thường một cách rõ ràng, codeknit bỏ qua việc lọc tệp bỏ qua và tệp kiểm thử cho nó. Các mẫu bao gồm/loại trừ và giới hạn kích thước vẫn được áp dụng.

## Lấy JSON cho các script

```bash
codeknit parse ./src --output-mode inline --format json --edges
```

Tài liệu JSON có các trường cấp cao nhất là `files`, `symbols`, `edges` (bị bỏ qua khi không có), và `errors` (bị bỏ qua khi không có). Một ví dụ nhỏ:

```json
{
  "files": ["app.go"],
  "symbols": [
    {
      "id": "app.go::User",
      "short_id": "S1",
      "name": "User",
      "file": "app.go",
      "category": "type",
      "kind": "struct",
      "signature": "type User struct",
      "span": [3, 3]
    },
    {
      "id": "app.go::Save",
      "short_id": "S2",
      "name": "Save",
      "file": "app.go",
      "category": "callable",
      "kind": "function",
      "signature": "Save(u: S1)",
      "span": [5, 5]
    }
  ],
  "edges": [
    {
      "from": "app.go::Save",
      "from_short": "S2",
      "to": "app.go::User",
      "to_short": "S1",
      "kind": "references"
    }
  ]
}
```

Để ghi JSON vào đĩa thay vì stdout, hãy bỏ `--output-mode inline`. Trong các chế độ thư mục, JSON được ghi vào `codeknit.json`. JSON là một tài liệu duy nhất và không được chia theo số dòng.

## Bao gồm các mối quan hệ

Theo mặc định, đầu ra chỉ có phần `[symbols]`, giúp tiết kiệm token. Thêm `--edges` khi bạn cần biết cái gì gọi, chứa, kế thừa từ, hoặc tham chiếu đến cái gì:

```bash
codeknit parse ./src --edges
```

### Đọc các mối quan hệ không chắc chắn

Codeknit phân tích cú pháp mà không cần trình biên dịch, vì vậy một số mối quan hệ không thể được thiết lập. Mỗi cạnh cho biết mức độ chắc chắn của nó:

```skt
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

- `--kind-->` là một mối quan hệ đã được giải quyết.
- `--kind[unresolved]-->` có nghĩa là codeknit không thể thiết lập mục tiêu.
- `--kind[ambiguous]-->` có nghĩa là có nhiều hơn một mục tiêu có thể. ID ứng viên được liệt kê khi có sẵn.

Ứng viên là các mục tiêu có thể, không phải là các phụ thuộc đã được xác nhận. Các điểm cuối đã biết sử dụng ShortID. Một mục tiêu chưa được giải quyết không có ký hiệu phù hợp có thể xuất hiện dưới dạng tên, tương đối với tệp nguồn.

Trong JSON, các cạnh không chắc chắn mang `resolution` (`"unresolved"` hoặc `"ambiguous"`) và, khi có sẵn, `candidates` dưới dạng ID ký hiệu đầy đủ. Chúng bỏ qua `to_short`.

Đối với C và C++, các chỉ thị `#include` trở thành các ký hiệu `meta/file` và `meta/include`, vì vậy các cạnh của chúng có ShortID. Điều này ghi lại chỉ thị như đã viết. Nó không có nghĩa là codeknit đã tìm thấy tệp tiêu đề.

Các số liệu đồ thị loại trừ các cạnh không chắc chắn và báo cáo số lượng đã bị loại trừ. Trình xem đồ thị HTML hiển thị số lượng các mối quan hệ không chắc chắn và không vẽ chúng. Để kiểm tra các cạnh không chắc chắn riêng lẻ, hãy sử dụng `--edges` trên `parse`. Xem [Lệnh đồ thị](../graph-commands/) để biết các báo cáo đó.

Ngữ pháp cạnh nằm trong [tài liệu tham khảo định dạng đầu ra](../../reference/output-format/). Các quy tắc và giới hạn giải quyết nằm trong [tài liệu tham khảo hỗ trợ ngôn ngữ](../../reference/language-support/#relationship-resolution).

## Giảm kích thước đầu ra

Để nén các token lặp lại, hãy thêm `--minify`:

```bash
codeknit parse ./src --minify
```

Thao tác này ghi một từ điển vào `dict.skt` (hoặc một phần `[dict]` inline). Thay thế các mã như `d0` và `d1` bằng các giá trị từ điển khi đọc. Các loại mối quan hệ cũng được mã hóa, nhưng các nhãn giải quyết như `[unresolved]` vẫn có thể đọc được.

Để kiểm soát kích thước phần nhỏ, hãy sử dụng `--max-lines` (mặc định `500`):

- Trong chế độ `directory-flat`, đây là mục tiêu. Một khối nguồn không bao giờ bị chia nhỏ, vì vậy một phần nhỏ có thể vượt quá nó bởi một khối.
- Trong chế độ `directory-tree`, đây là giới hạn cứng cho mỗi phần nhỏ nguồn. Giá trị tối thiểu là `3`. Mỗi phần nhỏ lặp lại các tiêu đề `[symbols]` và `## source-path` của nó.

## Phân tích các tệp tiêu đề C++ có tên `.h`

Codeknit phân tích các tệp `.h` dưới dạng C theo mặc định. Đối với các tệp tiêu đề C++ có phần mở rộng đó, hãy đặt `--header-language cpp`:

```bash
codeknit parse ./src --header-language cpp
```

Cờ này chỉ ảnh hưởng đến các tệp `.h`. Các tệp có tên `.hpp` và `.hxx` đã được phân tích dưới dạng C++, và các phần mở rộng khác giữ nguyên các trình phân tích thông thường của chúng.

## Kiểm soát các tệp được quét

Tất cả năm lệnh xử lý nguồn chấp nhận các cờ này. Chúng được liệt kê đầy đủ trong [tài liệu tham khảo cờ CLI](../../reference/cli-flags/).

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
```

- `--include GLOB` và `--exclude GLOB` có thể lặp lại, là các mẫu kiểu gitignore tương đối với thư mục đầu vào. Trích dẫn chúng để shell của bạn không mở rộng chúng. Các mẫu loại trừ thắng. Các mẫu bao gồm không ghi đè `.gitignore`, việc lọc tệp kiểm thử, hoặc các loại trừ tích hợp sẵn.
- `--max-file-size BYTES` bỏ qua các tệp lớn hơn. Mặc định là `10485760` (10 MiB). `0` vô hiệu hóa giới hạn.
- `--strict` thất bại trước khi ghi bất kỳ đầu ra nào nếu quá trình quét hoặc phân tích không hoàn chỉnh. Mặc định là `false`.

Codeknit bỏ qua và báo cáo các liên kết tượng trưng và các tệp không phải là tệp thông thường. Các quy tắc `.gitignore` của thư mục cha áp dụng độc lập với thư mục làm việc của bạn.

**Cạm bẫy: kết quả không đầy đủ.** Các đường dẫn không thể đọc được, các tệp quá lớn, các liên kết tượng trưng bị bỏ qua, và các cảnh báo cú pháp làm cho một phân tích không hoàn chỉnh. Nếu không có `--strict`, codeknit giữ lại các kết quả có thể sử dụng và báo cáo các vấn đề:

- trên stderr,
- trong trường `errors` của JSON, hoặc
- trong phần `[errors]` của SKT (inline, hoặc trong `warnings.skt` trong các chế độ thư mục).

Một đầu vào không có nguồn được hỗ trợ đã chọn, hoặc một lần chạy trong đó mọi nguồn đọc đều thất bại, luôn thoát không thành công. Để biết các đảm bảo chính xác, hãy xem [Độ tin cậy phân tích](../../reference/analysis-reliability/).

## Cờ phân tích

Các cờ này đặc thù cho `parse`. Các cờ chung được liệt kê trong [tài liệu tham khảo cờ CLI](../../reference/cli-flags/).

| Cờ                | Mặc định          | Mô tả                                                                                       |
| ----------------- | ---------------- | ------------------------------------------------------------------------------------------- |
| `--output-mode`   | `directory-flat` | Chế độ đầu ra: `inline`, `directory-flat`, hoặc `directory-tree`.                           |
| `--format`        | `skt`            | Định dạng đầu ra: `skt` hoặc `json`.                                                        |
| `--header-language`| `c`             | Trình phân tích cho các tệp `.h`: `c` hoặc `cpp`. Các phần mở rộng khác không thay đổi.      |
| `--max-lines`     | `500`            | Mục tiêu dòng cho đầu ra phẳng; giới hạn cứng (tối thiểu 3) cho mỗi phần nhỏ nguồn cây.      |
| `--collect-test`  | `false`          | Bao gồm các tệp kiểm thử.                                                                   |
| `--minify`        | `false`          | Sử dụng nén dựa trên từ điển.                                                               |
| `--edges`         | `false`          | Bao gồm phần `[edges]` (tắt theo mặc định để tiết kiệm token).                              |
| `--clean`         | `false`          | Thay thế đầu ra được tạo trước đó, chỉ sau khi quá trình tạo thành công.                     |
| `--workers`       | `NumCPU`         | Số lượng goroutine phân tích đồng thời tối đa.                                              |
| `--verbose`       | `false`          | In thông tin tiến trình.                                                                     |

## Mẹo

- Sử dụng `directory-flat` cho hầu hết các dự án, và chỉ đọc các phần nhỏ bạn cần.
- Sử dụng `--minify` trên các cơ sở mã lớn để tiết kiệm token.
- Chỉ truyền `--edges` khi bạn cần dữ liệu mối quan hệ.
- Kiểm tra [ma trận khả năng đã thử nghiệm và các giới hạn đã biết](../../reference/language-support/) trước khi dựa vào việc trích xuất cho một ngôn ngữ cụ thể.

## Xem thêm

- [Chế độ đầu ra](../output-modes/)
- [Tài liệu tham khảo định dạng đầu ra](../../reference/output-format/)
- [Độ tin cậy phân tích](../../reference/analysis-reliability/)
- [Lệnh đồ thị](../graph-commands/)