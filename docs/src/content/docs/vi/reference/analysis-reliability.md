---
title: Độ tin cậy của phân tích
description: Lựa chọn nguồn, phân tích không hoàn chỉnh, xuất bản đầu ra và đảm bảo phát hiện trùng lặp.
---

<!-- Được duy trì trong docs/reference/analysis-reliability.md; sao chép nguyên văn bởi docs/gen. -->

Các hành vi này mô tả những thay đổi Go/CLI chưa phát hành sau phiên bản 0.5.0.

## Lựa chọn nguồn và chẩn đoán

Tất cả năm lệnh xử lý nguồn (`parse`, `fingerprint`, `graph show`, `graph analyze`, và `graph hotspots`) chấp nhận:

| Cờ                      | Mặc định              | Ý nghĩa                                                              |
| ----------------------- | --------------------- | -------------------------------------------------------------------- |
| `--include GLOB`        | Tất cả nguồn hỗ trợ   | Bao gồm các đường dẫn nguồn phù hợp; lặp lại để chọn nhiều mẫu.      |
| `--exclude GLOB`        | Không có              | Loại trừ các đường dẫn hoặc thư mục phù hợp; lặp lại cho nhiều mẫu.  |
| `--max-file-size BYTES` | `10485760` (10 MiB)   | Bỏ qua các tệp nguồn lớn hơn; `0` vô hiệu hóa giới hạn này.          |
| `--strict`              | `false`               | Thất bại trước khi xuất kết quả nếu xảy ra chẩn đoán quét hoặc phân tích. |

Các mẫu sử dụng glob kiểu gitignore phân tách bằng dấu gạch chéo, tương đối với thư mục đầu vào đã chọn. Đối với đầu vào là một tệp duy nhất, các mẫu khớp với tên cơ sở của nó. Trích dẫn các mẫu để shell không mở rộng chúng. Các bao gồm tạo thành một hợp; các loại trừ được ưu tiên. Các bao gồm không ghi đè `.gitignore`, lọc kiểm thử hoặc các loại trừ tích hợp sẵn. Mẫu phủ định (`!pattern`) không được chấp nhận trong các cờ này; sử dụng các điều khiển bao gồm và loại trừ riêng biệt.

```bash
codeknit parse ./src --include '**/*.go' --exclude '**/generated/' --strict
codeknit fingerprint ./src --max-file-size 2097152 --include '**/*.ts'
codeknit parse ./src/app.go --output-mode inline --format json
```

Quét thư mục tuân thủ các tệp `.gitignore` ở thư mục cha và thư mục con tương đối với vị trí dự án của chúng, độc lập với thư mục làm việc của tiến trình. Thư mục `.git` hoặc dấu hiệu worktree gần nhất xác định ranh giới kho lưu trữ. Bên ngoài kho lưu trữ, các tệp ignore ở thư mục cha vẫn được áp dụng. Một thư mục con được chọn bên dưới một thư mục bị ignore vẫn bị ignore. Việc chọn rõ ràng một tệp nguồn thông thường sẽ bỏ qua ignore và lọc kiểm thử, nhưng vẫn áp dụng giới hạn bao gồm/loại trừ và kích thước.

Các liên kết tượng trưng nguồn, liên kết tượng trưng thư mục và các nguồn không thông thường bị bỏ qua và báo cáo. Việc đọc bị giới hạn trong gốc quét và giới hạn kích thước được thực thi lại trong khi đọc để một tệp phát triển sau khi phát hiện không thể vượt qua giới hạn. Các loại trừ thư mục tích hợp sẵn bao gồm `.git`, `.hg`, `.svn`, `node_modules`, bộ nhớ đệm Python và các môi trường ảo Python phổ biến. Các tệp kiểm thử vẫn bị loại trừ trừ khi `--collect-test` được thiết lập.

Các thư mục/tệp không thể đọc được, đọc tệp ignore thất bại, liên kết tượng trưng bị bỏ qua, các nguồn quá kích thước và cảnh báo cú pháp khiến phân tích không hoàn chỉnh. Theo mặc định, các kết quả một phần có thể sử dụng được vẫn được giữ lại và chẩn đoán xuất hiện trên stderr. JSON phân tích bao gồm các bản ghi đường dẫn/lý do trong `errors`; SKT bao gồm `[errors]` nội tuyến hoặc trong `warnings.skt`. Các lệnh đồ thị/fingerprint cũng báo cáo chẩn đoán trên stderr. Các tệp bị bỏ qua hoặc lọc một cách cố ý không phải là cảnh báo. `--strict` từ chối các kết quả không hoàn chỉnh trước khi xuất bản đầu ra. Một đầu vào không có nguồn hỗ trợ được chọn hoặc một lần chạy trong đó mọi nguồn đọc đều thất bại luôn thoát không thành công.

Giao diện người dùng văn bản (TUI) sử dụng cùng chính sách liên kết tượng trưng và mặc định 10 MiB. Các biểu mẫu của nó chưa hiển thị các cờ lựa chọn và strict mới.

## Tên đầu ra và xuất bản

Đầu ra dạng cây thêm `.skt` vào tên nguồn hoàn chỉnh: `src/foo.c` trở thành `src/foo.c.skt`, trong khi `src/foo.h` trở thành `src/foo.h.skt`. Đầu ra chia nhỏ sử dụng `src/foo.c_part1.skt`, `src/foo.c_part2.skt`, v.v. Các tệp từ điển và chẩn đoán giữ nguyên các tên dự trữ `dict.skt` và `warnings.skt`. Các đường dẫn đầu ra xung đột, bao gồm cả xung đột chỉ khác chữ hoa/thường, đều thất bại rõ ràng.

