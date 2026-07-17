package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(downCmd)
	downCmd.Flags().BoolP("all", "a", false, "Check all tunnels")
}

var downCmd = &cobra.Command{
	Use:   "down [name]",
	Short: "[limited] get stop instructions for a tunnel (cannot stop a remote session)",
	Long: `Tunnels run in the foreground and stop when you exit the CLI session.

Press [q] or Ctrl+C in the running 'bitrok up' or 'bitrok http' session to stop a tunnel.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		c, err := client.NewAPIClient()
		if err != nil {
			return err
		}

		if all {
			tuns, err := c.ListTunnels()
			if err != nil {
				return err
			}
			fmt.Println()
			anyActive := false
			for _, t := range tuns {
				if t.Active {
					icon := lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render("●")
					fmt.Printf("  %s %s is active\n", icon,
						lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(t.Name))
					ui.Hint("Press [q] or Ctrl+C in the running session to stop")
					anyActive = true
				}
			}
			if !anyActive {
				ui.Info("No active tunnels.")
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("tunnel name required (or use --all)")
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

		tun, err := c.GetTunnel(t.ID)
		if err != nil {
			return err
		}

		fmt.Println()
		if tun.Active {
			icon := lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render("●")
			fmt.Printf("  %s %s is active\n", icon,
				lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(name))
			ui.Hint("Press [q] or Ctrl+C in the running session to stop it")
		} else {
			icon := lipgloss.NewStyle().Foreground(ui.Gray).Render("○")
			fmt.Printf("  %s %s is not active\n", icon,
				lipgloss.NewStyle().Foreground(ui.White).Render(name))
		}
		return nil
	},
}
