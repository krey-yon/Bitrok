package relay

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Session represents an active WebSocket connection for a tunnel.
type Session struct {
	TunnelID string
	Conn     *websocket.Conn
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
	defer h.mu.Unlock()
	h.sessions[s.TunnelID] = s
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
