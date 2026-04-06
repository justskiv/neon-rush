package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var colorOverlay = color.RGBA{0, 0, 0, 160}

// HUDData contains all data the HUD needs to render.
type HUDData struct {
	Score           int
	ScrollSpeed     float64
	Fuel            float64
	NitroCharges    int
	NitroActive     bool
	ComboMultiplier int
	ComboTimer      int
}

// DrawHUD renders score, speed, fuel bar, combo, and nitro on screen.
func DrawHUD(screen *ebiten.Image, data HUDData) {
	// Semi-transparent top bar.
	DrawRect(screen, 0, 0, ScreenWidth, 42, color.RGBA{0, 0, 0, 120})

	speedKmh := int(data.ScrollSpeed * 30)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SCORE: %d", data.Score), 10, 4)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SPD: %d", speedKmh), 10, 20)

	// Combo.
	if data.ComboMultiplier > 1 {
		comboText := fmt.Sprintf("x%d COMBO", data.ComboMultiplier)
		cx := ScreenWidth/2 - 30
		// Pulse: dim combo bar behind text.
		barAlpha := uint8(180)
		if data.ComboTimer > 0 && data.ComboTimer%20 < 10 {
			barAlpha = 100
		}
		DrawRect(screen, float64(cx)-4, 2, float64(len(comboText))*6+8, 14,
			color.RGBA{0x20, 0x80, 0x10, barAlpha})
		ebitenutil.DebugPrintAt(screen, comboText, cx, 4)
	}

	// Fuel bar.
	barX, barY := float64(ScreenWidth-100), 6.0
	barW, barH := 80.0, 8.0
	DrawRect(screen, barX, barY, barW, barH, color.RGBA{0x33, 0x33, 0x33, 0xFF})
	fill := data.Fuel / FuelMax * barW
	if fill < 0 {
		fill = 0
	}
	fuelClr := color.RGBA{0x00, 0xCC, 0x00, 0xFF} // green
	if data.Fuel < 60 {
		fuelClr = color.RGBA{0xCC, 0xCC, 0x00, 0xFF} // yellow
	}
	if data.Fuel < 30 {
		fuelClr = color.RGBA{0xCC, 0x00, 0x00, 0xFF} // red
	}
	DrawRect(screen, barX, barY, fill, barH, fuelClr)
	ebitenutil.DebugPrintAt(screen, "FUEL", int(barX), 18)

	// Nitro charges.
	nitroX := int(barX) + 40
	for i := range data.NitroCharges {
		nx := float64(nitroX + i*10)
		DrawRect(screen, nx, 20, 6, 6, color.RGBA{0xFF, 0xDD, 0x00, 0xFF})
	}
	if data.NitroActive {
		ebitenutil.DebugPrintAt(screen, "NITRO!", int(barX)+40, 30)
	}
}

// GameOverData holds all info for the game over screen.
type GameOverData struct {
	Score          int
	HighScore      int
	IsNewHighScore bool
	NearMisses     int
	BestCombo      int
	Distance       float64
	ZoneName       string
	Selection      int // 0=RETRY, 1=MENU
	NewUnlocks     []string
	TotalScore     int
	Busted         bool
}

// DrawGameOver renders the game over overlay with stats.
func DrawGameOver(screen *ebiten.Image, data GameOverData) {
	DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180})

	cx := 100
	cy := 130
	title := "G A M E   O V E R"
	if data.Busted {
		title = "B U S T E D !"
	}
	ebitenutil.DebugPrintAt(screen, title, cx, cy)

	cy += 35
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score:       %d", data.Score), cx, cy)
	if data.IsNewHighScore {
		ebitenutil.DebugPrintAt(screen, "NEW!", cx+170, cy)
	}
	cy += 18
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best:        %d", data.HighScore), cx, cy)
	cy += 18
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Near Misses: %d", data.NearMisses), cx, cy)
	cy += 18
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Best Combo:  x%d", data.BestCombo), cx, cy)
	cy += 18
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Distance:    %.1f km", data.Distance), cx, cy)
	cy += 18
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Zone:        %s", data.ZoneName), cx, cy)

	// Selection.
	cy += 35
	retryMarker, menuMarker := "  ", "  "
	if data.Selection == 0 {
		retryMarker = "> "
	} else {
		menuMarker = "> "
	}
	ebitenutil.DebugPrintAt(screen, retryMarker+"RETRY", cx+20, cy)
	ebitenutil.DebugPrintAt(screen, menuMarker+"MENU", cx+120, cy)

	// Total score.
	cy += 30
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Total Score: %d", data.TotalScore), cx, cy)

	// Unlock notifications.
	for i, name := range data.NewUnlocks {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("\"%s\" unlocked!", name), cx, cy+18+i*16)
	}

	ebitenutil.DebugPrintAt(screen, "Left/Right - select   Enter - confirm", 40, ScreenHeight-30)
}
