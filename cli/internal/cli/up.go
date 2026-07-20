package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/runstate"
	"github.com/bitrok/bitrok/cli/internal/ui"
	"github.com/bitrok/bitrok/cli/internal/util"
)

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringP("config", "c", "", "Path to bitrok.yml (default: ./bitrok.yml)")
	upCmd.Flags().BoolP("detach", "d", false, "Run tunnels in background")
	upCmd.Flags().Bool("open", false, "Open the first public URL in browser")
	upCmd.Flags().Bool("no-anim", false, "Disable animations")
}

var upCmd = &cobra.Command{
	Use:   "up [name]",
	Short: "Start tunnels defined in bitrok.yml",
	Long: `Start multi-tunnel configs from bitrok.yml:

  tunnels:
    api:
      port: 3000
      subdomain: myapp-api
    web:
      port: 5173
      subdomain: myapp-web

  bitrok up          # start all
  bitrok up api      # start one named entry
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUp,
}

func runUp(cmd *cobra.Command, args []string) error {
	configFile, _ := cmd.Flags().GetString("config")
	detach, _ := cmd.Flags().GetBool("detach")
	openFirst, _ := cmd.Flags().GetBool("open")
	noAnim, _ := cmd.Flags().GetBool("no-anim")
	if noAnim {
		ui.NoAnim = true
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	path, err := resolveYAMLPath(configFile)
	if err != nil {
		return err
	}
	yamlCfg, err := config.LoadYAML(path)
	if err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}
	if yamlCfg.Server != "" {
		cfg.ServerURL = yamlCfg.Server
	}
	if yamlCfg.Token != "" {
		cfg.Token = yamlCfg.Token
	}
	if len(yamlCfg.Tunnels) == 0 {
		return fmt.Errorf("no tunnels defined in %s", path)
	}

	username := cfg.Username
	if username == "" {
		username, err = util.UsernameFromToken(cfg.Token)
		if err != nil {
			return fmt.Errorf("could not read auth token: %w", err)
		}
	}
	domain := cfg.DefaultDomain
	if domain == "" {
		domain = "bitrok.tech"
	}

	entries := yamlCfg.Tunnels
	if len(args) == 1 {
		want := args[0]
		t, ok := entries[want]
		if !ok {
			return fmt.Errorf("tunnel %q not found in %s", want, path)
		}
		entries = map[string]config.YAMLTunnel{want: t}
	}

	if !noAnim {
		ui.PrintAnimatedBootBanner("v0.3.0")
	}
	ui.Section("starting from " + filepath.Base(path))

	// Single named tunnel in foreground with TUI.
	if len(entries) == 1 && !detach {
		for key, t := range entries {
			return startYAMLEntry(cfg, key, t, username, domain, startFlags{
				Open:   openFirst,
				NoAnim: noAnim,
			})
		}
	}

	// Multi-tunnel or -d: spawn each as detached child `bitrok <name> <port> -d --host …`
	var firstURL string
	var firstErr error
	for key, t := range entries {
		appName, host, port, err := yamlIdentity(key, t, username, domain)
		if err != nil {
			ui.ErrorOut(fmt.Sprintf("%s: %v", key, err))
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := spawnDetachedTunnel(appName, host, port); err != nil {
			ui.ErrorOut(fmt.Sprintf("%s: %v", key, err))
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		// Wait briefly for meta.
		pub := waitMetaURL(appName, 12*time.Second)
		if pub == "" {
			ui.Warn(fmt.Sprintf("%s: started but URL not ready yet", appName))
			continue
		}
		ui.Success(fmt.Sprintf("%s  %s", appName, pub))
		if firstURL == "" {
			firstURL = pub
		}
	}

	if openFirst && firstURL != "" {
		util.OpenBrowser(firstURL)
	}
	if firstErr != nil {
		return firstErr
	}
	if firstURL == "" {
		return fmt.Errorf("no tunnels started")
	}
	ui.Hint("bitrok list   ·   bitrok stop <name>")
	return nil
}

func startYAMLEntry(cfg *config.CLIConfig, key string, t config.YAMLTunnel, username, domain string, flags startFlags) error {
	appName, host, port, err := yamlIdentity(key, t, username, domain)
	if err != nil {
		return err
	}
	flags.Host = host
	_ = cfg
	return runStart(appName, port, flags)
}

func yamlIdentity(key string, t config.YAMLTunnel, username, domain string) (name, host string, port int, err error) {
	if t.Port == 0 {
		return "", "", 0, fmt.Errorf("port is required")
	}
	port = t.Port
	sub := t.Subdomain
	if sub == "" {
		sub = key
	}
	name = util.Slugify(sub)
	if name == "" {
		name = util.Slugify(key)
	}
	if t.Host != "" {
		host = t.Host
	} else {
		user := util.Slugify(username)
		if user == "" {
			host = name + "." + domain
		} else {
			host = name + "-" + user + "." + domain
		}
	}
	return name, host, port, nil
}

func spawnDetachedTunnel(name, host string, port int) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	_ = runstate.RemoveMeta(name)

	if err := os.MkdirAll(runstate.RunDir(), 0700); err != nil {
		return err
	}
	logFile, err := os.OpenFile(runstate.LogPath(name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	// Child: bitrok __start name port --host host  (with BITROK_DETACHED=1)
	cmd := exec.Command(exe, "__start", name, fmt.Sprintf("%d", port), "--host", host, "--no-anim")
	cmd.Env = append(os.Environ(), runstate.DetachEnv+"=1")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	configureDetachedCommand(cmd)
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return err
	}
	logFile.Close()
	return nil
}

func waitMetaURL(name string, timeout time.Duration) string {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		m, err := runstate.ReadMeta(name)
		if err == nil && m != nil && m.PublicURL != "" && runstate.Alive(m) {
			return m.PublicURL
		}
		time.Sleep(120 * time.Millisecond)
	}
	return ""
}

func resolveYAMLPath(flag string) (string, error) {
	if flag != "" {
		if _, err := os.Stat(flag); err != nil {
			return "", fmt.Errorf("config file: %w", err)
		}
		return flag, nil
	}
	for _, c := range []string{"bitrok.yml", "bitrok.yaml", ".bitrok.yml"} {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("no bitrok.yml found (looked for bitrok.yml, bitrok.yaml); pass --config")
}
