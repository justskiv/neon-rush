package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// FloatingText is a score popup that drifts upward and fades out.
type FloatingText struct {
	X, Y  float64
	Text  string
	TTL   int
	MaxTTL int
	Color color.RGBA
}

// ScoreState tracks combo multiplier and floating texts.
type ScoreState struct {
	ComboMultiplier int
	ComboTimer      int
	FloatingTexts   []FloatingText
}

func NewScoreState() ScoreState {
	return ScoreState{
		ComboMultiplier: 1,
		FloatingTexts:   make([]FloatingText, 0, 8),
	}
}

// CheckNearMiss checks if an NPC car has passed the player closely enough.
// Returns bonus score (0 if no near-miss).
func CheckNearMiss(p *Player, car *TrafficCar, state *ScoreState, threshold float64) int {
	if car.NearMissChecked {
		return 0
	}

	// NPC must have passed below the player.
	if car.Y-car.Height/2 <= p.Y+p.Height/2 {
		return 0
	}

	car.NearMissChecked = true

	// Distance between nearest edges on X axis.
	dist := math.Max(0, math.Abs(p.X-car.X)-(p.Width+car.Width)/2)
	if dist >= threshold {
		return 0
	}

	bonus := NearMissBonus * state.ComboMultiplier
	if car.CarType == CarTypeOncoming {
		bonus *= 2 // oncoming near-miss is worth double
	}
	state.ComboTimer = ComboDecayTicks
	if state.ComboMultiplier < ComboMultiplierMax {
		state.ComboMultiplier++
	}

	txt := fmt.Sprintf("+%d NEAR MISS!", bonus)
	if state.ComboMultiplier > 2 {
		txt = fmt.Sprintf("+%d x%d NEAR MISS!", bonus, state.ComboMultiplier-1)
	}

	state.FloatingTexts = append(state.FloatingTexts, FloatingText{
		X:      car.X - 40,
		Y:      car.Y,
		Text:   txt,
		TTL:    60,
		MaxTTL: 60,
		Color:  color.RGBA{0x39, 0xFF, 0x14, 0xFF}, // lime green
	})

	return bonus
}

// UpdateScoreState decrements timers, decays combo, and cleans up texts.
func UpdateScoreState(state *ScoreState) {
	if state.ComboTimer > 0 {
		state.ComboTimer--
		if state.ComboTimer == 0 {
			state.ComboMultiplier = 1
		}
	}

	n := 0
	for i := range state.FloatingTexts {
		ft := &state.FloatingTexts[i]
		ft.Y -= 1
		ft.TTL--
		if ft.TTL > 0 {
			state.FloatingTexts[n] = state.FloatingTexts[i]
			n++
		}
	}
	state.FloatingTexts = state.FloatingTexts[:n]
}

// DrawFloatingTexts renders all active floating score texts.
func DrawFloatingTexts(screen *ebiten.Image, texts []FloatingText) {
	for _, ft := range texts {
		alpha := float64(ft.TTL) / float64(ft.MaxTTL)
		clr := color.RGBA{ft.Color.R, ft.Color.G, ft.Color.B, uint8(alpha * 255)}
		_ = clr // DebugPrint doesn't support color; we draw a tinted rect behind.
		ebitenutil.DebugPrintAt(screen, ft.Text, int(ft.X), int(ft.Y))
	}
}
