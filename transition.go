package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Transition handles fade-to-black screen transitions.
type Transition struct {
	active   bool
	alpha    float64 // 0..255
	phase    int     // 0=fade-in, 1=switch, 2=fade-out
	speed    float64 // alpha increment per tick
	onSwitch func()
}

// Start begins a fade transition. Speed is alpha per tick (higher=faster).
func (tr *Transition) Start(speed float64, onSwitch func()) {
	tr.active = true
	tr.alpha = 0
	tr.phase = 0
	tr.speed = speed
	tr.onSwitch = onSwitch
}

// Active returns whether a transition is in progress.
func (tr *Transition) Active() bool {
	return tr.active
}

// Update advances the transition state. Call once per tick.
func (tr *Transition) Update() {
	if !tr.active {
		return
	}
	switch tr.phase {
	case 0: // fade in (darken)
		tr.alpha += tr.speed
		if tr.alpha >= 255 {
			tr.alpha = 255
			tr.phase = 1
		}
	case 1: // switch
		if tr.onSwitch != nil {
			tr.onSwitch()
			tr.onSwitch = nil
		}
		tr.phase = 2
	case 2: // fade out (brighten)
		tr.alpha -= tr.speed
		if tr.alpha <= 0 {
			tr.alpha = 0
			tr.active = false
		}
	}
}

// Draw renders the fade overlay.
func (tr *Transition) Draw(screen *ebiten.Image) {
	if !tr.active || tr.alpha < 1 {
		return
	}
	a := uint8(tr.alpha)
	DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, a})
}
