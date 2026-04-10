package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// MenuState holds main menu state.
type MenuState struct {
	Selection    int // 0=PLAY, 1=DAILY, 2=GARAGE, 3=SETTINGS
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

var titleColors = [8]color.RGBA{
	{0x00, 0xAA, 0xFF, 0xFF}, // N
	{0xFF, 0x14, 0x93, 0xFF}, // E
	{0x39, 0xFF, 0x14, 0xFF}, // O
	{0xFF, 0xD7, 0x00, 0xFF}, // N
	{0xFF, 0x44, 0x44, 0xFF}, // R
	{0xAA, 0x44, 0xFF, 0xFF}, // U
	{0xFF, 0x66, 0x00, 0xFF}, // S
	{0x00, 0xDD, 0xCC, 0xFF}, // H
}

var colorSelected = color.RGBA{0xFF, 0xDD, 0x00, 0xFF}
var colorUnselected = color.RGBA{0x88, 0x88, 0x88, 0xFF}

// DrawMenu renders the main menu screen.
func DrawMenu(screen *ebiten.Image, menu *MenuState, save *SaveData) {
	menu.BgRoad.Draw(screen, zonePalettes[ZoneNightCity], ZoneNightCity,
		func(float64) float64 { return 0 })

	DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, colorOverlay)

	// Colored title "NEON RUSH".
	title := "NEONRUSH"
	charW := 7 * 3 // 7px char × scale 3
	totalW := len(title) * charW
	startX := (ScreenWidth - totalW) / 2
	for i, ch := range title {
		DrawTextScaled(screen, string(ch), startX+i*charW, 110, 3.0, titleColors[i])
	}

	// Menu items.
	daily := TodayChallenge()
	dailyLabel := fmt.Sprintf("DAILY: %s", daily.Name)
	if save.DailyDone == TodayDateStr() {
		dailyLabel += " DONE"
	}
	items := []string{"PLAY", dailyLabel, "GARAGE", "SETTINGS"}
	for i, item := range items {
		clr := colorUnselected
		prefix := "  "
		if i == menu.Selection {
			clr = colorSelected
			prefix = "> "
		}
		DrawTextColor(screen, prefix+item, ScreenWidth/2-50, 250+i*22, clr)
	}

	// High score.
	if save.HighScore > 0 {
		DrawTextColor(screen,
			fmt.Sprintf("HIGH SCORE: %d", save.HighScore),
			ScreenWidth/2-55, ScreenHeight-60,
			color.RGBA{0xAA, 0xAA, 0xAA, 0xFF})
	}

	DrawTextColor(screen, "Up/Down  Enter", ScreenWidth/2-50, ScreenHeight-30,
		color.RGBA{0x66, 0x66, 0x66, 0xFF})
}
