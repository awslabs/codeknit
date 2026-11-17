---
title: Sử dụng với Trợ lý AI
description: Cài đặt plugin codeknit cho Claude Code, Codex và Kiro.
---
Cài đặt plugin codeknit trong Claude Code, Codex hoặc Kiro, giữ cho nó luôn cập nhật và đặt các câu hỏi về cấu trúc mã và mã trùng lặp cho trợ lý. Plugin chứa một kỹ năng, `codeknit`. Kỹ năng này hướng dẫn trợ lý khi nào chạy `codeknit parse`, `codeknit graph analyze` và `codeknit fingerprint`, cũng như cách đọc đầu ra `.skt`.

Bạn không cần tự cài đặt binary `codeknit`. Lần đầu tiên kỹ năng chạy, nó sẽ chạy `scripts/install-codeknit.sh` đi kèm, thực hiện các bước sau:

1. Thoát ngay lập tức nếu `codeknit` đã có trên `PATH` hoặc trong `~/.local/bin`.
2. Nếu không, tải xuống bản phát hành GitHub mới nhất cho hệ điều hành và kiến trúc của bạn.
3. Xác minh tải xuống với `checksums.txt` của bản phát hành.
4. Cài đặt binary vào `~/.local/bin` và in đường dẫn tuyệt đối của nó trên dòng cuối cùng của đầu ra.

Nếu `~/.local/bin` không có trên `PATH`, trợ lý sẽ gọi codeknit bằng đường dẫn đầy đủ đó trong suốt phiên làm việc.

Để kiểm soát việc cài đặt, hãy đặt các biến môi trường sau:

| Biến | Tác dụng |
| --- | --- |
| `CODEKNIT_VERSION` | Ghim một bản phát hành, ví dụ `0.6.0`. |
| `CODEKNIT_INSTALL_DIR` | Thay đổi vị trí cài đặt. |

Để tự cài đặt binary, xem [Cài đặt](../../installation/).

## Cài đặt plugin

Chọn phần phù hợp với trợ lý của bạn.

### Claude Code

Chạy các lệnh sau trong shell của bạn. Trong phiên làm việc Claude Code, chạy các lệnh tương tự dưới dạng `/plugin ...`.

```bash
claude plugin marketplace add awslabs/codeknit --sparse .claude-plugin skills
claude plugin install codeknit@codeknit
```

### Codex

```bash
codex plugin marketplace add awslabs/codeknit --sparse .agents --sparse .codex-plugin --sparse skills
codex plugin add codeknit@codeknit
```

### Kiro IDE

1. Mở bảng Powers.
2. Chọn **Add Custom Power**, sau đó **Import power from GitHub**.
3. Nhập `https://github.com/awslabs/codeknit`.

### Kiro CLI

Kiro CLI cài đặt từ đường dẫn cục bộ, vì vậy hãy clone kho lưu trữ trước:

```bash
git clone --depth 1 --filter=blob:none --sparse https://github.com/awslabs/codeknit ~/.codeknit-power
git -C ~/.codeknit-power sparse-checkout set skills
kiro-cli powers install ~/.codeknit-power
```

### Về `--sparse`

Tùy chọn `--sparse` là tùy chọn. Nó giới hạn việc clone chỉ đến các tệp kê khai plugin và `skills/` thay vì toàn bộ cây nguồn.

### Cài đặt kỹ năng thủ công

Nếu trợ lý của bạn đọc Agent Skills nhưng không hỗ trợ plugin, hãy sao chép `skills/codeknit/` vào thư mục kỹ năng của trợ lý:

- `~/.claude/skills/`
- `~/.codex/skills/`
- `~/.kiro/skills/`

## Cập nhật plugin

| Trợ lý | Lệnh hoặc hành động |
| --- | --- |
| Claude Code | `claude plugin marketplace update codeknit` |
| Codex | `codex plugin marketplace upgrade` |
| Kiro IDE | Chọn **Check for updates** trong bảng Powers. |
| Kiro CLI | Chạy `git -C ~/.codeknit-power pull`, sau đó chạy lại `kiro-cli powers install ~/.codeknit-power`. |

Việc cập nhật plugin không cập nhật binary đã cài đặt. Để thay đổi phiên bản binary, hãy đặt `CODEKNIT_VERSION` như mô tả ở trên.

Nếu bạn nâng cấp codeknit và việc trích xuất thay đổi, ID ký hiệu cũng có thể thay đổi. Yêu cầu trợ lý tái tạo tất cả đầu ra thay vì trộn các phần từ các lần chạy khác nhau.

## Những gì kỹ năng dạy trợ lý

Kỹ năng cung cấp cho trợ lý các quy tắc sau, vì vậy bạn không cần phải lặp lại chúng trong mỗi yêu cầu:

