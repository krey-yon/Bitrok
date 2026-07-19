package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds server-side settings.
type Config struct {
	Port                int      `json:"port"`
	Domain              string   `json:"domain"`
	DBPath              string   `json:"db_path"`
	AuthTokens          []string `json:"auth_tokens"` // static tokens (backward compat)
	JWTSecret           string   `json:"jwt_secret"`  // for OAuth tokens from webapp
	LogLevel            string   `json:"log_level"`
	TLSCertPath         string   `json:"tls_cert_path"`
	TLSKeyPath          string   `json:"tls_key_path"`
	JWTExpectedAudience string   `json:"jwt_expected_audience"`
	JWTExpectedIssuer   string   `json:"jwt_expected_issuer"`
	AllowInsecureWS     bool     `json:"allow_insecure_ws"`
	MaxRequestBodyBytes int64    `json:"max_request_body_bytes"`

	// Timeouts (seconds)
	ReadTimeout       int `json:"read_timeout"`
	WriteTimeout      int `json:"write_timeout"`
	IdleTimeout       int `json:"idle_timeout"`
	ReadHeaderTimeout int `json:"read_header_timeout"`
	ShutdownTimeout   int `json:"shutdown_timeout"`

	// Rate limiter
	RateLimitCapacity int `json:"rate_limit_capacity"`
	RateLimitWindow   int `json:"rate_limit_window"`

	// WebSocket
	WSMaxMessageSizeMB int `json:"ws_max_message_size_mb"`
	WSHelloTimeoutSec  int `json:"ws_hello_timeout_sec"`
	WSPingIntervalSec  int `json:"ws_ping_interval_sec"`
	WSReadTimeoutSec   int `json:"ws_read_timeout_sec"`
	WSWriteTimeoutSec  int `json:"ws_write_timeout_sec"`

	// SQLite pool
	DBMaxOpenConns int `json:"db_max_open_conns"`
	DBMaxIdleConns int `json:"db_max_idle_conns"`
	DBConnLifetime int `json:"db_conn_lifetime"`
	DBConnIdleTime int `json:"db_conn_idle_time"`
}

// Default returns a sensible default configuration.
func Default() *Config {
	return &Config{
		Port:                8080,
		Domain:              "bitrok.tech",
		DBPath:              "./bitrok.db",
		AuthTokens:          []string{},
		LogLevel:            "info",
		JWTExpectedAudience: "bitrok-cli",
		MaxRequestBodyBytes: 1 << 20, // 1 MB
		// Timeouts
		ReadTimeout:       10,
		WriteTimeout:      30,
		IdleTimeout:       120,
		ReadHeaderTimeout: 5,
		ShutdownTimeout:   15,
		// Rate limiter (control-plane only: /api/*, /tunnel/*).
		// capacity tokens per windowSeconds ≈ steady rate capacity/window.
		// Defaults: 600 req / 60s = 10/s sustained, burst 600 — enough for CLI.
		RateLimitCapacity: 600,
		RateLimitWindow:   60,
		// WebSocket
		WSMaxMessageSizeMB: 10,
		WSHelloTimeoutSec:  10,
		WSPingIntervalSec:  30,
		WSReadTimeoutSec:   60,
		WSWriteTimeoutSec:  10,
		// SQLite pool
		DBMaxOpenConns: 1,
		DBMaxIdleConns: 1,
		DBConnLifetime: 300,
		DBConnIdleTime: 120,
	}
}

// Load reads configuration from a JSON file and overlays environment variables.
func Load(path string) (*Config, error) {
	cfg := Default()

	if path != "" {
		if err := cfg.loadFile(path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("load config file: %w", err)
		}
	}

	cfg.loadEnv()

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) loadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(c)
}

