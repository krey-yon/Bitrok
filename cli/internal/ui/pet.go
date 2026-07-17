package ui

import (
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Pet is a small animated ASCII mascot that lives in the dashboard footer.
// The bitrok mascot is "Bit" — a tiny fox that idles, blinks, and reacts to traffic.
//
// The pet is stateless per-render: given the current tick + traffic state,
// it returns the appropriate frame. All timing is derived from tick counts,
// so the caller drives animation via bubbletea Tick messages.

// PetMood controls which frame set the pet uses.
type PetMood int

const (
	PetIdle    PetMood = iota // slow breathing, occasional blink
	PetHappy                  // reacts to a request landing
	PetAlert                  // reacts to an error / 5xx
)

// petFramesIdle is the pet's default breathing loop.
// Frame widths must match so the surrounding layout doesn't jitter.
var petFramesIdle = []string{
	" /\\_/\\   ",
	"( o.o )  ",
	" > ^ <   ",
}

var petFramesBlink = []string{
	" /\\_/\\   ",
	"( -.- )  ",
	" > ^ <   ",
}

var petFramesHappy = []string{
	" /\\_/\\   ",
	"( ^.^ )  ",
	" > w <   ",
}

var petFramesAlert = []string{
	" /\\_/\\   ",
	"( O.O )! ",
	" > ! <   ",
}

// RenderPet returns the multi-line pet frame for the current mood + tick.
// Tick is a monotonically-increasing counter (e.g. seconds since start).
func RenderPet(mood PetMood, tick int) string {
	var frames []string
	switch mood {
	case PetHappy:
		frames = petFramesHappy
	case PetAlert:
		frames = petFramesAlert
	default:
		// Idle: blink every ~7 ticks for a single tick
		if tick%7 == 0 {
			frames = petFramesBlink
		} else {
			frames = petFramesIdle
		}
	}

	style := lipgloss.NewStyle().Foreground(AmberLight)
	if mood == PetAlert {
		style = lipgloss.NewStyle().Foreground(Red).Bold(true)
	} else if mood == PetHappy {
		style = lipgloss.NewStyle().Foreground(Green).Bold(true)
	}

	var lines []string
	for _, f := range frames {
		lines = append(lines, style.Render(f))
	}
	return strings.Join(lines, "\n")
}

// PetGreeting returns a random cute welcome line for the boot sequence.
func PetGreeting() string {
	greetings := []string{
		"purring the packets…",
		"chasing the ports…",
		"pouncing on requests…",
		"licking latency low…",
		"grooming the tunnel…",
		"stretching before the run…",
	}
	// Seed with time so consecutive runs vary
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return greetings[r.Intn(len(greetings))]
}
