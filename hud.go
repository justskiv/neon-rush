package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var colorOverlay = color.RGBA{0, 0, 0, 160}

// HUDData contains all data the HUD needs to render.
type HUDData struct {
	Score           int
	HighScore       int
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
	NeonAccent      color.RGBA
	DailyName       string
}

// DrawHUD renders score, speed, fuel bar, combo, and nitro on screen.
func DrawHUD(screen *ebiten.Image, data HUDData) {
	// Semi-transparent top bar.
	DrawRect(screen, 0, 0, ScreenWidth, 42, color.RGBA{0, 0, 0, 120})

	speedKmh := int(data.ScrollSpeed * 30)
	DebugPrintScaled(screen, fmt.Sprintf("SCORE: %d", data.Score), 10, 4)

	// Live Best Score progress bar.
	if data.HighScore > 0 {
		bsX, bsY := 10.0, 16.0
		bsW := 80.0
		DrawRect(screen, bsX, bsY, bsW, 2, color.RGBA{0x33, 0x33, 0x33, 0xFF})
		ratio := float64(data.Score) / float64(data.HighScore)
		if ratio > 1 {
			ratio = 1
		}
		fillW := bsW * ratio
		barClr := data.NeonAccent
		if barClr.A == 0 {
			barClr = color.RGBA{0x00, 0xCC, 0xFF, 0xFF}
		}
		// Pulsation when > 90%.
		if ratio > 0.9 && data.Score < data.HighScore {
			pulse := uint8(180 + int(75*math.Sin(float64(data.TickCount)*0.15)))
			barClr.A = pulse
		}
		DrawRect(screen, bsX, bsY, fillW, 2, barClr)
		if data.Score > data.HighScore {
			if (data.TickCount/10)%2 == 0 {
				DrawRect(screen, bsX, bsY, bsW, 2, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF})
			}
			DebugPrintScaled(screen, "NEW RECORD!", 95, 4)
		}
	}

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

	// Daily challenge indicator.
	if data.DailyName != "" {
		DebugPrintScaled(screen, "* "+data.DailyName, ScreenWidth-120, 32)
	}
}

// GameOverData holds all info for the game over screen.
type GameOverData struct {
	Score          int
	HighScore      int
	IsNewHighScore bool
	NearMisses     int
	BestNearMisses int
	BestCombo      int
	SaveBestCombo  int
	TopSpeed       float64
	BestTopSpeed   float64
	Distance       float64
	ZoneName       string
	ZonesReached   int
	BestZone       int
	Selection      int // 0=RETRY, 1=MENU
	NewUnlocks     []string
	TotalScore     int
	Busted         bool
	FuelEmpty      bool
	NextUnlockName string
	NextUnlockPct  float64
}

// DrawGameOver renders the game over overlay with stats.
func DrawGameOver(screen *ebiten.Image, data GameOverData) {
	DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180})

	cx := 80
	cy := 120
	title := "G A M E   O V E R"
	if data.Busted {
		title = "B U S T E D !"
	} else if data.FuelEmpty {
		title = "O U T  O F  F U E L"
	}
	DebugPrintScaled(screen, title, cx+20, cy)

	// Near-record message.
	if !data.IsNewHighScore && data.HighScore > 0 {
		ratio := float64(data.Score) / float64(data.HighScore)
		if ratio > 0.7 {
			delta := data.HighScore - data.Score
			DebugPrintScaled(screen,
				fmt.Sprintf("%d PTS FROM RECORD!", delta),
				cx, cy+18)
		}
	}

	cy += 38
	DebugPrintScaled(screen, fmt.Sprintf("Score:       %d", data.Score), cx, cy)
	if data.IsNewHighScore {
		DebugPrintScaled(screen, "NEW!", cx+170, cy)
	}
	cy += 16
	DebugPrintScaled(screen, fmt.Sprintf("Best:        %d", data.HighScore), cx, cy)

	cy += 16
	nmLine := fmt.Sprintf("Near Misses: %d", data.NearMisses)
	DebugPrintScaled(screen, nmLine, cx, cy)
	if data.NearMisses > data.BestNearMisses && data.BestNearMisses > 0 {
		DebugPrintScaled(screen, "BEST!", cx+170, cy)
	}

	cy += 16
	comboLine := fmt.Sprintf("Best Combo:  x%d", data.BestCombo)
	DebugPrintScaled(screen, comboLine, cx, cy)
	if data.BestCombo > data.SaveBestCombo && data.SaveBestCombo > 0 {
		DebugPrintScaled(screen, "BEST!", cx+170, cy)
	}

	cy += 16
	topSpeedKmh := int(data.TopSpeed * 30)
	bestSpeedKmh := int(data.BestTopSpeed * 30)
	DebugPrintScaled(screen, fmt.Sprintf("Top Speed:   %d km/h", topSpeedKmh), cx, cy)
	if topSpeedKmh > bestSpeedKmh && bestSpeedKmh > 0 {
		DebugPrintScaled(screen, "BEST!", cx+170, cy)
	}

	cy += 16
	DebugPrintScaled(screen, fmt.Sprintf("Distance:    %.1f km", data.Distance), cx, cy)
	cy += 16
	zoneLine := fmt.Sprintf("Zone:        %s (%d)", data.ZoneName, data.ZonesReached)
	if data.BestZone > 0 {
		zoneLine = fmt.Sprintf("Zone:    %s (%d, best %d)",
			data.ZoneName, data.ZonesReached, data.BestZone)
	}
	DebugPrintScaled(screen, zoneLine, cx, cy)

	// Unlock progress bar.
	if data.NextUnlockName != "" {
		cy += 20
		DrawRect(screen, float64(cx)-2, float64(cy)-2, 244, 26,
			color.RGBA{0x22, 0x22, 0x33, 0xCC})
		pct := int(data.NextUnlockPct * 100)
		DebugPrintScaled(screen,
			fmt.Sprintf("Next: %s  %d%%", data.NextUnlockName, pct),
			cx, cy)
		barY := float64(cy + 14)
		barW := 240.0
		DrawRect(screen, float64(cx), barY, barW, 6,
			color.RGBA{0x33, 0x33, 0x33, 0xFF})
		DrawRect(screen, float64(cx), barY, barW*data.NextUnlockPct, 6,
			color.RGBA{0x00, 0xCC, 0x00, 0xFF})
		cy += 28
	}

	// Selection.
	cy += 12
	retryMarker, menuMarker := "  ", "  "
	if data.Selection == 0 {
		retryMarker = "> "
	} else {
		menuMarker = "> "
	}
	DebugPrintScaled(screen, retryMarker+"RETRY", cx+40, cy)
	DebugPrintScaled(screen, menuMarker+"MENU", cx+140, cy)

	// Total score.
	cy += 25
	DebugPrintScaled(screen, fmt.Sprintf("Total Score: %d", data.TotalScore), cx, cy)

	// Unlock notifications.
	for i, name := range data.NewUnlocks {
		DebugPrintScaled(screen, fmt.Sprintf("\"%s\" unlocked!", name), cx, cy+16+i*14)
	}

	DebugPrintScaled(screen, "Left/Right - select   Enter - confirm", 40, ScreenHeight-30)
}
