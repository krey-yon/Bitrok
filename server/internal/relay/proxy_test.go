package relay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitrok/bitrok/pkg/api"
	"github.com/bitrok/bitrok/pkg/protocol"
	"github.com/bitrok/bitrok/server/internal/store"
)

type proxyStoreStub struct {
	store.Store
	tunnel *api.Tunnel
}

func TestHandleResponseIsBoundToTunnel(t *testing.T) {
	ch := make(chan protocol.ProxyResponse, 1)
	waitersMu.Lock()
	waiters["request-1"] = responseWaiter{tunnelID: "tunnel-a", ch: ch}
	waitersMu.Unlock()
	t.Cleanup(func() {
		waitersMu.Lock()
		delete(waiters, "request-1")
		waitersMu.Unlock()
	})

	resp := protocol.ProxyResponse{ReqID: "request-1", Status: http.StatusOK}
	HandleResponse("tunnel-b", resp)
	select {
	case <-ch:
		t.Fatal("response from another tunnel reached the waiter")
	default:
	}
	HandleResponse("tunnel-a", resp)
	select {
	case <-ch:
	default:
		t.Fatal("response from the owning tunnel did not reach the waiter")
	}
}

func TestSanitizeResponseHeaders(t *testing.T) {
	got := sanitizeResponseHeaders(map[string]string{
		"Content-Type":      "text/plain",
		"Connection":        "X-Internal",
		"X-Internal":        "secret",
		"Transfer-Encoding": "chunked",
		"Content-Length":    "999999",
		"X-Bad":             "one\r\ntwo",
	})
	if got["Content-Type"] != "text/plain" {
		t.Fatalf("content type missing: %#v", got)
	}
	for _, forbidden := range []string{"Connection", "X-Internal", "Transfer-Encoding", "Content-Length", "X-Bad"} {
		if _, ok := got[forbidden]; ok {
			t.Fatalf("unsafe header %q survived: %#v", forbidden, got)
		}
	}
	if validHeaderName("Bad Header") || !validHeaderName("X-Good_Header") {
		t.Fatal("header-name validation returned an unexpected result")
	}
	if strings.Contains(generateReqID(), " ") {
		t.Fatal("request id contains whitespace")
	}
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
