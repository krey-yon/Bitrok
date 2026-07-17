package api

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"

	"github.com/bitrok/bitrok/server/internal/config"
	"github.com/bitrok/bitrok/server/internal/relay"
	"github.com/bitrok/bitrok/server/internal/store"

	bitrokapi "github.com/bitrok/bitrok/pkg/api"
)

// Version is set at build time via ldflags.
var Version = "dev"

// NewRouter wires all handlers and middleware.
func NewRouter(cfg *config.Config, st store.Store, hub *relay.Hub) (*chi.Mux, *rateLimiter) {
	r := chi.NewRouter()

	// Recovery first so panics are caught and logged
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(requestLogger)

	// Rate limiter per IP for public endpoints
	rateLimiter := newRateLimiter(cfg.RateLimitCapacity, cfg.RateLimitWindow)
	r.Use(rateLimiter.Middleware)

	// Proxy middleware: intercepts non-API traffic and routes through active tunnels
	proxy := relay.NewProxyHandler(hub, st)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/health" || strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/tunnel/") || strings.HasPrefix(path, "/.well-known/") {
				next.ServeHTTP(w, r)
				return
			}
			proxy.ServeHTTP(w, r)
		})
	})

	// Public health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := st.Ping(ctx); err != nil {
			slog.Error("health check db ping failed", "error", err)
			JSON(w, http.StatusServiceUnavailable, bitrokapi.HealthResponse{Status: "unhealthy", Version: Version})
			return
		}
		JSON(w, http.StatusOK, bitrokapi.HealthResponse{Status: "ok", Version: Version})
	})

	// Public uptime status (for status page / dashboard)
	r.Get("/api/uptime", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		checks, err := st.GetUptimeHistory(ctx, 24*time.Hour)
		if err != nil {
			slog.Error("uptime query failed", "error", err)
			Error(w, http.StatusInternalServerError, "database error")
			return
		}

		resp := bucketUptime(checks)
		JSON(w, http.StatusOK, resp)
	})

	// Authenticated API + WS
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(cfg))

		th := &TunnelHandler{Store: st, Hub: hub, Config: cfg}
		th.Register(r)

		lh := &LogHandler{Store: st}
		lh.Register(r)

		// WebSocket tunnel endpoint
		r.Get("/tunnel/{id}/connect", WSConnectHandler(cfg, st, hub))
	})

	return r, rateLimiter
}

// bucketUptime aggregates raw checks into 1-hour buckets.
func bucketUptime(checks []bitrokapi.UptimeCheck) bitrokapi.UptimeResponse {
	if len(checks) == 0 {
		return bitrokapi.UptimeResponse{WindowHours: 24, OverallUptimePercent: 0, Buckets: []bitrokapi.UptimeBucket{}}
	}

	// Group by hour
	buckets := make(map[int64]*bitrokapi.UptimeBucket)
	var totalUp, totalChecks int

	for _, c := range checks {
		hour := c.TS.Truncate(time.Hour).Unix()
		b, ok := buckets[hour]
		if !ok {
			b = &bitrokapi.UptimeBucket{Hour: c.TS.Truncate(time.Hour)}
			buckets[hour] = b
		}
		b.Checks++
		if c.Status == http.StatusOK {
			b.Up++
			totalUp++
		} else {
			b.Down++
		}
		b.AvgLatencyMs += c.LatencyMs
		totalChecks++
	}

	// Convert map to sorted slice
	var hours []int64
	for h := range buckets {
		hours = append(hours, h)
	}
	sort.Slice(hours, func(i, j int) bool { return hours[i] < hours[j] })

	var out []bitrokapi.UptimeBucket
	for _, h := range hours {
		b := buckets[h]
		if b.Checks > 0 {
			b.AvgLatencyMs = b.AvgLatencyMs / b.Checks
			b.UptimePercent = float64(b.Up) / float64(b.Checks) * 100
		}
		out = append(out, *b)
	}

	overall := 0.0
	if totalChecks > 0 {
		overall = float64(totalUp) / float64(totalChecks) * 100
	}

	return bitrokapi.UptimeResponse{
		WindowHours:          24,
		OverallUptimePercent: overall,
		Buckets:              out,
	}
}
