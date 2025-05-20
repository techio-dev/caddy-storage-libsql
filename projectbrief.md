# Project Brief

## Mục tiêu cốt lõi
- Phát triển một module Caddy để lưu trữ chứng chỉ SSL (và có thể các dữ liệu lưu trữ khác của Caddy) bằng cách sử dụng LibSQL (Turso) làm backend.

## Yêu cầu chính
- Module phải tích hợp liền mạch với kiến trúc lưu trữ của Caddy.
- Sử dụng thư viện LibSQL Go để tương tác với cơ sở dữ liệu Turso.
- Đảm bảo lưu trữ và truy xuất dữ liệu an toàn và hiệu quả.
- Cung cấp tài liệu rõ ràng về cách cấu hình và sử dụng module.

## Phạm vi dự án
- **Trong phạm vi:**
    - Triển khai giao diện `StorageConverter` của Caddy.
    - Triển khai các phương thức `Store`, `Load`, `Delete`, `Exists`, `List`, `Stat`.
    - Xử lý kết nối và xác thực với Turso.
    - Thiết kế schema cơ sở dữ liệu cho LibSQL.
    - Viết unit test cơ bản.
- **Ngoài phạm vi (ban đầu):**
    - Giao diện người dùng quản trị phức tạp cho việc quản lý lưu trữ.
    - Hỗ trợ nhiều cơ sở dữ liệu LibSQL khác nhau trong một cấu hình Caddy duy nhất (trừ khi được yêu cầu cụ thể).
    - Các tính năng tối ưu hóa hiệu suất nâng cao (sẽ được xem xét sau khi các chức năng cốt lõi hoạt động).
