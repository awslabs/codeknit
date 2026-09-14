---
title: Hỗ trợ ngôn ngữ và độ chính xác API
description: Khả năng trích xuất đã được kiểm tra, lựa chọn trình phân tích cú pháp và các giới hạn ngôn ngữ đã biết.
---

<!-- Được duy trì trong docs/reference/language-support.md; sao chép nguyên văn bởi docs/gen. -->

Codeknit trích xuất cú pháp mà không cần chạy trình biên dịch, đánh giá macro hoặc tải các phụ thuộc của dự án. Ma trận dưới đây mô tả các cấu trúc đã được kiểm tra, không phải là sự tuân thủ ngôn ngữ hoàn chỉnh hoặc đảm bảo phân giải tên trên nhiều tệp.

Mỗi ngôn ngữ đều có một khẳng định chính xác về tham số/cấu trúc trả về cơ bản trong `internal/plugin/api_fidelity_test.go`. Độ phủ khai báo cũng được kiểm tra trong `internal/plugin/coverage_test.go` và các gói kiểm tra ngôn ngữ riêng lẻ.

| Ngôn ngữ   | Tiện ích mở rộng                                  | Cơ sở chữ ký                                                                                           | Độ phủ khai báo và vùng chứa                                                                         |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Tham số và kiểu trả về có tên trong định nghĩa và nguyên mẫu                                              | Struct, union, enum, typedef, macro; khai báo bên trong header guards                                |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Tham số và kiểu trả về có tên trong hàm, nguyên mẫu và khai báo phương thức                          | Lớp, struct, phương thức, không gian tên, mẫu; header guards và wrapper `extern "C"`                  |
| C#         | `.cs`                                       | Tham số có kiểu và kiểu trả về rõ ràng                                                             | Lớp, struct, giao diện, phương thức, trường, delegate, không gian tên                                       |
| Go         | `.go`                                       | Tham số nhóm, không tên, biến đổi, kiểu hàm và tham số tổng quát; kết quả có tên/nhiều kết quả; bộ nhận tổng quát | Trường struct với kiểu và thẻ, trường nhúng, chữ ký phương thức giao diện, tham số kiểu, gói |
| Java       | `.java`                                     | Tham số có kiểu và kiểu trả về rõ ràng                                                             | Lớp, giao diện, hàm tạo, phương thức, trường, lớp lồng nhau, gói                               |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Tham số có tên; không có kiểu trả về suy luận                                                                    | Hàm, hàm mũi tên, lớp, trường, phương thức, khai báo chứa JSX                          |
| PHP        | `.php`                                      | Tham số có kiểu và kiểu trả về rõ ràng                                                             | Lớp, giao diện, trait, enum, phương thức, thuộc tính, không gian tên                                        |
| Python     | `.py`, `.pyi`                               | Chú thích, giá trị mặc định, `*args`, `**kwargs`, phân tách vị trí/từ khóa, chú thích trả về                | Hàm, lớp, lớp lồng nhau, phương thức được trang trí; bỏ qua bộ nhận liên kết thông thường                |
| Ruby       | `.rb`                                       | Tham số có tên; không có kiểu trả về suy luận                                                                    | Lớp, mô-đun, phương thức instance/singleton, hằng số                                                    |
| Rust       | `.rs`                                       | Mẫu có kiểu, `self`/`&self`/`&mut self`, liên kết có thể thay đổi, kiểu trả về rõ ràng                          | Mô-đun nội tuyến và lồng nhau, trường struct có tên, phương thức trait và phương thức triển khai                   |
| Scala      | `.scala`, `.sc`                             | Tham số có kiểu và kiểu trả về rõ ràng                                                             | Lớp, trait, đối tượng, phương thức, enum, gói                                                         |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Tham số có kiểu và chú thích trả về rõ ràng                                                       | Hàm, hàm mũi tên, giao diện, lớp, trường, không gian tên; TSX sử dụng ngữ pháp riêng              |

## Lựa chọn ngôn ngữ cho tệp tiêu đề

Tất cả các lệnh CLI xử lý nguồn đều chấp nhận `--header-language c|cpp`. Lựa chọn này chỉ áp dụng cho tệp `.h`. Các tiện ích mở rộng khác vẫn sử dụng trình phân tích cú pháp thông thường của chúng.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

Mặc định là `c`, giữ lại việc trích xuất macro, typedef và union của C. Sử dụng `cpp` cho các tệp tiêu đề C++ có tên `.h`; các tệp có tên `.hpp` và `.hxx` đã sử dụng C++. Lựa chọn này được chỉ định rõ ràng vì cả hai ngữ pháp đều có thể chấp nhận một số cú pháp của ngôn ngữ khác mà không trích xuất đầy đủ cấu trúc của nó.

