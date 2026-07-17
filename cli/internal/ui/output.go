package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Output kit — the single source of truth for CLI output styling.
// Every command renders through these helpers so the CLI feels like
// one product, not thirteen.

// KVRow is a label/value pair for detail cards.
type KVRow struct {
	Label string
	Value string
}

// Success prints a green ✓ + bold message.
func Success(msg string) {
	icon := lipgloss.NewStyle().Foreground(Green).Bold(true).Render("✓")
	fmt.Printf("  %s %s\n", icon, lipgloss.NewStyle().Foreground(White).Bold(true).Render(msg))
}

// Warn prints an amber ! + message.
func Warn(msg string) {
	icon := lipgloss.NewStyle().Foreground(Amber).Bold(true).Render("!")
	fmt.Printf("  %s %s\n", icon, lipgloss.NewStyle().Foreground(AmberLight).Render(msg))
}

// ErrorOut prints a red ✗ + message to stderr.
func ErrorOut(msg string) {
	icon := lipgloss.NewStyle().Foreground(Red).Bold(true).Render("✗")
	fmt.Fprintf(os.Stderr, "  %s %s\n", icon, lipgloss.NewStyle().Foreground(Red).Render(msg))
}

// Info prints a gray informational line.
func Info(msg string) {
	fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(Gray).Render(msg))
}

// Hint prints a dim hint (for follow-up actions, keybindings).
func Hint(msg string) {
	fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(DarkGray).Render(msg))
}

// Section prints a gradient amber section header, preceded by a blank line.
func Section(title string) {
	fmt.Println()
	fmt.Printf("  %s\n", GradientAmber(strings.ToUpper(title)))
}

// KV renders a single label/value row, label right-aligned in a fixed column.
func KV(label, value string) string {
	return fmt.Sprintf("  %s  %s",
		lipgloss.NewStyle().Width(10).Foreground(Gray).Render(label),
		lipgloss.NewStyle().Foreground(White).Render(value))
}

// DetailCard renders a bordered card with a title and KV rows.
func DetailCard(title string, rows []KVRow) string {
	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(AmberLight).Render(title))
	lines = append(lines, lipgloss.NewStyle().Foreground(DarkGray).Render(strings.Repeat("─", 38)))
	for _, r := range rows {
		lines = append(lines, KV(r.Label, r.Value))
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Amber).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
}

// Confirm asks a yes/no question and returns the answer.
func Confirm(msg string) bool {
	icon := lipgloss.NewStyle().Foreground(Amber).Bold(true).Render("?")
	prompt := lipgloss.NewStyle().Foreground(White).Render(msg)
	fmt.Printf("  %s %s [%s/%s] ",
		icon, prompt,
		lipgloss.NewStyle().Foreground(Green).Bold(true).Render("y"),
		lipgloss.NewStyle().Foreground(Gray).Render("N"))
	var resp string
	fmt.Scanln(&resp)
	return strings.HasPrefix(strings.ToLower(resp), "y")
}
