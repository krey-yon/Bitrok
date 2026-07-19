package ui

import "fmt"

// rgb is a truecolor RGB triple.
type rgb struct{ r, g, b uint8 }

// gradientStops sweeps dim lime → primary → bright, matching the dashboard accent.
var gradientStops = []rgb{
	{0x4c, 0x72, 0x00}, // ring / deep lime
	{0xb8, 0xf3, 0x4a}, // accent #b8f34a
	{0xd2, 0xff, 0x7e}, // accent-light
}

// secondaryStops for orange highlights (secondary brand).
var secondaryStops = []rgb{
	{0xd9, 0x4d, 0x21},
	{0xff, 0x70, 0x43},
	{0xff, 0x8b, 0x66},
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

// GradientAccent renders s as a truecolor lime gradient (dim→bright).
func GradientAccent(s string) string {
	return gradientWith(s, gradientStops)
}

// GradientAmber is kept as an alias for call sites that still use the old name.
func GradientAmber(s string) string {
	return GradientAccent(s)
}

// GradientSecondary renders s with the orange secondary gradient.
func GradientSecondary(s string) string {
	return gradientWith(s, secondaryStops)
}

func gradientWith(s string, stops []rgb) string {
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
		c := pickColor(stops, t)
		b += fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c", c.r, c.g, c.b, r)
	}
	return b + "\x1b[0m"
}
