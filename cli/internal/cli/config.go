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
			{Label: "Server", Value: masked.ServerURL},
			{Label: "Token", Value: masked.Token},
			{Label: "Domain", Value: masked.DefaultDomain},
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
		case "server_url":
			cfg.ServerURL = value
		case "token":
			ui.Warn("token will be stored in plaintext — prefer 'bitrok auth' or BITROK_TOKEN env var")
			cfg.Token = value
		case "default_domain":
			cfg.DefaultDomain = value
		default:
			return fmt.Errorf("unknown config key: %s", key)
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
