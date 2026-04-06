package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// MenuState holds main menu state.
type MenuState struct {
	Selection    int // 0=PLAY, 1=GARAGE
	BgRoad       Road
	BgScrollTick int
}

func NewMenuState() MenuState {
	return MenuState{BgRoad: NewRoad()}
}

func (m *MenuState) Update() {
	m.BgScrollTick++
	m.BgRoad.Update(BaseScrollSpeed * 0.5)
}

// DrawMenu renders the main menu screen.
func DrawMenu(screen *ebiten.Image, menu *MenuState, save *SaveData) {
	// Scrolling road background.
	menu.BgRoad.Draw(screen, zonePalettes[ZoneNightCity], ZoneNightCity)

	// Dark overlay.
	DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, colorOverlay)

	// Title.
	ebitenutil.DebugPrintAt(screen, "N  E  O  N", ScreenWidth/2-32, 120)
	ebitenutil.DebugPrintAt(screen, "R  U  S  H", ScreenWidth/2-32, 140)

	// Menu items.
	items := []string{"PLAY", "GARAGE"}
	for i, item := range items {
		marker := "  "
		if i == menu.Selection {
			marker = "> "
		}
		ebitenutil.DebugPrintAt(screen, marker+item, ScreenWidth/2-30, 260+i*22)
	}

	// High score.
	if save.HighScore > 0 {
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("HIGH SCORE: %d", save.HighScore),
			ScreenWidth/2-55, ScreenHeight-60)
	}

	ebitenutil.DebugPrintAt(screen, "Up/Down - select   Enter - confirm", 50, ScreenHeight-30)
}
