package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/ui"
	"github.com/bitrok/bitrok/cli/internal/util"
	"github.com/bitrok/bitrok/pkg/api"
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("host", "H", "", "New host")
	updateCmd.Flags().IntP("port", "p", 0, "New port")
}

var updateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update a tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")

		c, err := client.NewAPIClient()
		if err != nil {
			return err
		}

		reg, err := config.LoadRegistry()
		if err != nil {
			return fmt.Errorf("failed to load local tunnel registry: %w", err)
		}
		t := reg.FindByName(name)
		if t == nil {
			return fmt.Errorf("no tunnel found with name %s", name)
		}

		req := api.TunnelUpdateRequest{}
		if host != "" {
			if err := util.ValidateHostname(host); err != nil {
				return err
			}
			req.Host = &host
		}
		if port != 0 {
			if err := util.ValidatePort(port); err != nil {
				return err
			}
			req.Port = &port
		}

		tun, err := c.UpdateTunnel(t.ID, req)
		if err != nil {
			return err
		}

		// Update local cache
		if host != "" {
			t.Host = host
		}
		if port != 0 {
			t.Port = port
		}
		reg.Upsert(*t)
		_ = config.SaveRegistry(reg)

		ui.Success("Updated tunnel " + name)
		if host != "" {
			fmt.Println(ui.KV("Host", host))
		}
		if port != 0 {
			fmt.Println(ui.KV("Port", fmt.Sprintf("%d", port)))
		}
		_ = tun
		return nil
	},
}
