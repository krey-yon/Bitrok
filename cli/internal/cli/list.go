package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/client"
	"github.com/bitrok/bitrok/cli/internal/runstate"
	"github.com/bitrok/bitrok/cli/internal/ui"
)

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	listCmd.Flags().Bool("server", false, "List tunnels registered on the server")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show active local tunnels",
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")
		server, _ := cmd.Flags().GetBool("server")
		if server {
			return listServer(asJSON)
		}
		return listLocal(asJSON)
	},
}

func listLocal(asJSON bool) error {
	metas, err := runstate.ListMeta()
	if err != nil {
		return err
	}

	var live []*runstate.TunnelMeta
	for _, m := range metas {
		if runstate.Alive(m) {
			live = append(live, m)
		} else {
			_ = runstate.RemoveMeta(m.Name)
		}
	}
	sort.Slice(live, func(i, j int) bool { return live[i].Name < live[j].Name })

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(live)
	}

	fmt.Println()
	fmt.Printf("  %s  %s\n",
		ui.Icon(ui.IconList, ui.Accent),
		lipgloss.NewStyle().Bold(true).Foreground(ui.White).Render("LOCAL TUNNELS"))
	fmt.Printf("  %s\n", ui.BorderLine(52))

	if len(live) == 0 {
		ui.Info("no active tunnels")
		ui.Hint("start one:  bitrok myapp 3000")
		fmt.Println()
		return nil
	}

	for _, m := range live {
		mode := "fg"
		if m.Detached {
			mode = "bg"
		}
		uptime := time.Since(m.StartedAt).Round(time.Second)
		fmt.Printf("  %s %s  %s\n",
			ui.StatusDot(true),
			lipgloss.NewStyle().Bold(true).Foreground(ui.White).Width(14).Render(m.Name),
			lipgloss.NewStyle().Foreground(ui.AccentLight).Render(m.PublicURL),
		)
		fmt.Printf("    %s\n", lipgloss.NewStyle().Foreground(ui.Gray).Render(
			fmt.Sprintf("localhost:%d  %s  up %s  pid %d  %d reqs",
				m.Port, mode, uptime, m.PID, m.Requests)))
	}
	fmt.Println()
	ui.Hint("bitrok stop <name>   ·   bitrok status <name>")
	fmt.Println()
	return nil
}

func listServer(asJSON bool) error {
	c, err := client.NewAPIClient()
	if err != nil {
		return err
	}
	tuns, err := c.ListTunnels()
	if err != nil {
		return err
	}
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(tuns)
	}
	fmt.Println()
	fmt.Printf("  %s  %s\n",
		ui.Icon(ui.IconGlobe, ui.Accent),
		lipgloss.NewStyle().Bold(true).Foreground(ui.White).Render("SERVER TUNNELS"))
	fmt.Printf("  %s\n", ui.BorderLine(52))
	fmt.Println(ui.RenderTable(tuns))
	fmt.Println()
	return nil
}
