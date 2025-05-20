# Kế hoạch Phát triển Module Caddy với LibSQL (Turso) Storage

**Giai đoạn 1: Thiết lập và Cấu trúc Cơ bản**

1.  **Thiết lập Môi trường Go:**
    *   Đảm bảo bạn đã cài đặt Go phiên bản mới nhất.
    *   Khởi tạo một Go module mới cho dự án của bạn (ví dụ: `go mod init github.com/yourusername/caddy-libsql-storage`).
2.  **Cấu trúc Thư mục Module Caddy:**
    *   Tạo cấu trúc thư mục cơ bản cho module. Thông thường, mã nguồn sẽ nằm trong một thư mục con (ví dụ: `libsqlstorage`).
3.  **Thêm Dependencies:**
    *   Thêm Caddy làm dependency: `go get github.com/caddyserver/caddy/v2`
    *   Thêm thư viện LibSQL Go client: `go get github.com/tursodatabase/libsql-client-go/libsql` (hoặc `github.com/libsql/libsql-client-go` tùy thuộc vào thư viện bạn chọn).

**Giai đoạn 2: Triển khai Interface Cốt lõi của Caddy**

4.  **Định nghĩa Module Struct:**
    *   Tạo một struct (ví dụ: `LibSQLStorage`) sẽ đại diện cho module lưu trữ của bạn.
    *   Struct này sẽ chứa các trường cấu hình cần thiết như URL kết nối Turso, token xác thực, tên bảng, v.v.
5.  **Triển khai `caddy.Module`:**
    *   Cung cấp một hàm `init()` để đăng ký module với Caddy bằng cách sử dụng `caddy.RegisterModule()`.
    *   Triển khai phương thức `CaddyModule()` để trả về thông tin module (ID, hàm khởi tạo).
6.  **Triển khai `caddy.Provisioner`:**
    *   Triển khai phương thức `Provision(ctx caddy.Context) error`.
    *   Trong phương thức này, bạn sẽ thiết lập kết nối đến cơ sở dữ liệu LibSQL/Turso dựa trên cấu hình được cung cấp.
    *   Khởi tạo client LibSQL.
7.  **Triển khai `caddy.StorageConverter` (và `certmagic.Storage`):**
    *   Module của bạn cần triển khai interface `caddy.StorageConverter`. Phương thức `Convert(moduleID string, an interface{}) (certmagic.Storage, error)` sẽ trả về một instance của `certmagic.Storage`.
    *   Tạo một struct riêng (ví dụ: `TursoStorage`) triển khai interface `certmagic.Storage`. Struct này sẽ sử dụng client LibSQL đã được provision.
    *   Triển khai các phương thức của `certmagic.Storage`:
        *   `Store(key string, value []byte) error`
        *   `Load(key string) ([]byte, error)`
        *   `Delete(key string) error`
        *   `Exists(key string) bool`
        *   `List(prefix string, recursive bool) ([]string, error)`
        *   `Stat(key string) (certmagic.KeyInfo, error)`
        *   `Lock(ctx context.Context, key string) error`
        *   `Unlock(key string) error`

**Giai đoạn 3: Tương tác Cơ sở dữ liệu và Schema**

8.  **Thiết kế Schema Cơ sở dữ liệu LibSQL:**
    *   Quyết định cách bạn sẽ lưu trữ dữ liệu chứng chỉ trong Turso. Một bảng đơn giản có thể bao gồm các cột như:
        *   `key` (TEXT, PRIMARY KEY): Khóa lưu trữ của Caddy/CertMagic.
        *   `value` (BLOB): Dữ liệu được lưu trữ.
        *   `modified_at` (TIMESTAMP): Thời gian sửa đổi cuối cùng.
        *   `size` (INTEGER): Kích thước của dữ liệu.
    *   Cân nhắc việc tạo bảng tự động nếu nó chưa tồn tồn tại trong phương thức `Provision` hoặc khi kết nối lần đầu.
9.  **Triển khai Logic Tương tác LibSQL:**
    *   Viết các hàm Go sử dụng client LibSQL để thực hiện các truy vấn SQL (CREATE TABLE IF NOT EXISTS, INSERT, SELECT, DELETE, LIST) tương ứng với các phương thức của `certmagic.Storage`.
    *   Xử lý lỗi kết nối và truy vấn một cách cẩn thận.

**Giai đoạn 4: Cấu hình, Đóng gói và Kiểm thử**

10. **Xử lý Cấu hình Module:**
    *   Xác định cách người dùng sẽ cấu hình module của bạn trong Caddyfile hoặc JSON config. Ví dụ:
        ```json
        {
            "storage": {
                "module": "libsql",
                "url": "libsql://your-db.turso.io",
                "auth_token": "YOUR_TURSO_AUTH_TOKEN",
                "table_name": "caddy_storage" // Tùy chọn
            }
        }
        ```
    *   Struct `LibSQLStorage` của bạn sẽ unmarshal cấu hình này.
11. **Đăng ký Module với Caddy:**
    *   Đảm bảo module của bạn được đăng ký chính xác để Caddy có thể tìm thấy và sử dụng nó.
12. **Viết Unit Tests:**
    *   Viết các unit test cho các phương thức lưu trữ, sử dụng cơ sở dữ liệu LibSQL cục bộ (ví dụ: sử dụng `libsql-server` hoặc file-based database cho testing) hoặc mock client LibSQL.
13. **Tài liệu (Documentation):**
    *   Viết tài liệu hướng dẫn cách cài đặt, cấu hình và sử dụng module của bạn. Bao gồm các ví dụ cấu hình.
14. **Đóng gói (Build):**
    *   Bạn có thể build Caddy với module của mình bằng cách sử dụng `xcaddy`.

**Sơ đồ kiến trúc tổng quan:**

```mermaid
graph TD
    A[Caddy Core] -- Uses --> B(Storage Interface)
    B -- Implemented by --> C{Your LibSQL Module}
    C -- Implements --> D(caddy.Module)
    C -- Implements --> E(caddy.Provisioner)
    C -- Implements --> F(caddy.StorageConverter)
    F -- Returns instance of --> G(certmagic.Storage)
    C -- Configured with --> H{Turso Config (URL, Token)}
    G -- Interacts with --> I[LibSQL Go Client]
    I -- Communicates with --> J[(Turso Database)]

    subgraph Your Module
        direction LR
        D
        E
        F
        G
    end
