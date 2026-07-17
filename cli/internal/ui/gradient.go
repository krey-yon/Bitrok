package ui

import "fmt"

// rgb is a truecolor RGB triple.
type rgb struct{ r, g, b uint8 }

// gradientStops sweeps dim → primary → bright amber, matching the web's amber glow.
var gradientStops = []rgb{
	{0xb4, 0x53, 0x09}, // amber-700 — dim
	{0xfb, 0xbf, 0x24}, // amber-400 — primary
	{0xfc, 0xd3, 0x4d}, // amber-300 — bright
}

// lerp linearly interpolates between two uint8 channels.
func lerp(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + (float64(b)-float64(a))*t)
}

// pickColor returns the interpolated color at position t (0..1) across stops.
func pickColor(stops []rgb, t float64) rgb {
	if len(stops) == 1 {
		return stops[0]
	}
	if t <= 0 {
		return stops[0]
	}
	if t >= 1 {
		return stops[len(stops)-1]
	}
	seg := 1.0 / float64(len(stops)-1)
	idx := int(t / seg)
	if idx >= len(stops)-1 {
		idx = len(stops) - 2
	}
	localT := (t - float64(idx)*seg) / seg
	return rgb{
		r: lerp(stops[idx].r, stops[idx+1].r, localT),
		g: lerp(stops[idx].g, stops[idx+1].g, localT),
		b: lerp(stops[idx].b, stops[idx+1].b, localT),
	}
}

// GradientAmber renders s as a truecolor amber gradient (dim→bright) using
// raw ANSI escape codes. Self-contained — no external dependency.
func GradientAmber(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	n := len(runes)
	var b string
	for i, r := range runes {
		t := 0.0
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		c := pickColor(gradientStops, t)
		b += fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c", c.r, c.g, c.b, r)
	}
	return b + "\x1b[0m"
}
