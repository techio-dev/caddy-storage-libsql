// Package libsqlstorage implements a Caddy storage module using LibSQL (Turso) as backend.
package libsqlstorage

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/certmagic"
)

func init() {
	caddy.RegisterModule(LibSQLStorage{})
}

// LibSQLStorage implements Caddy module and storage converter.
type LibSQLStorage struct {
	// Configurable fields
URL            string `json:"url,omitempty"`
LockTTLSeconds int    `json:"lock_ttl,omitempty"` // TTL cho lock (giây), mặc định 60

// Internal DB client
	db *sql.DB
}

// CaddyModule returns the Caddy module information.
func (LibSQLStorage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.storage.libsql",
		New: func() caddy.Module { return new(LibSQLStorage) },
	}
}

 // Provision sets up the LibSQL client and ensures DB schema.
func (s *LibSQLStorage) Provision(ctx caddy.Context) error {
	if s.URL == "" {
		return fmt.Errorf("LibSQLStorage: missing URL")
	}
    dsn := s.URL
    var err error
    s.db, err = sql.Open("libsql", dsn)
    if err != nil {
        return fmt.Errorf("LibSQLStorage: failed to open libsql DB: %w", err)
    }

createTable := `
CREATE TABLE IF NOT EXISTS caddy_storage (
key TEXT PRIMARY KEY,
value BLOB,
modified_at TIMESTAMP,
size INTEGER
)`
_, err = s.db.ExecContext(context.Background(), createTable)
if err != nil {
return fmt.Errorf("LibSQLStorage: failed to create table: %w", err)
}

    // Tạo bảng caddy_resource_locks nếu chưa có
    createLockTable := `
CREATE TABLE IF NOT EXISTS caddy_resource_locks (
key TEXT PRIMARY KEY,
expire_at TIMESTAMP
)`
    _, err = s.db.ExecContext(context.Background(), createLockTable)
    if err != nil {
        return fmt.Errorf("LibSQLStorage: failed to create lock table: %w", err)
    }
	return nil
}

// certmagic.Storage interface methods (scaffold)
func (s *LibSQLStorage) Store(ctx context.Context, key string, value []byte) error {
	now := fmt.Sprintf("%d", ctx.Value("now"))
	if now == "" {
		now = fmt.Sprintf("%d", int64(0))
	}
_, err := s.db.ExecContext(
ctx,
"INSERT OR REPLACE INTO caddy_storage (key, value, modified_at, size) VALUES (?, ?, CURRENT_TIMESTAMP, ?)",
key, value, len(value),
)
	if err != nil {
		return fmt.Errorf("LibSQLStorage: Store failed: %w", err)
	}
	return nil
}

func (s *LibSQLStorage) Load(ctx context.Context, key string) ([]byte, error) {
    var value []byte
    err := s.db.QueryRowContext(
        ctx,
        "SELECT value FROM caddy_storage WHERE key = ?",
        key,
    ).Scan(&value)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("LibSQLStorage: key not found")
    }
    if err != nil {
        return nil, fmt.Errorf("LibSQLStorage: Load failed: %w", err)
    }
    return value, nil
}

func (s *LibSQLStorage) Delete(ctx context.Context, key string) error {
    res, err := s.db.ExecContext(
        ctx,
        "DELETE FROM caddy_storage WHERE key = ?",
        key,
    )
    if err != nil {
        return fmt.Errorf("LibSQLStorage: Delete failed: %w", err)
    }
    n, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("LibSQLStorage: Delete rows affected error: %w", err)
    }
    if n == 0 {
        return fmt.Errorf("LibSQLStorage: key not found")
    }
    return nil
}

func (s *LibSQLStorage) Exists(ctx context.Context, key string) bool {
    var exists bool
    err := s.db.QueryRowContext(
        ctx,
        "SELECT EXISTS(SELECT 1 FROM caddy_storage WHERE key = ?)",
        key,
    ).Scan(&exists)
    return err == nil && exists
}