Các header guard và nhánh điều kiện được duyệt mà không đánh giá điều kiện của chúng. Do đó, kết quả có thể bao gồm các khai báo từ các nhánh loại trừ lẫn nhau. Việc mở rộng macro, phân giải include và cờ xây dựng không được áp dụng. Giao diện người dùng sử dụng lựa chọn C mặc định cho các tệp `.h`.

## Bảo toàn hình dạng API

Ví dụ, với Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

sẽ tạo ra:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Các phương thức giao diện Go và các trường có tên của Go/Rust xuất hiện dưới dạng [symbols] trong cả chế độ bình thường và chế độ fingerprint. Các phương thức và trường giữ lại loại sở hữu của chúng. Các mô-đun Rust lồng nhau giữ lại tên đủ điều kiện như `service.storage.read`, bao gồm cả khi một mô-đun khác khai báo một hàm có cùng tên.

Rust cho phép các trường và phương thức có cùng tên. Tên phạm vi của trường mang hậu tố `#field` nội bộ để ID và các cạnh chứa của chúng vẫn riêng biệt; tên hiển thị và chữ ký giữ nguyên tên trường ban đầu.

Chữ ký Python giữ lại chú thích, giá trị mặc định, splat và các dấu phân cách quy ước gọi. Một tham số `self` hoặc `cls` đầu tiên thông thường được bỏ qua trong phương thức liên kết, nhưng được giữ lại trong hàm tự do hoặc phương thức tĩnh.

Chữ ký nhiều dòng giữ lại ngắt dòng nguồn trong JSON. SKT thoát các ngắt dòng thành `\n` và `\r` để mỗi [symbol] vẫn nằm trên một dòng đầu ra vật lý.

## Phân giải mối quan hệ

Các mối quan hệ giữ lại sự đủ điều kiện của bộ nhận và quyền sở hữu từ vựng. Các import tương đối và import đã đổi tên của JavaScript/TypeScript và Python cung cấp ngữ cảnh mục tiêu. Khai báo gói Go và Java có thể phân giải trên các tệp trong cùng thư mục. Một import không khớp không bao giờ quay lại một khai báo không liên quan, và các ứng viên cạnh tranh vẫn mơ hồ.

Các liên kết bộ nhận cục bộ được ưu tiên hơn các import không gian tên, bao gồm cả khi một tham số che khuất một không gian tên đã import. Các import gói Python xem xét các khai báo trong mô-đun `__init__` của gói, không phải các mô-đun con tùy ý. Các re-export không thể được thiết lập từ mô-đun đó vẫn không được phân giải.

Việc trích xuất callable của Java, C++, C#, Scala và TypeScript giữ lại phạm vi byte khai báo để các phần thân overload có thể có người gọi và phạm vi bí danh riêng biệt, bao gồm các khai báo trên cùng một dòng. Việc chọn một mục tiêu overload vẫn mang tính bảo thủ. Các mối quan hệ override rõ ràng trong Java, C++, C#, Scala và TypeScript tìm kiếm khai báo kế thừa gần nhất; các cơ sở thiếu, chu kỳ hoặc các ứng viên cạnh tranh giữ lại sự không chắc chắn thay vì tạo ra các liên kết tự tham chiếu.

Đây là phân tích dựa trên cú pháp, không phải trình biên dịch hoặc đồ thị gọi thời gian chạy. Các chỉ định gói JavaScript đơn giản, re-export gói, cấu hình xây dựng, các kiểu bộ nhận phức tạp và điều phối động có thể vẫn không được phân giải. Một tên duy nhất ở nơi khác trong kho lưu trữ không đủ bằng chứng về sự phụ thuộc.

JSON giữ lại các cạnh không chắc chắn với `resolution: "unresolved"` hoặc `"ambiguous"` và bao gồm các ID ứng viên khi có sẵn. Các cạnh này không có short ID mục tiêu.

SKT sử dụng cùng một ký hiệu cạnh cho tất cả các mối quan hệ, với một dấu phân giải tùy chọn:

```text
S8 --references--> S5
S10 --references[unresolved]--> string, bool
S11 --calls[ambiguous]--> Helper [candidates=S12, S15]
```

