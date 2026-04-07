package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// CarSpecial defines unique car abilities.
type CarSpecial int

const (
	SpecialNone CarSpecial = iota
	SpecialWiderNearMiss   // FURY: threshold 12 instead of 10
	SpecialFuelSaver       // PHANTOM: fuel consumption * 0.8
	SpecialGhostShield     // GHOST: first collision is free
)

// PlayerCarDef defines a player car's properties.
type PlayerCarDef struct {
	Name        string
	Color       color.RGBA
	SpeedMod    float64    // multiplier for SpeedIncrement
	ManeuverMod float64    // multiplier for move speed and acceleration
	MaxNitro    int
	UnlockScore int
	Special     CarSpecial
}

var PlayerCars = [5]PlayerCarDef{
	{"STARTER", color.RGBA{0x00, 0xAA, 0xFF, 0xFF}, 1.0, 1.0, 1, 0, SpecialNone},
	{"SWIFT", color.RGBA{0x39, 0xFF, 0x14, 0xFF}, 1.15, 1.2, 2, 5000, SpecialNone},
	{"FURY", color.RGBA{0xFF, 0x14, 0x93, 0xFF}, 1.3, 0.9, 2, 20000, SpecialWiderNearMiss},
	{"PHANTOM", color.RGBA{0xFF, 0x66, 0x00, 0xFF}, 1.1, 1.1, 3, 50000, SpecialFuelSaver},
	{"GHOST", color.RGBA{0xEE, 0xEE, 0xFF, 0xFF}, 1.0, 1.0, 2, 100000, SpecialGhostShield},
}

var specialNames = map[CarSpecial]string{
	SpecialNone:          "-",
	SpecialWiderNearMiss: "Wide Near-Miss",
	SpecialFuelSaver:     "Fuel Saver",
	SpecialGhostShield:   "Ghost Shield",
}

// DrawGarage renders the garage screen.
func DrawGarage(screen *ebiten.Image, selectedIdx int, save *SaveData, sprites *SpriteCache) {
	screen.Fill(color.RGBA{0x0D, 0x0D, 0x1A, 0xFF})

	DebugPrintScaled(screen, "G A R A G E", ScreenWidth/2-35, 30)

	car := PlayerCars[selectedIdx]
	unlocked := save.IsCarUnlocked(selectedIdx)

	// Car preview (vector sprite, displayed at 2x logical size).
	previewCX := float64(ScreenWidth) / 2
	previewCY := 80.0 + float64(PlayerSpriteH)
	if unlocked {
		rs := renderScaleGlobal
		op := &ebiten.DrawImageOptions{}
		img := sprites.PlayerCars[selectedIdx]
		iw, ih := img.Bounds().Dx(), img.Bounds().Dy()
		// Show at 2x logical size, scaled to render resolution.
		previewScale := 2.0 * rs / SpriteScale
		op.GeoM.Translate(-float64(iw)/2, -float64(ih)/2)
		op.GeoM.Scale(previewScale, previewScale)
		op.GeoM.Translate(previewCX*rs, previewCY*rs)
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(img, op)
	} else {
		previewW, previewH := 60.0, 100.0
		DrawRect(screen, previewCX-previewW/2, previewCY-previewH/2, previewW, previewH,
			color.RGBA{0x44, 0x44, 0x44, 0xFF})
	}

	// Arrows.
	DebugPrintScaled(screen, "<", int(previewCX)-50, int(previewCY))
	DebugPrintScaled(screen, ">", int(previewCX)+40, int(previewCY))

	// Car name.
	nameY := int(previewCY) + PlayerSpriteH + 10
	DebugPrintScaled(screen, fmt.Sprintf("\"%s\"", car.Name), ScreenWidth/2-30, nameY)

	// Stats.
	statsY := nameY + 25
	drawStatBar(screen, "SPD", car.SpeedMod, 1.3, 100, statsY)
	drawStatBar(screen, "MGN", car.ManeuverMod, 1.2, 100, statsY+16)
	drawStatBar(screen, "NOS", float64(car.MaxNitro)/3.0, 1.0, 100, statsY+32)
	DebugPrintScaled(screen, fmt.Sprintf("SPC: %s", specialNames[car.Special]), 100, statsY+48)

	// Unlock status.
	statusY := statsY + 72
	if unlocked {
		if save.SelectedCar == selectedIdx {
			DebugPrintScaled(screen, "[SELECTED]", ScreenWidth/2-32, statusY)
		} else {
			DebugPrintScaled(screen, "[ENTER to SELECT]", ScreenWidth/2-52, statusY)
		}
	} else {
		DebugPrintScaled(screen,
			fmt.Sprintf("LOCKED - %d pts required", car.UnlockScore),
			60, statusY)
	}

	// Total score.
	DebugPrintScaled(screen,
		fmt.Sprintf("Total Score: %d", save.TotalScore),
		ScreenWidth/2-55, ScreenHeight-60)

	DebugPrintScaled(screen, "Left/Right - browse   Esc - back", 50, ScreenHeight-30)
}

func drawStatBar(screen *ebiten.Image, label string, value, maxVal float64, x, y int) {
	DebugPrintScaled(screen, label+":", x, y)
	barX := float64(x + 35)
	barW := 100.0
	DrawRect(screen, barX, float64(y+2), barW, 8, color.RGBA{0x33, 0x33, 0x33, 0xFF})
	fill := (value / maxVal) * barW
	if fill > barW {
		fill = barW
	}
	DrawRect(screen, barX, float64(y+2), fill, 8, color.RGBA{0x00, 0xCC, 0x00, 0xFF})
}
