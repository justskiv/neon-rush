package main

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// Player represents the player's car.
type Player struct {
	X, Y            float64
	Width, Height   float64
	LateralVelocity float64
	Color           color.RGBA
	MoveSpeed       float64
	Acceleration    float64
	SpinTimer       int
	SpinForce       float64
	CarIndex        int
	Blink           bool
	Damaged         bool
	RepairGlowTimer int
	Drift           DriftState
}

func NewPlayer() Player {
	return newPlayerFromCar(PlayerCars[0])
}

func newPlayerFromCar(car PlayerCarDef) Player {
	idx := 0
	for i, c := range PlayerCars {
		if c.Name == car.Name {
			idx = i
			break
		}
	}
	return Player{
		X:            PlayerStartX,
		Y:            PlayerStartY,
		Width:        PlayerWidth,
		Height:       PlayerHeight,
		Color:        car.Color,
		MoveSpeed:    PlayerMoveSpeed * car.Handling,
		Acceleration: PlayerAcceleration * car.Handling,
		CarIndex:     idx,
	}
}

func (p *Player) Update(curveOffset float64, mirror bool) {
	input := GetHorizontalInput()
	if mirror {
		input = -input
	}

	p.LateralVelocity += input * p.Acceleration
	p.LateralVelocity *= PlayerFriction

	if p.LateralVelocity > p.MoveSpeed {
		p.LateralVelocity = p.MoveSpeed
	}
	if p.LateralVelocity < -p.MoveSpeed {
		p.LateralVelocity = -p.MoveSpeed
	}

	// Oil spin effect.
	if p.SpinTimer > 0 {
		p.LateralVelocity += p.SpinForce
		p.SpinTimer--
	}

	p.X += p.LateralVelocity

	// Barrier: hard stop at road edge + 30px.
	halfW := p.Width / 2
	barrierL := float64(RoadLeft) + curveOffset - 30
	barrierR := float64(RoadRight) + curveOffset + 30
	if p.X-halfW < barrierL {
		p.X = barrierL + halfW
		p.LateralVelocity = 0
	}
	if p.X+halfW > barrierR {
		p.X = barrierR - halfW
		p.LateralVelocity = 0
	}
}

// IsOnShoulder returns true if the player is outside the road surface.
func (p *Player) IsOnShoulder(curveOffset float64) bool {
	halfW := p.Width / 2
	roadL := float64(RoadLeft) + curveOffset
	roadR := float64(RoadRight) + curveOffset
	return p.X-halfW < roadL || p.X+halfW > roadR
}

// IsAtBarrier returns true if the player is pressed against the barrier.
func (p *Player) IsAtBarrier(curveOffset float64) bool {
	halfW := p.Width / 2
	barrierL := float64(RoadLeft) + curveOffset - 30
	barrierR := float64(RoadRight) + curveOffset + 30
	return p.X-halfW <= barrierL+1 || p.X+halfW >= barrierR-1
}

func (p *Player) Draw(screen *ebiten.Image, sprites *SpriteCache, tick int, braking bool) {
	if p.Blink {
		return
	}

	rot := p.Drift.Rotation
	carImg := sprites.PlayerCars[p.CarIndex]
	glowImg := sprites.PlayerGlow[p.CarIndex]

	if p.Damaged {
		glowA := float32(0.3 + 0.25*math.Sin(float64(tick)*0.31) +
			0.15*math.Sin(float64(tick)*0.73))
		drawSpriteAlpha(screen, glowImg, p.X, p.Y, glowA)
		if rot != 0 {
			drawSpriteRotated(screen, carImg, p.X, p.Y, rot)
		} else {
			drawSpriteTinted(screen, carImg, p.X, p.Y, 1.0, 0.6, 0.6)
		}
	} else {
		glowA := float32(0.5)
		if p.Drift.Active {
			glowA = 0.5 + 0.4*float32(p.Drift.HeatLevel)
		} else if p.RepairGlowTimer > 0 {
			glowA = 0.5 + 0.4*float32(p.RepairGlowTimer)/20.0
			p.RepairGlowTimer--
		}
		drawSpriteAlpha(screen, glowImg, p.X, p.Y, glowA)
		if rot != 0 {
			drawSpriteRotated(screen, carImg, p.X, p.Y, rot)
		} else {
			drawSprite(screen, carImg, p.X, p.Y)
		}
	}

	// Brake lights: two red rectangles at rear bumper.
	if braking {
		a := uint8(200)
		clr := color.RGBA{0xFF, 0x22, 0x00, a}
		DrawRect(screen, p.X-9, p.Y+p.Height/2-3, 4, 3, clr)
		DrawRect(screen, p.X+5, p.Y+p.Height/2-3, 4, 3, clr)
	}
}

// ApplyOilSpin causes the player to lose control for 60 ticks.
func (p *Player) ApplyOilSpin() {
	p.SpinTimer = 60
	p.SpinForce = (rand.Float64()*6 - 3) * 0.05 // gentle random drift
}

func (p *Player) Bounds() Rect {
	// Hitbox is smaller than visual sprite (curved body is narrower).
	return NewRect(p.X, p.Y, p.Width*0.8, p.Height*0.9)
}