func (s *LibSQLStorage) List(ctx context.Context, prefix string, recursive bool) ([]string, error) {
    var rows *sql.Rows
    var err error
    if recursive {
        rows, err = s.db.QueryContext(
            ctx,
            "SELECT key FROM caddy_storage WHERE key LIKE ?",
            prefix+"%",
        )
    } else {
        rows, err = s.db.QueryContext(
            ctx,
            "SELECT key FROM caddy_storage WHERE key LIKE ?",
            prefix+"%",
        )
    }
    if err != nil {
        return nil, fmt.Errorf("LibSQLStorage: List failed: %w", err)
    }
    defer rows.Close()
    var keys []string
    for rows.Next() {
        var key string
        if err := rows.Scan(&key); err != nil {
            return nil, fmt.Errorf("LibSQLStorage: List scan failed: %w", err)
        }
        if !recursive {
            remain := key[len(prefix):]
            if len(remain) == 0 {
                continue
            }
            if idx := indexOf(remain, "/"); idx >= 0 {
                continue // Bỏ qua nếu có dấu / sau prefix
            }
        }
        keys = append(keys, key)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("LibSQLStorage: List rows error: %w", err)
    }
    return keys, nil
}

// indexOf trả về vị trí đầu tiên của sep trong s, -1 nếu không tìm thấy
func indexOf(s, sep string) int {
    for i := 0; i+len(sep) <= len(s); i++ {
        if s[i:i+len(sep)] == sep {
            return i
        }
    }
    return -1
}

func (s *LibSQLStorage) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
    var size int64
    var modifiedAtStr string
    err := s.db.QueryRowContext(
        ctx,
        "SELECT size, modified_at FROM caddy_storage WHERE key = ?",
        key,
    ).Scan(&size, &modifiedAtStr)
    if err == sql.ErrNoRows {
        return certmagic.KeyInfo{}, fmt.Errorf("LibSQLStorage: key not found")
    }
    if err != nil {
        return certmagic.KeyInfo{}, fmt.Errorf("LibSQLStorage: Stat failed: %w", err)
    }
    var modifiedAt time.Time
    if modifiedAtStr != "" {
        modifiedAt, err = time.Parse("2006-01-02 15:04:05", modifiedAtStr)
        if err != nil {
            modifiedAt = time.Time{}
        }
    }
    return certmagic.KeyInfo{
        Key:        key,
        Size:       size,
        Modified:   modifiedAt,
        IsTerminal: true,
    }, nil
}

func (s *LibSQLStorage) Lock(ctx context.Context, key string) error {
	ttl := s.LockTTLSeconds
	if ttl <= 0 {
		ttl = 60
	}
	// Xóa các lock đã hết hạn
    _, _ = s.db.ExecContext(ctx, "DELETE FROM caddy_resource_locks WHERE expire_at <= CURRENT_TIMESTAMP")

    // Tính expire_at
    expireAt := time.Now().Add(time.Duration(ttl) * time.Second).Format("2006-01-02 15:04:05")
    // Cố gắng insert lock mới
    _, err := s.db.ExecContext(ctx,
        "INSERT INTO caddy_resource_locks (key, expire_at) VALUES (?, ?)",
        key, expireAt,
    )
    if err != nil {
        // Nếu đã tồn tại, kiểm tra lock còn hạn không
        var dbExpire string
        row := s.db.QueryRowContext(ctx, "SELECT expire_at FROM caddy_resource_locks WHERE key = ?", key)
        if err2 := row.Scan(&dbExpire); err2 == nil {
            // Thử parse theo RFC3339 trước, nếu lỗi thì thử định dạng cũ
            t, err3 := time.Parse(time.RFC3339, dbExpire)
            if err3 != nil {
                t, err3 = time.Parse("2006-01-02 15:04:05", dbExpire)
            }
            if err3 == nil && t.After(time.Now().UTC()) {
                return fmt.Errorf("LibSQLStorage: key is locked")
            }
            // Nếu lock đã hết hạn, xóa và thử insert lại
            _, _ = s.db.ExecContext(ctx, "DELETE FROM caddy_resource_locks WHERE key = ?", key)
            _, err4 := s.db.ExecContext(ctx,
                "INSERT INTO caddy_resource_locks (key, expire_at) VALUES (?, ?)",
                key, expireAt,
            )
            if err4 != nil {
                return fmt.Errorf("LibSQLStorage: lock failed after cleanup: %w", err4)
            }
            return nil
        }
        return fmt.Errorf("LibSQLStorage: lock failed: %w", err)
    }
    return nil
}

func (s *LibSQLStorage) Unlock(ctx context.Context, key string) error {
    _, _ = s.db.ExecContext(ctx, "DELETE FROM caddy_resource_locks WHERE key = ?", key)
    // Luôn trả về thành công
    return nil
}

var (
_ caddy.Module      = (*LibSQLStorage)(nil)
_ caddy.Provisioner = (*LibSQLStorage)(nil)
_ certmagic.Storage = (*LibSQLStorage)(nil)
)
