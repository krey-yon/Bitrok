package ui

import "github.com/charmbracelet/lipgloss"

// Glyph set — monochrome Nerd-ish unicode that reads well in JetBrains Mono /
// system mono (same family as the dashboard font-mono token).
const (
	IconTunnel = "◎"
	IconLive   = "●"
	IconDown   = "○"
	IconArrow  = "→"
	IconLink   = "↗"
	IconCheck  = "✓"
	IconCross  = "✗"
	IconWarn   = "!"
	IconPrompt = "?"
	IconBolt   = "⚡"
	IconClock  = "◷"
	IconGlobe  = "◈"
	IconLock   = "▣"
	IconRocket = "▸"
	IconStop   = "■"
	IconList   = "☰"
	IconStat   = "▦"
	IconQR     = "▦"
	IconCopy   = "⎘"
	IconOpen   = "↗"
	IconDetach = "⇕"
	IconIP     = "⬡"
)

// Icon paints a glyph in the given color (defaults to Accent).
func Icon(glyph string, color ...lipgloss.Color) string {
	c := Accent
	if len(color) > 0 {
		c = color[0]
	}
	return lipgloss.NewStyle().Foreground(c).Bold(true).Render(glyph)
}

// StatusDot returns a live/down indicator.
func StatusDot(live bool) string {
	if live {
		return Icon(IconLive, Green)
	}
	return Icon(IconDown, Gray)
}

// Pill renders a compact status chip: " ● LIVE ".
func Pill(text string, fg, bg lipgloss.Color) string {
	return lipgloss.NewStyle().
		Foreground(fg).
		Background(bg).
		Bold(true).
		Padding(0, 1).
		Render(text)
}

// LivePill is the green live chip used in headers.
func LivePill() string {
	return Pill("● LIVE", Green, lipgloss.Color("#0f1a12"))
}

// OfflinePill is the muted offline chip.
func OfflinePill() string {
	return Pill("○ OFF", LightGray, BgCard)
}

// BorderLine returns a full-width hairline in the muted border color.
func BorderLine(width int) string {
	if width < 4 {
		width = 40
	}
	return lipgloss.NewStyle().Foreground(DarkGray).Render(repeat("─", width))
}

// AccentBorder returns a full-width line in accent color.
func AccentBorder(width int) string {
	if width < 4 {
		width = 40
	}
	return lipgloss.NewStyle().Foreground(AccentDim).Render(repeat("─", width))
}

func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}