func (c *Config) loadEnv() {
	if v := os.Getenv("BITROK_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Port = p
		}
	}
	if v := os.Getenv("BITROK_DOMAIN"); v != "" {
		c.Domain = v
	}
	if v := os.Getenv("BITROK_DB_PATH"); v != "" {
		c.DBPath = v
	}
	if v := os.Getenv("BITROK_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	// BITROK_AUTH_TOKENS is a comma-separated list of valid tokens
	if v := os.Getenv("BITROK_AUTH_TOKENS"); v != "" {
		c.AuthTokens = splitTokens(v)
	}
	if v := os.Getenv("BITROK_JWT_SECRET"); v != "" {
		c.JWTSecret = v
	}
	if v := os.Getenv("BITROK_TLS_CERT"); v != "" {
		c.TLSCertPath = v
	}
	if v := os.Getenv("BITROK_TLS_KEY"); v != "" {
		c.TLSKeyPath = v
	}
	if v := os.Getenv("BITROK_JWT_AUDIENCE"); v != "" {
		c.JWTExpectedAudience = v
	}
	if v := os.Getenv("BITROK_JWT_ISSUER"); v != "" {
		c.JWTExpectedIssuer = v
	}
	if v := os.Getenv("BITROK_ALLOW_INSECURE_WS"); v != "" {
		c.AllowInsecureWS, _ = strconv.ParseBool(v)
	}
	if v := os.Getenv("BITROK_MAX_REQUEST_BODY_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.MaxRequestBodyBytes = n
		}
	}
	// Timeouts
	if v := os.Getenv("BITROK_READ_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ReadTimeout = n
		}
	}
	if v := os.Getenv("BITROK_WRITE_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.WriteTimeout = n
		}
	}
	if v := os.Getenv("BITROK_IDLE_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.IdleTimeout = n
		}
	}
	if v := os.Getenv("BITROK_READ_HEADER_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ReadHeaderTimeout = n
		}
	}
	if v := os.Getenv("BITROK_SHUTDOWN_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ShutdownTimeout = n
		}
	}
	// Rate limiter
	if v := os.Getenv("BITROK_RATE_LIMIT_CAPACITY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.RateLimitCapacity = n
		}
	}
	if v := os.Getenv("BITROK_RATE_LIMIT_WINDOW"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.RateLimitWindow = n
		}
	}
	// WebSocket
	if v := os.Getenv("BITROK_WS_MAX_MESSAGE_SIZE_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.WSMaxMessageSizeMB = n
		}
	}
	if v := os.Getenv("BITROK_WS_HELLO_TIMEOUT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.WSHelloTimeoutSec = n
		}
	}
	if v := os.Getenv("BITROK_WS_PING_INTERVAL_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.WSPingIntervalSec = n
		}
	}
	if v := os.Getenv("BITROK_WS_READ_TIMEOUT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.WSReadTimeoutSec = n
		}
	}
	if v := os.Getenv("BITROK_WS_WRITE_TIMEOUT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.WSWriteTimeoutSec = n
		}
	}
	// SQLite pool
	if v := os.Getenv("BITROK_DB_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.DBMaxOpenConns = n
		}
	}
	if v := os.Getenv("BITROK_DB_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.DBMaxIdleConns = n
		}
	}
	if v := os.Getenv("BITROK_DB_CONN_LIFETIME"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.DBConnLifetime = n
		}
	}
	if v := os.Getenv("BITROK_DB_CONN_IDLE_TIME"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.DBConnIdleTime = n
		}
	}
}

func splitTokens(v string) []string {
	var out []string
	for _, t := range strings.Split(v, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// Validate ensures the config has the minimum required fields.
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.DBPath == "" {
		return fmt.Errorf("db_path cannot be empty")
	}
	if c.Domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	if len(c.AuthTokens) == 0 && c.JWTSecret == "" {
		return fmt.Errorf("at least one auth token or jwt_secret is required")
	}
	if c.JWTExpectedIssuer == "" {
		c.JWTExpectedIssuer = "bitrok"
	}
	return nil
}
