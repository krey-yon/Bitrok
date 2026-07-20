package relay

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bitrok/bitrok/pkg/protocol"
	"github.com/bitrok/bitrok/server/internal/store"
)

var reqIDCounter atomic.Uint64

const inactiveTunnelMessage = "This tunnel is not active yet. Ask the tunnel owner to start or refresh it from the Bitrok CLI."

// ProxyHandler handles incoming HTTP requests and routes them through WS sessions.
type ProxyHandler struct {
	Hub               *Hub
	Store             store.Store
	TrustProxyHeaders bool
}

// NewProxyHandler creates the proxy handler.
func NewProxyHandler(hub *Hub, st store.Store, trustProxyHeaders ...bool) *ProxyHandler {
	trust := false
	if len(trustProxyHeaders) > 0 {
		trust = trustProxyHeaders[0]
	}
	return &ProxyHandler{Hub: hub, Store: st, TrustProxyHeaders: trust}
}

// ServeHTTP implements the http.Handler interface.
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Behind Coolify/Traefik, r.Host is often the *service* domain (api.bitrok.tech)
	// while the public tunnel hostname is only in X-Forwarded-Host.
	// Prefer forwarded host so GetTunnelByHost matches the reserved tunnel URL.
	host := resolvePublicHost(r, p.TrustProxyHeaders)

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
		http.Error(w, inactiveTunnelMessage, http.StatusServiceUnavailable)
		return
	}

	session := p.Hub.Get(tun.ID)
	if session == nil {
		slog.Warn("proxy tunnel not active", "host", host, "tunnel_id", tun.ID, "path", r.URL.Path)
		http.Error(w, inactiveTunnelMessage, http.StatusServiceUnavailable)
		return
	}
	if !session.TryAcquireRequest() {
		http.Error(w, "too many concurrent requests for this tunnel", http.StatusServiceUnavailable)
		return
	}
	defer session.ReleaseRequest()

	// Serialize HTTP request to WS frame
	reqID := generateReqID()
	bodyBytes, err := readBody(r)
	if err != nil {
		slog.Warn("proxy request body too large", "host", host, "error", err)
		http.Error(w, "request too large", http.StatusRequestEntityTooLarge)
		return
	}

	headers := flattenHeaders(r.Header)
	if !p.TrustProxyHeaders {
		delete(headers, "X-Forwarded-For")
		delete(headers, "X-Forwarded-Host")
		delete(headers, "X-Forwarded-Proto")
	}
	// Strip hop-by-hop headers, including names nominated by Connection.
	connectionHeader := headers["Connection"]
	if connectionHeader != "" {
		for _, k := range strings.Split(connectionHeader, ",") {
			delete(headers, http.CanonicalHeaderKey(strings.TrimSpace(k)))
		}
	}
	hopByHop := map[string]bool{
		"Connection": true, "Keep-Alive": true, "Proxy-Authenticate": true,
		"Proxy-Authorization": true, "TE": true, "Trailer": true,
		"Transfer-Encoding": true, "Upgrade": true,
	}
	for k := range hopByHop {
		delete(headers, k)
	}
	// Add forwarding headers so the backend sees the original client
	if fwd := headers["X-Forwarded-For"]; fwd != "" {
		headers["X-Forwarded-For"] = fwd + ", " + stripRemotePort(r.RemoteAddr)
	} else {
		headers["X-Forwarded-For"] = stripRemotePort(r.RemoteAddr)
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
	waiters[reqID] = responseWaiter{tunnelID: tun.ID, ch: respChan}
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
		p.logRequest(tun.ID, r.Method, r.URL.Path, http.StatusBadGateway, int(time.Since(start).Milliseconds()), bytesIn, 0)
		return
	}

	slog.Debug("proxy request forwarded", "req_id", reqID, "tunnel_id", tun.ID, "method", r.Method, "path", r.URL.Path)

	// Wait for response
	select {
	case resp := <-respChan:
		latencyMs := int(time.Since(start).Milliseconds())
		body, err := base64.StdEncoding.DecodeString(resp.BodyB64)
		if err != nil || len(body) > protocol.MaxBodyBytes {
			slog.Error("proxy invalid response body", "req_id", reqID, "error", err)
			http.Error(w, "invalid relay response", http.StatusBadGateway)
			p.logRequest(tun.ID, r.Method, r.URL.Path, http.StatusBadGateway, latencyMs, bytesIn, 0)
			return
		}
		if resp.Status < 200 || resp.Status > 599 {
			slog.Error("proxy invalid response status", "req_id", reqID, "status", resp.Status)
			http.Error(w, "invalid relay response", http.StatusBadGateway)
			p.logRequest(tun.ID, r.Method, r.URL.Path, http.StatusBadGateway, latencyMs, bytesIn, 0)
			return
		}
		for k, v := range sanitizeResponseHeaders(resp.Headers) {
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
		w.Write(body)
		p.logRequest(tun.ID, r.Method, r.URL.Path, resp.Status, latencyMs, bytesIn, len(body))
	case <-ctx.Done():
		latencyMs := int(time.Since(start).Milliseconds())
		slog.Warn("proxy gateway timeout", "req_id", reqID, "tunnel_id", tun.ID)
		http.Error(w, "gateway timeout", http.StatusGatewayTimeout)
		p.logRequest(tun.ID, r.Method, r.URL.Path, http.StatusGatewayTimeout, latencyMs, bytesIn, 0)
	}
}

