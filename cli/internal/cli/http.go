package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/config"
	"github.com/bitrok/bitrok/cli/internal/util"
	"github.com/bitrok/bitrok/pkg/api"
)

func init() {
	rootCmd.AddCommand(httpCmd)
	httpCmd.Flags().StringP("subdomain", "s", "", "Subdomain (e.g. 'api' for api.bitrok.tech)")
	httpCmd.Flags().StringP("host", "H", "", "Full host (overrides --subdomain)")
}

var httpCmd = &cobra.Command{
	Use:   "http <port>",
	Short: "Start an ad-hoc tunnel with an auto-assigned subdomain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := util.ValidatePortString(args[0])
		if err != nil {
			return err
		}

		subdomain, _ := cmd.Flags().GetString("subdomain")
		hostFlag, _ := cmd.Flags().GetString("host")

		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.Validate(); err != nil {
			return err
		}

		// Determine the public host
		var host string
		if hostFlag != "" {
			host = hostFlag
		} else if subdomain != "" {
			domain := cfg.DefaultDomain
			if domain == "" {
				domain = "bitrok.tech"
			}
			host = subdomain + "." + domain
		} else {
			host = randomSubdomain(cfg.DefaultDomain)
		}

		if err := util.ValidateHostname(host); err != nil {
			return err
		}

		if err := util.ResolveLocalAddr(port); err != nil {
			return err
		}

		c, err := client.NewAPIClient()
		if err != nil {
			return err
		}

		name := "adhoc-" + strings.Split(host, ".")[0]
		tun, err := c.CreateTunnel(api.TunnelCreateRequest{Name: name, Host: host, Port: port})
		if err != nil {
			return err
		}

		localAddr := fmt.Sprintf("localhost:%d", port)

		// Cleanup: delete the temp tunnel on exit
		cleanup := func() {
			_ = c.DeleteTunnel(tun.ID)
		}

		return runTunnel(cfg.ServerURL, cfg.Token, tun.ID, tun.Host, localAddr, cleanup)
	},
}

func randomSubdomain(domain string) string {
	if domain == "" {
		domain = "bitrok.tech"
	}
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b) + "." + domain
}
