package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bitrok/bitrok/pkg/api"
	"github.com/bitrok/bitrok/cli/internal/config"
)

// APIClient talks to the bitrok-server REST API.
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
	return &APIClient{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
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

	req, err := http.NewRequest(method, c.cfg.ServerURL+path, bodyReader)
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
	if res.StatusCode >= 400 {
		var e api.ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return fmt.Errorf("HTTP %d", res.StatusCode)
		}
		return fmt.Errorf("%s", e.Error)
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}
	return nil
}

// CreateTunnel registers a new tunnel on the server.
func (c *APIClient) CreateTunnel(req api.TunnelCreateRequest) (*api.Tunnel, error) {
	res, err := c.request(http.MethodPost, "/api/tunnels", req)
	if err != nil {
		return nil, err
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
		return nil, err
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
