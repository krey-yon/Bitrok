package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CLIConfig holds user-level configuration for the bitrok CLI.
type CLIConfig struct {
	ServerURL     string `json:"server_url"`
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
			return &CLIConfig{DefaultDomain: "bitrok.tech"}, nil
		}
		return nil, err
	}
	defer f.Close()

	var cfg CLIConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save persists the config file with restricted permissions (0600).
func Save(cfg *CLIConfig) error {
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

// Validate ensures the config has the minimum required fields.
func (c *CLIConfig) Validate() error {
	if c.ServerURL == "" {
		return fmt.Errorf("server URL not configured; run 'bitrok auth'")
	}
	if c.Token == "" {
		return fmt.Errorf("auth token not configured; run 'bitrok auth'")
	}
	return nil
}
