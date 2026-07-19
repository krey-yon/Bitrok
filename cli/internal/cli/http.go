package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/util"
)

func init() {
	rootCmd.AddCommand(httpCmd)
	bindStartFlags(httpCmd)
	httpCmd.Flags().StringP("name", "n", "", "App name (skips the prompt)")
}

// httpCmd is an ngrok-style alias:
//
//	bitrok http 3000
//	bitrok http 3000 myapp
//
// Prefer the primary form: bitrok myapp 3000
var httpCmd = &cobra.Command{
	Use:   "http [port] [name]",
	Short: "Start an ad-hoc tunnel (alias of bitrok <name> <port>)",
	Args:  cobra.RangeArgs(0, 2),
	RunE:  runHTTP,
}

func runHTTP(cmd *cobra.Command, args []string) error {
	flags, err := readStartFlags(cmd)
	if err != nil {
		return err
	}
	nameFlag, _ := cmd.Flags().GetString("name")

	var port int
	var name string

	switch len(args) {
	case 0:
		p := promptOr("port", "")
		if p == "" {
			return fmt.Errorf("port is required")
		}
		port, err = util.ValidatePortString(p)
		if err != nil {
			return err
		}
		name = nameFlag
		if name == "" {
			name = promptOr("app name", "")
		}
	case 1:
		port, err = util.ValidatePortString(args[0])
		if err != nil {
			return err
		}
		name = nameFlag
		if name == "" {
			name = promptOr("app name", "")
		}
	default:
		port, err = util.ValidatePortString(args[0])
		if err != nil {
			return err
		}
		name = args[1]
		if nameFlag != "" {
			name = nameFlag
		}
	}

	if name == "" {
		return fmt.Errorf("app name is required")
	}
	return runStart(name, port, flags)
}

func promptOr(label, dflt string) string {
	// Lazy import ui to keep RunE simple — ui.Prompt is the interactive prompt.
	return promptUI(label, dflt)
}
