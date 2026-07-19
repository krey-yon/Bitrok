package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// CLIConfig holds user-level configuration for the bitrok CLI.
//
// ServerURL is the Go *relay* API (tunnels + WebSocket), e.g. http://localhost:8080.
// WebURL is the dashboard used only for browser login (token minting), e.g. http://localhost:3000.
// They are different processes — pointing ServerURL at the Next.js app causes
// "Unauthorized" because the web /api/tunnels route expects a session cookie, not a CLI JWT.
type CLIConfig struct {
	ServerURL     string `json:"server_url"`
	WebURL        string `json:"web_url,omitempty"`
	Token         string `json:"token"`
	DefaultDomain string `json:"default_domain"`
}

// Dir returns the configuration directory path.
func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "bitrok")
}

// ConfigPath returns the full path to config.json.
func ConfigPath() string {
	return filepath.Join(Dir(), "config.json")
}

// Load reads the config file or returns defaults.
func Load() (*CLIConfig, error) {
	path := ConfigPath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &CLIConfig{DefaultDomain: DefaultDomain}, nil
		}
		return nil, err
	}
	defer f.Close()

	var cfg CLIConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	cfg.Normalize()
	return &cfg, nil
}

// Save persists the config file with restricted permissions (0600).
func Save(cfg *CLIConfig) error {
	cfg.Normalize()
	dir := Dir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := ConfigPath()
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

// Normalize trims whitespace and trailing slashes on URLs.
func (c *CLIConfig) Normalize() {
	c.ServerURL = NormalizeURL(c.ServerURL)
	c.WebURL = NormalizeURL(c.WebURL)
	c.Token = strings.TrimSpace(c.Token)
	c.DefaultDomain = strings.TrimSpace(c.DefaultDomain)
}

// Validate ensures the config has the minimum required fields.
func (c *CLIConfig) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("relay server URL not configured; run 'bitrok login' or 'bitrok auth --server <relay-url>'")
	}
	if c.Token == "" {
		return fmt.Errorf("auth token not configured; run 'bitrok login'")
	}
	return nil
}

// NormalizeURL trims space and trailing slashes so path joins never become //.
func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return strings.TrimRight(raw, "/")
}

// Production hosts.
const (
	DefaultRelayURL = "https://api.bitrok.tech"
	DefaultWebURL   = "https://bitrok.tech"
	DefaultDomain   = "bitrok.tech"
)

// LooksLikeWebDashboard reports whether url is probably the Next.js app
// rather than the Go relay. Used to prevent the classic misconfig.
func LooksLikeWebDashboard(raw string) bool {
	u, err := url.Parse(NormalizeURL(raw))
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	// Local Next.js default.
	if port == "3000" {
		return true
	}
	// Production dashboard (Vercel) — not the Coolify relay.
	if host == "bitrok.tech" || host == "www.bitrok.tech" {
		return true
	}
	// Vercel previews.
	if strings.Contains(host, "vercel.app") {
		return true
	}
	return false
}

// DefaultRelayFromWeb maps a dashboard URL to the matching Go relay.
// Returns "" if no safe default is known.
func DefaultRelayFromWeb(web string) string {
	u, err := url.Parse(NormalizeURL(web))
	if err != nil || u.Host == "" {
		return ""
	}
	host := strings.ToLower(u.Hostname())

	// Production: bitrok.tech (web) → api.bitrok.tech (relay).
	if host == "bitrok.tech" || host == "www.bitrok.tech" {
		return DefaultRelayURL
	}

	// Local Next.js → local Go relay.
	if u.Port() == "3000" || (u.Port() == "" && (host == "localhost" || host == "127.0.0.1")) {
		scheme := u.Scheme
		if scheme == "" {
			scheme = "http"
		}
		h := u.Hostname()
		if h == "" {
			h = "localhost"
		}
		return fmt.Sprintf("%s://%s:8080", scheme, h)
	}
	return ""
}

// DefaultWebFromRelay maps a relay URL to the matching dashboard.
func DefaultWebFromRelay(relay string) string {
	u, err := url.Parse(NormalizeURL(relay))
	if err != nil || u.Host == "" {
		return ""
	}
	host := strings.ToLower(u.Hostname())

	// Production relay → production web.
	if host == "api.bitrok.tech" {
		return DefaultWebURL
	}

	// Local Go relay → local Next.js.
	if u.Port() == "8080" {
		scheme := u.Scheme
		if scheme == "" {
			scheme = "http"
		}
		h := u.Hostname()
		if h == "" {
			h = "localhost"
		}
		return fmt.Sprintf("%s://%s:3000", scheme, h)
	}
	return ""
}

// ResolveRelayURL normalizes and rewrites known web URLs to the relay.
func ResolveRelayURL(raw string) string {
	raw = NormalizeURL(raw)
	if raw == "" {
		return DefaultRelayURL
	}
	if LooksLikeWebDashboard(raw) {
		if fixed := DefaultRelayFromWeb(raw); fixed != "" {
			return fixed
		}
	}
	return raw
}