SKT dạng cây yêu cầu `--max-lines` ít nhất là 3. Mỗi khối ký hiệu lặp lại tiêu đề `[symbols]` và `## source-path` của nó. Các khối chỉ có cạnh mang `[edges]`. Mọi khối nguồn đều tuân thủ giới hạn; các tệp từ điển và chẩn đoán là đầu ra phụ trợ riêng biệt. SKT dạng phẳng giữ nguyên các khối nguồn, vì vậy giới hạn dòng của nó là mục tiêu chia nhỏ thay vì giới hạn cứng cho một khối nguồn lớn. JSON là một tài liệu duy nhất và không bị chia nhỏ theo số dòng.

Việc tạo thư mục hoàn thành trong một thư mục tạm thời cùng cấp. Nếu không có `--clean`, đầu ra đã tạo hiện có cho định dạng đã chọn sẽ gây ra lỗi. Với `--clean`, các tệp `.skt` lỗi thời (chế độ SKT) hoặc `codeknit.json` (chế độ JSON) chỉ được thay thế sau khi tạo thành công. Các mục không liên quan được giữ nguyên; xung đột với một mục sẽ gây ra lỗi. Không sử dụng thư mục nguồn hoặc thư mục cha của nó làm thư mục đầu ra.

Việc xuất bản giữ một khóa dành riêng cho thư mục, sao lưu đầu ra trước đó và khôi phục nó nếu thay thế thất bại. Nếu một tiến trình kết thúc giữa hai lần đổi tên thư mục, lần gọi tiếp theo sẽ khôi phục bản sao lưu trước khi tạo bất kỳ thứ gì. Tệp `.NAME.codeknit.lock` cùng cấp cố ý vẫn còn để điều phối các nhà xuất bản trong tương lai; không xóa khóa đang hoạt động hoặc bản sao lưu khôi phục. Có một khoảng thời gian ngắn giữa hai lần đổi tên thư mục, vì vậy những người đọc yêu cầu truy cập không bị gián đoạn nên điều phối với việc tạo. Điều này cung cấp khả năng khôi phục sau lỗi tiến trình, không phải đảm bảo chống lại lỗi lưu trữ hoặc mất điện.

Các báo cáo riêng lẻ được ghi và đồng bộ hóa vào một tệp tạm thời cùng cấp trước khi thay thế đích đến. Các tệp hiện có không bị cắt ngắn trong quá trình tạo. Đầu ra stdout nội tuyến không thể khôi phục nếu người tiêu dùng của nó thất bại.

## Đảm bảo phát hiện trùng lặp

So khớp chính xác có nghĩa là các luồng token chuẩn hóa hoàn chỉnh giống hệt nhau trong cùng danh mục ký hiệu giữa các khai báo đủ điều kiện để tạo fingerprint. Nó không có nghĩa là văn bản nguồn giống hệt nhau hoặc chứng minh được tính tương đương về mặt ngữ nghĩa. Các nhóm ứng viên chính xác sử dụng chính các luồng token, thay vì sự bằng nhau của fuzzy hashing.

Dưới 500 ký hiệu đủ điều kiện, phát hiện cấu trúc so sánh tất cả các cặp cùng danh mục. Từ 500 trở lên, tất cả các cặp token chính xác được giữ lại độc lập với việc truy xuất ứng viên, và các cặp khác sử dụng khoảng 50 láng giềng cấu trúc gần đúng cho mỗi ký hiệu. Do đó, việc truy xuất gần trùng lặp không đầy đủ trên các đầu vào lớn. Việc truy xuất vẫn quét tất cả các vector; nó không làm cho thuật toán tổng thể trở nên dưới bậc hai.

Việc lọc phạm vi độ tương đồng vẫn được áp dụng. Phạm vi mặc định là 65–95%, loại trừ các kết quả trùng khớp 100%. Để bao gồm các kết quả trùng khớp chính xác, hãy sử dụng:

```bash
codeknit fingerprint ./src --min-similarity 100 --max-similarity 100
```

Cờ `--rerank` tùy chọn thêm các ứng viên ngữ nghĩa, áp dụng sự đồng thuận cosine và lọc theo điểm số kết hợp; nó có thể loại bỏ một ứng viên chính xác về mặt cấu trúc. Văn bản nhúng đến từ cùng một ảnh chụp nguồn giới hạn được sử dụng để trích xuất. Các đầu vào ổn định và điểm số cố định tạo ra thứ tự xác định, bao gồm cả các trường hợp hòa; các phản hồi của mô hình tự chúng là đầu vào bên ngoài.

`internal/emitter/calibration_test.go` đánh giá một bộ kiểm tra nguồn được gắn nhãn nhỏ ở ngưỡng 70%, báo cáo true/false positive và negative, độ chính xác và độ nhạy riêng biệt với điểm số độ tương đồng. Bộ kiểm tra có một cặp chính xác, hai cặp gần trùng lặp và ba cặp không liên quan, được kiểm tra dưới và trên ranh giới 500 ký hiệu. Đây là một kiểm tra hồi quy, không phải là ước tính đại diện cho độ chính xác trong thế giới thực. Các ngân sách hiệu chuẩn và hiệu suất đa ngôn ngữ lớn hơn vẫn là công việc trong lộ trình.