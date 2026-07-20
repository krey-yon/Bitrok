package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/ui"
	"github.com/bitrok/bitrok/cli/internal/util"
)

var rootCmd = &cobra.Command{
	Use:   "bitrok [name] [port]",
	Short: "Deterministic tunnels, zero bullshit",
	Long: ui.Banner + `
Bitrok carves stable HTTPS URLs to localhost.

  bitrok myapp 3000          start a tunnel (myapp-<you>.bitrok.tech)
  bitrok myapp 3000 -d       background mode
  bitrok stop myapp          stop a local tunnel
  bitrok list                active local tunnels
  bitrok status myapp        uptime + traffic stats
  bitrok up                  start tunnels from bitrok.yml
`,
	Args:                  cobra.ArbitraryArgs,
	DisableFlagsInUseLine: false,
	SilenceUsage:          true, // don't dump flag help on runtime errors
	SilenceErrors:         true, // main already prints the error
	RunE:                  runRoot,
}

func init() {
	// Global-ish start flags on root so `bitrok myapp 3000 -d --qr` works.
	bindStartFlags(rootCmd)
}

// Execute is the CLI entrypoint.
func Execute() error {
	// Rewrite free-form `bitrok <name> <port>` before cobra treats <name> as
	// an unknown subcommand. Known commands pass through untouched.
	if rewritten := maybeRewriteStartArgs(os.Args); rewritten != nil {
		os.Args = rewritten
	}
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	return rootCmd.Execute()
}

// knownCommands are first-arg tokens that are real subcommands / reserved.
var knownCommands = map[string]bool{
	"help": true, "completion": true,
	"auth": true, "login": true, "config": true,
	"create": true, "delete": true, "update": true, "inspect": true,
	"up": true, "down": true, "stop": true,
	"list": true, "status": true,
	"http":    true,
	"version": true,
}

// maybeRewriteStartArgs converts `bitrok myapp 3000 [flags]` into
// `bitrok __start myapp 3000 [flags]` so cobra routes to the hidden start cmd.
func maybeRewriteStartArgs(args []string) []string {
	if len(args) < 3 {
		return nil
	}
	// args[0] is binary
	first := args[1]
	if strings.HasPrefix(first, "-") {
		return nil
	}
	if knownCommands[first] {
		return nil
	}
	// Second must look like a port (or we leave it — root RunE will error).
	if _, err := strconv.Atoi(args[2]); err != nil {
		// Could be `bitrok myapp --port 3000` later; for now require positional port.
		return nil
	}
	out := make([]string, 0, len(args)+1)
	out = append(out, args[0], "__start")
	out = append(out, args[1:]...)
	return out
}

func runRoot(cmd *cobra.Command, args []string) error {
	// No args → brand splash + help.
	if len(args) == 0 {
		ui.PrintBanner(Version)
		fmt.Println(cmd.UsageString())
		return nil
	}
	// Fallback: name + port without rewrite (shouldn't happen often).
	if len(args) == 2 {
		port, err := util.ValidatePortString(args[1])
		if err != nil {
			return fmt.Errorf("unknown command %q\n\nRun 'bitrok --help' for usage", args[0])
		}
		flags, err := readStartFlags(cmd)
		if err != nil {
			return err
		}
		return runStart(args[0], port, flags)
	}
	return fmt.Errorf("unknown command %q\n\nRun 'bitrok --help' for usage", args[0])
}

// hidden start command used after argv rewrite.
var startCmd = &cobra.Command{
	Use:    "__start [name] [port]",
	Hidden: true,
	Args:   cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := util.ValidatePortString(args[1])
		if err != nil {
			return err
		}
		flags, err := readStartFlags(cmd)
		if err != nil {
			return err
		}
		return runStart(args[0], port, flags)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	bindStartFlags(startCmd)
}
