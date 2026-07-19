package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/bitrok/bitrok/cli/internal/util"
	"github.com/bitrok/bitrok/pkg/protocol"
)

// hop-by-hop headers that must be stripped when relaying in either direction
var hopByHop = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"TE":                  true,
	"Trailer":             true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

// RequestLog records a single proxied request for live display.
type RequestLog struct {
	Time      time.Time
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
	ReqID     string
	ReqBytes  int
	RespBytes int
}

// TunnelSession manages a WebSocket connection to the server for a single tunnel.
type TunnelSession struct {
	ServerURL  string
	Token      string
	TunnelID   string
	LocalAddr  string
	AllowIPs   []string // optional CIDR allowlist (client-side filter)
	conn       *websocket.Conn
	done       chan struct{}
	stopOnce   sync.Once
	reconnect  *Reconnect
	sem        chan struct{} // limits concurrent handleRequest goroutines
	httpClient *http.Client
	writeMu    sync.Mutex // protects conn.WriteJSON from concurrent goroutines
	Logs       chan RequestLog
	// Stats hook — called after each request (for PID meta updates).
	OnStats func(total int64, p50ms int64, bytesIn, bytesOut int64)

	// in-memory counters for OnStats
	statMu     sync.Mutex
	statCount  int64
	statIn     int64
	statOut    int64
	statLats   []time.Duration
}

// NewTunnelSession creates a new tunnel session.
func NewTunnelSession(serverURL, token, tunnelID, localAddr string) *TunnelSession {
	return &TunnelSession{
		ServerURL: serverURL,
		Token:     token,
		TunnelID:  tunnelID,
		LocalAddr: localAddr,
		done:      make(chan struct{}),
		reconnect: NewReconnect(5),
		sem:       make(chan struct{}, 50), // max 50 concurrent requests
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			// Don't follow redirects — a tunnel must pass 301/302 through to the
			// upstream client untouched.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Start dials the server and begins the relay loop.
func (t *TunnelSession) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Propagate done (Stop) into context cancellation
	go func() {
		select {
		case <-t.done:
			cancel()
		case <-ctx.Done():
		}
	}()

	var lastErr error
	for {
		if err := t.connect(); err != nil {
			lastErr = err
			// User hit q / Ctrl+C — don't report as reconnect failure.
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			if !t.reconnect.SleepContext(ctx) {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("could not connect to relay after %d attempts: %w", t.reconnect.MaxRetries, lastErr)
			}
			continue
		}
		t.reconnect.Reset()
		lastErr = nil

		err := t.readLoop()
		if err == io.EOF {
			return nil // graceful shutdown
		}
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if err != nil {
			lastErr = err
		}
		if !t.reconnect.SleepContext(ctx) {
			if ctx.Err() != nil {
				return nil
			}
			if lastErr != nil {
				return fmt.Errorf("connection lost after %d reconnects: %w", t.reconnect.MaxRetries, lastErr)
			}
			return fmt.Errorf("connection lost after %d reconnects", t.reconnect.MaxRetries)
		}
	}
}

// Stop closes the connection gracefully. Safe to call multiple times.
func (t *TunnelSession) Stop() {
	t.stopOnce.Do(func() {
		close(t.done)
		if t.conn != nil {
			t.conn.Close()
		}
		if t.Logs != nil {
			close(t.Logs)
		}
	})
}

func (t *TunnelSession) connect() error {
	raw := t.ServerURL
	// Ensure the URL has a scheme; default to https so the token isn't sent
	// over cleartext.
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	case "wss", "ws":
		// already correct
	default:
		return fmt.Errorf("unsupported server URL scheme: %q", u.Scheme)
	}
	u.Path = fmt.Sprintf("/tunnel/%s/connect", t.TunnelID)

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+t.Token)

	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		// Surface the server's response body when available so a 401/403
		// doesn't show as a generic dial error.
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			msg := strings.TrimSpace(string(body))
			if msg != "" {
				return fmt.Errorf("connect failed (HTTP %d): %s", resp.StatusCode, msg)
			}
			return fmt.Errorf("connect failed (HTTP %d)", resp.StatusCode)
		}
		return err
	}

	// Send hello
	hello := protocol.Hello{
		Type:     string(protocol.TypeHello),
		Token:    t.Token,
		TunnelID: t.TunnelID,
	}
	if err := conn.WriteJSON(hello); err != nil {
		conn.Close()
		return err
	}

	t.conn = conn
	return nil
}

func (t *TunnelSession) readLoop() error {
	for {
		select {
		case <-t.done:
			return io.EOF
		default:
		}

		if err := t.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			return err
		}
		_, data, err := t.conn.ReadMessage()
		if err != nil {
			return err
		}

		var msg protocol.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case string(protocol.TypePing):
			t.writeMu.Lock()
			_ = t.conn.WriteJSON(protocol.Pong{Type: string(protocol.TypePong)})
			t.writeMu.Unlock()
		case string(protocol.TypeRequest):
			var req protocol.ProxyRequest
			if err := json.Unmarshal(data, &req); err != nil {
				continue
			}
			select {
			case t.sem <- struct{}{}:
				go func(r protocol.ProxyRequest) {
					defer func() { <-t.sem }()
					t.handleRequest(r)
				}(req)
			default:
				// Too many concurrent requests, send 503
				t.sendResponse(req.ReqID, 503, nil, "")
			}
		}
	}
}

