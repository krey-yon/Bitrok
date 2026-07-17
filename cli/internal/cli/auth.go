package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.Flags().StringP("server", "s", "", "Server URL (e.g. https://bitrok.yourdomain.com)")
	authCmd.Flags().StringP("token", "t", "", "Auth token (prefer BITROK_TOKEN env var to avoid shell history)")
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with a bitrok server",
	Long: `Authenticate with a Bitrok server. 

The token can be provided via --token flag or the BITROK_TOKEN environment variable.
Using the environment variable is recommended to avoid exposing the token in shell history.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server, _ := cmd.Flags().GetString("server")
		token, _ := cmd.Flags().GetString("token")

		// BITROK_TOKEN env var as fallback (avoids shell history exposure)
		if token == "" {
			token = os.Getenv("BITROK_TOKEN")
		}

		if server == "" {
			return fmt.Errorf("server URL is required; use --server flag")
		}
		if token == "" {
			return fmt.Errorf("auth token is required; use --token flag or BITROK_TOKEN env var")
		}

		cfg, _ := config.Load()
		cfg.ServerURL = server
		cfg.Token = token
		if cfg.DefaultDomain == "" {
			cfg.DefaultDomain = "bitrok.tech"
		}

		if err := config.Save(cfg); err != nil {
			return err
		}
		ui.Success("Authenticated successfully")
		fmt.Println(ui.KV("Server", server))
		return nil
	},
}
