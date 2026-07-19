package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
	downCmd.Flags().BoolP("all", "a", false, "Stop all local tunnels")
}

// downCmd is an alias of stop.
var downCmd = &cobra.Command{
	Use:   "down [name]",
	Short: "Stop a local tunnel (alias of stop)",
	Long:  `Stop a named local tunnel process. Same as bitrok stop.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if all {
			_ = stopCmd.Flags().Set("all", "true")
			return stopCmd.RunE(stopCmd, nil)
		}
		if len(args) == 0 {
			return fmt.Errorf("tunnel name required (or use --all)\n\n  bitrok stop myapp")
		}
		return stopCmd.RunE(stopCmd, args)
	},
}
