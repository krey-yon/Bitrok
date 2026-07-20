package runstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrok/bitrok/cli/internal/config"
)

// RunDir is ~/.config/bitrok/run — PID + meta for background tunnels.
func RunDir() string {
	return filepath.Join(config.Dir(), "run")
}

// TunnelMeta is persisted for every running (foreground or -d) tunnel so
// list / status / stop can resolve by name without talking to the server.
type TunnelMeta struct {
	Name       string    `json:"name"`
	PID        int       `json:"pid"`
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	PublicURL  string    `json:"public_url"`
	TunnelID   string    `json:"tunnel_id"`
	StartedAt  time.Time `json:"started_at"`
	Detached   bool      `json:"detached"`
	LogPath    string    `json:"log_path,omitempty"`
	Requests   int64     `json:"requests"`
	LatencyP50 int64     `json:"latency_p50_ms"`
	BytesIn    int64     `json:"bytes_in"`
	BytesOut   int64     `json:"bytes_out"`
	AllowIPs   []string  `json:"allow_ips,omitempty"`
	Executable string    `json:"executable,omitempty"`
}

func metaPath(name string) string {
	return filepath.Join(RunDir(), name+".json")
}

// LogPath returns the default log file for a detached tunnel.
func LogPath(name string) string {
	return filepath.Join(RunDir(), name+".log")
}

// WriteMeta persists tunnel metadata (atomic replace).
func WriteMeta(m *TunnelMeta) error {
	if err := os.MkdirAll(RunDir(), 0700); err != nil {
		return err
	}
	path := metaPath(m.Name)
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// ReadMeta loads metadata for a named tunnel. Returns (nil, nil) if missing.
func ReadMeta(name string) (*TunnelMeta, error) {
	path := metaPath(name)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var m TunnelMeta
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// RemoveMeta deletes the PID/meta file for a tunnel.
func RemoveMeta(name string) error {
	err := os.Remove(metaPath(name))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ListMeta returns all known local tunnel metas (including dead PIDs).
func ListMeta() ([]*TunnelMeta, error) {
	dir := RunDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*TunnelMeta
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		m, err := ReadMeta(e.Name()[:len(e.Name())-5]) // strip .json
		if err != nil || m == nil {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

// Alive reports whether the process for this meta is still running.
func Alive(m *TunnelMeta) bool {
	if m == nil || m.PID <= 0 {
		return false
	}
	proc, err := os.FindProcess(m.PID)
	if err != nil {
		return false
	}
	if !processAlive(proc) {
		return false
	}
	return m.Executable == "" || processMatches(m.PID, m.Executable)
}

// Stop kills a named tunnel's local process (SIGTERM, then SIGKILL after 3s).
// Does not delete the server-side tunnel registration.
func Stop(name string) error {
	m, err := ReadMeta(name)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("no local tunnel named %q", name)
	}
	if !Alive(m) {
		_ = RemoveMeta(name)
		return fmt.Errorf("tunnel %q is not running (stale pid %d)", name, m.PID)
	}
	proc, err := os.FindProcess(m.PID)
	if err != nil {
		return err
	}
	expectedExecutable := m.Executable
	if expectedExecutable == "" {
		expectedExecutable, _ = os.Executable()
	}
	if expectedExecutable != "" && !processMatches(m.PID, expectedExecutable) {
		_ = RemoveMeta(name)
		return fmt.Errorf("refusing to stop pid %d: it is not the recorded Bitrok process", m.PID)
	}
	if err := requestProcessStop(proc); err != nil {
		return fmt.Errorf("signal %d: %w", m.PID, err)
	}
	// Wait up to 3s for graceful exit.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !Alive(m) {
			_ = RemoveMeta(name)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = forceProcessStop(proc)
	_ = RemoveMeta(name)
	return nil
}

// UpdateStats patches request counters on an existing meta file.
func UpdateStats(name string, requests, p50ms, bytesIn, bytesOut int64) error {
	m, err := ReadMeta(name)
	if err != nil || m == nil {
		return err
	}
	m.Requests = requests
	m.LatencyP50 = p50ms
	m.BytesIn = bytesIn
	m.BytesOut = bytesOut
	return WriteMeta(m)
}

// Register writes initial meta for the current process.
func Register(name, host, publicURL, tunnelID string, port int, detached bool, allowIPs []string) (*TunnelMeta, error) {
	// Refuse double-start of same name if already alive.
	if existing, _ := ReadMeta(name); existing != nil && Alive(existing) {
		return nil, fmt.Errorf("tunnel %q is already running (pid %d)", name, existing.PID)
	}
	executable, _ := os.Executable()
	m := &TunnelMeta{
		Name:       name,
		PID:        os.Getpid(),
		Host:       host,
		Port:       port,
		PublicURL:  publicURL,
		TunnelID:   tunnelID,
		StartedAt:  time.Now().UTC(),
		Detached:   detached,
		AllowIPs:   allowIPs,
		Executable: executable,
	}
	if detached {
		m.LogPath = LogPath(name)
	}
	if err := WriteMeta(m); err != nil {
		return nil, err
	}
	return m, nil
}
