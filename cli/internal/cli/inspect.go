package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().StringP("host", "H", "", "Inspect by host instead of name")
	inspectCmd.Flags().BoolP("json", "j", false, "Output raw JSON")
}

var inspectCmd = &cobra.Command{
	Use:   "inspect [name]",
	Short: "Inspect a single tunnel",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		asJSON, _ := cmd.Flags().GetBool("json")

		c, err := client.NewAPIClient()
		if err != nil {
			return err
		}

		reg, err := config.LoadRegistry()
		if err != nil {
			return fmt.Errorf("failed to load local tunnel registry: %w", err)
		}
		var id string
		if host != "" {
			t := reg.FindByHost(host)
			if t == nil {
				return fmt.Errorf("no tunnel found for host %s", host)
			}
			id = t.ID
		} else {
			if len(args) == 0 {
				return fmt.Errorf("name or --host required")
			}
			t := reg.FindByName(args[0])
			if t == nil {
				return fmt.Errorf("no tunnel found with name %s", args[0])
			}
			id = t.ID
		}

		tun, err := c.GetTunnel(id)
		if err != nil {
			return err
		}

		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tun)
		}

		status := "○ down"
		if tun.Active {
			status = "● up"
		}
		fmt.Println()
		fmt.Println(ui.DetailCard(tun.Name, []ui.KVRow{
			{Label: "Host", Value: tun.Host},
			{Label: "Port", Value: fmt.Sprintf("%d", tun.Port)},
			{Label: "Status", Value: status},
			{Label: "ID", Value: tun.ID},
			{Label: "Created", Value: tun.CreatedAt.Format("2006-01-02 15:04")},
			{Label: "Updated", Value: tun.UpdatedAt.Format("2006-01-02 15:04")},
		}))
		fmt.Println()
		return nil
	},
}
