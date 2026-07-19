package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLConfig is the structure of bitrok.yml / bitrok.yaml.
type YAMLConfig struct {
	Server  string                `yaml:"server"`
	Token   string                `yaml:"token"`
	Tunnels map[string]YAMLTunnel `yaml:"tunnels"`
}

// YAMLTunnel defines a tunnel inside bitrok.yml.
//
// Prefer `subdomain` (app label) — host is built as:
//
//	{subdomain}-{username}.bitrok.tech
//
// `host` overrides the full public hostname when set.
type YAMLTunnel struct {
	Host      string `yaml:"host,omitempty"`
	Port      int    `yaml:"port"`
	Subdomain string `yaml:"subdomain,omitempty"`
}

// LoadYAML reads a bitrok.yml file.
func LoadYAML(path string) (*YAMLConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg YAMLConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
