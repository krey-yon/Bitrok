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
	icon := lipgloss.NewStyle().Foreground(Green).Bold(true).Render(IconCheck)
	fmt.Printf("  %s %s\n", icon, lipgloss.NewStyle().Foreground(White).Bold(true).Render(msg))
}

// Warn prints an orange ! + message.
func Warn(msg string) {
	icon := lipgloss.NewStyle().Foreground(Secondary).Bold(true).Render(IconWarn)
	fmt.Printf("  %s %s\n", icon, lipgloss.NewStyle().Foreground(SecondaryLight).Render(msg))
}

// ErrorOut prints a red ✗ + message to stderr.
func ErrorOut(msg string) {
	icon := lipgloss.NewStyle().Foreground(Red).Bold(true).Render(IconCross)
	fmt.Fprintf(os.Stderr, "  %s %s\n", icon, lipgloss.NewStyle().Foreground(Red).Render(msg))
}

// Info prints a muted informational line.
func Info(msg string) {
	fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(LightGray).Render(msg))
}

// Hint prints a dim hint (for follow-up actions, keybindings).
func Hint(msg string) {
	fmt.Printf("  %s\n", lipgloss.NewStyle().Foreground(Gray).Render(msg))
}

// Section prints a gradient lime section header, preceded by a blank line.
func Section(title string) {
	fmt.Println()
	fmt.Printf("  %s\n", GradientAccent(strings.ToUpper(title)))
	fmt.Printf("  %s\n", BorderLine(40))
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
	lines = append(lines, lipgloss.NewStyle().Bold(true).Foreground(AccentLight).Render(title))
	lines = append(lines, lipgloss.NewStyle().Foreground(DarkGray).Render(strings.Repeat("─", 38)))
	for _, r := range rows {
		lines = append(lines, KV(r.Label, r.Value))
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Accent).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
}

// Confirm asks a yes/no question and returns the answer.
func Confirm(msg string) bool {
	icon := lipgloss.NewStyle().Foreground(Accent).Bold(true).Render(IconPrompt)
	prompt := lipgloss.NewStyle().Foreground(White).Render(msg)
	fmt.Printf("  %s %s [%s/%s] ",
		icon, prompt,
		lipgloss.NewStyle().Foreground(Green).Bold(true).Render("y"),
		lipgloss.NewStyle().Foreground(Gray).Render("N"))
	var resp string
	fmt.Scanln(&resp)
	return strings.HasPrefix(strings.ToLower(resp), "y")
}

// Prompt asks a question and reads a single line of input, returning the trim.
func Prompt(label, dflt string) string {
	icon := lipgloss.NewStyle().Foreground(Accent).Bold(true).Render(IconPrompt)
	prompt := lipgloss.NewStyle().Foreground(White).Render(label)
	if dflt != "" {
		fmt.Printf("  %s %s [%s] ",
			icon, prompt,
			lipgloss.NewStyle().Foreground(Gray).Render(dflt))
	} else {
		fmt.Printf("  %s %s ", icon, prompt)
	}
	var resp string
	fmt.Scanln(&resp)
	resp = strings.TrimSpace(resp)
	if resp == "" {
		return dflt
	}
	return resp
}

// RevealURL prints the public URL → local address line with a flourish.
func RevealURL(pubURL, localAddr string, copied bool) {
	fmt.Println()
	fmt.Printf("  %s %s %s %s\n",
		lipgloss.NewStyle().Foreground(Accent).Bold(true).Render(IconGlobe),
		lipgloss.NewStyle().Foreground(AccentLight).Bold(true).Underline(true).Render(pubURL),
		lipgloss.NewStyle().Foreground(DarkGray).Render(IconArrow),
		lipgloss.NewStyle().Foreground(White).Render(localAddr))
	if copied {
		fmt.Printf("  %s %s\n",
			lipgloss.NewStyle().Foreground(Gray).Render(IconCopy),
			lipgloss.NewStyle().Foreground(Gray).Render("copied to clipboard"))
	}
	fmt.Println()
}
