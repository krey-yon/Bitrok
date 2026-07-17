package ui

import "github.com/charmbracelet/lipgloss"

// Bitrok color palette — matched to the web dashboard.
// Primary accent is amber-400 on near-black; no cyan (web has none).
var (
	Amber      = lipgloss.Color("#fbbf24") // amber-400 — primary accent
	AmberLight = lipgloss.Color("#fcd34d") // amber-300 — highlight/hover
	AmberDim   = lipgloss.Color("#b45309") // amber-700 — muted accent
	Bg         = lipgloss.Color("#0a0a0a") // near-black background
	BgCard     = lipgloss.Color("#171717") // card surface
	White      = lipgloss.Color("#ededed") // off-white foreground
	LightGray  = lipgloss.Color("#a3a3a3") // neutral-400
	Gray       = lipgloss.Color("#737373") // neutral-500 — muted
	DarkGray   = lipgloss.Color("#404040") // neutral-700 — borders
	Green      = lipgloss.Color("#22c55e") // green-500 — up/success
	Red        = lipgloss.Color("#ef4444") // red-500 — down/error
)
