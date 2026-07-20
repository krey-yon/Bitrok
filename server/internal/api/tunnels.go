package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
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

var (
	tunnelNamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	hostLabelPattern  = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
)

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
	if err := decodeSingleJSON(r.Body, &req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			Error(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Name = strings.ToLower(strings.TrimSpace(req.Name))
	req.Host = normalizeTunnelHost(req.Host)
	if err := validateTunnelName(req.Name); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := validateTunnelHost(req.Host, h.Config.Domain, getUsername(r.Context())); err != nil {
		Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Port < 1 || req.Port > 65535 {
		Error(w, http.StatusBadRequest, "port must be between 1 and 65535")
		return
	}
	if h.Config.MaxTunnelsPerUser > 0 {
		existing, err := h.Store.ListTunnels(r.Context(), userID)
		if err != nil {
			Error(w, http.StatusInternalServerError, "database error")
			return
		}
		if len(existing) >= h.Config.MaxTunnelsPerUser {
			Error(w, http.StatusTooManyRequests, "tunnel limit reached")
			return
		}
	}

	tun, err := h.Store.CreateTunnel(r.Context(), userID, req.Name, req.Host, req.Port)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			Error(w, http.StatusConflict, "tunnel already exists")
			return
		}
		slog.Error("create tunnel failed", "user_id", userID, "error", err)
		Error(w, http.StatusInternalServerError, "database error")
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
	if err := decodeSingleJSON(r.Body, &req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			Error(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name != nil {
		normalized := strings.ToLower(strings.TrimSpace(*req.Name))
		if err := validateTunnelName(normalized); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		req.Name = &normalized
	}
	if req.Host != nil {
		normalized := normalizeTunnelHost(*req.Host)
		if err := validateTunnelHost(normalized, h.Config.Domain, getUsername(r.Context())); err != nil {
			Error(w, http.StatusBadRequest, err.Error())
			return
		}
		req.Host = &normalized
	}
	if req.Port != nil && (*req.Port < 1 || *req.Port > 65535) {
		Error(w, http.StatusBadRequest, "port must be between 1 and 65535")
		return
	}

	tun, err := h.Store.UpdateTunnel(r.Context(), userID, id, req.Name, req.Host, req.Port)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			Error(w, http.StatusConflict, "tunnel already exists")
			return
		}
		slog.Error("update tunnel failed", "user_id", userID, "tunnel_id", id, "error", err)
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
	if session := h.Hub.Get(id); session != nil {
		h.Hub.Unregister(session)
		_ = session.Close()
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

func validateTunnelName(name string) error {
	if !tunnelNamePattern.MatchString(name) {
		return fmt.Errorf("name must be a lowercase DNS label between 1 and 63 characters")
	}
	return nil
}

func validateTunnelHost(host, platformDomain, username string) error {
	if host == "" || len(host) > 253 || strings.ContainsAny(host, "\x00\r\n\t ") {
		return fmt.Errorf("host must be a valid DNS name")
	}
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return fmt.Errorf("host must be a fully qualified DNS name")
	}
	for _, label := range labels {
		if !hostLabelPattern.MatchString(label) {
			return fmt.Errorf("host must be a valid DNS name")
		}
	}

	platformDomain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(platformDomain), "."))
	if host == platformDomain {
		return fmt.Errorf("the platform domain cannot be used as a tunnel host")
	}
	if strings.Contains(host, "."+platformDomain+".") {
		return fmt.Errorf("host must not embed the platform domain")
	}
	if !strings.HasSuffix(host, "."+platformDomain) {
		return fmt.Errorf("host must be a subdomain of %s", platformDomain)
	}
	if username == "" {
		return fmt.Errorf("token must include a username to reserve a platform host")
	}
	prefix := strings.TrimSuffix(host, "."+platformDomain)
	if strings.Contains(prefix, ".") {
		return fmt.Errorf("platform hosts must use exactly one subdomain label")
	}
	if !strings.HasSuffix(prefix, "-"+username) || prefix == username {
		return fmt.Errorf("platform hosts must end with -%s.%s", username, platformDomain)
	}
	return nil
}

func decodeSingleJSON(r io.Reader, dst any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}
