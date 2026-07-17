package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/bitrok/bitrok/pkg/api"
)

// RenderTable formats a list of tunnels as an aligned table with a
// gradient header and colored status indicators.
func RenderTable(tunnels []api.Tunnel) string {
	if len(tunnels) == 0 {
		return lipgloss.NewStyle().Foreground(Gray).Render("  No tunnels found.")
	}

	nameWidth := 20
	hostWidth := 36
	portWidth := 8
	statusWidth := 12

	header := fmt.Sprintf("  %-*s %-*s %-*s %-*s %s",
		nameWidth, "NAME",
		hostWidth, "HOST",
		portWidth, "PORT",
		statusWidth, "STATUS",
		"CREATED",
	)

	var lines []string
	lines = append(lines, GradientAmber(header))
	lines = append(lines, "  "+lipgloss.NewStyle().Foreground(DarkGray).Render(strings.Repeat("─", len(header)-2)))

	for _, t := range tunnels {
		var status, statusRendered string
		if t.Active {
			status = "● up"
			statusRendered = lipgloss.NewStyle().Foreground(Green).Bold(true).Render(status)
		} else {
			status = "○ down"
			statusRendered = lipgloss.NewStyle().Foreground(Gray).Render(status)
		}
		line := fmt.Sprintf("  %-*s %-*s %-*d %-*s %s",
			nameWidth, truncate(t.Name, nameWidth-1),
			hostWidth, truncate(t.Host, hostWidth-1),
			portWidth, t.Port,
			statusWidth, status,
			humanize(t.CreatedAt),
		)
		line = strings.Replace(line, status, statusRendered, 1)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func humanize(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
