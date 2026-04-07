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

const particlePoolSize = 400

// Particle is a single visual effect element.
type Particle struct {
	Active         bool
	Type           ParticleType
	X, Y, VX, VY  float64
	W, H           float64
	TTL, MaxTTL    int
	Color          color.RGBA
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
	return nil
}

// EmitSparks creates collision spark particles at the given position.
func (ps *ParticleSystem) EmitSparks(x, y float64, count int) {
	for range count {
		p := ps.spawn()
		if p == nil {
			return
		}
		ttl := 15 + rand.IntN(11) // 15-25
		*p = Particle{
			Active: true,
			Type:   ParticleSpark,
			X: x, Y: y,
			VX: (rand.Float64()*6 - 3),
			VY: (rand.Float64()*6 - 3),
			W: 2, H: 2,
			TTL: ttl, MaxTTL: ttl,
			Color: sparkColor(),
		}
	}
}

func sparkColor() color.RGBA {
	if rand.IntN(2) == 0 {
		return color.RGBA{0xFF, 0xDD, 0x00, 0xFF} // yellow
	}
	return color.RGBA{0xFF, 0x88, 0x00, 0xFF} // orange
}

// EmitFlash creates a near-miss burst: small cyan particles flying outward.
func (ps *ParticleSystem) EmitFlash(x, y float64) {
	for range 8 {
		p := ps.spawn()
		if p == nil {
			return
		}
		angle := rand.Float64() * 6.283
		speed := 2.0 + rand.Float64()*3.0
		ttl := 12 + rand.IntN(8)
		*p = Particle{
			Active: true,
			Type:   ParticleFlash,
			X: x + (rand.Float64()*10 - 5),
			Y: y + (rand.Float64()*10 - 5),
			VX: speed * math.Cos(angle),
			VY: speed * math.Sin(angle),
			W: 3, H: 3,
			TTL: ttl, MaxTTL: ttl,
			Color: color.RGBA{0x00, 0xFF, 0xCC, 0xC0},
		}
	}
}

// EmitNitroFlame creates nitro exhaust particles behind the player.
func (ps *ParticleSystem) EmitNitroFlame(x, y float64) {
	for range 3 {
		p := ps.spawn()
		if p == nil {
			return
		}
		ttl := 8 + rand.IntN(5) // 8-12
		w := 4 + rand.Float64()*3
		h := 6 + rand.Float64()*5
		*p = Particle{
			Active: true,
			Type:   ParticleNitro,
			X: x + (rand.Float64()*6 - 3), Y: y,
			VX: (rand.Float64() - 0.5),
			VY: 1 + rand.Float64()*2,
			W: w, H: h,
			TTL: ttl, MaxTTL: ttl,
			Color: nitroFlameColor(),
		}
	}
}

func nitroFlameColor() color.RGBA {
	if rand.IntN(2) == 0 {
		return color.RGBA{0xFF, 0xAA, 0x00, 0xFF} // orange
	}
	return color.RGBA{0xFF, 0xDD, 0x00, 0xFF} // yellow
}

// EmitBrakeTrails creates brake trail particles behind the player.
func (ps *ParticleSystem) EmitBrakeTrails(x, y float64) {
	for _, offsetX := range []float64{-8, 8} {
		p := ps.spawn()
		if p == nil {
			return
		}
		*p = Particle{
			Active: true,
			Type:   ParticleBrake,
			X: x + offsetX, Y: y,
			W: 2, H: 8,
			TTL: 20, MaxTTL: 20,
			Color: color.RGBA{0x88, 0x88, 0x88, 0x50},
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
			Active: true,
			Type:   ParticleRain,
			X:      rand.Float64() * screenWidth,
			Y:      -10,
			VX:     -0.3,
			VY:     8 + rand.Float64()*4,
			W:      1, H: 8 + rand.Float64()*4,
			TTL: ttl, MaxTTL: ttl,
			Color: color.RGBA{0xCC, 0xDD, 0xFF, 0x40},
		}
	}
}

// Update moves all active particles and deactivates expired ones.
func (ps *ParticleSystem) Update() {
	for i := range ps.Pool {
		p := &ps.Pool[i]
		if !p.Active {
			continue
		}
		p.X += p.VX
		p.Y += p.VY
		p.TTL--
		if p.TTL <= 0 {
			p.Active = false
		}
	}
}

// Draw renders all active particles.
func (ps *ParticleSystem) Draw(screen *ebiten.Image) {
	for i := range ps.Pool {
		p := &ps.Pool[i]
		if !p.Active {
			continue
		}
		alpha := uint8(float64(p.Color.A) * float64(p.TTL) / float64(p.MaxTTL))
		clr := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		DrawRect(screen, p.X-p.W/2, p.Y-p.H/2, p.W, p.H, clr)
	}
}
