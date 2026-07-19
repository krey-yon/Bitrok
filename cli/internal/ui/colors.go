package ui

import "github.com/charmbracelet/lipgloss"

// Bitrok CLI palette — matched to the web dashboard "Signal System"
// (dark mode): acid-lime accent, safety-orange secondary, near-black surfaces.
//
// Web tokens (globals.css .dark):
//   --accent: #b8f34a   --secondary: #ff7043
//   --background: #0b0d0a   --card: #12150f   --foreground: #f2f3ea
//   --success: #63d68b   --warning: #ffbd5a   --danger: #ff6b6b
var (
	// Accent (acid lime) — primary brand signal
	Accent      = lipgloss.Color("#b8f34a")
	AccentSoft  = lipgloss.Color("#c7fa69")
	AccentLight = lipgloss.Color("#d2ff7e")
	AccentDim   = lipgloss.Color("#77a900")
	AccentFg    = lipgloss.Color("#111600") // text on accent fills

	// Secondary (safety orange)
	Secondary      = lipgloss.Color("#ff7043")
	SecondaryLight = lipgloss.Color("#ff8b66")

	// Surfaces
	Bg       = lipgloss.Color("#0b0d0a")
	BgCard   = lipgloss.Color("#12150f")
	BgAlt    = lipgloss.Color("#171a13")
	BgSelect = lipgloss.Color("#1a1f14")

	// Text
	White     = lipgloss.Color("#f2f3ea") // foreground
	LightGray = lipgloss.Color("#a4aa9b") // muted-foreground
	Gray      = lipgloss.Color("#565b50") // muted
	DarkGray  = lipgloss.Color("#34392e") // border
	Hairline  = lipgloss.Color("#262a22")

	// Semantic
	Green  = lipgloss.Color("#63d68b") // success
	Red    = lipgloss.Color("#ff6b6b") // danger
	Yellow = lipgloss.Color("#ffbd5a") // warning

	// ── Compatibility aliases (amber → lime) ──────────────────────────
	// Existing code still references Amber*; map them so nothing breaks
	// while components migrate to Accent*.
	Amber      = Accent
	AmberLight = AccentLight
	AmberDim   = AccentDim
)
