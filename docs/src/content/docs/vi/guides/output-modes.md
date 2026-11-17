---
title: Chế độ đầu ra
description: Chọn chế độ đầu ra phù hợp với kích thước dự án và quy trình làm việc của bạn.
---

`codeknit` tách biệt hai quyết định: nơi đầu ra được ghi (chế độ **đầu ra**) và cách nó được mã hóa (**định dạng đầu ra**). Trang này giải thích lý do tồn tại của từng tùy chọn, chi phí và thời điểm phù hợp. Đối với ngữ pháp `.skt`, hãy xem [tài liệu tham chiếu định dạng đầu ra](../../reference/output-format/). Đối với các lệnh, hãy xem [hướng dẫn lệnh phân tích](../parse-command/).

## Hai lựa chọn độc lập

Cờ `--output-mode` chọn một trong ba chế độ: `directory-flat`, `directory-tree` hoặc `inline`. Cờ `--format` chọn `skt` (mặc định) hoặc `json`. Bất kỳ chế độ nào cũng có thể kết hợp với bất kỳ định dạng nào.

Hai lựa chọn này tách biệt vì chúng trả lời các câu hỏi khác nhau. Chế độ quyết định cách đầu ra được phân chia và định vị trên đĩa, điều này quan trọng đối với người hoặc LLM đọc nó. Định dạng quyết định ai sẽ tiêu thụ nó: LLM hoặc người đọc, hay một tập lệnh.

## directory-flat

`directory-flat` là chế độ mặc định. Nó ghi các tệp chunk được đánh số, `map_001.skt`, `map_002.skt`, v.v. vào `./skeleton/` trừ khi bạn truyền một thư mục đầu ra khác.

**Lý do tồn tại.** Một cơ sở mã lớn tạo ra nhiều cấu trúc mã hơn mức vừa với một cửa sổ ngữ cảnh. Chunking cho phép người đọc chỉ mở các tệp cần thiết thay vì tải tất cả.

**Cách phân chia hoạt động.** `codeknit` giữ nguyên khối của mỗi tệp nguồn và bắt đầu một chunk mới giữa các khối khi một chunk đạt khoảng `--max-lines` (mặc định là 500). Giới hạn này là mục tiêu, không phải ràng buộc cứng: một tệp nguồn có khối rất lớn có thể tạo ra một chunk dài hơn 500 dòng. Giữ nguyên các khối có nghĩa là một ký hiệu không bao giờ bị tách khỏi các ký hiệu anh em trong cùng một tệp.

**Đánh đổi.** Các chunk không được tổ chức theo đường dẫn nguồn. Để tìm cấu trúc của một tệp, bạn tìm kiếm các chunk cho tiêu đề `## source-path` của nó. Nếu bạn thường xuyên tra cứu các tệp cụ thể, `directory-tree` sẽ phù hợp hơn.

