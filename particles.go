package main

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// ParticleType identifies the visual behavior of a particle.
type ParticleType int

const (
	ParticleSpark ParticleType = iota
	ParticleFlash
	ParticleNitro
	ParticleBrake
	ParticleRain
)

// ParticleShape controls how a particle is rendered.
type ParticleShape int

const (
	ShapeSquare  ParticleShape = iota
	ShapeCircle                // visually same as square at small sizes
	ShapeLine                  // tall narrow rect with rotation
	ShapeDiamond               // square rotated 45 degrees
	ShapeStar                  // 3 overlapping thin rects
)

const particlePoolSize = 800

// Particle is a single visual effect element.
type Particle struct {
	Active        bool
	Type          ParticleType
	Shape         ParticleShape
	X, Y, VX, VY float64
	W, H          float64
	TTL, MaxTTL   int
	Color         color.RGBA
	Rotation      float64
	RotSpeed      float64
	Gravity       float64
	FadeOut       bool
	Shrink        bool
}

// ParticleSystem manages a fixed pool of particles.
type ParticleSystem struct {
	Pool [particlePoolSize]Particle
}

func NewParticleSystem() ParticleSystem {
	return ParticleSystem{}
}

func (ps *ParticleSystem) spawn() *Particle {
	for i := range ps.Pool {
		if !ps.Pool[i].Active {
			return &ps.Pool[i]
		}
	}
	for i := range ps.Pool {
		if ps.Pool[i].Type == ParticleRain {
			ps.Pool[i].Active = false
			return &ps.Pool[i]
		}
	}
	return nil
}

// --- Emit functions ---

// EmitFlash creates a directional near-miss "whoosh" toward the NPC side.
func (ps *ParticleSystem) EmitFlash(x, y, npcX float64, tier NearMissTier) {
	if tier <= TierNone || int(tier) >= len(flashConfigs) {
		return
	}
	cfg := flashConfigs[tier]

	// Determine which side the NPC is on.
	side := 1.0 // NPC on right
	if npcX < x {
		side = -1.0 // NPC on left
	}

	// Streak lines — the directional "whoosh" trailing past.
	for range cfg.Count {
		p := ps.spawn()
		if p == nil {
			return
		}
		ttl := 12 + rand.IntN(8)
		*p = Particle{
			Active: true, Type: ParticleFlash, Shape: ShapeLine,
			X: x + side*12 + (rand.Float64()*8 - 4),
			Y: y + (rand.Float64()*30 - 15),
			VX: -side * (1.0 + rand.Float64()*1.5),
			VY: 3.0 + rand.Float64()*4.0,
			W: 1.5, H: 8 + rand.Float64()*4,
			TTL: ttl, MaxTTL: ttl,
			Color: cfg.Color, FadeOut: true,
			Rotation: math.Pi/2 + side*0.3,
		}
	}

	// Diamond glints — small accent sparkles.
	glintCount := cfg.Count / 4
	if glintCount < 2 {
		glintCount = 2
	}
	baseAngle := math.Pi // away from right
	if side < 0 {
		baseAngle = 0 // away from left
	}
	for range glintCount {
		p := ps.spawn()
		if p == nil {
			return
		}
		angle := baseAngle + (rand.Float64()*0.8 - 0.4)
		speed := 1.5 + rand.Float64()*2.0
		ttl := 10 + rand.IntN(6)
		*p = Particle{
			Active: true, Type: ParticleFlash, Shape: ShapeDiamond,
			X: x + side*10, Y: y + (rand.Float64()*10 - 5),
			VX: speed * math.Cos(angle), VY: speed*math.Sin(angle) + 1.5,
			W: cfg.Size * 0.7, H: cfg.Size * 0.7,
			TTL: ttl, MaxTTL: ttl,
			Color: cfg.Color, FadeOut: true, Shrink: true,
			RotSpeed: (rand.Float64() - 0.5) * 0.15,
		}
	}
}

var flashConfigs = [5]struct {
	Count    int
	Size     float64
	SpeedMin float64
	SpeedMax float64
	Color    color.RGBA
}{
	{},
	{8, 3, 2.0, 5.0, color.RGBA{0x00, 0xFF, 0xCC, 0xC0}},
	{12, 3, 2.5, 5.5, color.RGBA{0xFF, 0xFF, 0x00, 0xC0}},
	{16, 4, 3.0, 6.0, color.RGBA{0xFF, 0x88, 0x00, 0xD0}},
	{22, 4, 3.5, 7.0, color.RGBA{0xFF, 0x44, 0xFF, 0xE0}},
}

