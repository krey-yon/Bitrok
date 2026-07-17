package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/bitrok/bitrok/cli/internal/util"
)

// RequestLogMsg is a tea.Msg carrying a single proxied request log entry.
type RequestLogMsg struct {
	Time      time.Time
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
	ReqBytes  int
	RespBytes int
}

// LogEntry represents a single proxied request for the logs pane.
type LogEntry struct {
	Time      time.Time
	Method    string
	Path      string
	Status    int
	Latency   time.Duration
	ReqBytes  int
	RespBytes int
}

// DashboardModel is the Bubbletea model for the active tunnel TUI.
type DashboardModel struct {
	PublicURL string
	LocalAddr string
	StartTime time.Time
	Spinner   spinner.Model
	Logs      []LogEntry
	Width     int
	Height    int
	Quitting  bool

	// Stats
	TotalRequests  int
	TotalReqBytes  int64
	TotalRespBytes int64
	latencies      []time.Duration

	// Pet + animation
	tick         int     // seconds since start (drives pet animation)
	petMood      PetMood // current pet mood
	petMoodUntil int     // tick at which mood reverts to idle
}

func NewDashboard(publicURL, localAddr string) DashboardModel {
	return DashboardModel{
		PublicURL: publicURL,
		LocalAddr: localAddr,
		StartTime: time.Now(),
		Spinner:   NewSpinner(),
		Logs:      make([]LogEntry, 0),
		petMood:   PetIdle,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(m.Spinner.Tick, tickCmd())
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "o":
			if m.PublicURL != "" {
				util.OpenBrowser(m.PublicURL)
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case RequestLogMsg:
		m.TotalRequests++
		m.TotalReqBytes += int64(msg.ReqBytes)
		m.TotalRespBytes += int64(msg.RespBytes)
		m.latencies = append(m.latencies, msg.Latency)
		if len(m.latencies) > 200 {
			m.latencies = m.latencies[len(m.latencies)-200:]
		}
		entry := LogEntry{
			Time:      msg.Time,
			Method:    msg.Method,
			Path:      msg.Path,
			Status:    msg.Status,
			Latency:   msg.Latency,
			ReqBytes:  msg.ReqBytes,
			RespBytes: msg.RespBytes,
		}
		m.Logs = append(m.Logs, entry)
		if len(m.Logs) > 500 {
			m.Logs = m.Logs[len(m.Logs)-500:]
		}
		// React: alert on 5xx, happy otherwise. Mood decays after 3 ticks.
		if msg.Status >= 500 {
			m.petMood = PetAlert
		} else {
			m.petMood = PetHappy
		}
		m.petMoodUntil = m.tick + 3
		return m, nil
	case tickMsg:
		m.tick++
		// Decay mood back to idle
		if m.tick >= m.petMoodUntil && m.petMood != PetIdle {
			m.petMood = PetIdle
		}
		return m, tickCmd()
	default:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// p50Latency returns the median of recent latencies.
func (m DashboardModel) p50Latency() time.Duration {
	if len(m.latencies) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(m.latencies))
	copy(sorted, m.latencies)
	// Simple insertion sort — n ≤ 200
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func (m DashboardModel) View() string {
	if m.Quitting {
		return ""
	}

	// Width fallback for pre-window-size render
	width := m.Width - 2
	if width < 60 {
		width = 78
	}

	uptime := time.Since(m.StartTime).Round(time.Second)

	// ── HEADER PANEL ────────────────────────────────────────────────
	// Left: gradient logo text  Right: status pill + uptime
	statusPill := lipgloss.NewStyle().
		Foreground(Green).
		Bold(true).
		Render("● LIVE")
	uptimeStr := lipgloss.NewStyle().
		Foreground(Gray).
		Render(fmt.Sprintf("up %s", uptime))

	logoText := GradientAmber("BITROK")
	headerLeft := "  " + logoText + "  " + lipgloss.NewStyle().Foreground(DarkGray).Render("│") + "  " + statusPill
	headerRight := uptimeStr + "  "
	header := padBetween(headerLeft, headerRight, width)

	// ── URL PANEL ───────────────────────────────────────────────────
	urlLine := fmt.Sprintf("  %s %s %s %s",
		m.Spinner.View(),
		lipgloss.NewStyle().Foreground(AmberLight).Bold(true).Underline(true).Render(m.PublicURL),
		lipgloss.NewStyle().Foreground(DarkGray).Render("→"),
		lipgloss.NewStyle().Foreground(White).Render(m.LocalAddr),
	)

	// ── STATS PANEL ─────────────────────────────────────────────────
	p50 := m.p50Latency()
	stats := fmt.Sprintf("  %s   %s   %s   %s   %s",
		statLabel("Requests", fmt.Sprintf("%d", m.TotalRequests)),
		statLabel("↑ Out", util.FormatBytes(m.TotalReqBytes)),
		statLabel("↓ In", util.FormatBytes(m.TotalRespBytes)),
		statLabel("p50", fmt.Sprintf("%dms", p50.Milliseconds())),
		statLabel("2xx/4xx/5xx",
			fmt.Sprintf("%s/%s/%s",
				lipgloss.NewStyle().Foreground(Green).Render(fmt.Sprintf("%d", m.countStatus(200, 299))),
				lipgloss.NewStyle().Foreground(AmberLight).Render(fmt.Sprintf("%d", m.countStatus(400, 499))),
				lipgloss.NewStyle().Foreground(Red).Render(fmt.Sprintf("%d", m.countStatus(500, 599))),
			)),
	)

	// ── LOGS PANEL ──────────────────────────────────────────────────
	logsHeader := "  " + GradientAmber("TRAFFIC") + "  " +
		lipgloss.NewStyle().Foreground(DarkGray).Render(strings.Repeat("─", max(0, width-14)))

	var logLines []string
	// Reserve rows: header(1) + url(1) + gap + stats(1) + gap + logsHeader(1) + gap + footer(3 for pet)
	maxLogs := m.Height - 16
	if maxLogs < 3 {
		maxLogs = 3
	}
	start := len(m.Logs) - maxLogs
	if start < 0 {
		start = 0
	}
	for _, l := range m.Logs[start:] {
		statusColor := Green
		if l.Status >= 500 {
			statusColor = Red
		} else if l.Status >= 400 {
			statusColor = AmberLight
		} else if l.Status >= 300 {
			statusColor = Amber
		}
		logLines = append(logLines, fmt.Sprintf("  %s  %s %s  %s  %s",
			lipgloss.NewStyle().Foreground(Gray).Render(l.Time.Format("15:04:05")),
			lipgloss.NewStyle().Foreground(AmberLight).Bold(true).Render(fmt.Sprintf("%-6s", l.Method)),
			lipgloss.NewStyle().Foreground(White).Render(truncate(l.Path, 30)),
			lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(fmt.Sprintf("%3d", l.Status)),
			lipgloss.NewStyle().Foreground(Gray).Render(fmt.Sprintf("%dms", l.Latency.Milliseconds())),
		))
	}
	if len(logLines) == 0 {
		logLines = append(logLines, "  "+lipgloss.NewStyle().Foreground(Gray).Italic(true).Render("Waiting for traffic…"))
	}

	// ── FOOTER PANEL — Pet + keys ──────────────────────────────────
	petBlock := RenderPet(m.petMood, m.tick)
	petLines := strings.Split(petBlock, "\n")
	// Pet name + speech
	petName := lipgloss.NewStyle().Foreground(AmberLight).Bold(true).Render("Bit")
	petMsg := lipgloss.NewStyle().Foreground(Gray).Italic(true).Render(petSaying(m.petMood, m.TotalRequests))

	keys := lipgloss.NewStyle().
		Foreground(DarkGray).
		Render("[o] open   [q] quit")

	// Compose footer: pet on left, message + keys stacked on right
	footerRight := []string{
		"",
		"  " + petName + "  " + petMsg,
		"  " + keys,
	}
	footerLines := make([]string, 3)
	for i := 0; i < 3; i++ {
		left := ""
		if i < len(petLines) {
			left = petLines[i]
		}
		right := ""
		if i < len(footerRight) {
			right = footerRight[i]
		}
		footerLines[i] = "  " + left + right
	}
	footer := strings.Join(footerLines, "\n")

	// ── ASSEMBLE — with dividers between panels ────────────────────
	divider := "  " + lipgloss.NewStyle().Foreground(DarkGray).Render(strings.Repeat("─", max(0, width-4)))

	// Breathing border: subtly shift border color between Amber ↔ AmberDim
	borderColor := Amber
	if m.tick%4 == 2 {
		borderColor = AmberDim
	}

	body := strings.Join([]string{
		header,
		divider,
		urlLine,
		divider,
		stats,
		"",
		logsHeader,
		strings.Join(logLines, "\n"),
		"",
		divider,
		footer,
	}, "\n")

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Render(body)
}

func (m DashboardModel) countStatus(lo, hi int) int {
	n := 0
	for _, l := range m.Logs {
		if l.Status >= lo && l.Status <= hi {
			n++
		}
	}
	return n
}

func statLabel(label, value string) string {
	return fmt.Sprintf("%s %s",
		lipgloss.NewStyle().Foreground(Gray).Render(label),
		lipgloss.NewStyle().Foreground(White).Bold(true).Render(value))
}

// petSaying returns a short line from the mascot based on mood + traffic.
func petSaying(mood PetMood, totalRequests int) string {
	switch mood {
	case PetHappy:
		return "nice one!"
	case PetAlert:
		return "eek! something 5xx'd"
	default:
		if totalRequests == 0 {
			return "napping, poke me with a request"
		}
		return "watching the wire…"
	}
}

// padBetween places left + right on the same line, padded to fit width.
func padBetween(left, right string, width int) string {
	// lipgloss.Width strips ANSI codes for accurate width
	lw := lipgloss.Width(left)
	rw := lipgloss.Width(right)
	gap := width - lw - rw
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
