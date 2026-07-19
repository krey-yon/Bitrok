package cli

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
	"github.com/bitrok/bitrok/cli/internal/util"
)

func init() {
	rootCmd.AddCommand(loginCmd)
	// --server is the Go *relay* (API + WebSocket). --web is the dashboard for minting tokens.
	loginCmd.Flags().StringP("server", "s", "", "Relay server URL (default: https://api.bitrok.tech)")
	loginCmd.Flags().String("web", "", "Dashboard URL for token page (default: https://bitrok.tech)")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login via browser — copy & paste token",
	Long: `Opens your browser to the Bitrok web dashboard token page.

After signing in, click "Generate CLI Token", copy the token, and paste it
back here.

The CLI talks to the Go relay (api.bitrok.tech), not the Next.js dashboard.

  Production (default):
    bitrok login
    # web  → https://bitrok.tech
    # relay → https://api.bitrok.tech

  Local dev:
    bitrok login --web http://localhost:3000 --server http://localhost:8080
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		webURL, _ := cmd.Flags().GetString("web")

		if serverURL == "" {
			if env := os.Getenv("BITROK_SERVER"); env != "" {
				serverURL = env
			}
		}
		if webURL == "" {
			if env := os.Getenv("BITROK_WEB"); env != "" {
				webURL = env
			}
		}

		// Fall back to saved config.
		cfg, _ := config.Load()
		if serverURL == "" && cfg != nil && cfg.ServerURL != "" {
			serverURL = cfg.ServerURL
		}
		if webURL == "" && cfg != nil && cfg.WebURL != "" {
			webURL = cfg.WebURL
		}

		// Production defaults when nothing is set.
		if serverURL == "" && webURL == "" {
			serverURL = config.DefaultRelayURL
			webURL = config.DefaultWebURL
		}

		// Smart pairing: web-only or relay-only.
		if serverURL == "" && webURL != "" {
			if relay := config.DefaultRelayFromWeb(webURL); relay != "" {
				serverURL = relay
			} else {
				serverURL = webURL
			}
		}
		if webURL == "" && serverURL != "" {
			if web := config.DefaultWebFromRelay(serverURL); web != "" {
				webURL = web
			} else {
				webURL = config.DefaultWebURL
			}
		}

		// Rewrite accidental web URLs stored as --server / BITROK_SERVER.
		if config.LooksLikeWebDashboard(serverURL) {
			if fixed := config.DefaultRelayFromWeb(serverURL); fixed != "" {
				ui.Warn(fmt.Sprintf("%s is the web dashboard, not the relay", serverURL))
				ui.Info(fmt.Sprintf("using relay %s", fixed))
				if webURL == "" || webURL == serverURL {
					webURL = serverURL
					if config.NormalizeURL(webURL) == config.DefaultRelayURL {
						webURL = config.DefaultWebURL
					}
					// If they passed bitrok.tech as server, web is that host.
					if host := config.NormalizeURL(serverURL); host != "" {
						webURL = host
					}
				}
				serverURL = fixed
			}
		}

		serverURL = config.NormalizeURL(serverURL)
		webURL = config.NormalizeURL(webURL)
		if serverURL == "" {
			serverURL = config.DefaultRelayURL
		}
		if webURL == "" {
			webURL = config.DefaultWebURL
		}

		return copyPasteLogin(serverURL, webURL)
	},
}

func copyPasteLogin(relayURL, webURL string) error {
	authURL := strings.TrimRight(webURL, "/") + "/dashboard/cli-token"

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(ui.Accent).Bold(true).Render("  Opening browser for authentication..."))
	fmt.Println()
	ui.Info("If your browser doesn't open, visit:")
	fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(ui.AccentLight).Underline(true).Render(authURL))
	fmt.Println()
	ui.Hint(fmt.Sprintf("relay · %s", relayURL))
	fmt.Println()
	util.OpenBrowser(authURL)

	ui.Info("After clicking \"Generate CLI Token\", copy the token and paste it below.")
	fmt.Println()

	promptStyle := lipgloss.NewStyle().Foreground(ui.Accent).Bold(true)
	fmt.Print(promptStyle.Render("  Paste token: "))

	reader := bufio.NewReader(os.Stdin)
	raw, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	token := strings.TrimSpace(raw)
	if token == "" {
		return fmt.Errorf("no token provided")
	}

	if err := verifyRelayToken(relayURL, token); err != nil {
		return err
	}

	username, _ := util.UsernameFromToken(token)

	cfg, _ := config.Load()
	if cfg == nil {
		cfg = &config.CLIConfig{}
	}
	cfg.ServerURL = relayURL
	cfg.WebURL = webURL
	cfg.Token = token
	if cfg.DefaultDomain == "" || cfg.DefaultDomain == "example.com" {
		cfg.DefaultDomain = config.DefaultDomain
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	ui.Success("Authenticated successfully")
	fmt.Println(ui.KV("relay", relayURL))
	fmt.Println(ui.KV("web", webURL))
	fmt.Println(ui.KV("domain", cfg.DefaultDomain))
	if username != "" {
		fmt.Println(ui.KV("user", username))
		ui.Hint(fmt.Sprintf("tunnels look like · myapp-%s.%s", username, cfg.DefaultDomain))
	} else {
		ui.Warn("token has no username — regenerate after deploying the latest dashboard")
		ui.Hint("or use: bitrok myapp 3000 --host myapp-you.bitrok.tech")
	}
	fmt.Println()
	ui.Hint("start a tunnel · bitrok myapp 3000")
	fmt.Println()
	return nil
}

// verifyRelayToken hits GET /api/tunnels with the JWT.
func verifyRelayToken(relayURL, token string) error {
	url := config.NormalizeURL(relayURL) + "/api/tunnels"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 12 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach relay at %s: %w\n\n  Production relay is https://api.bitrok.tech", relayURL, err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
	msg := strings.TrimSpace(string(body))

	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		if config.LooksLikeWebDashboard(relayURL) || strings.Contains(msg, `"Unauthorized"`) {
			return fmt.Errorf("relay rejected the token (HTTP %d) at %s\n\n  Point server_url at the Go relay, not the web app:\n    bitrok login\n    # → web https://bitrok.tech  ·  relay https://api.bitrok.tech\n\n  Response: %s", res.StatusCode, url, truncate(msg, 120))
		}
		return fmt.Errorf("relay rejected the token (HTTP %d) at %s\n\n  Token may be expired, or BITROK_JWT_SECRET on the web dashboard\n  does not match the secret on api.bitrok.tech.\n  Generate a fresh token at %s/dashboard/cli-token\n  Response: %s",
			res.StatusCode, url, config.DefaultWebURL, truncate(msg, 120))
	default:
		return fmt.Errorf("unexpected response from relay (HTTP %d) at %s: %s", res.StatusCode, url, truncate(msg, 120))
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
