package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(wsCmd)
	bindStartFlags(wsCmd)
	wsCmd.Flags().StringP("name", "n", "", "App name (skips the prompt)")
}

// wsCmd is an alias of `bitrok http` for now: the relay is request/response
// and strips Upgrade/Connection headers, so true WebSocket tunneling is a
// separate server-side effort.
var wsCmd = &cobra.Command{
	Use:   "ws [port] [name]",
	Short: "Start an ad-hoc tunnel for WebSocket traffic (alias of http)",
	Args:  cobra.RangeArgs(0, 2),
	RunE:  runHTTP,
}
