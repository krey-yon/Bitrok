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
		lipgloss.NewStyle().Bold(true).Foreground(AmberLight).Render(version),
		lipgloss.NewStyle().Foreground(DarkGray).Render("—"),
		lipgloss.NewStyle().Foreground(Gray).Render("deterministic tunnels, zero bullshit"))
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
