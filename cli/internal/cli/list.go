package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("json", "j", false, "Output as JSON")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tunnels",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewAPIClient()
		if err != nil {
			return err
		}

		tuns, err := c.ListTunnels()
		if err != nil {
			return err
		}

		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(tuns)
		}

		fmt.Println(ui.RenderTable(tuns))
		return nil
	},
}
