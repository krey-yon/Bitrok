package relay

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bitrok/bitrok/pkg/protocol"
	"github.com/bitrok/bitrok/server/internal/store"
)

var reqIDCounter atomic.Uint64

// ProxyHandler handles incoming HTTP requests and routes them through WS sessions.
type ProxyHandler struct {
	Hub   *Hub
	Store store.Store
}

// NewProxyHandler creates the proxy handler.
func NewProxyHandler(hub *Hub, st store.Store) *ProxyHandler {
	return &ProxyHandler{Hub: hub, Store: st}
}

// ServeHTTP implements the http.Handler interface.
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Behind Coolify/Traefik, r.Host is often the *service* domain (api.bitrok.tech)
	// while the public tunnel hostname is only in X-Forwarded-Host.
	// Prefer forwarded host so GetTunnelByHost matches the reserved tunnel URL.
	host := resolvePublicHost(r)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	tun, err := p.Store.GetTunnelByHost(ctx, host)
	if err != nil {
		slog.Error("proxy db error", "host", host, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tun == nil {
		slog.Warn("proxy tunnel not found",
			"host", host,
			"r_host", r.Host,
			"x_forwarded_host", r.Header.Get("X-Forwarded-Host"),
			"path", r.URL.Path,
		)
		http.Error(w, "tunnel not found", http.StatusNotFound)
		return
	}

	session := p.Hub.Get(tun.ID)
	if session == nil {
		slog.Warn("proxy tunnel not active", "host", host, "tunnel_id", tun.ID, "path", r.URL.Path)
		http.Error(w, "tunnel not active", http.StatusServiceUnavailable)
		return
	}

	// Serialize HTTP request to WS frame
	reqID := generateReqID()
	bodyBytes, err := readBody(r)
	if err != nil {
		slog.Warn("proxy request body too large", "host", host, "error", err)
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	headers := flattenHeaders(r.Header)
	// Strip hop-by-hop headers per RFC 2616 Section 13.5.1
	hopByHop := map[string]bool{
		"Connection": true, "Keep-Alive": true, "Proxy-Authenticate": true,
		"Proxy-Authorization": true, "TE": true, "Trailer": true,
		"Transfer-Encoding": true, "Upgrade": true,
	}
	for k := range hopByHop {
		delete(headers, k)
	}
	if connHdr := headers["Connection"]; connHdr != "" {
		for _, k := range strings.Split(connHdr, ",") {
			delete(headers, strings.TrimSpace(k))
		}
	}
	// Add forwarding headers so the backend sees the original client
	if fwd := headers["X-Forwarded-For"]; fwd != "" {
		headers["X-Forwarded-For"] = fwd + ", " + r.RemoteAddr
	} else {
		headers["X-Forwarded-For"] = r.RemoteAddr
	}
	// Respect X-Forwarded-Proto from the trusted reverse proxy (Coolify/Traefik/nginx).
	// They terminate TLS and set this header to the original client protocol.
	// Only fall back to r.TLS when no proxy is in front (direct connection / dev),
	// otherwise backends doing HTTP→HTTPS redirects would infinite-loop.
	if proto := headers["X-Forwarded-Proto"]; proto == "" {
		if r.TLS != nil {
			headers["X-Forwarded-Proto"] = "https"
		} else {
			headers["X-Forwarded-Proto"] = "http"
		}
	}
	// Always present the *public* tunnel host to the local app, not the
	// internal service name Traefik may have put in r.Host.
	headers["X-Forwarded-Host"] = host
	headers["Host"] = host

	frame := protocol.ProxyRequest{
		Type:    string(protocol.TypeRequest),
		ReqID:   reqID,
		Method:  r.Method,
		Path:    r.URL.RequestURI(),
		Host:    host,
		Headers: headers,
		BodyB64: base64.StdEncoding.EncodeToString(bodyBytes),
	}

	start := time.Now()
	bytesIn := len(bodyBytes)

	// Set up response waiter with context-aware cleanup
	respChan := make(chan protocol.ProxyResponse, 1)
	waitersMu.Lock()
	waiters[reqID] = respChan
	waitersMu.Unlock()

	// Ensure cleanup happens even if we time out or panic
	defer func() {
		waitersMu.Lock()
		delete(waiters, reqID)
		waitersMu.Unlock()
	}()

	if err := session.WriteJSON(frame); err != nil {
		slog.Error("proxy websocket write failed", "tunnel_id", tun.ID, "error", err)
		http.Error(w, "relay error", http.StatusBadGateway)
		_ = p.Store.LogRequest(ctx, tun.ID, r.Method, r.URL.Path, http.StatusBadGateway, int(time.Since(start).Milliseconds()), bytesIn, 0)
		return
	}

	slog.Debug("proxy request forwarded", "req_id", reqID, "tunnel_id", tun.ID, "method", r.Method, "path", r.URL.Path)

	// Wait for response
	select {
	case resp := <-respChan:
		latencyMs := int(time.Since(start).Milliseconds())
		for k, v := range resp.Headers {
			// Set-Cookie values are joined by newlines (preserving multi-value semantics)
			// Re-wrap them as individual header entries
			if k == "Set-Cookie" {
				for _, cookie := range strings.Split(v, "\n") {
					cookie = strings.TrimSpace(cookie)
					if cookie != "" {
						w.Header().Add("Set-Cookie", cookie)
					}
				}
			} else {
				w.Header().Set(k, v)
			}
		}
		w.WriteHeader(resp.Status)
		body, err := base64.StdEncoding.DecodeString(resp.BodyB64)
		if err != nil {
			slog.Error("proxy invalid response body", "req_id", reqID, "error", err)
			http.Error(w, "invalid relay response", http.StatusBadGateway)
			_ = p.Store.LogRequest(ctx, tun.ID, r.Method, r.URL.Path, http.StatusBadGateway, latencyMs, bytesIn, 0)
			return
		}
		w.Write(body)
		_ = p.Store.LogRequest(ctx, tun.ID, r.Method, r.URL.Path, resp.Status, latencyMs, bytesIn, len(body))
	case <-ctx.Done():
		latencyMs := int(time.Since(start).Milliseconds())
		slog.Warn("proxy gateway timeout", "req_id", reqID, "tunnel_id", tun.ID)
		http.Error(w, "gateway timeout", http.StatusGatewayTimeout)
		_ = p.Store.LogRequest(ctx, tun.ID, r.Method, r.URL.Path, http.StatusGatewayTimeout, latencyMs, bytesIn, 0)
	}
}

var (
	waitersMu sync.Mutex
	waiters   = make(map[string]chan protocol.ProxyResponse)
)

// HandleResponse routes a proxy response frame to the waiting HTTP handler.
func HandleResponse(resp protocol.ProxyResponse) {
	waitersMu.Lock()
	ch, ok := waiters[resp.ReqID]
	waitersMu.Unlock()
	if ok {
		ch <- resp
	}
}

func generateReqID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), reqIDCounter.Add(1))
}

