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
	host := r.Host
	if host == "" {
		host = r.Header.Get("X-Forwarded-Host")
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	tun, err := p.Store.GetTunnelByHost(ctx, host)
	if err != nil {
		slog.Error("proxy db error", "host", host, "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if tun == nil {
		slog.Debug("proxy tunnel not found", "host", host)
		http.Error(w, "tunnel not found", http.StatusNotFound)
		return
	}

	session := p.Hub.Get(tun.ID)
	if session == nil {
		slog.Debug("proxy tunnel not active", "host", host, "tunnel_id", tun.ID)
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
	if r.TLS != nil {
		headers["X-Forwarded-Proto"] = "https"
	} else {
		headers["X-Forwarded-Proto"] = "http"
	}
	headers["X-Forwarded-Host"] = r.Host

	frame := protocol.ProxyRequest{
		Type:    string(protocol.TypeRequest),
		ReqID:   reqID,
		Method:  r.Method,
		Path:    r.URL.RequestURI(),
		Host:    r.Host,
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

	if err := session.Conn.WriteJSON(frame); err != nil {
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
