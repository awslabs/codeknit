---
title: Hỗ trợ ngôn ngữ và độ chính xác API
description: Khả năng trích xuất đã được kiểm thử, lựa chọn trình phân tích cú pháp và các giới hạn ngôn ngữ đã biết.
---

<!-- Được duy trì trong docs/reference/language-support.md; sao chép nguyên văn bởi docs/gen. -->

Codeknit trích xuất cú pháp mà không cần chạy trình biên dịch, đánh giá macro hoặc tải các phụ thuộc của dự án. Ma trận dưới đây mô tả các cấu trúc đã được kiểm thử, không phải là sự tuân thủ ngôn ngữ hoàn chỉnh hoặc đảm bảo phân giải tên trên nhiều tệp.

Mỗi ngôn ngữ đều có một xác nhận tham số/kiểu trả về cơ bản chính xác trong `internal/plugin/api_fidelity_test.go`. Độ phủ khai báo cũng được kiểm tra trong `internal/plugin/coverage_test.go` và các gói kiểm thử ngôn ngữ riêng lẻ.

| Ngôn ngữ   | Phần mở rộng                                | Cơ sở chữ ký                                                                                                          | Độ phủ khai báo và vùng chứa                                                                        |
| ---------- | ------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| C          | `.c`, `.h`                                  | Tham số được đặt tên và kiểu trả về trong định nghĩa và nguyên mẫu                                                    | Struct, union, enum, typedef, macro; khai báo bên trong guard của header                            |
| C++        | `.cpp`, `.cc`, `.cxx`, `.hpp`, `.hxx`, `.h` | Tham số được đặt tên và kiểu trả về trong hàm, nguyên mẫu và khai báo phương thức                                     | Lớp, struct, phương thức, không gian tên, template; guard của header và wrapper `extern "C"`        |
| C#         | `.cs`                                       | Tham số được đặt tên và kiểu rõ ràng                                                                                  | Lớp, struct, interface, phương thức, trường, delegate, không gian tên                               |
| Go         | `.go`                                       | Tham số được nhóm, không đặt tên, variadic, kiểu hàm và generic; kết quả được đặt tên/nhiều kết quả; receiver generic | Trường struct với kiểu và tag, trường nhúng, chữ ký phương thức interface, tham số kiểu, gói        |
| Java       | `.java`                                     | Tham số được đặt tên và kiểu rõ ràng                                                                                  | Lớp, interface, constructor, phương thức, trường, lớp lồng nhau, gói                                |
| JavaScript | `.js`, `.jsx`, `.mjs`, `.cjs`               | Tham số được đặt tên; không suy luận kiểu trả về                                                                      | Hàm, hàm mũi tên, lớp, trường, phương thức, khai báo chứa JSX                                       |
| PHP        | `.php`                                      | Tham số được đặt tên và kiểu rõ ràng                                                                                  | Lớp, interface, trait, enum, phương thức, thuộc tính, không gian tên                                |
| Python     | `.py`, `.pyi`                               | Chú thích, giá trị mặc định, `*args`, `**kwargs`, dấu phân cách vị trí/từ khóa, chú thích trả về                      | Hàm, lớp, lớp lồng nhau, phương thức được trang trí; bỏ qua receiver liên kết thông thường          |
| Ruby       | `.rb`                                       | Tham số được đặt tên; không suy luận kiểu trả về                                                                      | Lớp, module, phương thức instance/singleton, hằng số                                                |
| Rust       | `.rs`                                       | Mẫu được gõ, `self`/`&self`/`&mut self`, liên kết có thể thay đổi, kiểu trả về rõ ràng                                | Module inline và lồng nhau, trường struct được đặt tên, phương thức trait và phương thức triển khai |
| Scala      | `.scala`, `.sc`                             | Tham số được đặt tên và kiểu rõ ràng                                                                                  | Lớp, trait, đối tượng, phương thức, enum, gói                                                       |
| TypeScript | `.ts`, `.tsx`, `.mts`, `.cts`               | Tham số được đặt tên và kiểu rõ ràng và chú thích trả về                                                              | Hàm, hàm mũi tên, interface, lớp, trường, không gian tên; TSX sử dụng ngữ pháp riêng                |

