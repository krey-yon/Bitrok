package api

import "time"

// Tunnel represents a persisted tunnel configuration.
type Tunnel struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TunnelCreateRequest is the payload for POST /api/tunnels.
type TunnelCreateRequest struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int    `json:"port"`
}

// TunnelUpdateRequest is the payload for PATCH /api/tunnels/:id.
type TunnelUpdateRequest struct {
	Name *string `json:"name,omitempty"`
	Host *string `json:"host,omitempty"`
	Port *int    `json:"port,omitempty"`
}

// TunnelListResponse wraps a list of tunnels.
type TunnelListResponse struct {
	Tunnels []Tunnel `json:"tunnels"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HealthResponse is returned by GET /health.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// TunnelLog records a single proxied request, as persisted by the server.
// ID is the autoincrement row id (SQLite). TunnelName is joined from tunnels
// for display convenience and is only populated by read/list queries.
type TunnelLog struct {
	ID         int64     `json:"id"`
	TunnelID   string    `json:"tunnel_id"`
	TunnelName string    `json:"tunnel_name,omitempty"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Status     int       `json:"status"`
	LatencyMs  int       `json:"latency_ms"`
	BytesIn    int       `json:"bytes_in"`
	BytesOut   int       `json:"bytes_out"`
	TS         time.Time `json:"ts"`
}

// LogListResponse is returned by GET /api/logs. Total is the user-scoped
// count of all request logs (for the dashboard KPI); Logs is the bounded
// recent slice (ordered by ts DESC).
type LogListResponse struct {
	Total int         `json:"total"`
	Logs  []TunnelLog `json:"logs"`
}

// UptimeCheck records a single health-check result.
type UptimeCheck struct {
	TS         time.Time `json:"ts"`
	Status     int       `json:"status"`
	LatencyMs  int       `json:"latency_ms"`
	Error      string    `json:"error,omitempty"`
}

// UptimeBucket aggregates checks into a time window.
type UptimeBucket struct {
	Hour            time.Time `json:"hour"`
	Checks          int       `json:"checks"`
	Up              int       `json:"up"`
	Down            int       `json:"down"`
	AvgLatencyMs    int       `json:"avg_latency_ms"`
	UptimePercent   float64   `json:"uptime_percent"`
}

// UptimeResponse is returned by GET /api/uptime.
type UptimeResponse struct {
	WindowHours          int            `json:"window_hours"`
	OverallUptimePercent float64        `json:"overall_uptime_percent"`
	Buckets              []UptimeBucket `json:"buckets"`
}
