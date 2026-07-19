package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Banner is the raw ASCII logo (no color) — safe for cobra help text.
const Banner = `
  ██████╗ ██╗████████╗██████╗  ██████╗ ██╗  ██╗
  ██╔══██╗██║╚══██╔══╝██╔══██╗██╔═══██╗██║ ██╔╝
  ██████╔╝██║   ██║   ██████╔╝██║   ██║█████╔╝
  ██╔══██╗██║   ██║   ██╔══██╗██║   ██║██╔═██╗
  ██████╔╝██║   ██║   ██╔  ██║╚██████╔╝██║  ██╗
  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝
`

// logoLines are the banner art lines used for the colored render + animation.
var logoLines = []string{
	"  ██████╗ ██╗████████╗██████╗  ██████╗ ██╗  ██╗",
	"  ██╔══██╗██║╚══██╔══╝██╔══██╗██╔═══██╗██║ ██╔╝",
	"  ██████╔╝██║   ██║   ██████╔╝██║   ██║█████╔╝",
	"  ██╔══██╗██║   ██║   ██╔══██╗██║   ██║██╔═██╗",
	"  ██████╔╝██║   ██║   ██╔  ██║╚██████╔╝██║  ██╗",
	"  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝",
}

// tagline renders the version + tagline line in the bitrok palette.
func tagline(version string) string {
	return fmt.Sprintf("  %s %s %s",
		lipgloss.NewStyle().Bold(true).Foreground(AccentLight).Render(version),
		lipgloss.NewStyle().Foreground(DarkGray).Render("—"),
		lipgloss.NewStyle().Foreground(LightGray).Render("deterministic tunnels, zero bullshit"))
}

// PrintBanner outputs the gradient-colored logo with a tagline (instant).
func PrintBanner(version string) {
	for _, line := range logoLines {
		fmt.Println(GradientAmber(line))
	}
	fmt.Println()
	fmt.Println(tagline(version))
	fmt.Println()
}

// PrintBootBanner animates the logo drawing in line-by-line, then the tagline.
// Total runtime ~330ms. Use for high-intent moments (e.g. `bitrok up`).
func PrintBootBanner(version string) {
	for _, line := range logoLines {
		fmt.Println(GradientAmber(line))
		time.Sleep(55 * time.Millisecond)
	}
	fmt.Println()
	fmt.Println(tagline(version))
	fmt.Println()
}

// PrintAnimatedBootBanner draws the logo line-by-line, then blinks the Bit
// mascot in next to the tagline. ~1.3s total. Falls back to the instant
// PrintBanner when animation is disabled (non-TTY / NO_COLOR / --no-anim).
func PrintAnimatedBootBanner(version string) {
	if !AnimationsEnabled() {
		PrintBanner(version)
		return
	}

	// 1. Logo draws in line-by-line.
	for _, line := range logoLines {
		fmt.Println(GradientAmber(line))
		time.Sleep(45 * time.Millisecond)
	}
	fmt.Println()

	// 2. Mascot blinks in: ears → eyes → full → blink → full.
	mascotFrames := [][]string{
		{"  /\\_/\\  "},
		{"  /\\_/\\  ", " ( o.o ) "},
		{"  /\\_/\\  ", " ( o.o ) ", "  > ^ < "},
		{"  /\\_/\\  ", " ( -.- ) ", "  > ^ < "},
		{"  /\\_/\\  ", " ( ^.^ ) ", "  > w < "},
	}
	for _, f := range mascotFrames {
		clearLines(len(f))
		style := lipgloss.NewStyle().Foreground(AccentLight)
		for _, line := range f {
			fmt.Println("  " + style.Render(line))
		}
		time.Sleep(90 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println(tagline(version))
	fmt.Println()
}
