package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().IntP("tail", "t", 50, "Number of lines to show")
	logsCmd.Flags().BoolP("all", "a", false, "Show logs for all tunnels")
}

var logsCmd = &cobra.Command{
	Use:   "logs [name]",
	Short: "[not implemented] show request logs for a tunnel",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if all {
			return fmt.Errorf("--all not yet implemented; use 'bitrok logs <name>' for a single tunnel")
		}
		if len(args) == 0 {
			return fmt.Errorf("tunnel name required")
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

		fmt.Println()
		ui.Warn("Live log streaming (SSE) is not yet implemented")
		ui.Info("Run 'bitrok up " + name + "' and watch the dashboard TUI for live traffic.")
		fmt.Println()
		return nil
	},
}
