package cli

import "github.com/bitrok/bitrok/cli/internal/ui"

func promptUI(label, dflt string) string {
	return ui.Prompt(label, dflt)
}