func (p *ProxyHandler) logRequest(tunnelID, method, path string, status, latencyMs, bytesIn, bytesOut int) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Store.LogRequest(ctx, tunnelID, method, path, status, latencyMs, bytesIn, bytesOut); err != nil {
		slog.Warn("proxy request log failed", "tunnel_id", tunnelID, "error", err)
	}
}

var (
	waitersMu sync.Mutex
	waiters   = make(map[string]responseWaiter)
)

type responseWaiter struct {
	tunnelID string
	ch       chan protocol.ProxyResponse
}

// HandleResponse routes a proxy response frame to the waiting HTTP handler.
func HandleResponse(tunnelID string, resp protocol.ProxyResponse) {
	waitersMu.Lock()
	waiter, ok := waiters[resp.ReqID]
	waitersMu.Unlock()
	if ok && waiter.tunnelID == tunnelID {
		select {
		case waiter.ch <- resp:
		default:
		}
	}
}

func generateReqID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err == nil {
		return "req_" + hex.EncodeToString(b)
	}
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), reqIDCounter.Add(1))
}

// resolvePublicHost returns the hostname the visitor used (for tunnel lookup).
//
// Priority:
//  1. X-Forwarded-Host (first value) — set by Coolify/Traefik/Cloudflare
//  2. r.Host — direct connections / when the proxy preserves the original host
//
// Strips :port and lowercases so DB host matches are stable.
func resolvePublicHost(r *http.Request, trustProxyHeaders bool) string {
	raw := ""
	if trustProxyHeaders {
		raw = strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	}
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

func sanitizeResponseHeaders(headers map[string]string) map[string]string {
	connectionTokens := make(map[string]bool)
	for k, v := range headers {
		if strings.EqualFold(k, "Connection") {
			for _, token := range strings.Split(v, ",") {
				connectionTokens[strings.ToLower(strings.TrimSpace(token))] = true
			}
		}
	}

	out := make(map[string]string, len(headers))
	for k, v := range headers {
		canonical := http.CanonicalHeaderKey(strings.TrimSpace(k))
		if !validHeaderName(canonical) || isHopByHopHeader(canonical) || strings.EqualFold(canonical, "Content-Length") || connectionTokens[strings.ToLower(canonical)] {
			continue
		}
		if strings.ContainsRune(v, '\r') || (canonical != "Set-Cookie" && strings.ContainsRune(v, '\n')) {
			continue
		}
		out[canonical] = v
	}
	return out
}

func isHopByHopHeader(name string) bool {
	switch strings.ToLower(name) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}

func validHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') && !strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c)) {
			return false
		}
	}
	return true
}

func stripRemotePort(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, int64(protocol.MaxBodyBytes)+1))
	if err != nil {
		return nil, err
	}
	if len(body) > protocol.MaxBodyBytes {
		return nil, fmt.Errorf("body too large")
	}
	return body, nil
}
