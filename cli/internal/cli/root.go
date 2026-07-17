package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "bitrok",
	Short: "Deterministic tunnels, zero bullshit",
	Long:  ui.Banner + "\nBitrok is a self-hosted tunneling CLI with custom subdomains and CRUD proxy management.",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintBanner("v0.1.0")
		fmt.Println(cmd.UsageString())
	},
}

func Execute() error {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	return rootCmd.Execute()
}
