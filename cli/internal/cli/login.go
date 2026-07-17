package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
	"github.com/bitrok/bitrok/cli/internal/util"
)

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("server", "s", "", "Server URL (e.g. https://bitrok.tech)")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login via browser — copy & paste token",
	Long: `Opens your browser to the Bitrok web dashboard token page.

After signing in, click "Generate CLI Token", copy the token, and paste it
back here. No local server, no ports — works in SSH and headless too.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")
		if serverURL == "" {
			// BITROK_SERVER env var overrides the saved config (e.g. point at
			// http://localhost:3000 for local dev without editing config.json).
			if envServer := os.Getenv("BITROK_SERVER"); envServer != "" {
				serverURL = envServer
			}
		}
		if serverURL == "" {
			cfg, err := config.Load()
			if err == nil && cfg.ServerURL != "" {
				serverURL = cfg.ServerURL
			}
		}
		if serverURL == "" {
			return fmt.Errorf("server URL required; use --server, BITROK_SERVER env var, or run 'bitrok auth' first")
		}
		return copyPasteLogin(serverURL)
	},
}

func copyPasteLogin(serverURL string) error {
	authURL := strings.TrimRight(serverURL, "/") + "/dashboard/cli-token"

	fmt.Println()
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Foreground(ui.Amber).Bold(true).Render("  Opening browser for authentication..."))
	fmt.Println()
	ui.Info("If your browser doesn't open, visit:")
	fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(ui.AmberLight).Underline(true).Render(authURL))
	fmt.Println()
	util.OpenBrowser(authURL)

	// Step 2: Prompt for token paste
	ui.Info("After clicking \"Generate CLI Token\", copy the token and paste it below.")
	fmt.Println()

	promptStyle := lipgloss.NewStyle().Foreground(ui.Amber).Bold(true)

	fmt.Print(promptStyle.Render("  Paste token: "))

	// Read the pasted token
	reader := bufio.NewReader(os.Stdin)
	raw, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}

	token := strings.TrimSpace(raw)
	if token == "" {
		return fmt.Errorf("no token provided")
	}

	// Step 3: Save to config
	cfg, _ := config.Load()
	cfg.ServerURL = serverURL
	cfg.Token = token
	if cfg.DefaultDomain == "" {
		cfg.DefaultDomain = "bitrok.tech"
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	ui.Success("Authenticated successfully")
	fmt.Println()
	ui.Hint("You can now use 'bitrok create', 'bitrok up', 'bitrok http', etc.")
	fmt.Println()
	return nil
}
