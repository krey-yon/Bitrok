package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check active tunnels",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewAPIClient()
		if err != nil {
			return err
		}
		tuns, err := c.ListTunnels()
		if err != nil {
			return err
		}
		active := 0
		for _, t := range tuns {
			if t.Active {
				active++
			}
		}

		icon := lipgloss.NewStyle().Foreground(ui.Gray).Render("○")
		if active > 0 {
			icon = lipgloss.NewStyle().Foreground(ui.Green).Bold(true).Render("●")
		}
		count := lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(fmt.Sprintf("%d/%d", active, len(tuns)))
		fmt.Printf("  %s %s tunnels active\n", icon, count)
		return nil
	},
}
