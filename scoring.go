package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// NearMissTier classifies the proximity of a near-miss.
type NearMissTier int

const (
	TierNone      NearMissTier = iota
	TierNear                   // far-ish pass
	TierClose                  // medium proximity
	TierVeryClose              // very tight
	TierInsane                 // pixel-perfect
)

// NearMissResult holds the outcome of a near-miss check.
type NearMissResult struct {
	Bonus int
	Tier  NearMissTier
	X, Y  float64 // position for effects
}

// FloatingText is a score popup that drifts upward with bounce animation.
type FloatingText struct {
	X, Y       float64
	Text       string
	TTL        int
	MaxTTL     int
	Color      color.RGBA
	VY         float64 // vertical speed (negative = up)
	Scale      float64 // current scale
	ScaleStart float64 // initial scale for bounce
	ScaleEnd   float64 // target scale after bounce
	ScaleTicks int     // ticks for bounce animation
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

// Tier colors (base).
var tierColors = [5]color.RGBA{
	{},                                  // TierNone
	{0x39, 0xFF, 0x14, 0xFF},           // TierNear — lime green
	{0xFF, 0xFF, 0x00, 0xFF},           // TierClose — yellow
	{0xFF, 0x66, 0x00, 0xFF},           // TierVeryClose — hot orange
	{0xFF, 0x00, 0x88, 0xFF},           // TierInsane — neon magenta
}

var tierLabels = [5]string{
	"", "", " CLOSE!", " VERY CLOSE!", " INSANE!",
}

// CheckNearMiss checks if an NPC car has passed the player closely enough.
// scrollSpeed is used for the speed multiplier. offsetFn returns per-Y offset.
func CheckNearMiss(p *Player, car *TrafficCar, state *ScoreState, threshold, scrollSpeed float64, offsetFn func(float64) float64) NearMissResult {
	if car.NearMissChecked {
		return NearMissResult{}
	}

	if car.Y-car.Height/2 <= p.Y+p.Height/2 {
		return NearMissResult{}
	}

	car.NearMissChecked = true

	dist := math.Max(0, math.Abs(p.X-(car.X+offsetFn(car.Y)))-(p.Width+car.Width)/2)
	if dist >= threshold {
		return NearMissResult{}
	}

	// Proximity: 0 = at threshold edge, 1 = touching.
	proximity := 1.0 - dist/threshold
	proxMult := 1.0 + proximity*proximity*2.0 // 1..3

	// Speed: 1× at BaseScrollSpeed, 2.5× at MaxScrollSpeed.
	speedNorm := (scrollSpeed - BaseScrollSpeed) / (MaxScrollSpeed - BaseScrollSpeed)
	if speedNorm < 0 {
		speedNorm = 0
	}
	if speedNorm > 1 {
		speedNorm = 1
	}
	speedMult := 1.0 + speedNorm*1.5 // 1..2.5

	totalMult := proxMult * speedMult * float64(state.ComboMultiplier)
	if car.CarType == CarTypeOncoming {
		totalMult *= 2
	}
	bonus := int(math.Round(float64(NearMissBonus) * totalMult))

	state.ComboTimer = ComboDecayTicks
	if state.ComboMultiplier < ComboMultiplierMax {
		state.ComboMultiplier++
	}

	// Determine tier from proximity.
	tier := TierNear
	switch {
	case proximity >= 0.9:
		tier = TierInsane
	case proximity >= 0.7:
		tier = TierVeryClose
	case proximity >= 0.4:
		tier = TierClose
	}

	// Build text.
	txt := fmt.Sprintf("+%d%s", bonus, tierLabels[tier])
	if totalMult >= 2.0 {
		txt = fmt.Sprintf("%s x%.1f", txt, totalMult)
	}

	// Color: tier base, shifted toward white by speed.
	baseClr := tierColors[tier]
	clr := lerpColorWhite(baseClr, speedNorm*0.3)

	// Scale by tier — bigger text for closer passes.
	type tierScale struct {
		start, end float64
		ticks      int
		vy         float64
	}
	scales := [5]tierScale{
		{},
		{2.5, 1.5, 10, -1.8},  // TierNear
		{3.0, 1.8, 12, -2.0},  // TierClose
		{3.5, 2.0, 14, -2.2},  // TierVeryClose
		{4.5, 2.5, 16, -2.5},  // TierInsane
	}
	ts := scales[tier]
	// Extra boost for high multipliers.
	if totalMult >= 10 {
		ts.start += 0.5
		ts.end += 0.5
	}

	state.FloatingTexts = append(state.FloatingTexts, FloatingText{
		X: car.X - 50, Y: car.Y,
		Text: txt, TTL: 60, MaxTTL: 60,
		Color: clr, VY: ts.vy,
		ScaleStart: ts.start, ScaleEnd: ts.end, ScaleTicks: ts.ticks,
	})

	return NearMissResult{Bonus: bonus, Tier: tier, X: p.X, Y: p.Y}
}

// lerpColorWhite shifts a color toward white by factor (0..1).
func lerpColorWhite(c color.RGBA, f float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) + (255-float64(c.R))*f),
		G: uint8(float64(c.G) + (255-float64(c.G))*f),
		B: uint8(float64(c.B) + (255-float64(c.B))*f),
		A: c.A,
	}
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
		vy := ft.VY
		if vy == 0 {
			vy = -1 // default for backward compatibility
		}
		ft.Y += vy
		ft.TTL--
		if ft.TTL > 0 {
			state.FloatingTexts[n] = state.FloatingTexts[i]
			n++
		}
	}
	state.FloatingTexts = state.FloatingTexts[:n]
}

// DrawFloatingTexts renders all active floating score texts with colored background.
func DrawFloatingTexts(screen *ebiten.Image, texts []FloatingText) {
	for i := range texts {
		ft := &texts[i]
		alpha := float64(ft.TTL) / float64(ft.MaxTTL)
		a := uint8(alpha * 180)
		clr := color.RGBA{ft.Color.R, ft.Color.G, ft.Color.B, a}

		// Compute bounce scale.
		scale := ft.ScaleEnd
		if ft.ScaleTicks > 0 {
			elapsed := ft.MaxTTL - ft.TTL
			if elapsed < ft.ScaleTicks {
				t := float64(elapsed) / float64(ft.ScaleTicks)
				bounce := 1.0 - (1.0-t)*(1.0-t) // quadratic ease-out
				scale = ft.ScaleStart + (ft.ScaleEnd-ft.ScaleStart)*bounce
			}
		}
		if scale < 0.5 {
			scale = 1.0
		}

		tw := float64(len(ft.Text)*6+4) * scale
		DrawRect(screen, ft.X-1, ft.Y-1, tw, 16*scale, color.RGBA{0, 0, 0, a / 2})
		DrawRect(screen, ft.X-1, ft.Y-1, 2, 16*scale, clr)
		DrawTextScaled(screen, ft.Text, int(ft.X+3), int(ft.Y), scale, clr)
	}
}
