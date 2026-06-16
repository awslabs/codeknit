---
title: Sử dụng với Trợ lý AI
description: Thiết lập codeknit như một kỹ năng cho Kiro, Claude Code và các trợ lý lập trình AI khác.
---

codeknit đi kèm với các kỹ năng được tạo sẵn giúp dạy các trợ lý lập trình AI cách sử dụng nó một cách hiệu quả. Những kỹ năng này cho phép trợ lý trích xuất cấu trúc mã, phát hiện các đoạn trùng lặp và thực hiện phân tích cấu trúc mà không cần prompting thủ công.

## Tổng quan về kỹ năng

codeknit cung cấp hai kỹ năng:

- **`codeknit-parse`**: Dạy trợ lý cách trích xuất cấu trúc mã (hàm, lớp, phương thức, biến) và các mối quan hệ (lời gọi, kế thừa, bao chứa) vào các tệp `.skt`.
- **`codeknit-fingerprint`**: Dạy trợ lý cách phát hiện mã trùng lặp và gần trùng lặp bằng cách sử dụng fuzzy hashing.

Mỗi kỹ năng bao gồm tài liệu mà trợ lý đọc theo yêu cầu để hiểu cách sử dụng, cờ lệnh, chế độ đầu ra và quy trình làm việc.

## Cài đặt

Sao chép các thư mục kỹ năng vào thư mục kỹ năng của trợ lý.

Đối với **Kiro**:

```bash
cp -r skills/codeknit-parse ~/.kiro/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.kiro/skills/codeknit-fingerprint
```

Đối với **Claude Code**:

```bash
cp -r skills/codeknit-parse ~/.claude/skills/codeknit-parse
cp -r skills/codeknit-fingerprint ~/.claude/skills/codeknit-fingerprint
```

Sau khi cài đặt, trợ lý sẽ tự động biết cách gọi các lệnh codeknit, chọn các cờ lệnh phù hợp và diễn giải đầu ra `.skt`.

## Nội dung mỗi kỹ năng dạy

### codeknit-parse

Kỹ năng `codeknit-parse` dạy trợ lý cách:

- Chạy `codeknit parse` với các cờ lệnh phù hợp cho các tình huống khác nhau
- Chọn chế độ đầu ra phù hợp:
  - `directory-flat` (mặc định) cho hầu hết các dự án
  - `inline` cho các tệp đơn hoặc đầu vào nhỏ
  - `directory-tree` để phản ánh cấu trúc nguồn
- Đọc và diễn giải các tệp đầu ra `.skt`, bao gồm các phần `[symbols]`, `[edges]` và tùy chọn `[dict]`
- Sử dụng dữ liệu cấu trúc cho việc tái cấu trúc, lập bản đồ phụ thuộc và đánh giá mã
- Chạy `codeknit graph analyze` để có cái nhìn sâu hơn về chất lượng mã (phụ thuộc vòng, ký hiệu trung tâm, god classes, v.v.)

### codeknit-fingerprint

Kỹ năng `codeknit-fingerprint` dạy trợ lý cách:

- Sử dụng `codeknit fingerprint` để phát hiện trùng lặp, kiểm tra DRY và xác định các đoạn mã cần tái cấu trúc
- Chọn phạm vi độ tương đồng phù hợp (`--min-similarity`, `--max-similarity`)
- Đọc phần `[duplicates]` để xác định mã gần trùng lặp
- Hiểu rằng fingerprints đo lường hình dạng cấu trúc, không phải ý định ngữ nghĩa
- Sử dụng `--rerank` với các embeddings Ollama để giảm dương tính giả khi cần thiết

## Ví dụ về quy trình làm việc

### Phân tích cấu trúc

1. Yêu cầu trợ lý phân tích cấu trúc codebase của bạn
2. Trợ lý chạy `codeknit parse ./src` và đọc các tệp `.skt` kết quả
3. Trợ lý trả lời các câu hỏi về cấu trúc: phụ thuộc, chuỗi gọi, dead code
4. Để có cái nhìn sâu hơn, trợ lý chạy `codeknit graph analyze ./src` và diễn giải báo cáo

```skt
[symbols]
## src/service.go
S1 type/struct L5-L8 AuthService {}
S2 callable/method L10-L15 Authenticate(token: string) {receiver=*AuthService}

[edges]
S1 --contains--> S2
```

### Phát hiện trùng lặp

1. Yêu cầu trợ lý tìm mã trùng lặp
2. Trợ lý chạy `codeknit fingerprint ./src`
3. Trợ lý đọc phần `[duplicates]` trong đầu ra
4. Trợ lý điều tra các cặp được gắn cờ và đề xuất hợp nhất

```skt
[duplicates]
S1, S2: 87% độ tương đồng
S3, S4: 76% độ tương đồng
```

## Mẹo

- **Luôn đọc tệp `.skt`, không phải mã nguồn thô, cho các câu hỏi về cấu trúc** — chúng chứa cấu trúc đã trích xuất ở định dạng nhỏ gọn và đáng tin cậy
- Sử dụng `codeknit graph analyze` để phát hiện các vấn đề về chất lượng mã như phụ thuộc vòng, ký hiệu trung tâm và chuỗi kế thừa sâu
- Chạy `codeknit fingerprint` trước khi tái cấu trúc lớn để xác định mã sao chép cần được hợp nhất
- Định dạng `.skt` được thiết kế để tiết kiệm token, làm cho nó trở nên lý tưởng cho các cửa sổ ngữ cảnh LLM
- Sử dụng `--minify` để giảm thêm việc sử dụng token khi xử lý các codebase lớn
