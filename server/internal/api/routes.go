package api

import (
	"context"
	_ "embed"
	"net/http"
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

//go:embed install.sh
var installScript string

//go:embed install.ps1
var installPowerShellScript string

// Version is set at build time via ldflags.
var Version = "dev"

// NewRouter wires all handlers and middleware.
func NewRouter(cfg *config.Config, st store.Store, hub *relay.Hub) (*chi.Mux, *rateLimiter) {
	r := chi.NewRouter()

	// Recovery first so panics are caught and logged
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(requestLogger)

	// Rate limiter: control-plane only (/api/*, /tunnel/*). Tunnel visitor
	// proxy traffic is not limited (would break real sites behind Traefik).
	rateLimiter := newRateLimiter(cfg.RateLimitCapacity, cfg.RateLimitWindow, cfg.TrustProxyHeaders)
	r.Use(rateLimiter.Middleware)

	// Proxy middleware: intercepts non-API traffic and routes through active tunnels
	proxy := relay.NewProxyHandler(hub, st, cfg.TrustProxyHeaders)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if path == "/health" || path == "/install" || path == "/install.ps1" || strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/tunnel/") || strings.HasPrefix(path, "/.well-known/") {
				next.ServeHTTP(w, r)
				return
			}
			proxy.ServeHTTP(w, r)
		})
	})

	// Public install script: curl -fsSL https://bitrok.tech/install | sh
	r.Get("/install", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=\"install.sh\"")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write([]byte(installScript))
	})
	r.Get("/install.ps1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=\"install.ps1\"")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write([]byte(installPowerShellScript))
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
