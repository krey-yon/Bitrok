package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// BootStep is one line of the startup sequence.
type BootStep struct {
	Label    string
	Duration time.Duration // how long the "loading" phase animates before flipping to done
}

// BootSequence prints an animated multi-step startup. Each step spins briefly,
// then marks itself as done with a green check. Total runtime is the sum of
// step durations. Fire-and-forget: this blocks the caller.
//
// Non-goal: this is intentionally fake-timed — the CLI can't actually know
// when the WebSocket handshake completes from here. The point is delight, not
// truth. Real state updates happen in the dashboard after this returns.
func BootSequence(steps []BootStep) {
	spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	spinStyle := lipgloss.NewStyle().Foreground(Amber)
	checkStyle := lipgloss.NewStyle().Foreground(Green).Bold(true)
	labelDone := lipgloss.NewStyle().Foreground(White)
	labelPending := lipgloss.NewStyle().Foreground(Gray)

	// Track completed lines so they persist above the current spinning line.
	var completed []string

	for _, step := range steps {
		start := time.Now()
		frameIdx := 0

		// Animate this step for its Duration
		for time.Since(start) < step.Duration {
			// Redraw completed lines + current spinning line
			// Move cursor up by (completed + 1) and clear each
			clearLines(len(completed) + 1)
			for _, done := range completed {
				fmt.Println(done)
			}
			fmt.Printf("  %s %s\n",
				spinStyle.Render(spinChars[frameIdx%len(spinChars)]),
				labelPending.Render(step.Label))
			frameIdx++
			time.Sleep(80 * time.Millisecond)
		}

		// Mark this step done — replace the spinning line with a check
		completed = append(completed, fmt.Sprintf("  %s %s",
			checkStyle.Render("✓"),
			labelDone.Render(step.Label)))
	}

	// Final render — all lines with checks
	clearLines(len(completed) + 1) // clear the last still-showing line + steps
	for _, done := range completed {
		fmt.Println(done)
	}
}

// clearLines moves the cursor up n lines and clears them.
func clearLines(n int) {
	if n <= 0 {
		return
	}
	// Move up n lines, clear each line down
	fmt.Printf("\x1b[%dA", n)
	for i := 0; i < n; i++ {
		fmt.Print("\x1b[2K") // clear entire line
		if i < n-1 {
			fmt.Print("\x1b[1B") // move down one
		}
	}
	// Move back up to the start
	if n > 1 {
		fmt.Printf("\x1b[%dA", n-1)
	}
	fmt.Print("\r")
}

// DefaultBootSteps returns the standard boot sequence for `bitrok up`.
func DefaultBootSteps(tunnelName string) []BootStep {
	return []BootStep{
		{Label: "Loading local registry", Duration: 250 * time.Millisecond},
		{Label: fmt.Sprintf("Resolving tunnel %s", strings.TrimSpace(tunnelName)), Duration: 350 * time.Millisecond},
		{Label: "Authenticating with server", Duration: 400 * time.Millisecond},
		{Label: fmt.Sprintf("Waking Bit — %s", PetGreeting()), Duration: 400 * time.Millisecond},
	}
}
