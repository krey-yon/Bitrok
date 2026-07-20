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
// ServerURL is the Go relay API and authenticated WebSocket control channel,
// e.g. http://localhost:8080.
// WebURL is the dashboard used only for browser login (token minting), e.g. http://localhost:3000.
// They are different processes — pointing ServerURL at the Next.js app causes
// "Unauthorized" because the web /api/tunnels route expects a session cookie, not a CLI JWT.
type CLIConfig struct {
	ServerURL     string `json:"server_url"`
	WebURL        string `json:"web_url,omitempty"`
	Token         string `json:"token"`
	Username      string `json:"username,omitempty"`
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
	if err := os.Chmod(dir, 0700); err != nil {
		return err
	}
	path := ConfigPath()
	return writeJSONFile(path, cfg)
}

func writeJSONFile(path string, value any) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".bitrok-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err := f.Chmod(0600); err != nil {
		_ = f.Close()
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Normalize trims whitespace and trailing slashes on URLs.
func (c *CLIConfig) Normalize() {
	c.ServerURL = NormalizeURL(c.ServerURL)
	c.WebURL = NormalizeURL(c.WebURL)
	c.Token = strings.TrimSpace(c.Token)
	c.Username = strings.TrimSpace(c.Username)
	c.DefaultDomain = strings.TrimSpace(c.DefaultDomain)
}

// Validate ensures the config has the minimum required fields.
func (c *CLIConfig) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("relay server URL not configured; run 'bitrok login'")
	}
	if c.Token == "" {
		return fmt.Errorf("auth token not configured; run 'bitrok login'")
	}
	if err := validateServiceURL("relay server", c.ServerURL); err != nil {
		return err
	}
	if c.WebURL != "" {
		if err := validateServiceURL("web dashboard", c.WebURL); err != nil {
			return err
		}
	}
	return nil
}

// NormalizeURL trims space and trailing slashes so path joins never become //.
func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, "://") {
		host := raw
		if i := strings.IndexByte(host, '/'); i >= 0 {
			host = host[:i]
		}
		scheme := "https://"
		if parsed, err := url.Parse("//" + host); err == nil && isLoopbackHost(parsed.Hostname()) {
			scheme = "http://"
		}
		raw = scheme + raw
	}
	return strings.TrimRight(raw, "/")
}

func validateServiceURL(label, raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("%s must be an http or https URL", label)
	}
	if u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return fmt.Errorf("%s URL must not contain credentials, a path, query, or fragment", label)
	}
	if u.Scheme == "http" && !isLoopbackHost(u.Hostname()) {
		return fmt.Errorf("%s must use HTTPS unless it points to localhost", label)
	}
	return nil
}

func isLoopbackHost(host string) bool {
	host = strings.Trim(strings.ToLower(host), "[]")
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
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
