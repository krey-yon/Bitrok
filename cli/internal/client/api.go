package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/pkg/api"
)

// APIClient talks to the bitrok-server REST API (Go relay), not the Next.js app.
type APIClient struct {
	cfg    *config.CLIConfig
	client *http.Client
}

// NewAPIClient creates a client from the current CLI config.
func NewAPIClient() (*APIClient, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	// Auto-correct the classic footgun: server_url points at the web dashboard.
	if config.LooksLikeWebDashboard(cfg.ServerURL) {
		fixed := config.ResolveRelayURL(cfg.ServerURL)
		return nil, fmt.Errorf(
			"server_url is set to the web dashboard (%s), not the Go relay\n\n  Fix:\n    bitrok config set server_url %s\n  Or:\n    bitrok login\n  (relay is https://api.bitrok.tech in production)",
			cfg.ServerURL, fixed,
		)
	}
	return &APIClient{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *APIClient) base() string {
	return config.NormalizeURL(c.cfg.ServerURL)
}

func (c *APIClient) request(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	url := c.base() + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	return c.client.Do(req)
}

func (c *APIClient) decode(res *http.Response, out any) error {
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	if res.StatusCode >= 400 {
		var e api.ErrorResponse
		if err := json.Unmarshal(raw, &e); err == nil && e.Error != "" {
			return c.authHint(res.StatusCode, e.Error)
		}
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", res.StatusCode)
		}
		return c.authHint(res.StatusCode, msg)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *APIClient) authHint(status int, msg string) error {
	if status == http.StatusUnauthorized || strings.EqualFold(msg, "Unauthorized") || strings.Contains(msg, "invalid token") {
		return fmt.Errorf("%s\n\n  Relay: %s\n  CLI JWTs are validated by the Go relay, not the web dashboard.\n  • production relay: https://api.bitrok.tech\n  • production web:    https://bitrok.tech\n  • re-login: bitrok login\n  • BITROK_JWT_SECRET must match between web (Vercel) and api.bitrok.tech",
			msg, c.base())
	}
	return fmt.Errorf("%s", msg)
}

// CreateTunnel registers a new tunnel on the server.
func (c *APIClient) CreateTunnel(req api.TunnelCreateRequest) (*api.Tunnel, error) {
	res, err := c.request(http.MethodPost, "/api/tunnels", req)
	if err != nil {
		return nil, fmt.Errorf("create tunnel: %w (is the relay running at %s?)", err, c.base())
	}
	var tun api.Tunnel
	if err := c.decode(res, &tun); err != nil {
		return nil, err
	}
	return &tun, nil
}

// ListTunnels fetches all tunnels.
func (c *APIClient) ListTunnels() ([]api.Tunnel, error) {
	res, err := c.request(http.MethodGet, "/api/tunnels", nil)
	if err != nil {
		return nil, fmt.Errorf("list tunnels: %w (is the relay running at %s?)", err, c.base())
	}
	var out api.TunnelListResponse
	if err := c.decode(res, &out); err != nil {
		return nil, err
	}
	return out.Tunnels, nil
}

// GetTunnel fetches a single tunnel by ID.
func (c *APIClient) GetTunnel(id string) (*api.Tunnel, error) {
	res, err := c.request(http.MethodGet, "/api/tunnels/"+id, nil)
	if err != nil {
		return nil, err
	}
	var tun api.Tunnel
	if err := c.decode(res, &tun); err != nil {
		return nil, err
	}
	return &tun, nil
}

// UpdateTunnel patches a tunnel.
func (c *APIClient) UpdateTunnel(id string, req api.TunnelUpdateRequest) (*api.Tunnel, error) {
	res, err := c.request(http.MethodPatch, "/api/tunnels/"+id, req)
	if err != nil {
		return nil, err
	}
	var tun api.Tunnel
	if err := c.decode(res, &tun); err != nil {
		return nil, err
	}
	return &tun, nil
}

// DeleteTunnel removes a tunnel.
func (c *APIClient) DeleteTunnel(id string) error {
	res, err := c.request(http.MethodDelete, "/api/tunnels/"+id, nil)
	if err != nil {
		return err
	}
	return c.decode(res, nil)
}
