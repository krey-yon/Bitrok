package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/bitrok/bitrok/server/internal/store"
)

// LogHandler exposes read-only access to proxied request logs.
type LogHandler struct {
	Store store.Store
}

func (h *LogHandler) Register(r chi.Router) {
	r.Get("/api/logs", h.List)
}

// List returns a bounded recent slice of the user's request logs plus a total.
//
// GET /api/logs?limit=10  (default 10, clamped to [1, 200])
func (h *LogHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	resp, err := h.Store.ListLogs(r.Context(), userID, limit)
	if err != nil {
		slog.Error("list logs failed", "user_id", userID, "error", err)
		Error(w, http.StatusInternalServerError, "database error")
		return
	}
	JSON(w, http.StatusOK, resp)
}
