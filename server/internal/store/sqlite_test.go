package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func newTestSQLite(t *testing.T) *SQLite {
	t.Helper()
	db, err := NewSQLite(filepath.Join(t.TempDir(), "bitrok.db"), 1, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestDeleteTunnelDoesNotDeleteAnotherUsersLogs(t *testing.T) {
	ctx := context.Background()
	db := newTestSQLite(t)
	tunnel, err := db.CreateTunnel(ctx, "user-b", "app-b", "app-b.bitrok.test", 3000)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.LogRequest(ctx, tunnel.ID, "GET", "/", 200, 1, 0, 10); err != nil {
		t.Fatal(err)
	}

	if err := db.DeleteTunnel(ctx, "user-a", tunnel.ID); err != nil {
		t.Fatal(err)
	}
	logs, err := db.ListLogs(ctx, "user-b", 10)
	if err != nil {
		t.Fatal(err)
	}
	if logs.Total != 1 {
		t.Fatalf("other user's logs were modified: total=%d", logs.Total)
	}
}

func TestCreateTunnelReturnsConflictSentinel(t *testing.T) {
	ctx := context.Background()
	db := newTestSQLite(t)
	if _, err := db.CreateTunnel(ctx, "user-a", "app-a", "shared.bitrok.test", 3000); err != nil {
		t.Fatal(err)
	}
	_, err := db.CreateTunnel(ctx, "user-b", "app-b", "shared.bitrok.test", 3001)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate host error = %v, want ErrConflict", err)
	}
}

func TestTunnelNamesAreScopedToUsers(t *testing.T) {
	ctx := context.Background()
	db := newTestSQLite(t)
	if _, err := db.CreateTunnel(ctx, "user-a", "app", "app-user-a.bitrok.test", 3000); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateTunnel(ctx, "user-b", "app", "app-user-b.bitrok.test", 3001); err != nil {
		t.Fatalf("same app name should be allowed for another user: %v", err)
	}
	_, err := db.CreateTunnel(ctx, "user-a", "app", "another-user-a.bitrok.test", 3002)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate name for one user error = %v, want ErrConflict", err)
	}
}

func TestDatabaseFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode bits are not applicable")
	}
	path := filepath.Join(t.TempDir(), "bitrok.db")
	db, err := NewSQLite(path, 1, 1, time.Minute, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("database mode = %o, want 600", got)
	}
}
