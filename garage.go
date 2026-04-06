package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
func DrawGarage(screen *ebiten.Image, selectedIdx int, save *SaveData) {
	screen.Fill(color.RGBA{0x0D, 0x0D, 0x1A, 0xFF})

	ebitenutil.DebugPrintAt(screen, "G A R A G E", ScreenWidth/2-35, 30)

	car := PlayerCars[selectedIdx]
	unlocked := save.IsCarUnlocked(selectedIdx)

	// Car preview (large rectangle).
	previewW, previewH := 60.0, 100.0
	px := float64(ScreenWidth)/2 - previewW/2
	py := 80.0
	previewClr := car.Color
	if !unlocked {
		previewClr = color.RGBA{0x44, 0x44, 0x44, 0xFF}
	}
	DrawRect(screen, px, py, previewW, previewH, previewClr)

	// Arrows.
	ebitenutil.DebugPrintAt(screen, "<", int(px)-20, int(py)+40)
	ebitenutil.DebugPrintAt(screen, ">", int(px+previewW)+10, int(py)+40)

	// Car name.
	nameY := int(py + previewH + 15)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("\"%s\"", car.Name), ScreenWidth/2-30, nameY)

	// Stats.
	statsY := nameY + 25
	drawStatBar(screen, "SPD", car.SpeedMod, 1.3, 100, statsY)
	drawStatBar(screen, "MGN", car.ManeuverMod, 1.2, 100, statsY+16)
	drawStatBar(screen, "NOS", float64(car.MaxNitro)/3.0, 1.0, 100, statsY+32)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SPC: %s", specialNames[car.Special]), 100, statsY+48)

	// Unlock status.
	statusY := statsY + 72
	if unlocked {
		if save.SelectedCar == selectedIdx {
			ebitenutil.DebugPrintAt(screen, "[SELECTED]", ScreenWidth/2-32, statusY)
		} else {
			ebitenutil.DebugPrintAt(screen, "[ENTER to SELECT]", ScreenWidth/2-52, statusY)
		}
	} else {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("LOCKED - %d pts required", car.UnlockScore),
			60, statusY)
	}

	// Total score.
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Total Score: %d", save.TotalScore),
		ScreenWidth/2-55, ScreenHeight-60)

	ebitenutil.DebugPrintAt(screen, "Left/Right - browse   Esc - back", 50, ScreenHeight-30)
}

func drawStatBar(screen *ebiten.Image, label string, value, maxVal float64, x, y int) {
	ebitenutil.DebugPrintAt(screen, label+":", x, y)
	barX := float64(x + 35)
	barW := 100.0
	DrawRect(screen, barX, float64(y+2), barW, 8, color.RGBA{0x33, 0x33, 0x33, 0xFF})
	fill := (value / maxVal) * barW
	if fill > barW {
		fill = barW
	}
	DrawRect(screen, barX, float64(y+2), fill, 8, color.RGBA{0x00, 0xCC, 0x00, 0xFF})
}
