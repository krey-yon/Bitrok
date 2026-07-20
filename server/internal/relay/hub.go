package relay

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Session represents an active WebSocket connection for a tunnel.
type Session struct {
	TunnelID     string
	Conn         *websocket.Conn
	writeMu      sync.Mutex
	writeTimeout time.Duration
	requests     chan struct{}
}

// NewSession constructs a tunnel session with the given write timeout.
func NewSession(tunnelID string, conn *websocket.Conn, writeTimeout time.Duration) *Session {
	return &Session{
		TunnelID:     tunnelID,
		Conn:         conn,
		writeTimeout: writeTimeout,
		requests:     make(chan struct{}, 50),
	}
}

// Close terminates the underlying tunnel connection.
func (s *Session) Close() error {
	return s.Conn.Close()
}

// TryAcquireRequest bounds the number of public HTTP handlers waiting on one
// tunnel. The CLI has the same cap, but enforcing it here prevents unbounded
// server goroutines before a request reaches the client.
func (s *Session) TryAcquireRequest() bool {
	select {
	case s.requests <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Session) ReleaseRequest() {
	<-s.requests
}

// WriteJSON serializes writes to the underlying WebSocket and resets the write
// deadline before each frame. gorilla/websocket permits one concurrent writer
// per conn and a single absolute write deadline — without the mutex, concurrent
// proxy + ping writes corrupt the stream; without resetting the deadline, the
// ping loop's SetWriteDeadline(now+writeTimeout) expires and every later proxy
// write fails with i/o timeout.
func (s *Session) WriteJSON(v any) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_ = s.Conn.SetWriteDeadline(time.Now().Add(s.writeTimeout))
	return s.Conn.WriteJSON(v)
}

// Hub tracks active tunnel sessions.
type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*Session // tunnelID -> session
}

// NewHub creates a new session hub.
func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]*Session),
	}
}

// Register adds a session to the hub.
func (h *Hub) Register(s *Session) {
	h.mu.Lock()
	previous := h.sessions[s.TunnelID]
	h.sessions[s.TunnelID] = s
	h.mu.Unlock()
	if previous != nil && previous != s {
		_ = previous.Close()
	}
}

// Unregister removes a session only if it is the currently registered one.
// This prevents a stale connection from unregistering a newer reconnect.
func (h *Hub) Unregister(s *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if existing, ok := h.sessions[s.TunnelID]; ok && existing == s {
		delete(h.sessions, s.TunnelID)
	}
}

// Get looks up a session by tunnel ID.
func (h *Hub) Get(tunnelID string) *Session {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.sessions[tunnelID]
}
