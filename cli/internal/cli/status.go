package cli

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/bitrok/bitrok/cli/internal/runstate"
	"github.com/bitrok/bitrok/cli/internal/ui"
	"github.com/bitrok/bitrok/cli/internal/util"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show uptime, requests, and latency for a local tunnel",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return statusOverview()
		}
		return statusOne(args[0])
	},
}

func statusOverview() error {
	metas, err := runstate.ListMeta()
	if err != nil {
		return err
	}
	live := 0
	for _, m := range metas {
		if runstate.Alive(m) {
			live++
		}
	}
	fmt.Println()
	dot := ui.StatusDot(live > 0)
	count := lipgloss.NewStyle().Foreground(ui.White).Bold(true).Render(fmt.Sprintf("%d", live))
	fmt.Printf("  %s %s local tunnel(s) running\n", dot, count)
	if live > 0 {
		ui.Hint("bitrok status <name> for details")
	}
	fmt.Println()
	return nil
}

func statusOne(name string) error {
	m, err := runstate.ReadMeta(name)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("no local tunnel named %q", name)
	}
	alive := runstate.Alive(m)

	fmt.Println()
	rows := []ui.KVRow{
		{Label: "name", Value: m.Name},
		{Label: "status", Value: map[bool]string{true: "● up", false: "○ down"}[alive]},
		{Label: "url", Value: m.PublicURL},
		{Label: "local", Value: fmt.Sprintf("localhost:%d", m.Port)},
		{Label: "pid", Value: fmt.Sprintf("%d", m.PID)},
		{Label: "mode", Value: map[bool]string{true: "background", false: "foreground"}[m.Detached]},
	}
	if alive {
		rows = append(rows,
			ui.KVRow{Label: "uptime", Value: time.Since(m.StartedAt).Round(time.Second).String()},
			ui.KVRow{Label: "requests", Value: fmt.Sprintf("%d", m.Requests)},
			ui.KVRow{Label: "p50", Value: fmt.Sprintf("%dms", m.LatencyP50)},
			ui.KVRow{Label: "↑ out", Value: util.FormatBytes(m.BytesOut)},
			ui.KVRow{Label: "↓ in", Value: util.FormatBytes(m.BytesIn)},
		)
	} else {
		rows = append(rows, ui.KVRow{Label: "note", Value: "process not running (stale meta)"})
	}
	if len(m.AllowIPs) > 0 {
		rows = append(rows, ui.KVRow{Label: "allow-ip", Value: fmt.Sprintf("%v", m.AllowIPs)})
	}
	fmt.Println(ui.DetailCard(name, rows))
	fmt.Println()
	return nil
}
