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
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("name", "n", "", "Tunnel name")
	createCmd.Flags().StringP("host", "H", "", "Proxy host (e.g. api.excalidraw.bitrok.tech)")
	createCmd.Flags().IntP("port", "p", 0, "Local port to forward")
	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("host")
	_ = createCmd.MarkFlagRequired("port")
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Register a new tunnel",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")

		if err := util.ValidateHostname(host); err != nil {
			return err
		}
		if err := util.ValidatePort(port); err != nil {
			return err
		}

		c, err := client.NewAPIClient()
		if err != nil {
			return err
		}

		tun, err := c.CreateTunnel(api.TunnelCreateRequest{Name: name, Host: host, Port: port})
		if err != nil {
			return err
		}

		// Cache locally
		reg, err := config.LoadRegistry()
		if err != nil {
			return fmt.Errorf("failed to load local tunnel registry: %w", err)
		}
		reg.Upsert(config.LocalTunnel{
			ID:        tun.ID,
			Name:      name,
			Host:      host,
			Port:      port,
			CreatedAt: tun.CreatedAt,
		})
		_ = config.SaveRegistry(reg)

		ui.Success("Created tunnel " + name)
		fmt.Println(ui.KV("Host", host))
		fmt.Println(ui.KV("Port", fmt.Sprintf("%d", port)))
		fmt.Println(ui.KV("ID", tun.ID))
		return nil
	},
}
