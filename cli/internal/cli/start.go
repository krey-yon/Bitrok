package cli

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/runstate"
	"github.com/bitrok/bitrok/cli/internal/ui"
	"github.com/bitrok/bitrok/cli/internal/util"
	"github.com/bitrok/bitrok/pkg/api"
)

// startFlags are shared by `bitrok <name> <port>`, `bitrok http`, and `bitrok ws`.
type startFlags struct {
	Detach   bool
	Open     bool
	QR       bool
	NoAnim   bool
	AllowIPs []string
	Host     string // full host override
}

func bindStartFlags(cmd *cobra.Command) {
	cmd.Flags().BoolP("detach", "d", false, "Run tunnel in background and free the terminal")
	cmd.Flags().Bool("open", false, "Open the public URL in your browser")
	cmd.Flags().Bool("qr", false, "Print a QR code for the public URL")
	cmd.Flags().Bool("no-anim", false, "Disable banner + inline animations")
	cmd.Flags().StringSlice("allow-ip", nil, "CIDR allowlist for visitor IPs (repeatable)")
	cmd.Flags().StringP("host", "H", "", "Full host override (skips app-user derivation)")
}

func readStartFlags(cmd *cobra.Command) (startFlags, error) {
	detach, _ := cmd.Flags().GetBool("detach")
	open, _ := cmd.Flags().GetBool("open")
	qr, _ := cmd.Flags().GetBool("qr")
	noAnim, _ := cmd.Flags().GetBool("no-anim")
	allow, _ := cmd.Flags().GetStringSlice("allow-ip")
	host, _ := cmd.Flags().GetString("host")
	if _, err := util.ParseAllowList(allow); err != nil {
		return startFlags{}, err
	}
	return startFlags{
		Detach:   detach,
		Open:     open,
		QR:       qr,
		NoAnim:   noAnim,
		AllowIPs: allow,
		Host:     host,
	}, nil
}

// runStart is the primary tunnel entry: name + port → deterministic host.
//
//	bitrok myapp 3000
//	→ https://myapp-<username>.bitrok.tech
func runStart(name string, port int, flags startFlags) error {
	if flags.NoAnim {
		ui.NoAnim = true
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	slug := util.Slugify(name)
	if slug == "" {
		return fmt.Errorf("app name must contain alphanumeric characters")
	}
	if err := util.ValidatePort(port); err != nil {
		return err
	}
	// Skip local port check in parent of -d; child will verify.
	if !flags.Detach || runstate.IsDetached() {
		if err := util.ResolveLocalAddr(port); err != nil {
			return err
		}
	}

	host := flags.Host
	if host == "" {
		username, err := util.UsernameFromToken(cfg.Token)
		if err != nil {
			return fmt.Errorf("could not read auth token: %w", err)
		}
		if username == "" {
			return fmt.Errorf("your token has no username claim\n\n  Generate a fresh token on the dashboard (after deploy that embeds username),\n  then:\n    bitrok login\n\n  Or force a host:\n    bitrok %s %d --host %s-you.bitrok.tech", slug, port, slug)
		}
		domain := cfg.DefaultDomain
		if domain == "" {
			domain = "bitrok.tech"
		}
		host = slug + "-" + util.Slugify(username) + "." + domain
	}
	if err := util.ValidateHostname(host); err != nil {
		return err
	}

	// Background mode: parent re-execs child, waits for meta, prints URL.
	if flags.Detach && !runstate.IsDetached() {
		return detachStart(slug, host, flags)
	}

	return foregroundStart(slug, port, host, cfg, flags)
}

func detachStart(name, host string, flags startFlags) error {
	// Remove stale meta so we can detect the child's fresh write.
	_ = runstate.RemoveMeta(name)

	pid, err := runstate.Detach(name, runstate.SelfArgv())
	if err != nil {
		return err
	}

	// Wait for child to register meta (up to 15s).
	var meta *runstate.TunnelMeta
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		m, err := runstate.ReadMeta(name)
		if err == nil && m != nil && m.PID == pid && m.PublicURL != "" {
			meta = m
			break
		}
		// Child may have died.
		if !pidAlive(pid) {
			return fmt.Errorf("background tunnel exited early — check %s", runstate.AbsLogPath(name))
		}
		time.Sleep(100 * time.Millisecond)
	}
	if meta == nil {
		return fmt.Errorf("timed out waiting for tunnel to start (pid %d) — check %s", pid, runstate.AbsLogPath(name))
	}

	ui.Success(fmt.Sprintf("tunnel %s detached (pid %d)", name, pid))
	fmt.Printf("  %s %s\n",
		ui.Icon(ui.IconGlobe, ui.Accent),
		lipgloss.NewStyle().Foreground(ui.AccentLight).Bold(true).Underline(true).Render(meta.PublicURL))
	fmt.Printf("  %s %s\n",
		ui.Icon(ui.IconArrow, ui.DarkGray),
		lipgloss.NewStyle().Foreground(ui.White).Render(fmt.Sprintf("localhost:%d", meta.Port)))
	ui.Hint(fmt.Sprintf("logs · %s", runstate.AbsLogPath(name)))
	ui.Hint(fmt.Sprintf("stop · bitrok stop %s", name))
	if flags.QR {
		fmt.Println()
		_ = util.PrintQR(meta.PublicURL)
	}
	if flags.Open {
		util.OpenBrowser(meta.PublicURL)
	}
	_ = host // host used by child via same argv
	return nil
}

func pidAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func foregroundStart(name string, port int, host string, cfg *config.CLIConfig, flags startFlags) error {
	detachedWorker := runstate.IsDetached()

	if !detachedWorker && !flags.NoAnim {
		ui.PrintAnimatedBootBanner("v0.3.0")
	}

	localAddr := fmt.Sprintf("localhost:%d", port)
	pubURL := publicURLFor(cfg.ServerURL, host)

	c, err := client.NewAPIClient()
	if err != nil {
		return err
	}

	var tun *api.Tunnel
	createFn := func() error {
		t, err := ensureTunnel(c, name, host, port)
		if err != nil {
			return err
		}
		tun = t
		return nil
	}

	if detachedWorker || flags.NoAnim {
		if err := createFn(); err != nil {
			return err
		}
	} else {
		if err := ui.RunAnimated("carving tunnel", ui.Fire, createFn); err != nil {
			return err
		}
	}

	_, err = runstate.Register(name, host, pubURL, tun.ID, port, detachedWorker, flags.AllowIPs)
	if err != nil {
		return err
	}
	defer func() { _ = runstate.RemoveMeta(name) }()

	if !detachedWorker {
		copied := util.CopyToClipboard(pubURL) == nil
		ui.RevealURL(pubURL, localAddr, copied)
		if flags.QR {
			_ = util.PrintQR(pubURL)
			fmt.Println()
		}
		if flags.Open {
			util.OpenBrowser(pubURL)
		}
	}

	// Ephemeral: delete server registration when the local process exits.
	cleanup := func() { _ = c.DeleteTunnel(tun.ID) }

	return runTunnelSession(tunnelOpts{
		ServerURL: cfg.ServerURL,
		Token:     cfg.Token,
		TunnelID:  tun.ID,
		Host:      host,
		LocalAddr: localAddr,
		Name:      name,
		AllowIPs:  flags.AllowIPs,
		ShowUI:    !detachedWorker,
		SkipIntro: true,
		Cleanup:   cleanup,
	})
}

// ensureTunnel creates the tunnel or reuses an existing one with the same host/name.
func ensureTunnel(c *client.APIClient, name, host string, port int) (*api.Tunnel, error) {
	tun, err := c.CreateTunnel(api.TunnelCreateRequest{
		Name: name,
		Host: host,
		Port: port,
	})
	if err == nil {
		return tun, nil
	}
	list, listErr := c.ListTunnels()
	if listErr != nil {
		return nil, err
	}
	for i := range list {
		if list[i].Host == host || list[i].Name == name {
			if list[i].Port != port {
				updated, uerr := c.UpdateTunnel(list[i].ID, api.TunnelUpdateRequest{Port: &port})
				if uerr != nil {
					return &list[i], nil
				}
				return updated, nil
			}
			return &list[i], nil
		}
	}
	return nil, err
}

type tunnelOpts struct {
	ServerURL string
	Token     string
	TunnelID  string
	Host      string
	LocalAddr string
	Name      string
	AllowIPs  []string
	ShowUI    bool
	SkipIntro bool
	Cleanup   func()
}

func runTunnelSession(opts tunnelOpts) error {
	pubURL := publicURLFor(opts.ServerURL, opts.Host)
	if opts.ShowUI && !opts.SkipIntro {
		ui.PrintBootBanner("v0.3.0")
		ui.BootSequence(ui.DefaultBootSteps(opts.Host))
		fmt.Printf("  %s %s %s %s\n",
			ui.Icon(ui.IconGlobe, ui.Accent),
			lipgloss.NewStyle().Foreground(ui.AccentLight).Render(pubURL),
			lipgloss.NewStyle().Foreground(ui.DarkGray).Render(ui.IconArrow),
			lipgloss.NewStyle().Foreground(ui.White).Render(opts.LocalAddr))
		fmt.Println()
	}

	session := client.NewTunnelSession(opts.ServerURL, opts.Token, opts.TunnelID, opts.LocalAddr)
	session.AllowIPs = opts.AllowIPs
	if opts.Name != "" {
		var statMu sync.Mutex
		var lastWrite time.Time
		session.OnStats = func(total, p50ms, bytesIn, bytesOut int64) {
			statMu.Lock()
			defer statMu.Unlock()
			if time.Since(lastWrite) < 2*time.Second && total%10 != 0 {
				return
			}
			lastWrite = time.Now()
			_ = runstate.UpdateStats(opts.Name, total, p50ms, bytesIn, bytesOut)
		}
	}

	var p *tea.Program
	if opts.ShowUI {
		dash := ui.NewDashboard(pubURL, opts.LocalAddr)
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

	var sessionErr error
	if opts.ShowUI {
		// Quit the TUI when the session dies (failed connect / reconnects exhausted).
		done := make(chan error, 1)
		go func() {
			err := <-errCh
			done <- err
			p.Quit()
		}()
		if _, err := p.Run(); err != nil {
			session.Stop()
			if opts.Cleanup != nil {
				opts.Cleanup()
			}
			return err
		}
		session.Stop()
		select {
		case sessionErr = <-done:
		case <-time.After(2 * time.Second):
		}
	} else {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sigCh:
			session.Stop()
			select {
			case sessionErr = <-errCh:
			case <-time.After(2 * time.Second):
			}
		case sessionErr = <-errCh:
		}
	}

	if opts.Cleanup != nil {
		opts.Cleanup()
	}

	if sessionErr != nil && !strings.Contains(sessionErr.Error(), "context canceled") {
		return sessionErr
	}

	if opts.ShowUI {
		ui.Info("tunnel stopped")
	}
	return nil
}

func publicURLFor(serverURL, host string) string {
	scheme := "https"
	if u, err := url.Parse(serverURL); err == nil && u.Scheme == "http" {
		scheme = "http"
	}
	return scheme + "://" + host
}
