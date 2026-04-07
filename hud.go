package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
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
	Damaged         bool
	RepairFlash     int // ticks remaining for repair flash
	TickCount       int
	Accelerating    bool
	Braking         bool
	DriftActive     bool
	DriftHeat       float64 // 0..1
	DriftOverheat   bool
}

// DrawHUD renders score, speed, fuel bar, combo, and nitro on screen.
func DrawHUD(screen *ebiten.Image, data HUDData) {
	// Semi-transparent top bar.
	DrawRect(screen, 0, 0, ScreenWidth, 42, color.RGBA{0, 0, 0, 120})

	speedKmh := int(data.ScrollSpeed * 30)
	DebugPrintScaled(screen, fmt.Sprintf("SCORE: %d", data.Score), 10, 4)

	speedLabel := fmt.Sprintf("SPD: %d", speedKmh)
	if data.Accelerating {
		speedLabel += " GAS"
	} else if data.Braking {
		speedLabel += " BRK"
	}
	DebugPrintScaled(screen, speedLabel, 10, 20)

	// Combo: text + decaying timer bar underneath.
	if data.ComboMultiplier > 1 {
		comboText := fmt.Sprintf("x%d COMBO", data.ComboMultiplier)
		cx := ScreenWidth/2 - 30
		DebugPrintScaled(screen, comboText, cx, 2)

		// Timer bar: shrinks from full to zero as combo decays.
		timerFrac := float64(data.ComboTimer) / float64(ComboDecayTicks)
		if timerFrac > 1 {
			timerFrac = 1
		}
		fullW := float64(len(comboText))*6 + 8
		barW := fullW * timerFrac
		// Color: green→yellow→red as timer runs out.
		var barClr color.RGBA
		switch {
		case timerFrac > 0.5:
			barClr = color.RGBA{0x20, 0xCC, 0x20, 0xCC}
		case timerFrac > 0.25:
			barClr = color.RGBA{0xCC, 0xCC, 0x00, 0xCC}
		default:
			barClr = color.RGBA{0xCC, 0x33, 0x00, 0xCC}
		}
		DrawRect(screen, float64(cx)-4, 16, barW, 3, barClr)
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
	DebugPrintScaled(screen, "FUEL", int(barX), 18)

	// Nitro charges.
	nitroX := int(barX) + 40
	for i := range data.NitroCharges {
		nx := float64(nitroX + i*10)
		DrawRect(screen, nx, 20, 6, 6, color.RGBA{0xFF, 0xDD, 0x00, 0xFF})
	}
	if data.NitroActive {
		DebugPrintScaled(screen, "NITRO!", int(barX)+40, 30)
	}

	// Damage indicator.
	if data.RepairFlash > 0 {
		DrawRect(screen, ScreenWidth/2-55, 44, 110, 14, color.RGBA{0x00, 0xDD, 0xFF, 0x50})
		DebugPrintScaled(screen, ">> REPAIRED <<", ScreenWidth/2-50, 46)
	} else if data.Damaged {
		DrawRect(screen, 8, 32, 78, 14, color.RGBA{0xFF, 0x22, 0x00, 0x88})
		DebugPrintScaled(screen, "! DAMAGED !", 10, 34)
	}

	// Drift heat bar (under the player car).
	if data.DriftActive || data.DriftHeat > 0.01 {
		bx := float64(ScreenWidth/2 - 20)
		by := float64(PlayerStartY + PlayerHeight/2 + 8)
		bw := 40.0
		bh := 4.0
		DrawRect(screen, bx, by, bw, bh, color.RGBA{0x33, 0x33, 0x33, 0xAA})
		fill := data.DriftHeat * bw
		heatClr := color.RGBA{0x00, 0xFF, 0x00, 0xCC}
		if data.DriftHeat > 0.6 {
			heatClr = color.RGBA{0xFF, 0xFF, 0x00, 0xCC}
		}
		if data.DriftHeat > 0.85 {
			heatClr = color.RGBA{0xFF, 0x33, 0x00, 0xCC}
		}
		DrawRect(screen, bx, by, fill, bh, heatClr)
	}
	if data.DriftOverheat {
		DebugPrintScaled(screen, "OVERHEAT!",
			ScreenWidth/2-28, PlayerStartY+PlayerHeight/2+14)
	} else if data.DriftActive {
		DebugPrintScaled(screen, "DRIFT",
			ScreenWidth/2-18, PlayerStartY+PlayerHeight/2+14)
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
	FuelEmpty      bool
}

// DrawGameOver renders the game over overlay with stats.
func DrawGameOver(screen *ebiten.Image, data GameOverData) {
	DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180})

	cx := 100
	cy := 130
	title := "G A M E   O V E R"
	if data.Busted {
		title = "B U S T E D !"
	} else if data.FuelEmpty {
		title = "O U T  O F  F U E L"
	}
	DebugPrintScaled(screen, title, cx, cy)

	cy += 35
	DebugPrintScaled(screen, fmt.Sprintf("Score:       %d", data.Score), cx, cy)
	if data.IsNewHighScore {
		DebugPrintScaled(screen, "NEW!", cx+170, cy)
	}
	cy += 18
	DebugPrintScaled(screen, fmt.Sprintf("Best:        %d", data.HighScore), cx, cy)
	cy += 18
	DebugPrintScaled(screen, fmt.Sprintf("Near Misses: %d", data.NearMisses), cx, cy)
	cy += 18
	DebugPrintScaled(screen, fmt.Sprintf("Best Combo:  x%d", data.BestCombo), cx, cy)
	cy += 18
	DebugPrintScaled(screen, fmt.Sprintf("Distance:    %.1f km", data.Distance), cx, cy)
	cy += 18
	DebugPrintScaled(screen, fmt.Sprintf("Zone:        %s", data.ZoneName), cx, cy)

	// Selection.
	cy += 35
	retryMarker, menuMarker := "  ", "  "
	if data.Selection == 0 {
		retryMarker = "> "
	} else {
		menuMarker = "> "
	}
	DebugPrintScaled(screen, retryMarker+"RETRY", cx+20, cy)
	DebugPrintScaled(screen, menuMarker+"MENU", cx+120, cy)

	// Total score.
	cy += 30
	DebugPrintScaled(screen, fmt.Sprintf("Total Score: %d", data.TotalScore), cx, cy)

	// Unlock notifications.
	for i, name := range data.NewUnlocks {
		DebugPrintScaled(screen, fmt.Sprintf("\"%s\" unlocked!", name), cx, cy+18+i*16)
	}

	DebugPrintScaled(screen, "Left/Right - select   Enter - confirm", 40, ScreenHeight-30)
}
