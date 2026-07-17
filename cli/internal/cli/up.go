package cli

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
	"github.com/bitrok/bitrok/cli/internal/util"
)

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.Flags().StringP("host", "H", "", "Ad-hoc host (no saved config)")
	upCmd.Flags().IntP("port", "p", 0, "Ad-hoc port (no saved config)")
	upCmd.Flags().StringP("config", "c", "", "Path to bitrok.yaml config")
}

var upCmd = &cobra.Command{
	Use:   "up [name]",
	Short: "Start forwarding traffic for a tunnel",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		configFile, _ := cmd.Flags().GetString("config")

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return err
		}

		if configFile != "" {
			return runFromYAML(cfg, configFile)
		}

		if host != "" && port != 0 {
			if err := util.ValidateHostname(host); err != nil {
				return err
			}
			if err := util.ValidatePort(port); err != nil {
				return err
			}
			return fmt.Errorf("ad-hoc mode requires server-side registration first; use 'bitrok create'")
		}

		if len(args) == 0 {
			return fmt.Errorf("tunnel name required (or use --config)")
		}
		name := args[0]

		reg, err := config.LoadRegistry()
		if err != nil {
			return fmt.Errorf("failed to load local tunnel registry: %w", err)
		}
		t := reg.FindByName(name)
		if t == nil {
			return fmt.Errorf("no tunnel found with name %s", name)
		}

		localAddr := fmt.Sprintf("localhost:%d", t.Port)
		if err := util.ResolveLocalAddr(t.Port); err != nil {
			return err
		}

		return runTunnel(cfg.ServerURL, cfg.Token, t.ID, t.Host, localAddr, nil)
	},
}

// runTunnel starts a tunnel with the TUI dashboard.
func runTunnel(serverURL, token, tunnelID, host, localAddr string, cleanup func()) error {
	return runTunnelWithUI(serverURL, token, tunnelID, host, localAddr, true, cleanup)
}

// runTunnelHeadless runs a tunnel without the TUI (YAML multi-tunnel mode).
func runTunnelHeadless(serverURL, token, tunnelID, host, localAddr string) error {
	return runTunnelWithUI(serverURL, token, tunnelID, host, localAddr, false, nil)
}

func publicURLFor(serverURL, host string) string {
	scheme := "https"
	if u, err := url.Parse(serverURL); err == nil && u.Scheme == "http" {
		scheme = "http"
	}
	return scheme + "://" + host
}

func runTunnelWithUI(serverURL, token, tunnelID, host, localAddr string, showUI bool, cleanup func()) error {
	pubURL := publicURLFor(serverURL, host)
	if showUI {
		ui.PrintBootBanner("v0.1.0")
		// Fake but delightful startup sequence — the truthful state lives in the dashboard.
		tunnelLabel := host
		ui.BootSequence(ui.DefaultBootSteps(tunnelLabel))
	}
	fmt.Printf("  %s %s %s %s\n",
		lipgloss.NewStyle().Foreground(ui.Amber).Render("⟢"),
		lipgloss.NewStyle().Foreground(ui.AmberLight).Render(pubURL),
		lipgloss.NewStyle().Foreground(ui.DarkGray).Render("→"),
		lipgloss.NewStyle().Foreground(ui.White).Render(localAddr))
	fmt.Println()

	session := client.NewTunnelSession(serverURL, token, tunnelID, localAddr)

	var p *tea.Program
	if showUI {
		dash := ui.NewDashboard(pubURL, localAddr)
		session.Logs = make(chan client.RequestLog, 256)
		p = tea.NewProgram(dash, tea.WithAltScreen())
		go func() {
			for log := range session.Logs {
				p.Send(ui.RequestLogMsg{
					Time:      log.Time,
					Method:    log.Method,
					Path:      log.Path,
					Status:    log.Status,
					Latency:   log.Latency,
					ReqBytes:  log.ReqBytes,
					RespBytes: log.RespBytes,
				})
			}
		}()
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- session.Start()
	}()

	if showUI {
		if _, err := p.Run(); err != nil {
			session.Stop()
			if cleanup != nil {
				cleanup()
			}
			return err
		}
		// TUI exited (user pressed q / ctrl+c) — stop the session
		session.Stop()
	} else {
		// Headless mode: single signal handler, one SIGINT exits
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh
		session.Stop()
	}

	// Run cleanup (e.g. delete temp tunnel for `bitrok http`)
	if cleanup != nil {
		cleanup()
	}

	select {
	case err := <-errCh:
		if err != nil && err.Error() != "context canceled" {
			return fmt.Errorf("tunnel error: %w", err)
		}
	default:
	}

	fmt.Println("Tunnel stopped.")
	return nil
}

func runFromYAML(cfg *config.CLIConfig, path string) error {
	yamlCfg, err := config.LoadYAML(path)
	if err != nil {
		return fmt.Errorf("load yaml: %w", err)
	}

	if yamlCfg.Server != "" {
		cfg.ServerURL = yamlCfg.Server
	}
	if yamlCfg.Token != "" {
		cfg.Token = yamlCfg.Token
	}

	reg, err := config.LoadRegistry()
	if err != nil {
		return fmt.Errorf("failed to load local tunnel registry: %w", err)
	}

	var wg sync.WaitGroup
	var firstErr error
	var mu sync.Mutex

	for name, t := range yamlCfg.Tunnels {
		lt := reg.FindByName(name)
		if lt == nil {
			fmt.Fprintf(os.Stderr, "tunnel %s not found in local registry, skipping\n", name)
			continue
		}
		wg.Add(1)
		go func(n, h, id string, p int) {
			defer wg.Done()
			if err := runTunnelHeadless(cfg.ServerURL, cfg.Token, id, h, fmt.Sprintf("localhost:%d", p)); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(name, t.Host, lt.ID, t.Port)
	}

	wg.Wait()
	return firstErr
}
