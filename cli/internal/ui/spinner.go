package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BrailleFrames are the default spinner animation frames.
var BrailleFrames = spinner.Spinner{
	Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	FPS:    time.Second / 10,
}

// MoonFrames animates through the lunar phases (soft, playful).
var MoonFrames = spinner.Spinner{
	Frames: []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"},
	FPS:    time.Second / 8,
}

// PulseFrames animates a filled circle growing and shrinking (great for
// "connecting" states — matches web's amber glow aesthetic).
var PulseFrames = spinner.Spinner{
	Frames: []string{"·", "•", "●", "•", "·", " "},
	FPS:    time.Second / 8,
}

// BounceFrames — a bouncing dot, feels like a heartbeat.
var BounceFrames = spinner.Spinner{
	Frames: []string{"⠁", "⠂", "⠄", "⠂"},
	FPS:    time.Second / 8,
}

// NewSpinner creates a default braille spinner.
func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = BrailleFrames
	s.Style = lipgloss.NewStyle().Foreground(Amber)
	return s
}

// NewPulseSpinner returns a pulse-style spinner in amber.
func NewPulseSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = PulseFrames
	s.Style = lipgloss.NewStyle().Foreground(Amber).Bold(true)
	return s
}

// SpinnerModel is a simple tea.Model for showing a spinner with a message.
type SpinnerModel struct {
	Spinner spinner.Model
	Message string
	Done    bool
}

func NewSpinnerModel(msg string) SpinnerModel {
	return SpinnerModel{
		Spinner: NewSpinner(),
		Message: msg,
	}
}

func (m SpinnerModel) Init() tea.Cmd {
	return m.Spinner.Tick
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	default:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m SpinnerModel) View() string {
	if m.Done {
		return fmt.Sprintf("✓ %s\n", m.Message)
	}
	return fmt.Sprintf("%s %s\n", m.Spinner.View(), m.Message)
}