## Lựa chọn ngôn ngữ cho header

Tất cả các lệnh CLI xử lý nguồn đều chấp nhận `--header-language c|cpp`. Lựa chọn này chỉ áp dụng cho tệp `.h`. Các phần mở rộng khác vẫn giữ trình phân tích cú pháp thông thường của chúng.

```bash
codeknit parse ./src --header-language cpp
codeknit fingerprint ./src --header-language c
codeknit graph analyze ./src --header-language cpp
```

Mặc định là `c`, giữ lại việc trích xuất macro, typedef và union của C. Sử dụng `cpp` cho các header C++ có tên `.h`; các tệp có tên `.hpp` và `.hxx` đã sử dụng C++. Lựa chọn này được chỉ định rõ ràng vì cả hai ngữ pháp đều có thể chấp nhận một số cú pháp của ngôn ngữ khác mà không trích xuất đầy đủ cấu trúc của nó.

Các guard của header và các nhánh điều kiện được duyệt mà không đánh giá điều kiện của chúng. Do đó, kết quả có thể bao gồm các khai báo từ các nhánh loại trừ lẫn nhau. Không áp dụng việc mở rộng macro, phân giải include và cờ build. Giao diện người dùng sử dụng lựa chọn C mặc định cho các tệp `.h`.

## Bảo toàn hình dạng API

Ví dụ, với Go:

```go
func Pair(a, b int, rest ...string) (string, error)
```

sẽ tạo ra:

```text
Pair(a: int, b: int, rest: ...string) -> (string, error)
```

Các phương thức interface của Go và các trường được đặt tên của Go/Rust xuất hiện dưới dạng **ký hiệu** trong cả chế độ bình thường và chế độ fingerprint. Các phương thức và trường giữ lại kiểu sở hữu của chúng. Các module Rust lồng nhau giữ lại tên đủ điều kiện như `service.storage.read`, bao gồm cả khi một module khác khai báo một hàm cùng tên.

Rust cho phép các trường và phương thức có cùng tên. Tên phạm vi trường mang hậu tố `#field` bên trong để ID và các **cạnh** chứa của chúng vẫn riêng biệt; tên hiển thị và chữ ký giữ nguyên tên trường ban đầu.

Chữ ký Python giữ lại chú thích, giá trị mặc định, splat và dấu phân cách quy ước gọi. Một tham số `self` hoặc `cls` đầu tiên thông thường được bỏ qua trong phương thức liên kết, nhưng được giữ lại trong hàm tự do hoặc phương thức tĩnh.

Chữ ký nhiều dòng giữ lại ngắt dòng nguồn trong JSON. SKT thoát ngắt dòng thành `\n` và `\r` để mỗi **ký hiệu** nằm trên một dòng đầu ra vật lý.

## Các giới hạn đã biết

- Các khai báo con trỏ, tham chiếu, mảng và con trỏ hàm phức tạp của C/C++ vẫn có thể có chữ ký không hoàn chỉnh. Không mở rộng các khai báo được tạo bởi macro.
- Việc xử lý constructor của TypeScript/JavaScript và một số dạng tham số tùy chọn, rest hoặc destructured chưa hoàn chỉnh.
- Các ràng buộc generic nâng cao và các dạng kiểu ngoài các trường hợp đã kiểm thử khác nhau theo ngôn ngữ. Không suy luận kiểu trả về.
- Không mở rộng macro của Rust. Ghi lại các khai báo module ngoài dòng; các tệp của chúng phải được bao gồm trong quá trình quét. Trường tuple, chi tiết biến thể enum và các mục liên kết có độ phủ một phần.
- Việc trích xuất cú pháp không phân giải dispatch động, overload, các phụ thuộc được nhập hoặc ngữ nghĩa cấu hình build với độ chính xác của trình biên dịch.
- Lỗi cú pháp một phần vẫn có thể bỏ qua các khai báo. Kiểm tra các cảnh báo được báo cáo và xác minh nguồn trước khi áp dụng refactor.

Việc trích xuất cú pháp hoàn chỉnh hơn sẽ thay đổi số lượng **ký hiệu** và ShortID. Tạo lại đầu ra skeleton hiện có khi nâng cấp thay vì trộn nó với đầu ra cũ.