- **Chọn lệnh theo câu hỏi.**
  - `parse` trả lời những gì tồn tại, vị trí của nó và những gì gọi, kế thừa hoặc chứa gì.
  - `graph analyze` tìm kiếm các phụ thuộc vòng, trung tâm, mã chết, god classes và nút thắt cổ chai.
  - `fingerprint` tìm kiếm mã sao chép và các ứng viên tái cấu trúc.
- **Mặc định sử dụng đầu ra `directory-flat`.** Trợ lý viết các tệp chunked vào `./skeleton/` và chỉ đọc các chunk cần thiết. Nó chỉ sử dụng đầu ra `inline` cho một tệp duy nhất hoặc một số ít tệp, vì `inline` đổ mọi thứ vào cửa sổ ngữ cảnh.
- **Đọc đầu ra `.skt`, không phải mã nguồn thô, cho các câu hỏi về cấu trúc.** Trợ lý chỉ mở các tệp nguồn để kiểm tra hoặc thay đổi chi tiết triển khai, sử dụng các phạm vi dòng từ đầu ra. Tệp `references/output-format.md` của kỹ năng chứa định dạng đầy đủ. Xem thêm [tài liệu tham khảo định dạng đầu ra](../../reference/output-format/).
- **Chọn phạm vi độ tương đồng và xử lý các kết quả trùng khớp như tín hiệu.** Fingerprints so sánh hình dạng cấu trúc, không phải ý định. Trợ lý đọc cả hai ký hiệu của một cặp trước khi đề xuất tái cấu trúc.

## Yêu cầu phân tích cấu trúc

1. Yêu cầu trợ lý phân tích cấu trúc codebase của bạn. Ví dụ: "Lập bản đồ cấu trúc của `./src` và cho tôi biết mô-đun nào phụ thuộc vào gói auth."
2. Trợ lý đảm bảo binary đã được cài đặt, sau đó chạy `codeknit parse` với các cạnh được bật và đọc đầu ra `.skt`.
3. Đặt các câu hỏi tiếp theo về các phụ thuộc, chuỗi gọi hoặc mã chết. Trợ lý trả lời từ đầu ra.
4. Để phân tích sâu hơn, yêu cầu đánh giá cấu trúc. Ví dụ: "Tìm các phụ thuộc vòng và god classes trong `./src`." Trợ lý chạy `codeknit graph analyze` và đọc báo cáo.

Các mối quan hệ được đánh dấu `[unresolved]` hoặc `[ambiguous]` trong đầu ra là không chắc chắn, không phải là các phụ thuộc đã được xác nhận. Xem các câu trả lời dựa trên chúng như các ứng viên cần xác minh.

Để biết các lệnh đằng sau các yêu cầu này, xem [hướng dẫn lệnh parse](../parse-command/) và [hướng dẫn các lệnh đồ thị](../graph-commands/).

## Yêu cầu phát hiện trùng lặp

1. Yêu cầu trợ lý tìm mã trùng lặp. Ví dụ: "Tìm các hàm sao chép trong `./src`."
2. Trợ lý chạy `codeknit fingerprint` và đọc phần `[duplicates]` của `./skeleton/fingerprints.skt`.
3. Trợ lý mở cả hai ký hiệu của mỗi cặp được gắn cờ, sau đó đề xuất tái cấu trúc.
4. Nếu kết quả có nhiều nhiễu, yêu cầu nó điều chỉnh phạm vi độ tương đồng hoặc thêm sắp xếp lại ngữ nghĩa. Việc sắp xếp lại yêu cầu một máy chủ Ollama cục bộ với mô hình nhúng đã được tải xuống.

Để biết các tùy chọn đằng sau các yêu cầu này, xem [hướng dẫn lệnh fingerprint](../fingerprint-command/).

## Mẹo

- Nhắc trợ lý đọc các tệp `.skt`, không phải mã nguồn thô, cho các câu hỏi về cấu trúc. Điều này giúp giữ ngữ cảnh nhỏ gọn, vì định dạng này nhỏ gọn và chứa các ký hiệu, mối quan hệ và phạm vi dòng.
- Sử dụng `graph analyze` để có cái nhìn sâu sắc về chất lượng mã. Các phát hiện là ứng viên đánh giá, vì các lời gọi bên ngoài và điều phối động có thể thay đổi kết quả.
- Chạy fingerprint trước khi tái cấu trúc lớn để phát hiện mã sao chép cần được hợp nhất.
- Nếu trợ lý chạy lại `parse` vào thư mục đã chứa đầu ra `.skt`, codeknit sẽ từ chối ghi. Yêu cầu nó truyền `--clean` để thay thế đầu ra trước đó.
- Nếu tải xuống binary thất bại (không kết nối mạng hoặc nền tảng không được hỗ trợ), hãy tải bản phát hành thủ công từ [trang releases](https://github.com/awslabs/codeknit/releases), hoặc xem [Cài đặt](../../installation/).
- Nếu một lần chạy báo cáo kết quả một phần, hãy kiểm tra stderr và phần `[errors]`. Xem [độ tin cậy phân tích](../../reference/analysis-reliability/) để biết chi tiết.