// EmitCollisionBurst creates debris (car color) + smoke on crash.
func (ps *ParticleSystem) EmitCollisionBurst(x, y float64, carClr color.RGBA) {
	// 20 debris squares.
	for range 20 {
		p := ps.spawn()
		if p == nil {
			return
		}
		angle := rand.Float64() * 2 * math.Pi
		speed := 2.0 + rand.Float64()*4.0
		size := 3.0 + rand.Float64()*3.0
		ttl := 30 + rand.IntN(15)
		*p = Particle{
			Active: true, Type: ParticleSpark, Shape: ShapeSquare,
			X: x, Y: y,
			VX: speed * math.Cos(angle), VY: speed * math.Sin(angle),
			W: size, H: size, TTL: ttl, MaxTTL: ttl,
			Color: carClr, Gravity: 0.15, FadeOut: true,
			Rotation: rand.Float64() * math.Pi, RotSpeed: (rand.Float64() - 0.5) * 0.3,
		}
	}
	// 8 smoke puffs.
	for range 8 {
		p := ps.spawn()
		if p == nil {
			return
		}
		angle := rand.Float64() * 2 * math.Pi
		speed := 0.5 + rand.Float64()*1.5
		ttl := 40 + rand.IntN(20)
		*p = Particle{
			Active: true, Type: ParticleSpark, Shape: ShapeCircle,
			X: x + (rand.Float64()*10 - 5), Y: y + (rand.Float64()*10 - 5),
			VX: speed * math.Cos(angle), VY: speed*math.Sin(angle) - 0.5,
			W: 6, H: 6, TTL: ttl, MaxTTL: ttl,
			Color: color.RGBA{0x88, 0x88, 0x88, 0x80}, FadeOut: true, Shrink: true,
		}
	}
}

// EmitCoinPickup creates gold star particles flying upward.
func (ps *ParticleSystem) EmitCoinPickup(x, y float64) {
	for range 6 {
		p := ps.spawn()
		if p == nil {
			return
		}
		ttl := 12 + rand.IntN(8)
		*p = Particle{
			Active: true, Type: ParticleFlash, Shape: ShapeStar,
			X: x + (rand.Float64()*6 - 3), Y: y,
			VX: (rand.Float64() - 0.5) * 2, VY: -2 - rand.Float64()*2,
			W: 3, H: 3, TTL: ttl, MaxTTL: ttl,
			Color: color.RGBA{0xFF, 0xD7, 0x00, 0xFF}, FadeOut: true,
		}
	}
}

// EmitFuelPickup creates green circle particles on fuel pickup.
func (ps *ParticleSystem) EmitFuelPickup(x, y float64) {
	for range 8 {
		p := ps.spawn()
		if p == nil {
			return
		}
		angle := rand.Float64() * 2 * math.Pi
		speed := 1.5 + rand.Float64()*2.0
		ttl := 10 + rand.IntN(6)
		*p = Particle{
			Active: true, Type: ParticleFlash, Shape: ShapeCircle,
			X: x, Y: y,
			VX: speed * math.Cos(angle), VY: speed*math.Sin(angle) - 1,
			W: 2, H: 2, TTL: ttl, MaxTTL: ttl,
			Color: color.RGBA{0x22, 0xDD, 0x44, 0xDD}, FadeOut: true, Shrink: true,
		}
	}
}

// EmitRepairBurst creates a cyan diamond burst on repair pickup.
func (ps *ParticleSystem) EmitRepairBurst(x, y float64) {
	for range 14 {
		p := ps.spawn()
		if p == nil {
			return
		}
		angle := rand.Float64() * 2 * math.Pi
		speed := 2.0 + rand.Float64()*2.0
		ttl := 18 + rand.IntN(8)
		clr := color.RGBA{0x00, 0xFF, 0xFF, 0xE0}
		if rand.IntN(2) == 0 {
			clr = color.RGBA{0xCC, 0xFF, 0xFF, 0xF0}
		}
		*p = Particle{
			Active: true, Type: ParticleFlash, Shape: ShapeDiamond,
			X: x, Y: y,
			VX: speed * math.Cos(angle), VY: speed * math.Sin(angle),
			W: 3.5, H: 3.5, TTL: ttl, MaxTTL: ttl,
			Color: clr, FadeOut: true, Shrink: true,
			RotSpeed: (rand.Float64() - 0.5) * 0.2,
		}
	}
}

// EmitSparks creates simple collision spark particles.
func (ps *ParticleSystem) EmitSparks(x, y float64, count int) {
	for range count {
		p := ps.spawn()
		if p == nil {
			return
		}
		ttl := 15 + rand.IntN(11)
		*p = Particle{
			Active: true, Type: ParticleSpark, Shape: ShapeSquare,
			X: x, Y: y,
			VX: (rand.Float64()*6 - 3), VY: (rand.Float64()*6 - 3),
			W: 2, H: 2, TTL: ttl, MaxTTL: ttl,
			Color: sparkColor(), FadeOut: true,
		}
	}
}

func sparkColor() color.RGBA {
	if rand.IntN(2) == 0 {
		return color.RGBA{0xFF, 0xDD, 0x00, 0xFF}
	}
	return color.RGBA{0xFF, 0x88, 0x00, 0xFF}
}