Các điểm cuối và ứng viên đã biết sử dụng short ID, bao gồm các điểm cuối của các mối quan hệ không được phân giải. Các mục tiêu không chắc chắn không có [symbol] sử dụng tên tương đối so với tệp nguồn. Các tên literal trông giống như short ID, và các định danh phức tạp khác, sử dụng full ID được trích dẫn để tránh mơ hồ. Các cạnh có cùng nguồn, loại, phân giải và tập ứng viên chia sẻ một danh sách mục tiêu. Bộ đọc SKT giữ lại phân giải và ứng viên, bao gồm các tham chiếu đến [symbols] trong các khối đầu ra sau. Đầu ra đã thu gọn mã hóa loại mối quan hệ trong khi giữ dấu phân giải có thể đọc được.

C/C++ bao gồm các [symbols] `meta/file` và `meta/include` rõ ràng. Các cách viết tiêu đề xuất hiện một lần trong danh sách [symbols]; các cạnh sử dụng short ID của chúng:

```text
[symbols]
## src/main.c
S1 meta/file L1-L4 main.c
S2 meta/include L1-L1 "config.h" {resolution=unresolved}
S3 meta/include L2-L2 "utils.h" {resolution=unresolved}
S4 meta/include L3-L3 <stdio.h> {resolution=unresolved}

[edges]
S1 --references[unresolved]--> S2, S3, S4
```

Một [symbol] include mô tả chỉ thị đã viết, không phải một khai báo tiêu đề đã phân giải. Các include lặp lại trong một tệp chia sẻ một điểm cuối. Siêu dữ liệu tệp/include bị loại trừ khỏi xếp hạng phụ thuộc và phát hiện mã chết; việc thêm ID không khẳng định rằng một tiêu đề đã được tìm thấy thông qua cấu hình tìm kiếm include của trình biên dịch.

Các số liệu phụ thuộc và liên kết đồ thị HTML chỉ sử dụng các cạnh đã phân giải. Các báo cáo hiển thị số lượng mối quan hệ bị loại trừ, và đồ thị HTML hiển thị số lượng không chắc chắn. Fan-in và fan-out đếm các lân cận riêng biệt. Phân tích kế thừa xem xét tất cả các cơ sở và giao diện; kế thừa vòng không có độ sâu được báo cáo. Khả năng tiếp cận bắt đầu từ các callable `main`, `Main` hoặc `init` có tên và theo dõi các cạnh phụ thuộc. Việc tiếp cận một thành viên làm cho vùng chứa của nó trở nên liên quan mà không làm cho mọi sibling có thể tiếp cận được. Nếu không có điểm vào được nhận dạng, các phát hiện về khả năng tiếp cận sẽ bị bỏ qua.

Các đối số callback và callable trả về là các tham chiếu, không phải bằng chứng về việc gọi. Phân giải bí danh giữ lại danh tính tệp, hàm và đối tượng; các phép gán xung đột vẫn không chắc chắn. Các phát hiện đồ thị là các ứng viên để xem xét: các trình gọi bên ngoài và hành vi động có thể thay đổi kết quả. Các quy tắc lớp dựa trên thư mục và trọng số lan truyền thay đổi là các heuristic, không phải xác suất đo được.

## Các giới hạn đã biết

- Các khai báo con trỏ, tham chiếu, mảng và con trỏ hàm phức tạp của C/C++ vẫn có thể có chữ ký không đầy đủ. Các khai báo được tạo bởi macro không được mở rộng.
- Việc xử lý hàm tạo của TypeScript/JavaScript và một số dạng tham số tùy chọn, rest hoặc destructured chưa hoàn chỉnh.
- Các ràng buộc generic nâng cao và các dạng kiểu ngoài các trường hợp đã kiểm tra khác nhau theo ngôn ngữ. Kiểu trả về không được suy luận.
- Các macro Rust không được mở rộng. Các khai báo mô-đun ngoài dòng được ghi lại; các tệp của chúng phải được bao gồm trong quá trình quét. Các trường tuple, chi tiết biến thể enum và các mục liên kết có độ phủ một phần.
- Việc trích xuất cú pháp không phân giải điều phối động, overload, các phụ thuộc đã import hoặc ngữ nghĩa cấu hình xây dựng với độ chính xác của trình biên dịch.
- Các lỗi cú pháp một phần vẫn có thể bỏ sót các khai báo. Kiểm tra các cảnh báo đã báo cáo và xác minh nguồn trước khi áp dụng tái cấu trúc.

Việc trích xuất hoàn chỉnh hơn thay đổi số lượng [symbols] và short ID. Tạo lại đầu ra skeleton hiện có khi nâng cấp thay vì trộn nó với đầu ra cũ hơn.