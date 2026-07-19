package api

import (
	"bufio"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/bitrok/bitrok/pkg/api"
	"github.com/bitrok/bitrok/server/internal/config"
)

// ctxKeyRequestID is used to retrieve the request ID from context.
type ctxKeyRequestID struct{}

func getReqID(ctx context.Context) string {
	if id, ok := ctx.Value(middleware.RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// ctxKeyUserID is used to store the authenticated user ID in context.
type ctxKeyUserID struct{}

func getUserID(ctx context.Context) string {
	if uid, ok := ctx.Value(ctxKeyUserID{}).(string); ok {
		return uid
	}
	return ""
}

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// Error writes a standard error response.
func Error(w http.ResponseWriter, status int, err string) {
	JSON(w, status, api.ErrorResponse{Error: err})
}

// AuthMiddleware validates the Bearer token against configured auth tokens.
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				slog.Warn("auth_failed", "ip", r.RemoteAddr, "path", r.URL.Path, "reason", "missing authorization header")
				Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}
			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				slog.Warn("auth_failed", "ip", r.RemoteAddr, "path", r.URL.Path, "reason", "invalid authorization header format")
				Error(w, http.StatusUnauthorized, "invalid authorization header")
				return
			}
			token := parts[1]
			valid := false
			userID := ""

			// Try static tokens first (backward compat)
			for _, t := range cfg.AuthTokens {
				if subtle.ConstantTimeCompare([]byte(t), []byte(token)) == 1 {
					valid = true
					userID = "legacy"
					break
				}
			}

			// Try JWT validation
			if !valid && cfg.JWTSecret != "" {
				var jwtOK bool
				userID, jwtOK = validateJWT(token, cfg.JWTSecret, cfg.JWTExpectedAudience, cfg.JWTExpectedIssuer)
				valid = jwtOK
			}

			if !valid {
				slog.Warn("auth_failed", "ip", r.RemoteAddr, "path", r.URL.Path, "reason", "invalid token")
				Error(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUserID{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requestLogger logs HTTP requests in structured format.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration", time.Since(start).String(),
			"remote_addr", r.RemoteAddr,
			"request_id", getReqID(r.Context()),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Hijack unwraps to the underlying Hijacker so WebSocket upgrades work.
// Without this, gorilla/websocket fails: "response does not implement http.Hijacker".
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("responseWriter: underlying ResponseWriter does not implement http.Hijacker")
	}
	return h.Hijack()
}

// Flush unwraps Flusher (SSE / streaming).
func (w *responseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap lets Go 1.20+ http.ResponseController reach the original writer.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// rateLimiter is a token-bucket rate limiter per client IP.
// It is applied only to control-plane paths (/api/*, /tunnel/*), never to
// visitor tunnel proxy traffic (that would throttle real sites behind Traefik
// where many clients share one edge IP).
type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity float64
	// tokens refilled per second = capacity / windowSeconds
	rate float64
	done chan struct{}
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
}

func newRateLimiter(capacity int, windowSeconds int) *rateLimiter {
	if capacity < 1 {
		capacity = 1
	}
	if windowSeconds < 1 {
		windowSeconds = 1
	}
	rl := &rateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: float64(capacity),
		rate:     float64(capacity) / float64(windowSeconds),
		done:     make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) Stop() {
	close(rl.done)
}

func (rl *rateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip rate limit for health, install, ACME, and all tunnel *proxy*
		// traffic (everything that is not /api/* or /tunnel/*).
		if !shouldRateLimit(path) {
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r)
		if !rl.allow(ip) {
			w.Header().Set("Retry-After", "1")
			Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// shouldRateLimit reports whether path is control-plane (CRUD / WS connect).
// Visitor traffic to tunnel hosts is path like / or /app.js — not rate limited.
func shouldRateLimit(path string) bool {
	if path == "/health" || path == "/install" || strings.HasPrefix(path, "/.well-known/") {
		return false
	}
	// Control plane only
	return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/tunnel/")
}

func clientIP(r *http.Request) string {
	// Prefer real client behind Coolify/Traefik/Cloudflare
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ip := strings.TrimSpace(strings.Split(xff, ",")[0])
		if ip != "" {
			return stripPort(ip)
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return stripPort(strings.TrimSpace(xri))
	}
	return stripPort(r.RemoteAddr)
}

func stripPort(hostport string) string {
	// [ipv6]:port or host:port
	if strings.HasPrefix(hostport, "[") {
		if i := strings.Index(hostport, "]:"); i >= 0 {
			return hostport[1:i]
		}
		return strings.Trim(hostport, "[]")
	}
	if host, _, err := net.SplitHostPort(hostport); err == nil {
		return host
	}
	return hostport
}

// allow implements a continuous token bucket: up to `capacity` tokens, refilled
// at `capacity/window` tokens per second (e.g. 600/min ≈ 10/s steady state).
func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.capacity - 1, lastRefill: now}
		return true
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens += elapsed * rl.rate
		if b.tokens > rl.capacity {
			b.tokens = rl.capacity
		}
		b.lastRefill = now
	}

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func (rl *rateLimiter) cleanup() {
	// Drop idle buckets after 10 minutes of inactivity.
	const idleTTL = 10 * time.Minute
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for k, b := range rl.buckets {
				if time.Since(b.lastRefill) > idleTTL {
					delete(rl.buckets, k)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}
