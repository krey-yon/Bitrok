package api

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/bitrok/bitrok/pkg/protocol"
	"github.com/bitrok/bitrok/server/internal/config"
	"github.com/bitrok/bitrok/server/internal/relay"
	"github.com/bitrok/bitrok/server/internal/store"
)

func newUpgrader(cfg *config.Config) *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if cfg.AllowInsecureWS {
				return true
			}
			origin := r.Header.Get("Origin")
			// Programmatic clients (e.g. the bitrok CLI) don't send an Origin
			// header. They authenticate via Bearer JWT, already validated by
			// AuthMiddleware before the WS upgrade. Accept them; only enforce
			// Origin for browser-initiated connections (CSRF protection).
			if origin == "" {
				return true
			}
			// Only trust configured domain, never r.Host (attacker-controlled)
			return origin == "https://"+cfg.Domain || origin == "http://"+cfg.Domain
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
}

// WSConnectHandler upgrades HTTP to WebSocket for a tunnel session.
func WSConnectHandler(cfg *config.Config, st store.Store, hub *relay.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getUserID(r.Context())
		id := chi.URLParam(r, "id")
		tun, err := st.GetTunnel(r.Context(), userID, id)
		if err != nil {
			slog.Error("db error looking up tunnel", "tunnel_id", id, "error", err)
			Error(w, http.StatusInternalServerError, "database error")
			return
		}
		if tun == nil {
			Error(w, http.StatusNotFound, "tunnel not found")
			return
		}

		conn, err := newUpgrader(cfg).Upgrade(w, r, nil)
		if err != nil {
			// Upgrade may have already written a status; log loudly so Coolify
			// shows why /tunnel/.../connect returns 4xx/5xx.
			slog.Error("websocket upgrade failed", "tunnel_id", id, "error", err)
			return
		}
		defer conn.Close()
		conn.SetReadLimit(int64(cfg.WSMaxMessageSizeMB) * 1024 * 1024)

		// Expect hello frame
		conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.WSHelloTimeoutSec) * time.Second))
		var hello protocol.Hello
		if err := conn.ReadJSON(&hello); err != nil {
			slog.Warn("failed to read hello frame", "tunnel_id", id, "error", err)
			_ = conn.WriteJSON(protocol.Message{Type: "error", Error: "invalid hello frame"})
			return
		}
		if hello.Type != string(protocol.TypeHello) {
			_ = conn.WriteJSON(protocol.Message{Type: "error", Error: "expected hello frame"})
			return
		}

		// Validate token from hello frame (defense in depth)
		if len(cfg.AuthTokens) > 0 {
			valid := false
			for _, t := range cfg.AuthTokens {
				if subtle.ConstantTimeCompare([]byte(t), []byte(hello.Token)) == 1 {
					valid = true
					break
				}
			}
			if !valid {
				slog.Warn("invalid token in hello frame", "tunnel_id", id)
				_ = conn.WriteJSON(protocol.Message{Type: "error", Error: "unauthorized"})
				return
			}
		}

		// Reset deadline
		conn.SetReadDeadline(time.Time{})

		// Register session
		session := &relay.Session{
			TunnelID: id,
			Conn:     conn,
		}
		hub.Register(session)
		defer hub.Unregister(session)

		slog.Info("tunnel_connected", "user_id", userID, "tunnel_id", id)

		// Ping ticker
		ticker := time.NewTicker(time.Duration(cfg.WSPingIntervalSec) * time.Second)
		defer ticker.Stop()

		// Read loop
		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.WSReadTimeoutSec) * time.Second))
				_, data, err := conn.ReadMessage()
				if err != nil {
					slog.Debug("websocket read error", "tunnel_id", id, "error", err)
					return
				}
				var msg protocol.Message
				if err := json.Unmarshal(data, &msg); err != nil {
					continue
				}
				switch msg.Type {
				case string(protocol.TypePong):
					// keepalive
				case string(protocol.TypeResponse):
					var resp protocol.ProxyResponse
					if err := json.Unmarshal(data, &resp); err != nil {
						continue
					}
					relay.HandleResponse(resp)
				}
			}
		}()

		// Write loop (ping + any queued messages)
		for {
			select {
			case <-ticker.C:
				if err := conn.SetWriteDeadline(time.Now().Add(time.Duration(cfg.WSWriteTimeoutSec) * time.Second)); err != nil {
					slog.Debug("websocket set write deadline failed", "tunnel_id", id, "error", err)
					return
				}
				if err := conn.WriteJSON(protocol.Ping{Type: string(protocol.TypePing)}); err != nil {
					slog.Debug("websocket ping failed", "tunnel_id", id, "error", err)
					return
				}
			case <-readDone:
				slog.Info("tunnel disconnected", "tunnel_id", id)
				return
			}
		}
	}
}
