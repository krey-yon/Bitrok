package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/runstate"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolP("all", "a", false, "Stop all local tunnels")
}

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a local tunnel by name (kills the process)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if all {
			metas, err := runstate.ListMeta()
			if err != nil {
				return err
			}
			n := 0
			for _, m := range metas {
				if !runstate.Alive(m) {
					_ = runstate.RemoveMeta(m.Name)
					continue
				}
				if err := runstate.Stop(m.Name); err != nil {
					ui.Warn(fmt.Sprintf("%s: %v", m.Name, err))
					continue
				}
				ui.Success(fmt.Sprintf("stopped %s", m.Name))
				n++
			}
			if n == 0 {
				ui.Info("no local tunnels running")
			}
			return nil
		}
		if len(args) == 0 {
			return fmt.Errorf("tunnel name required (or use --all)")
		}
		name := args[0]
		if err := runstate.Stop(name); err != nil {
			return err
		}
		ui.Success(fmt.Sprintf("stopped %s", name))
		return nil
	},
}
