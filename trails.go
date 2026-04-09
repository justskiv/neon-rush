package main

import (
	"image/color"
	"math"
)

// TrailDef defines an exhaust trail color variant.
type TrailDef struct {
	Name        string
	Color       color.RGBA
	UnlockScore int
}

// TrailDefs lists all available exhaust trail colors.
var TrailDefs = [7]TrailDef{
	{"DEFAULT", color.RGBA{0xCC, 0x99, 0x55, 0xFF}, 0},
	{"TOXIC", color.RGBA{0x39, 0xFF, 0x14, 0xFF}, 2000},
	{"ICE", color.RGBA{0x00, 0xDD, 0xFF, 0xFF}, 6000},
	{"ROYAL", color.RGBA{0xAA, 0x00, 0xFF, 0xFF}, 15000},
	{"SOLAR", color.RGBA{0xFF, 0xAA, 0x00, 0xFF}, 30000},
	{"VOID", color.RGBA{0x44, 0x00, 0x00, 0xFF}, 60000},
	{"RAINBOW", color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}, 120000},
}

// TrailColor returns the trail color for a given index and tick.
// VOID pulses red, RAINBOW cycles through hues.
func TrailColor(idx, tick int) color.RGBA {
	if idx < 0 || idx >= len(TrailDefs) {
		idx = 0
	}
	switch idx {
	case 5: // VOID: dark with red pulse
		pulse := uint8(0x22 + int(0x33*math.Abs(math.Sin(float64(tick)*0.08))))
		return color.RGBA{pulse, 0x00, 0x00, 0xFF}
	case 6: // RAINBOW: hue cycle
		hue := (tick * 3) % 360
		return hueToRGB(hue)
	default:
		return TrailDefs[idx].Color
	}
}

// hueToRGB converts a hue (0-359) to an RGB color (full saturation).
func hueToRGB(hue int) color.RGBA {
	h := float64(hue%360) / 60.0
	i := int(h)
	f := h - float64(i)
	q := uint8(255 * (1 - f))
	t := uint8(255 * f)
	switch i {
	case 0:
		return color.RGBA{0xFF, t, 0x00, 0xFF}
	case 1:
		return color.RGBA{q, 0xFF, 0x00, 0xFF}
	case 2:
		return color.RGBA{0x00, 0xFF, t, 0xFF}
	case 3:
		return color.RGBA{0x00, q, 0xFF, 0xFF}
	case 4:
		return color.RGBA{t, 0x00, 0xFF, 0xFF}
	default:
		return color.RGBA{0xFF, 0x00, q, 0xFF}
	}
}
