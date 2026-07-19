package ui

import (
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

// Inline animated indicators (pi-animations-style), amber-themed.
// Frame-based, ANSI true color, gated by TTY + NO_COLOR + a global NoAnim flag
// the CLI can set from --no-anim. One-line, fixed-width bars — tidy inline, not
// full-screen. ponytail: faux fire (single-line flicker, not true demoscene
// propagation); upgrade to multi-row fire if we want the real thing.

// NoAnim is set by the CLI from --no-anim / config to disable all animation.
var NoAnim bool

// Animation renders a single frame of an inline indicator at the given frame
// index and column width. Returns the (ANSI-colored) string for that frame.
type Animation func(frame, width int) string

// AnimationsEnabled reports whether inline animation should run.
func AnimationsEnabled() bool {
	if NoAnim {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// RunAnimated runs fn while an inline animation plays next to label.
// When animation is disabled (or fn returns an error), it degrades to a plain
// spinner-less line. On success it prints ✓ label; on error ✗ label.
func RunAnimated(label string, anim Animation, fn func() error) error {
	if !AnimationsEnabled() {
		fmt.Printf("  %s\n", lipglossGray(label))
		err := fn()
		if err != nil {
			ErrorOut(label + ": " + err.Error())
			return err
		}
		Success(label)
		return nil
	}

	const barWidth = 22
	const fps = 30 * time.Millisecond

	stop := make(chan struct{})
	var doneMu sync.Mutex
	done := false

	go func() {
		frame := 0
		for {
			select {
			case <-stop:
				return
			default:
			}
			bar := anim(frame, barWidth)
			// \r to start of line, render: "  ⠿ label  <bar>   " padded to overwrite
			fmt.Printf("\r  %s %s  %s   ",
				amberBold("⠿"), lipglossGray(label), bar)
			frame++
			time.Sleep(fps)
			doneMu.Lock()
			if done {
				doneMu.Unlock()
				return
			}
			doneMu.Unlock()
		}
	}()

	err := fn()
	doneMu.Lock()
	done = true
	doneMu.Unlock()
	close(stop)

	// Clear the animation line and print the final state.
	fmt.Printf("\r\x1b[2K")
	if err != nil {
		ErrorOut(label + ": " + err.Error())
		return err
	}
	Success(label)
	return nil
}

// ── Fire ────────────────────────────────────────────────────────────────
// A single-line amber flicker: each column's intensity is a slow noise wave
// riding a baseline, rendered with ░▒▓█ at increasing amber brightness.
// Not true demoscene propagation (that needs multiple rows); reads as fire.

var fireChars = []rune("░▒▓█")

func Fire(frame, width int) string {
	var b strings.Builder
	for x := 0; x < width; x++ {
		// Two layered sines + a phase offset = organic flicker.
		t := float64(frame) * 0.18
		v := math.Sin(float64(x)*0.55+t)*0.5 + 0.5
		v2 := math.Sin(float64(x)*0.31-t*1.4)*0.5 + 0.5
		heat := v*0.6 + v2*0.4
		// Bias toward the center so the bar glows brighter in the middle.
		glow := 1.0 - math.Abs(float64(x)-float64(width)/2)/float64(width)
		heat = heat*0.7 + glow*0.3
		if heat > 1 {
			heat = 1
		}
		idx := int(heat * float64(len(fireChars)))
		if idx >= len(fireChars) {
			idx = len(fireChars) - 1
		}
		// Amber gradient: dim → bright as heat rises.
		c := pickColor(gradientStops, heat)
		b.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c", c.r, c.g, c.b, fireChars[idx]))
	}
	b.WriteString("\x1b[0m")
	return b.String()
}

// ── Plasma ──────────────────────────────────────────────────────────────
// A classic plasma band: sum of two sines, mapped to amber brightness + block
// density. Smoother than fire — used for the "connecting" moment.

var plasmaChars = []rune("░▒▓█")

func Plasma(frame, width int) string {
	var b strings.Builder
	for x := 0; x < width; x++ {
		t := float64(frame) * 0.12
		v := math.Sin(float64(x)*0.35+t) +
			math.Sin(float64(x)*0.13-t*1.7) +
			math.Sin(float64(x)*0.07+t*0.6)
		// v ranges roughly [-3, 3]; normalize to [0, 1].
		heat := (v + 3) / 6
		if heat < 0 {
			heat = 0
		} else if heat > 1 {
			heat = 1
		}
		idx := int(heat * float64(len(plasmaChars)))
		if idx >= len(plasmaChars) {
			idx = len(plasmaChars) - 1
		}
		c := pickColor(gradientStops, heat)
		b.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c", c.r, c.g, c.b, plasmaChars[idx]))
	}
	b.WriteString("\x1b[0m")
	return b.String()
}

// ── small style helpers (kept local so this file is self-contained) ──────

func amberBold(s string) string {
	// Accent lime #b8f34a — kept name for call sites.
	return fmt.Sprintf("\x1b[1;38;2;%d;%d;%dm%s\x1b[0m", 184, 243, 74, s)
}

func lipglossGray(s string) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", 86, 91, 80, s)
}
