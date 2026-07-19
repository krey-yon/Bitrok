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

// rateLimiter is a simple token-bucket rate limiter per IP.
type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity int
	window   time.Duration
	done     chan struct{}
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

func newRateLimiter(capacity int, windowSeconds int) *rateLimiter {
	rl := &rateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: capacity,
		window:   time.Duration(windowSeconds) * time.Second,
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
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = strings.Split(fwd, ",")[0]
			ip = strings.TrimSpace(ip)
		}

		if !rl.allow(ip) {
			Error(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{tokens: rl.capacity - 1, lastRefill: time.Now()}
		rl.buckets[key] = b
		return true
	}

	// Refill tokens based on elapsed time
	elapsed := time.Since(b.lastRefill)
	refill := int(elapsed / rl.window)
	if refill > 0 {
		b.tokens = min(b.tokens+refill, rl.capacity)
		b.lastRefill = b.lastRefill.Add(time.Duration(refill) * rl.window)
	}

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for k, b := range rl.buckets {
				if time.Since(b.lastRefill) > rl.window*2 {
					delete(rl.buckets, k)
				}
			}
			rl.mu.Unlock()
		case <-rl.done:
			return
		}
	}
}
