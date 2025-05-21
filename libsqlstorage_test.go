//go:build !integration

package libsqlstorage

import (
	"context"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func newTestStorage(t *testing.T) *LibSQLStorage {
	// Xóa file DB cũ trước mỗi test để đảm bảo sạch
	dbName := "http://localhost:8080"
st := &LibSQLStorage{
URL: dbName,
}
var dummyCtx caddy.Context
if err := st.Provision(dummyCtx); err != nil {
t.Fatalf("Provision failed: %v", err)
}
// Truncate bảng trước mỗi test
_, _ = st.db.Exec("DELETE FROM caddy_storage")
_, _ = st.db.Exec("DELETE FROM caddy_resource_locks")
return st
}

func TestProvision(t *testing.T) {
	st := newTestStorage(t)
	_, err := st.db.Exec("SELECT 1 FROM caddy_storage")
	assert.NoError(t, err, "caddy_storage table should exist")
_, err = st.db.Exec("SELECT 1 FROM caddy_resource_locks")
assert.NoError(t, err, "caddy_resource_locks table should exist")
}

func TestStoreAndLoad(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	key := "testkey"
	val := []byte("testvalue")

	err := st.Store(ctx, key, val)
	require.NoError(t, err, "Store should not fail")

	got, err := st.Load(ctx, key)
	require.NoError(t, err, "Load should not fail")
	assert.Equal(t, val, got, "Loaded value should match stored value")
}

func TestExists(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	key := "testkey"
	val := []byte("testvalue")

	require.NoError(t, st.Store(ctx, key, val))
	assert.True(t, st.Exists(ctx, key), "Exists should return true after Store")

	require.NoError(t, st.Delete(ctx, key))
	assert.False(t, st.Exists(ctx, key), "Exists should return false after Delete")
}

func TestStat(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	key := "testkey"
	val := []byte("testvalue")

	require.NoError(t, st.Store(ctx, key, val))
	info, err := st.Stat(ctx, key)
	require.NoError(t, err, "Stat should not fail")
	assert.Equal(t, int64(len(val)), info.Size, "Stat size should match value length")
	assert.Equal(t, key, info.Key, "Stat key should match")

	_, err = st.Stat(ctx, "notfound")
	assert.Error(t, err, "Stat non-existent should return error")
	assert.EqualError(t, err, "LibSQLStorage: key not found")
}

func TestDelete(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	key := "testkey"
	val := []byte("testvalue")

	require.NoError(t, st.Store(ctx, key, val))
	require.NoError(t, st.Delete(ctx, key), "Delete should not fail")
	assert.False(t, st.Exists(ctx, key), "Exists should return false after Delete")

	err := st.Delete(ctx, key)
	assert.Error(t, err, "Delete non-existent should return error")
	assert.EqualError(t, err, "LibSQLStorage: key not found")
}

func TestList(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	keys := []string{"foo/one", "foo/two", "foo/bar/baz"}
	for _, k := range keys {
		require.NoError(t, st.Store(ctx, k, []byte("v")), "Store %q failed", k)
	}
	// List non-recursive
	got, err := st.List(ctx, "foo/", false)
	require.NoError(t, err, "List non-recursive failed")
	want := []string{"foo/one", "foo/two"}
	assert.ElementsMatch(t, want, got, "List non-recursive should match expected keys")

	// List recursive
	got, err = st.List(ctx, "foo/", true)
	require.NoError(t, err, "List recursive failed")
	assert.ElementsMatch(t, keys, got, "List recursive should match all keys")
}

func TestLockUnlock(t *testing.T) {
	st := newTestStorage(t)
	ctx := context.Background()
	key := "lockkey"

	require.NoError(t, st.Lock(ctx, key), "Lock should not fail")
	err := st.Lock(ctx, key)
	assert.Error(t, err, "Lock should fail if already locked")

	require.NoError(t, st.Unlock(ctx, key), "Unlock should not fail")
	require.NoError(t, st.Lock(ctx, key), "Lock after unlock should succeed")
}
