package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// MenuState holds main menu state.
type MenuState struct {
	Selection    int // 0=PLAY, 1=DAILY, 2=GARAGE
	BgRoad       Road
	BgScrollTick int
}

func NewMenuState() MenuState {
	return MenuState{BgRoad: NewRoad()}
}

func (m *MenuState) Update() {
	m.BgScrollTick++
	m.BgRoad.Update(BaseScrollSpeed*0.5, 0)
}

// DrawMenu renders the main menu screen.
func DrawMenu(screen *ebiten.Image, menu *MenuState, save *SaveData) {
	// Scrolling road background.
	menu.BgRoad.Draw(screen, zonePalettes[ZoneNightCity], ZoneNightCity, func(float64) float64 { return 0 })

	// Dark overlay.
	DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, colorOverlay)

	// Title.
	DebugPrintScaled(screen, "N  E  O  N", ScreenWidth/2-32, 120)
	DebugPrintScaled(screen, "R  U  S  H", ScreenWidth/2-32, 140)

	// Menu items.
	daily := TodayChallenge()
	dailyLabel := fmt.Sprintf("DAILY: %s", daily.Name)
	if save.DailyDone == TodayDateStr() {
		dailyLabel += " DONE"
	}
	items := []string{"PLAY", dailyLabel, "GARAGE"}
	for i, item := range items {
		marker := "  "
		if i == menu.Selection {
			marker = "> "
		}
		DebugPrintScaled(screen, marker+item, ScreenWidth/2-40, 260+i*22)
	}

	// High score.
	if save.HighScore > 0 {
		DebugPrintScaled(screen,
			fmt.Sprintf("HIGH SCORE: %d", save.HighScore),
			ScreenWidth/2-55, ScreenHeight-60)
	}

	DebugPrintScaled(screen, "Up/Down - select   Enter - confirm", 50, ScreenHeight-30)
}
