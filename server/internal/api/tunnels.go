package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/bitrok/bitrok/pkg/api"
	"github.com/bitrok/bitrok/server/internal/config"
	"github.com/bitrok/bitrok/server/internal/relay"
	"github.com/bitrok/bitrok/server/internal/store"
)

// TunnelHandler holds dependencies for tunnel CRUD.
type TunnelHandler struct {
	Store  store.Store
	Hub    *relay.Hub
	Config *config.Config
}

func (h *TunnelHandler) Register(r chi.Router) {
	r.Post("/api/tunnels", h.Create)
	r.Get("/api/tunnels", h.List)
	r.Get("/api/tunnels/{id}", h.Get)
	r.Patch("/api/tunnels/{id}", h.Update)
	r.Delete("/api/tunnels/{id}", h.Delete)
}

func (h *TunnelHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, h.Config.MaxRequestBodyBytes)
	var req api.TunnelCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			Error(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" || req.Host == "" {
		Error(w, http.StatusBadRequest, "name and host are required")
		return
	}
	if len(req.Name) > 100 {
		Error(w, http.StatusBadRequest, "name must be 100 characters or less")
		return
	}
	if len(req.Host) > 255 {
		Error(w, http.StatusBadRequest, "host must be 255 characters or less")
		return
	}
	if req.Port < 1 || req.Port > 65535 {
		Error(w, http.StatusBadRequest, "port must be between 1 and 65535")
		return
	}

	// Normalize host: lowercase, strip trailing dots / accidental scheme.
	req.Host = normalizeTunnelHost(req.Host)

	tun, err := h.Store.CreateTunnel(r.Context(), userID, req.Name, req.Host, req.Port)
	if err != nil {
		Error(w, http.StatusConflict, "tunnel already exists")
		return
	}
	slog.Info("tunnel_mutated", "action", "create", "user_id", userID, "tunnel_id", tun.ID)
	JSON(w, http.StatusCreated, tun)
}

func (h *TunnelHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	tuns, err := h.Store.ListTunnels(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	for i := range tuns {
		tuns[i].Active = h.Hub.Get(tuns[i].ID) != nil
	}
	JSON(w, http.StatusOK, api.TunnelListResponse{Tunnels: tuns})
}

func (h *TunnelHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	id := chi.URLParam(r, "id")
	tun, err := h.Store.GetTunnel(r.Context(), userID, id)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	if tun == nil {
		Error(w, http.StatusNotFound, "tunnel not found")
		return
	}
	tun.Active = h.Hub.Get(tun.ID) != nil
	JSON(w, http.StatusOK, tun)
}

func (h *TunnelHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	id := chi.URLParam(r, "id")
	r.Body = http.MaxBytesReader(w, r.Body, h.Config.MaxRequestBodyBytes)
	var req api.TunnelUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			Error(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name != nil && len(*req.Name) > 100 {
		Error(w, http.StatusBadRequest, "name must be 100 characters or less")
		return
	}
	if req.Host != nil && len(*req.Host) > 255 {
		Error(w, http.StatusBadRequest, "host must be 255 characters or less")
		return
	}
	if req.Port != nil && (*req.Port < 1 || *req.Port > 65535) {
		Error(w, http.StatusBadRequest, "port must be between 1 and 65535")
		return
	}

	tun, err := h.Store.UpdateTunnel(r.Context(), userID, id, req.Name, req.Host, req.Port)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	if tun == nil {
		Error(w, http.StatusNotFound, "tunnel not found")
		return
	}
	slog.Info("tunnel_mutated", "action", "update", "user_id", userID, "tunnel_id", tun.ID)
	JSON(w, http.StatusOK, tun)
}

func (h *TunnelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	id := chi.URLParam(r, "id")
	tun, err := h.Store.GetTunnel(r.Context(), userID, id)
	if err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	if tun == nil {
		Error(w, http.StatusNotFound, "tunnel not found")
		return
	}
	if err := h.Store.DeleteTunnel(r.Context(), userID, id); err != nil {
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	slog.Info("tunnel_mutated", "action", "delete", "user_id", userID, "tunnel_id", id)
	w.WriteHeader(http.StatusNoContent)
}

// normalizeTunnelHost lowercases and strips scheme/trailing dots so proxy
// lookups match (Traefik + browsers always send lowercase hosts).
func normalizeTunnelHost(host string) string {
	host = strings.TrimSpace(host)
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	host = strings.TrimSuffix(host, ".")
	return strings.ToLower(host)
}
