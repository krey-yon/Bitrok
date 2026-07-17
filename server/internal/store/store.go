package store

import (
	"context"
	"time"

	"github.com/bitrok/bitrok/pkg/api"
)

// Store defines persistence operations for tunnels and logs.
type Store interface {
	CreateTunnel(ctx context.Context, userID, name, host string, port int) (*api.Tunnel, error)
	ListTunnels(ctx context.Context, userID string) ([]api.Tunnel, error)
	GetTunnel(ctx context.Context, userID, id string) (*api.Tunnel, error)
	GetTunnelByName(ctx context.Context, userID, name string) (*api.Tunnel, error)
	GetTunnelByHost(ctx context.Context, host string) (*api.Tunnel, error)
	UpdateTunnel(ctx context.Context, userID, id string, name, host *string, port *int) (*api.Tunnel, error)
	DeleteTunnel(ctx context.Context, userID, id string) error

	LogRequest(ctx context.Context, tunnelID, method, path string, status, latencyMs, bytesIn, bytesOut int) error
	ListLogs(ctx context.Context, userID string, limit int) (*api.LogListResponse, error)

	Ping(ctx context.Context) error

	LogUptimeCheck(ctx context.Context, status, latencyMs int, errMsg string) error
	GetUptimeHistory(ctx context.Context, window time.Duration) ([]api.UptimeCheck, error)
	CleanupUptimeChecks(ctx context.Context, window time.Duration) error
}
