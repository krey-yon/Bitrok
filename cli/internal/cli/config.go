package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configResetCmd)
	configGetCmd.Flags().BoolP("json", "j", false, "Output raw JSON")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage bitrok CLI configuration",
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		asJSON, _ := cmd.Flags().GetBool("json")

		// Mask token for safety — printing a bearer token to stdout is a
		// footgun during screen shares and pipes.
		masked := *cfg
		if masked.Token != "" {
			masked.Token = maskToken(masked.Token)
		}

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(masked)
		}

		fmt.Println()
		fmt.Println(ui.DetailCard("bitrok config", []ui.KVRow{
			{Label: "relay", Value: masked.ServerURL},
			{Label: "web", Value: masked.WebURL},
			{Label: "token", Value: masked.Token},
			{Label: "domain", Value: masked.DefaultDomain},
		}))
		fmt.Println()
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		switch key {
		case "server_url", "server", "relay":
			cfg.ServerURL = config.NormalizeURL(value)
			if config.LooksLikeWebDashboard(cfg.ServerURL) {
				ui.Warn("this looks like the web dashboard URL — CLI needs the Go relay (often :8080)")
			}
		case "web_url", "web":
			cfg.WebURL = config.NormalizeURL(value)
		case "token":
			ui.Warn("token will be stored in plaintext — prefer 'bitrok login'")
			cfg.Token = value
		case "default_domain", "domain":
			cfg.DefaultDomain = value
		default:
			return fmt.Errorf("unknown config key: %s (try server_url, web_url, token, default_domain)", key)
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		ui.Success("Config updated")
		ui.Info(fmt.Sprintf("%s = %s", key, value))
		return nil
	},
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset configuration to defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !ui.Confirm("Erase server URL and auth token?") {
			ui.Info("Aborted.")
			return nil
		}
		if err := config.Save(&config.CLIConfig{DefaultDomain: "bitrok.tech"}); err != nil {
			return err
		}
		ui.Success("Configuration reset")
		return nil
	},
}

// maskToken returns a redacted form of the token for display.
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + "..." + token[len(token)-4:]
}