func (t *TunnelSession) handleRequest(req protocol.ProxyRequest) {
	start := time.Now()

	// Client-side IP allowlist (visitor IP from X-Forwarded-For).
	if len(t.AllowIPs) > 0 {
		al, err := util.ParseAllowList(t.AllowIPs)
		if err == nil && !al.Empty() {
			ip := util.ClientIPFromHeaders(req.Headers)
			if ip == nil || !al.Contains(ip) {
				t.sendResponse(req.ReqID, 403, map[string]string{"Content-Type": "text/plain"}, base64.StdEncoding.EncodeToString([]byte("forbidden: ip not allowlisted")))
				t.emitLog(RequestLog{Time: start, Method: req.Method, Path: req.Path, Status: 403, ReqID: req.ReqID, Latency: time.Since(start)})
				t.recordStats(0, 0, time.Since(start))
				return
			}
		}
	}

	body, err := base64.StdEncoding.DecodeString(req.BodyB64)
	if err != nil {
		t.sendResponse(req.ReqID, 400, nil, "")
		t.emitLog(RequestLog{Time: start, Method: req.Method, Path: req.Path, Status: 400, ReqID: req.ReqID})
		return
	}

	httpReq, err := http.NewRequest(req.Method, "http://"+t.LocalAddr+req.Path, nil)
	if err != nil {
		t.sendResponse(req.ReqID, 502, nil, "")
		t.emitLog(RequestLog{Time: start, Method: req.Method, Path: req.Path, Status: 502, ReqID: req.ReqID})
		return
	}
	// Strip hop-by-hop headers on the request side as well so the server's
	// Transfer-Encoding/Connection don't leak to localhost.
	for k, v := range req.Headers {
		if hopByHop[k] {
			continue
		}
		httpReq.Header.Set(k, v)
	}
	if req.Host != "" {
		httpReq.Host = req.Host
	}
	if len(body) > 0 {
		httpReq.Body = io.NopCloser(bytes.NewReader(body))
		httpReq.ContentLength = int64(len(body))
	}

	resp, err := t.httpClient.Do(httpReq)
	if err != nil {
		t.sendResponse(req.ReqID, 502, nil, "")
		t.emitLog(RequestLog{Time: start, Method: req.Method, Path: req.Path, Status: 502, ReqID: req.ReqID, ReqBytes: len(body), Latency: time.Since(start)})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.sendResponse(req.ReqID, 502, nil, "")
		t.emitLog(RequestLog{Time: start, Method: req.Method, Path: req.Path, Status: 502, ReqID: req.ReqID, ReqBytes: len(body), Latency: time.Since(start)})
		return
	}

	// Strip hop-by-hop headers and flatten to the format expected by the server
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if hopByHop[k] {
			continue
		}
		if len(v) == 0 {
			continue
		}
		// Set-Cookie should preserve multiple values with newlines as separator
		// (the server proxy will re-wrap them into individual headers)
		if k == "Set-Cookie" {
			respHeaders[k] = strings.Join(v, "\n")
		} else {
			respHeaders[k] = strings.Join(v, ", ")
		}
	}
	// Also strip any headers named in the Connection header
	if connHdr := respHeaders["Connection"]; connHdr != "" {
		for _, h := range strings.Split(connHdr, ",") {
			delete(respHeaders, strings.TrimSpace(h))
		}
	}

	t.sendResponse(req.ReqID, resp.StatusCode, respHeaders, base64.StdEncoding.EncodeToString(respBody))
	lat := time.Since(start)
	t.emitLog(RequestLog{Time: start, Method: req.Method, Path: req.Path, Status: resp.StatusCode, ReqID: req.ReqID, ReqBytes: len(body), RespBytes: len(respBody), Latency: lat})
	t.recordStats(int64(len(body)), int64(len(respBody)), lat)
}

// emitLog sends a request log to the TUI if a channel is wired.
// Non-blocking: a slow consumer never stalls the relay.
func (t *TunnelSession) emitLog(l RequestLog) {
	if t.Logs == nil {
		return
	}
	select {
	case t.Logs <- l:
	default:
	}
}

func (t *TunnelSession) recordStats(bytesIn, bytesOut int64, lat time.Duration) {
	t.statMu.Lock()
	t.statCount++
	t.statIn += bytesIn
	t.statOut += bytesOut
	t.statLats = append(t.statLats, lat)
	if len(t.statLats) > 200 {
		t.statLats = t.statLats[len(t.statLats)-200:]
	}
	count, in, out := t.statCount, t.statIn, t.statOut
	p50 := p50Of(t.statLats)
	t.statMu.Unlock()
	if t.OnStats != nil {
		t.OnStats(count, p50.Milliseconds(), in, out)
	}
}

func p50Of(lats []time.Duration) time.Duration {
	if len(lats) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(lats))
	copy(sorted, lats)
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func (t *TunnelSession) sendResponse(reqID string, status int, headers map[string]string, bodyB64 string) {
	resp := protocol.ProxyResponse{
		Type:    string(protocol.TypeResponse),
		ReqID:   reqID,
		Status:  status,
		Headers: headers,
		BodyB64: bodyB64,
	}
	// Serialize writes to prevent concurrent goroutines from interleaving JSON frames
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	_ = t.conn.WriteJSON(resp)
}
