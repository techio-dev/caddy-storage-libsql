//go:build !integration

package libsqlstorage

import (
	"context"
	"testing"

	"github.com/caddyserver/caddy/v2"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func newTestStorage(t *testing.T) *LibSQLStorage {
	// Xóa file DB cũ trước mỗi test để đảm bảo sạch
	dbName := "libsql://test-naicoi92.aws-us-west-2.turso.io?authToken=eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhIjoicnciLCJpYXQiOjE3NDc4MTkzMTcsImlkIjoiMzRlODg1MGQtYzRlNC00ODJhLWI2M2EtZGVmZWI2MmZiOTRiIiwicmlkIjoiYTMwMDdkMmMtNDA0Zi00Y2FhLTk3NmQtMmExNGRhMDBjYmUxIn0.qIPCKRZbpayl1dW8K8e_JDaUQBFdqP2LqbEFY0DYqqiajIbrMJcbqv6A5EgXidhoSfbdhkaq4v5Vpqa3V9KeCA"
st := &LibSQLStorage{
URL: dbName,
}
var dummyCtx caddy.Context
if err := st.Provision(dummyCtx); err != nil {
t.Fatalf("Provision failed: %v", err)
}
return st
}

func TestProvision(t *testing.T) {
	st := newTestStorage(t)
	// Check tables exist
	_, err := st.db.Exec("SELECT 1 FROM caddy_storage")
	if err != nil {
		t.Errorf("caddy_storage table not found: %v", err)
	}
	_, err = st.db.Exec("SELECT 1 FROM caddy_resource_locks")
	if err != nil {
		t.Errorf("caddy_resource_locks table not found: %v", err)
	}
}

func TestStoreLoadExistsStatDelete(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	key := "testkey"
	val := []byte("testvalue")

	// Store
	if err := st.Store(ctx, key, val); err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	// Exists
	if !st.Exists(ctx, key) {
		t.Error("Exists should return true after Store")
	}
	// Load
	got, err := st.Load(ctx, key)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if string(got) != string(val) {
		t.Errorf("Load got %q, want %q", got, val)
	}
// Stat
info, err := st.Stat(ctx, key)
if err != nil {
t.Fatalf("Stat failed: %v", err)
}
if info.Size != int64(len(val)) {
t.Errorf("Stat size got %d, want %d", info.Size, len(val))
}
if info.Key != key {
t.Errorf("Stat key got %q, want %q", info.Key, key)
}
// Stat non-existent
_, err = st.Stat(ctx, "notfound")
if err == nil {
t.Error("Stat non-existent should return error")
} else if err.Error() != "LibSQLStorage: key not found" {
t.Errorf("Stat non-existent error = %v, want 'LibSQLStorage: key not found'", err)
}
	// Delete
	if err := st.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if st.Exists(ctx, key) {
		t.Error("Exists should return false after Delete")
	}
// Delete non-existent
if err := st.Delete(ctx, key); err == nil {
t.Error("Delete non-existent should return error")
} else if err.Error() != "LibSQLStorage: key not found" {
t.Errorf("Delete non-existent error = %v, want 'LibSQLStorage: key not found'", err)
}
}

func TestList(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	keys := []string{"foo/one", "foo/two", "foo/bar/baz"}
	for _, k := range keys {
		if err := st.Store(ctx, k, []byte("v")); err != nil {
			t.Fatalf("Store %q failed: %v", k, err)
		}
	}
	// List non-recursive
	got, err := st.List(ctx, "foo/", false)
	if err != nil {
		t.Fatalf("List non-recursive failed: %v", err)
	}
	want := []string{"foo/one", "foo/two"}
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List non-recursive missing %q", w)
		}
	}
	// List recursive
	got, err = st.List(ctx, "foo/", true)
	if err != nil {
		t.Fatalf("List recursive failed: %v", err)
	}
	for _, w := range keys {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List recursive missing %q", w)
		}
	}
}

func TestLockUnlock(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	key := "lockkey"
	// Lock
	if err := st.Lock(ctx, key); err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	// Lock again (should fail)
	if err := st.Lock(ctx, key); err == nil {
		t.Error("Lock should fail if already locked")
	}
	// Unlock
	if err := st.Unlock(ctx, key); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	// Lock again (should succeed)
	if err := st.Lock(ctx, key); err != nil {
		t.Fatalf("Lock after unlock failed: %v", err)
	}
}
