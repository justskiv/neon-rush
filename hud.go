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
	DisplaySpeed    float64 // lerped speed for smooth display
	Fuel            float64
	NitroCharges    int
	NitroActive     bool
	ComboMultiplier int
	ComboTimer      int
	Damaged         bool
	RepairFlash     int // ticks remaining for repair flash
	TickCount       int
	ScoreFlash      int // ticks remaining for yellow score flash
	DriftActive     bool
	DriftHeat       float64 // 0..1
	DriftOverheat   bool
	NeonAccent      color.RGBA
	DailyName       string
}

// DrawHUD renders the in-game HUD overlay.
func DrawHUD(screen *ebiten.Image, data HUDData) {
	// Gradient background: 5 strips with decreasing alpha.
	alphas := [5]uint8{200, 160, 120, 80, 0}
	stripH := 7.0
	for i, a := range alphas {
		DrawRect(screen, 0, float64(i)*stripH, ScreenWidth, stripH,
			color.RGBA{0, 0, 0, a})
	}

	// Score: right side, with flash effect.
	scoreText := fmt.Sprintf("%d", data.Score)
	scoreClr := colorWhite
	if data.ScoreFlash > 0 {
		scoreClr = color.RGBA{0xFF, 0xDD, 0x00, 0xFF}
	}
	DrawTextColor(screen, scoreText, ScreenWidth-10-len(scoreText)*7, 2, scoreClr)
	DrawTextColor(screen, "SCORE", ScreenWidth-42, 14,
		color.RGBA{0x88, 0x88, 0x88, 0xFF})

	// Live Best Score bar under SCORE label.
	if data.HighScore > 0 {
		bsX := float64(ScreenWidth - 100)
		bsY := 24.0
		bsW := 90.0
		DrawRect(screen, bsX, bsY, bsW, 2, color.RGBA{0x33, 0x33, 0x33, 0xFF})
		ratio := float64(data.Score) / float64(data.HighScore)
		if ratio > 1 {
			ratio = 1
		}
		barClr := data.NeonAccent
		if barClr.A == 0 {
			barClr = color.RGBA{0x00, 0xCC, 0xFF, 0xFF}
		}
		if ratio > 0.9 && data.Score < data.HighScore {
			pulse := uint8(180 + int(75*math.Sin(float64(data.TickCount)*0.15)))
			barClr.A = pulse
		}
		DrawRect(screen, bsX, bsY, bsW*ratio, 2, barClr)
		if data.Score > data.HighScore && (data.TickCount/10)%2 == 0 {
			DrawRect(screen, bsX, bsY, bsW, 2, colorWhite)
		}
	}

	// Speed: smooth lerp display.
	speedKmh := int(data.DisplaySpeed * 30)
	DrawTextColor(screen, fmt.Sprintf("%d km/h", speedKmh), ScreenWidth-80, 28,
		color.RGBA{0xAA, 0xAA, 0xAA, 0xFF})

	// Combo: center top.
	if data.ComboMultiplier > 1 {
		comboText := fmt.Sprintf("x%d COMBO", data.ComboMultiplier)
		comboClr := color.RGBA{0x20, 0xCC, 0x20, 0xFF}
		DrawTextColor(screen, comboText, ScreenWidth/2-len(comboText)*7/2, 2, comboClr)

		timerFrac := float64(data.ComboTimer) / float64(ComboDecayTicks)
		if timerFrac > 1 {
			timerFrac = 1
		}
		fullW := float64(len(comboText))*7 + 4
		barW := fullW * timerFrac
		var barClr color.RGBA
		switch {
		case timerFrac > 0.5:
			barClr = color.RGBA{0x20, 0xCC, 0x20, 0xCC}
		case timerFrac > 0.25:
			barClr = color.RGBA{0xCC, 0xCC, 0x00, 0xCC}
		default:
			barClr = color.RGBA{0xCC, 0x33, 0x00, 0xCC}
			// Blink when < 25%.
			if (data.TickCount/6)%2 == 0 {
				barClr.A = 0x44
			}
		}
		cx := float64(ScreenWidth/2) - fullW/2
		DrawRect(screen, cx, 16, barW, 3, barClr)
	}

	// Nitro: left side with symbol.
	if data.NitroCharges > 0 || data.NitroActive {
		nitroText := fmt.Sprintf("N%d", data.NitroCharges)
		nitroClr := color.RGBA{0xFF, 0xDD, 0x00, 0xFF}
		if data.NitroActive {
			nitroText = "NITRO!"
		}
		DrawTextColor(screen, nitroText, 10, 2, nitroClr)
	}

	// Fuel bar: left side.
	fuelX, fuelY := 10.0, 18.0
	fuelW, fuelH := 80.0, 6.0
	DrawRect(screen, fuelX, fuelY, fuelW, fuelH, color.RGBA{0x33, 0x33, 0x33, 0xFF})
	fill := data.Fuel / FuelMax * fuelW
	if fill < 0 {
		fill = 0
	}
	fuelClr := color.RGBA{0x00, 0xCC, 0x00, 0xFF}
	if data.Fuel < 50 {
		fuelClr = color.RGBA{0xCC, 0xCC, 0x00, 0xFF}
	}
	if data.Fuel < 20 {
		fuelClr = color.RGBA{0xCC, 0x00, 0x00, 0xFF}
		// Blink when critical.
		a := uint8(100 + int(155*math.Abs(math.Sin(float64(data.TickCount)*0.2))))
		fuelClr.A = a
	}
	DrawRect(screen, fuelX, fuelY, fill, fuelH, fuelClr)
	DrawTextColor(screen, "FUEL", 10, 26, color.RGBA{0x88, 0x88, 0x88, 0xFF})

	// Damage indicator.
	if data.RepairFlash > 0 {
		DrawRect(screen, ScreenWidth/2-55, 44, 110, 14, color.RGBA{0x00, 0xDD, 0xFF, 0x50})
		DrawText(screen, ">> REPAIRED <<", ScreenWidth/2-50, 46)
	} else if data.Damaged {
		DrawRect(screen, 8, 36, 78, 14, color.RGBA{0xFF, 0x22, 0x00, 0x88})
		DrawTextColor(screen, "! DAMAGED !", 10, 38,
			color.RGBA{0xFF, 0x44, 0x00, 0xFF})
	}

	// Drift heat bar (under the player car).
	if data.DriftActive || data.DriftHeat > 0.01 {
		bx := float64(ScreenWidth/2 - 20)
		by := float64(PlayerStartY + PlayerHeight/2 + 8)
		bw := 40.0
		bh := 4.0
		DrawRect(screen, bx, by, bw, bh, color.RGBA{0x33, 0x33, 0x33, 0xAA})
		hfill := data.DriftHeat * bw
		heatClr := color.RGBA{0x00, 0xFF, 0x00, 0xCC}
		if data.DriftHeat > 0.6 {
			heatClr = color.RGBA{0xFF, 0xFF, 0x00, 0xCC}
		}
		if data.DriftHeat > 0.85 {
			heatClr = color.RGBA{0xFF, 0x33, 0x00, 0xCC}
		}
		DrawRect(screen, bx, by, hfill, bh, heatClr)
	}
	if data.DriftOverheat {
		DrawText(screen, "OVERHEAT!",
			ScreenWidth/2-28, PlayerStartY+PlayerHeight/2+14)
	} else if data.DriftActive {
		DrawText(screen, "DRIFT",
			ScreenWidth/2-18, PlayerStartY+PlayerHeight/2+14)
	}

	// Daily challenge indicator.
	if data.DailyName != "" {
		DrawTextColor(screen, "* "+data.DailyName, ScreenWidth-120, 34,
			color.RGBA{0xFF, 0xAA, 0x00, 0xFF})
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
	DrawText(screen, title, cx+20, cy)

	// Near-record message.
	if !data.IsNewHighScore && data.HighScore > 0 {
		ratio := float64(data.Score) / float64(data.HighScore)
		if ratio > 0.7 {
			delta := data.HighScore - data.Score
			DrawText(screen,
				fmt.Sprintf("%d PTS FROM RECORD!", delta),
				cx, cy+18)
		}
	}

	cy += 38
	DrawText(screen, fmt.Sprintf("Score:       %d", data.Score), cx, cy)
	if data.IsNewHighScore {
		DrawText(screen, "NEW!", cx+170, cy)
	}
	cy += 16
	DrawText(screen, fmt.Sprintf("Best:        %d", data.HighScore), cx, cy)

	cy += 16
	nmLine := fmt.Sprintf("Near Misses: %d", data.NearMisses)
	DrawText(screen, nmLine, cx, cy)
	if data.NearMisses > data.BestNearMisses && data.BestNearMisses > 0 {
		DrawText(screen, "BEST!", cx+170, cy)
	}

	cy += 16
	comboLine := fmt.Sprintf("Best Combo:  x%d", data.BestCombo)
	DrawText(screen, comboLine, cx, cy)
	if data.BestCombo > data.SaveBestCombo && data.SaveBestCombo > 0 {
		DrawText(screen, "BEST!", cx+170, cy)
	}

	cy += 16
	topSpeedKmh := int(data.TopSpeed * 30)
	bestSpeedKmh := int(data.BestTopSpeed * 30)
	DrawText(screen, fmt.Sprintf("Top Speed:   %d km/h", topSpeedKmh), cx, cy)
	if topSpeedKmh > bestSpeedKmh && bestSpeedKmh > 0 {
		DrawText(screen, "BEST!", cx+170, cy)
	}

	cy += 16
	DrawText(screen, fmt.Sprintf("Distance:    %.1f km", data.Distance), cx, cy)
	cy += 16
	zoneLine := fmt.Sprintf("Zone:        %s (%d)", data.ZoneName, data.ZonesReached)
	if data.BestZone > 0 {
		zoneLine = fmt.Sprintf("Zone:    %s (%d, best %d)",
			data.ZoneName, data.ZonesReached, data.BestZone)
	}
	DrawText(screen, zoneLine, cx, cy)

	// Unlock progress bar.
	if data.NextUnlockName != "" {
		cy += 20
		DrawRect(screen, float64(cx)-2, float64(cy)-2, 244, 26,
			color.RGBA{0x22, 0x22, 0x33, 0xCC})
		pct := int(data.NextUnlockPct * 100)
		DrawText(screen,
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
	DrawText(screen, retryMarker+"RETRY", cx+40, cy)
	DrawText(screen, menuMarker+"MENU", cx+140, cy)

	// Total score.
	cy += 25
	DrawText(screen, fmt.Sprintf("Total Score: %d", data.TotalScore), cx, cy)

	// Unlock notifications.
	for i, name := range data.NewUnlocks {
		DrawText(screen, fmt.Sprintf("\"%s\" unlocked!", name), cx, cy+16+i*14)
	}

	DrawText(screen, "Left/Right - select   Enter - confirm", 40, ScreenHeight-30)
}
