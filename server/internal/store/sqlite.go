package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"

	"github.com/bitrok/bitrok/pkg/api"
)

//go:embed migrations/001_init.sql
var migration001 string

//go:embed migrations/002_uptime.sql
var migration002 string

//go:embed migrations/003_auth.sql
var migration003 string

//go:embed migrations/004_tunnel_name_scope.sql
var migration004 string

//go:embed migrations/005_drop_self_uptime.sql
var migration005 string

// SQLite implements Store using SQLite.
type SQLite struct {
	db *sql.DB
}

// NewSQLite opens the database, runs migrations, and returns a Store.
func NewSQLite(dbPath string, maxOpenConns, maxIdleConns int, connMaxLifetime, connMaxIdleTime time.Duration) (*SQLite, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := os.Chmod(dbPath, 0600); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("secure database permissions: %w", err)
	}
	// SQLite serializes writes; keep pool small
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	s := &SQLite{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLite) migrate() error {
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	migrations := []struct {
		version int
		sql     string
	}{
		{1, migration001},
		{2, migration002},
		{3, migration003},
		{4, migration004},
		{5, migration005},
	}

	for _, m := range migrations {
		var exists int
		err := s.db.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, m.version).Scan(&exists)
		if err == nil {
			continue // already applied
		}
		if err != sql.ErrNoRows {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}

		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec(m.sql); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, m.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}
	return nil
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "tun_" + hex.EncodeToString(b), nil
}

// CreateTunnel inserts a new tunnel.
func (s *SQLite) CreateTunnel(ctx context.Context, userID, name, host string, port int) (*api.Tunnel, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO tunnels (id, user_id, name, host, port, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, userID, name, host, port, now, now,
	)
	if err != nil {
		return nil, classifyWriteError(err)
	}
	return s.GetTunnel(ctx, userID, id)
}

// ListTunnels returns all tunnels ordered by creation time for a user.
func (s *SQLite) ListTunnels(ctx context.Context, userID string) ([]api.Tunnel, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, name, host, port, created_at, updated_at FROM tunnels WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tunnels []api.Tunnel
	for rows.Next() {
		var t api.Tunnel
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Host, &t.Port, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tunnels = append(tunnels, t)
	}
	return tunnels, rows.Err()
}

// GetTunnel fetches a tunnel by ID scoped to a user.
func (s *SQLite) GetTunnel(ctx context.Context, userID, id string) (*api.Tunnel, error) {
	var t api.Tunnel
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, host, port, created_at, updated_at FROM tunnels WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Host, &t.Port, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// GetTunnelByName fetches a tunnel by name scoped to a user.
func (s *SQLite) GetTunnelByName(ctx context.Context, userID, name string) (*api.Tunnel, error) {
	var t api.Tunnel
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, host, port, created_at, updated_at FROM tunnels WHERE name = ? AND user_id = ?`, name, userID,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Host, &t.Port, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// GetTunnelByHost fetches a tunnel by host.
func (s *SQLite) GetTunnelByHost(ctx context.Context, host string) (*api.Tunnel, error) {
	var t api.Tunnel
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, host, port, created_at, updated_at FROM tunnels WHERE host = ?`, host,
	).Scan(&t.ID, &t.UserID, &t.Name, &t.Host, &t.Port, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

// UpdateTunnel modifies a tunnel's fields.
func (s *SQLite) UpdateTunnel(ctx context.Context, userID, id string, name, host *string, port *int) (*api.Tunnel, error) {
	var sets []string
	var args []any
	if name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *name)
	}
	if host != nil {
		sets = append(sets, "host = ?")
		args = append(args, *host)
	}
	if port != nil {
		sets = append(sets, "port = ?")
		args = append(args, *port)
	}
	if len(sets) == 0 {
		return s.GetTunnel(ctx, userID, id)
	}

	args = append(args, time.Now().UTC(), id, userID)
	query := "UPDATE tunnels SET " + strings.Join(sets, ", ") + ", updated_at = ? WHERE id = ? AND user_id = ?"
	if _, err := s.db.ExecContext(ctx, query, args...); err != nil {
		return nil, classifyWriteError(err)
	}
	return s.GetTunnel(ctx, userID, id)
}

// DeleteTunnel removes a tunnel and its logs scoped to a user.
func (s *SQLite) DeleteTunnel(ctx context.Context, userID, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Scope the compatibility cleanup through the owned tunnel row. This keeps
	// the store invariant safe even when called outside the HTTP handler.
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM tunnel_logs
		WHERE tunnel_id IN (SELECT id FROM tunnels WHERE id = ? AND user_id = ?)`, id, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM tunnels WHERE id = ? AND user_id = ?`, id, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// LogRequest records a proxied request.
func (s *SQLite) LogRequest(ctx context.Context, tunnelID, method, path string, status, latencyMs, bytesIn, bytesOut int) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO tunnel_logs (tunnel_id, method, path, status, latency_ms, bytes_in, bytes_out) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		tunnelID, method, path, status, latencyMs, bytesIn, bytesOut,
	)
	return err
}

// ListLogs returns a bounded recent slice of request logs plus the user-scoped
// total count. Results are ordered newest-first. limit is clamped to [1, 200].
// The query joins tunnels so TunnelName is populated and rows are scoped by
// tunnels.user_id (so a user can never see another user's traffic).
func (s *SQLite) ListLogs(ctx context.Context, userID string, limit int) (*api.LogListResponse, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}

	var total int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tunnel_logs l JOIN tunnels t ON t.id = l.tunnel_id WHERE t.user_id = ?`, userID,
	).Scan(&total); err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT l.id, l.tunnel_id, t.name, l.method, l.path, l.status, l.latency_ms, l.bytes_in, l.bytes_out, l.ts
		FROM tunnel_logs l
		JOIN tunnels t ON t.id = l.tunnel_id
		WHERE t.user_id = ?
		ORDER BY l.ts DESC
		LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]api.TunnelLog, 0, limit)
	for rows.Next() {
		var l api.TunnelLog
		if err := rows.Scan(&l.ID, &l.TunnelID, &l.TunnelName, &l.Method, &l.Path, &l.Status, &l.LatencyMs, &l.BytesIn, &l.BytesOut, &l.TS); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &api.LogListResponse{Total: total, Logs: logs}, nil
}

// CleanupTunnelLogs bounds disk usage by deleting request logs older than the
// configured retention window.
func (s *SQLite) CleanupTunnelLogs(ctx context.Context, window time.Duration) error {
	cutoff := time.Now().UTC().Add(-window)
	_, err := s.db.ExecContext(ctx, `DELETE FROM tunnel_logs WHERE ts < ?`, cutoff)
	return err
}

// Ping checks database connectivity.
func (s *SQLite) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the underlying database.
func (s *SQLite) Close() error {
	return s.db.Close()
}

func classifyWriteError(err error) error {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
		return fmt.Errorf("%w: %v", ErrConflict, err)
	}
	return err
}
