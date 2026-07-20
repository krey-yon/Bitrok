package relay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrok/bitrok/pkg/api"
	"github.com/bitrok/bitrok/server/internal/store"
)

type proxyStoreStub struct {
	store.Store
	tunnel *api.Tunnel
}

func (s proxyStoreStub) GetTunnelByHost(context.Context, string) (*api.Tunnel, error) {
	return s.tunnel, nil
}

func TestProxyUnavailableTunnelMessage(t *testing.T) {
	tests := []struct {
		name   string
		tunnel *api.Tunnel
	}{
		{name: "unregistered tunnel"},
		{name: "registered tunnel without session", tunnel: &api.Tunnel{ID: "tunnel-id"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewProxyHandler(NewHub(), proxyStoreStub{tunnel: tt.tunnel})
			req := httptest.NewRequest(http.MethodGet, "https://demo.bitrok.tech/", nil)
			resp := httptest.NewRecorder()

			handler.ServeHTTP(resp, req)

			if resp.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want %d", resp.Code, http.StatusServiceUnavailable)
			}
			if !strings.Contains(resp.Body.String(), inactiveTunnelMessage) {
				t.Fatalf("body = %q, want inactive tunnel message", resp.Body.String())
			}
		})
	}
}