**Tối giản hóa.** Với `--minify`, từ điển được ghi vào `dict.skt` bên cạnh các chunk. Xem [Tối giản hóa](#tối-giản-hóa).

## directory-tree

`directory-tree` phản chiếu cây thư mục nguồn của bạn dưới thư mục đầu ra. Mỗi tệp nguồn nhận một tệp `.skt` tại đường dẫn tương ứng.

**Đặt tên.** `codeknit` thêm `.skt` vào tên tệp nguồn hoàn chỉnh thay vì thay thế phần mở rộng. `src/foo.c` trở thành `src/foo.c.skt`, và `src/foo.h` trở thành `src/foo.h.skt`. Nếu không có quy tắc này, cả hai sẽ ánh xạ tới `foo.skt` và xung đột. Khi đầu ra của một tệp bị phân chia, các phần được đặt tên là `foo.c_part1.skt`, `foo.c_part2.skt`, v.v. Các đường dẫn đầu ra xung đột, bao gồm cả xung đột chỉ khác chữ hoa chữ thường, sẽ thất bại rõ ràng thay vì ghi đè lên nhau.

**Phân chia khác với chế độ flat.** Trong chế độ tree, `--max-lines` là giới hạn cứng đối với mỗi chunk nguồn, với tối thiểu là 3. Mỗi chunk ký hiệu lặp lại các tiêu đề `[symbols]` và `## source-path` để có thể đọc riêng lẻ. Các chunk chỉ chứa cạnh mang `[edges]`. Giới hạn áp dụng cho các chunk nguồn, không áp dụng cho các tệp phụ trợ như `dict.skt` và `warnings.skt`.

**Đánh đổi.** Đầu ra tree dễ dàng điều hướng cùng với nguồn, nhưng nó tạo ra nhiều tệp nhỏ. Một dự án lớn có thể tạo ra hàng nghìn tệp, và người đọc muốn có cái nhìn tổng quan phải mở nhiều tệp.

## inline

`inline` ghi tất cả vào stdout. Nó không tạo thư mục đầu ra hay khu vực staging.

**Lý do tồn tại.** Nó phù hợp với hai trường hợp: kiểm tra nhanh một tệp đơn lẻ và truyền đầu ra vào một công cụ khác. Đây là chế độ duy nhất không ghi gì vào đĩa.

**Đánh đổi.** Không có gì bị phân chia. Toàn bộ kết quả xuất hiện trên terminal hoặc trong ngữ cảnh của người tiêu thụ cùng một lúc, vì vậy nó có thể làm quá tải người đọc với bất kỳ thứ gì ngoài một vài tệp. Hãy coi nó phù hợp với một tệp đơn lẻ hoặc một dự án rất nhỏ, khoảng dưới năm tệp. Vì stdout không thể khôi phục, một người tiêu thụ thất bại có thể thấy đầu ra một phần. Các chế độ thư mục không gặp vấn đề này.

## Định dạng đầu ra

| Định dạng | Phù hợp nhất cho | Đầu ra |
| --- | --- | --- |
| `skt` | Ngữ cảnh LLM và kiểm tra bởi con người | Tệp `.skt` hoặc stdout |
| `json` | Tập lệnh và tích hợp có cấu trúc | `codeknit.json` trong chế độ thư mục, hoặc stdout trong chế độ inline |

**skt** gọn nhẹ và định hướng theo dòng. Mỗi ký hiệu nằm trên một dòng vật lý, với các ngắt dòng trong chữ ký được thoát, vì vậy người đọc có thể tải một phần đầu ra mà không làm hỏng cấu trúc của nó. Đó là lý do chunking hoạt động với `.skt`.

**json** tồn tại cho những người tiêu thụ cần các trường có kiểu thay vì văn bản để đọc. Nó là một tài liệu duy nhất với `files`, `symbols` và, khi bao gồm các cạnh, `edges`. Vì nó là một tài liệu duy nhất, nó không bị phân chia theo số dòng. Trong các chế độ thư mục `--max-lines` do đó không có tác dụng đối với JSON, và một dự án lớn tạo ra một `codeknit.json` lớn.

Một mẫu gọn nhẹ:

```json
{
  "files": ["app.go"],
  "symbols": [
    {"id": "app.go::User", "short_id": "S1", "name": "User", "file": "app.go",
     "category": "type", "kind": "struct", "signature": "type User struct", "span": [3, 3]},
    {"id": "app.go::Save", "short_id": "S2", "name": "Save", "file": "app.go",
     "category": "callable", "kind": "function", "signature": "Save(u: S1)", "span": [5, 5]}
  ],
  "edges": [
    {"from": "app.go::Save", "from_short": "S2", "to": "app.go::User",
     "to_short": "S1", "kind": "references"}
  ]
}
```

## Chọn chế độ

| Chế độ | Phù hợp nhất cho | Đầu ra |
| --- | --- | --- |
| `directory-flat` | Hầu hết các dự án (mặc định) | `./skeleton/map_001.skt`, `map_002.skt`, ... |
| `directory-tree` | Điều hướng cùng với nguồn | `./skeleton/<đường dẫn phản chiếu>.skt` |
| `inline` | Tệp đơn, truyền qua pipe | stdout |

Quy tắc chung:

- **Khi không chắc chắn**, hãy sử dụng `directory-flat`, chế độ mặc định.
- **Một tệp đơn**: `inline` là đủ.
- **Bất kỳ thứ gì ngoài một vài tệp**: sử dụng `directory-flat` hoặc `directory-tree`.
- **Cơ sở mã lớn**: thêm `--minify`.
- **Chạy lại vào cùng thư mục**: thêm `--clean`.

## Staging, chạy lại và `--clean`

Các chế độ thư mục không ghi vào thư mục đầu ra ngay lập tức. `codeknit` tạo đầu ra trong thư mục staging anh em và chỉ xuất bản nó sau khi tạo thành công. Một lần chạy thất bại do đó để lại đầu ra trước đó của bạn tại chỗ.

Theo mặc định, `codeknit` từ chối ghi khi đầu ra đã tạo cho định dạng đã chọn đã tồn tại. Lý do là các ID ký hiệu (`S1`, `S2`, ...) có thể thay đổi bất cứ khi nào trích xuất thay đổi. Trộn các chunk từ các lần chạy khác nhau sẽ tạo ra các cạnh trỏ đến các ký hiệu sai. `--clean` cho phép thay thế: sau khi tạo thành công, các tệp `.skt` lỗi thời (trong chế độ SKT) hoặc `codeknit.json` (trong chế độ JSON) được thay thế. Các mục không liên quan trong thư mục được giữ nguyên. Nếu một trong số chúng xung đột với đường dẫn đã tạo, lần chạy sẽ thất bại thay vì ghi đè lên nó.

Việc xuất bản sử dụng khóa cụ thể cho thư mục, sao lưu đầu ra trước đó và khôi phục nó nếu thay thế thất bại. Nếu quá trình bị gián đoạn giữa hai lần đổi tên thư mục, lần chạy tiếp theo sẽ khôi phục bản sao lưu trước khi tạo bất kỳ thứ gì. Điều này khắc phục sự cố khi quá trình thất bại. Nó không làm cho việc cập nhật thư mục trở nên tức thời: có một khoảng thời gian ngắn giữa các lần đổi tên, vì vậy người đọc cần truy cập không bị gián đoạn nên phối hợp với việc tạo. Nó cũng không bảo vệ chống lại sự cố lưu trữ hoặc mất điện.

Không chọn thư mục nguồn, hoặc thư mục tổ tiên của nó, làm thư mục đầu ra. Đối với các đảm bảo xuất bản đầy đủ, hãy xem [tài liệu tham chiếu độ tin cậy phân tích](../../reference/analysis-reliability/).

## Tối giản hóa

`--minify` áp dụng nén dựa trên từ điển. `codeknit` thay thế các token lặp lại, chẳng hạn như `callable/method` hoặc `exported`, bằng các mã ngắn (`d0`, `d1`, ...). Ánh xạ được lưu trữ trong phần `[dict]`, trong các chế độ thư mục là tệp `dict.skt`.

Việc tiết kiệm đến từ sự lặp lại. Hầu hết các dòng ký hiệu chia sẻ cùng loại, kiểu và thuộc tính, vì vậy một cơ sở mã lớn lặp lại các chuỗi này hàng nghìn lần. Chi phí là sự gián tiếp: người đọc phải mở rộng các mã bằng từ điển, và một chunk đơn lẻ không còn tự chứa. Từ điển là đầu ra phụ trợ riêng biệt, vì vậy nó không tính vào `--max-lines`.

Đầu ra đã tối giản mã hóa loại mối quan hệ nhưng giữ các dấu hiệu phân giải có thể đọc được, vì vậy các cạnh `unresolved` và `ambiguous` vẫn có thể phân biệt được. Đối với các dự án nhỏ, mức tiết kiệm khiêm tốn và bước bổ sung hiếm khi đáng giá. Đối với các dự án lớn, mức tiết kiệm về token có thể đáng kể.

## Các trang liên quan

- [Hướng dẫn lệnh phân tích](../parse-command/): các lệnh cho từng chế độ.
- [Tài liệu tham chiếu định dạng đầu ra](../../reference/output-format/): ngữ pháp `.skt` và JSON.
- [Tài liệu tham chiếu cờ CLI](../../reference/cli-flags/): mọi cờ và giá trị mặc định.
- [Tài liệu tham chiếu độ tin cậy phân tích](../../reference/analysis-reliability/): đặt tên đầu ra và đảm bảo xuất bản.