// EmitNitroFlame creates nitro exhaust: circles + line flames.
func (ps *ParticleSystem) EmitNitroFlame(x, y float64) {
	// 2 shrinking circles.
	for range 2 {
		p := ps.spawn()
		if p == nil {
			return
		}
		ttl := 10 + rand.IntN(5)
		*p = Particle{
			Active: true, Type: ParticleNitro, Shape: ShapeCircle,
			X: x + (rand.Float64()*6 - 3), Y: y,
			VX: (rand.Float64() - 0.5) * 0.8, VY: 1 + rand.Float64()*2,
			W: 4, H: 4, TTL: ttl, MaxTTL: ttl,
			Color: nitroFlameColor(), FadeOut: true, Shrink: true,
		}
	}
	// 1 line flame.
	p := ps.spawn()
	if p == nil {
		return
	}
	ttl := 6 + rand.IntN(4)
	*p = Particle{
		Active: true, Type: ParticleNitro, Shape: ShapeLine,
		X: x + (rand.Float64()*4 - 2), Y: y,
		VX: (rand.Float64() - 0.5) * 0.5, VY: 2 + rand.Float64()*2,
		W: 1.5, H: 5, TTL: ttl, MaxTTL: ttl,
		Color:    color.RGBA{0xFF, 0xFF, 0x88, 0xFF},
		Rotation: (rand.Float64() - 0.5) * 0.3, FadeOut: true,
	}
}

func nitroFlameColor() color.RGBA {
	if rand.IntN(2) == 0 {
		return color.RGBA{0xFF, 0xAA, 0x00, 0xFF}
	}
	return color.RGBA{0xFF, 0xDD, 0x00, 0xFF}
}

// EmitBrakeTrails creates brake trail particles behind the player.
func (ps *ParticleSystem) EmitBrakeTrails(x, y float64) {
	for _, offsetX := range []float64{-8, 8} {
		p := ps.spawn()
		if p == nil {
			return
		}
		*p = Particle{
			Active: true, Type: ParticleBrake,
			X: x + offsetX, Y: y,
			W: 2, H: 8, TTL: 20, MaxTTL: 20,
			Color: color.RGBA{0x88, 0x88, 0x88, 0x50}, FadeOut: true,
		}
	}
}

// EmitRain spawns rain drop particles across the screen.
func (ps *ParticleSystem) EmitRain(screenWidth float64) {
	for range 3 {
		p := ps.spawn()
		if p == nil {
			return
		}
		ttl := 12 + rand.IntN(8)
		*p = Particle{
			Active: true, Type: ParticleRain, Shape: ShapeLine,
			X: rand.Float64() * screenWidth, Y: -10,
			VX: -0.3, VY: 8 + rand.Float64()*4,
			W: 1, H: 8 + rand.Float64()*4,
			TTL: ttl, MaxTTL: ttl,
			Color:    color.RGBA{0xCC, 0xDD, 0xFF, 0x40},
			Rotation: -0.1, FadeOut: true,
		}
	}
}

// --- Update & Draw ---

func (ps *ParticleSystem) Update() {
	for i := range ps.Pool {
		p := &ps.Pool[i]
		if !p.Active {
			continue
		}
		p.X += p.VX
		p.Y += p.VY
		p.VY += p.Gravity
		p.Rotation += p.RotSpeed
		p.TTL--
		if p.TTL <= 0 {
			p.Active = false
		}
	}
}

func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for i := range ps.Pool {
		p := &ps.Pool[i]
		if !p.Active {
			continue
		}
		progress := float64(p.TTL) / float64(p.MaxTTL)

		alpha := p.Color.A
		if p.FadeOut {
			alpha = uint8(float64(p.Color.A) * progress)
		}
		clr := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}

		w, h := p.W, p.H
		if p.Shrink {
			w *= progress
			h *= progress
		}

		switch p.Shape {
		case ShapeSquare, ShapeCircle:
			if p.Rotation != 0 || p.RotSpeed != 0 {
				drawParticleRect(screen, p.X, p.Y, w, h, p.Rotation, clr)
			} else {
				DrawRect(screen, p.X-w/2, p.Y-h/2, w, h, clr)
			}
		case ShapeLine:
			drawParticleRect(screen, p.X, p.Y, w, h, p.Rotation, clr)
		case ShapeDiamond:
			drawParticleRect(screen, p.X, p.Y, w, h, p.Rotation+math.Pi/4, clr)
		case ShapeStar:
			drawParticleStar(screen, p.X, p.Y, w, clr)
		}
	}
}

func drawParticleRect(screen *ebiten.Image, x, y, w, h, rotation float64, clr color.RGBA) {
	rs := renderScaleGlobal
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w*rs, h*rs)
	op.GeoM.Translate(-w*rs/2, -h*rs/2)
	op.GeoM.Rotate(rotation)
	op.GeoM.Translate(x*rs, y*rs)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(pixel, op)
}

func drawParticleStar(screen *ebiten.Image, x, y, size float64, clr color.RGBA) {
	for i := range 3 {
		angle := float64(i) * math.Pi / 3
		drawParticleRect(screen, x, y, size*0.3, size, angle, clr)
	}
}
