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
	Name         string
	Color        color.RGBA
	MaxSpeed     float64
	Acceleration float64
	BrakeForce   float64
	Handling     float64 // lateral movement multiplier
	MaxNitro     int
	UnlockScore  int
	Special      CarSpecial
}

var PlayerCars = [5]PlayerCarDef{
	{"STARTER", color.RGBA{0x00, 0xAA, 0xFF, 0xFF}, 10.0, 0.06, 0.08, 1.0, 1, 0, SpecialNone},
	{"SWIFT", color.RGBA{0x39, 0xFF, 0x14, 0xFF}, 11.5, 0.09, 0.10, 1.2, 2, 5000, SpecialNone},
	{"FURY", color.RGBA{0xFF, 0x14, 0x93, 0xFF}, 13.0, 0.04, 0.05, 0.8, 2, 20000, SpecialWiderNearMiss},
	{"PHANTOM", color.RGBA{0xFF, 0x66, 0x00, 0xFF}, 11.0, 0.06, 0.12, 1.1, 3, 50000, SpecialFuelSaver},
	{"GHOST", color.RGBA{0xEE, 0xEE, 0xFF, 0xFF}, 10.0, 0.07, 0.08, 1.0, 2, 100000, SpecialGhostShield},
}

var specialNames = map[CarSpecial]string{
	SpecialNone:          "-",
	SpecialWiderNearMiss: "Wide Near-Miss",
	SpecialFuelSaver:     "Fuel Saver",
	SpecialGhostShield:   "Ghost Shield",
}

// DrawGarage renders the garage screen.
func DrawGarage(screen *ebiten.Image, selectedIdx, section, trailIdx int, save *SaveData, sprites *SpriteCache) {
	screen.Fill(color.RGBA{0x0D, 0x0D, 0x1A, 0xFF})

	DrawText(screen, "G A R A G E", ScreenWidth/2-35, 30)

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
	DrawText(screen, "<", int(previewCX)-50, int(previewCY))
	DrawText(screen, ">", int(previewCX)+40, int(previewCY))

	// Car name.
	nameY := int(previewCY) + PlayerSpriteH + 10
	DrawText(screen, fmt.Sprintf("\"%s\"", car.Name), ScreenWidth/2-30, nameY)

	// Stats.
	statsY := nameY + 25
	drawStatBar(screen, "SPD", car.MaxSpeed, 13.0, 100, statsY)
	drawStatBar(screen, "ACC", car.Acceleration, 0.09, 100, statsY+16)
	drawStatBar(screen, "BRK", car.BrakeForce, 0.12, 100, statsY+32)
	drawStatBar(screen, "HND", car.Handling, 1.2, 100, statsY+48)
	drawStatBar(screen, "NOS", float64(car.MaxNitro)/3.0, 1.0, 100, statsY+64)
	DrawText(screen, fmt.Sprintf("SPC: %s", specialNames[car.Special]), 100, statsY+80)

	// Car unlock/select status.
	statusY := statsY + 100
	carSectionClr := color.RGBA{0x33, 0x33, 0x44, 0xFF}
	if section == 0 {
		carSectionClr = color.RGBA{0x44, 0x44, 0x66, 0xFF}
	}
	DrawRect(screen, 58, float64(statusY)-4, 284, 18, carSectionClr)
	if unlocked {
		if save.SelectedCar == selectedIdx {
			DrawText(screen, "[SELECTED]", ScreenWidth/2-32, statusY)
		} else {
			DrawText(screen, "[ENTER to SELECT]", ScreenWidth/2-52, statusY)
		}
	} else {
		DrawText(screen,
			fmt.Sprintf("LOCKED - %d pts", car.UnlockScore), 80, statusY)
	}

	// Trail section.
	trailY := statusY + 28
	trailSectionClr := color.RGBA{0x33, 0x33, 0x44, 0xFF}
	if section == 1 {
		trailSectionClr = color.RGBA{0x44, 0x44, 0x66, 0xFF}
	}
	DrawRect(screen, 58, float64(trailY)-4, 284, 34, trailSectionClr)
	DrawText(screen, "TRAIL:", 65, trailY)

	// Trail color swatches.
	swatchX := 120
	for i := range TrailDefs {
		sx := float64(swatchX + i*22)
		sy := float64(trailY - 1)
		if save.IsTrailUnlocked(i) {
			DrawRect(screen, sx, sy, 16, 12, TrailDefs[i].Color)
		} else {
			DrawRect(screen, sx, sy, 16, 12, color.RGBA{0x33, 0x33, 0x33, 0xFF})
		}
		// Selection marker.
		if i == trailIdx && section == 1 {
			DrawRect(screen, sx-1, sy-1, 18, 14, color.RGBA{0xFF, 0xFF, 0xFF, 0x88})
		}
		// Active marker.
		if i == save.SelectedTrail {
			DrawRect(screen, sx+5, sy+13, 6, 2, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
		}
	}

	// Trail name.
	trailDef := TrailDefs[trailIdx]
	trailStatus := trailDef.Name
	if !save.IsTrailUnlocked(trailIdx) {
		trailStatus = fmt.Sprintf("%s - %d pts", trailDef.Name, trailDef.UnlockScore)
	}
	DrawText(screen, trailStatus, 120, trailY+16)

	// Total score.
	DrawText(screen,
		fmt.Sprintf("Total Score: %d", save.TotalScore),
		ScreenWidth/2-55, ScreenHeight-60)

	DrawText(screen, "L/R browse  U/D section  Esc back", 50, ScreenHeight-30)
}

func drawStatBar(screen *ebiten.Image, label string, value, maxVal float64, x, y int) {
	DrawText(screen, label+":", x, y)
	barX := float64(x + 35)
	barW := 100.0
	DrawRect(screen, barX, float64(y+2), barW, 8, color.RGBA{0x33, 0x33, 0x33, 0xFF})
	fill := (value / maxVal) * barW
	if fill > barW {
		fill = barW
	}
	DrawRect(screen, barX, float64(y+2), fill, 8, color.RGBA{0x00, 0xCC, 0x00, 0xFF})
}
