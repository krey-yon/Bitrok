package ui

import "github.com/charmbracelet/lipgloss"

// Common Lipgloss styles — matched to the web dashboard palette.
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Accent)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(AccentLight)

	LabelStyle = lipgloss.NewStyle().
			Foreground(Gray)

	ValueStyle = lipgloss.NewStyle().
			Foreground(White)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Green)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Red)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Accent).
			Padding(1, 2)
)
