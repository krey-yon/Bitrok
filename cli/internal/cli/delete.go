package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringP("host", "H", "", "Delete by host instead of name")
}

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a tunnel",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")

		c, err := client.NewAPIClient()
		if err != nil {
			return err
		}

		reg, err := config.LoadRegistry()
		if err != nil {
			return fmt.Errorf("failed to load local tunnel registry: %w", err)
		}
		var id, name string
		if host != "" {
			t := reg.FindByHost(host)
			if t == nil {
				return fmt.Errorf("no tunnel found for host %s", host)
			}
			id = t.ID
			name = t.Name
		} else {
			if len(args) == 0 {
				return fmt.Errorf("name or --host required")
			}
			name = args[0]
			t := reg.FindByName(name)
			if t == nil {
				return fmt.Errorf("no tunnel found with name %s", name)
			}
			id = t.ID
		}

		if !ui.Confirm("Delete tunnel " + name + "?") {
			ui.Info("Aborted.")
			return nil
		}

		if err := c.DeleteTunnel(id); err != nil {
			return err
		}

		reg.Delete(name)
		_ = config.SaveRegistry(reg)

		ui.Success("Deleted tunnel " + name)
		return nil
	},
}
