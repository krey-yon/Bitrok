package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLConfig is the structure of bitrok.yaml.
type YAMLConfig struct {
	Server  string                 `yaml:"server"`
	Token   string                 `yaml:"token"`
	Tunnels map[string]YAMLTunnel `yaml:"tunnels"`
}

// YAMLTunnel defines a tunnel inside bitrok.yaml.
type YAMLTunnel struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// LoadYAML reads a bitrok.yaml file.
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