// resolvePublicHost returns the hostname the visitor used (for tunnel lookup).
//
// Priority:
//  1. X-Forwarded-Host (first value) — set by Coolify/Traefik/Cloudflare
//  2. r.Host — direct connections / when the proxy preserves the original host
//
// Strips :port and lowercases so DB host matches are stable.
func resolvePublicHost(r *http.Request) string {
	raw := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if raw != "" {
		// "a.example.com, b.example.com" → first hop
		if i := strings.IndexByte(raw, ','); i >= 0 {
			raw = strings.TrimSpace(raw[:i])
		}
	}
	if raw == "" {
		raw = r.Host
	}
	raw = stripHostPort(raw)
	return strings.ToLower(strings.TrimSpace(raw))
}

// stripHostPort removes a trailing :port from host, leaving IPv6 [addr] alone.
func stripHostPort(host string) string {
	// bracketed IPv6: [2001:db8::1]:443
	if strings.HasPrefix(host, "[") {
		if end := strings.IndexByte(host, ']'); end >= 0 {
			return host[:end+1]
		}
		return host
	}
	// hostname:port or ipv4:port
	if i := strings.LastIndexByte(host, ':'); i >= 0 {
		port := host[i+1:]
		if port != "" && isAllDigits(port) {
			return host[:i]
		}
	}
	return host
}

func isAllDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			out[k] = strings.Join(v, ", ")
		}
	}
	return out
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	const maxBody = 10 * 1024 * 1024
	body, err := io.ReadAll(io.LimitReader(r.Body, int64(maxBody)+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBody {
		return nil, fmt.Errorf("body too large")
	}
	return body, nil
}
