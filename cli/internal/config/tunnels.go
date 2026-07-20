package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LocalTunnel is a cached tunnel definition on the client side.
type LocalTunnel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	CreatedAt time.Time `json:"created_at"`
}

// RegistryPath returns the path to the local tunnels cache.
func RegistryPath() string {
	return filepath.Join(Dir(), "tunnels.json")
}

// LoadRegistry reads the local tunnel cache.
func LoadRegistry() (*TunnelRegistry, error) {
	path := RegistryPath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TunnelRegistry{Tunnels: []LocalTunnel{}}, nil
		}
		return nil, err
	}
	defer f.Close()

	var reg TunnelRegistry
	if err := json.NewDecoder(f).Decode(&reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

// SaveRegistry persists the local tunnel cache with restricted permissions (0600).
func SaveRegistry(reg *TunnelRegistry) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	if err := os.Chmod(Dir(), 0700); err != nil {
		return err
	}
	return writeJSONFile(RegistryPath(), reg)
}

// TunnelRegistry is the top-level file structure.
type TunnelRegistry struct {
	Tunnels []LocalTunnel `json:"tunnels"`
}

// FindByName returns a tunnel by its local name.
func (r *TunnelRegistry) FindByName(name string) *LocalTunnel {
	for i := range r.Tunnels {
		if r.Tunnels[i].Name == name {
			return &r.Tunnels[i]
		}
	}
	return nil
}

// FindByHost returns a tunnel by its host.
func (r *TunnelRegistry) FindByHost(host string) *LocalTunnel {
	for i := range r.Tunnels {
		if r.Tunnels[i].Host == host {
			return &r.Tunnels[i]
		}
	}
	return nil
}

// Upsert adds or updates a tunnel in the registry.
func (r *TunnelRegistry) Upsert(t LocalTunnel) {
	for i := range r.Tunnels {
		if r.Tunnels[i].ID == t.ID {
			r.Tunnels[i] = t
			return
		}
	}
	r.Tunnels = append(r.Tunnels, t)
}

// Delete removes a tunnel by name.
func (r *TunnelRegistry) Delete(name string) bool {
	for i, t := range r.Tunnels {
		if t.Name == name {
			r.Tunnels = append(r.Tunnels[:i], r.Tunnels[i+1:]...)
			return true
		}
	}
	return false
}
