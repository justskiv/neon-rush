package main

import (
	"image/color"
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
		MoveSpeed:    PlayerMoveSpeed * car.ManeuverMod,
		Acceleration: PlayerAcceleration * car.ManeuverMod,
		CarIndex:     idx,
	}
}

func (p *Player) Update() {
	input := GetHorizontalInput()

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

	halfW := p.Width / 2
	if p.X-halfW < RoadLeft {
		p.X = RoadLeft + halfW
		p.LateralVelocity = 0
	}
	if p.X+halfW > RoadRight {
		p.X = RoadRight - halfW
		p.LateralVelocity = 0
	}
}

func (p *Player) Draw(screen *ebiten.Image, sprites *SpriteCache) {
	drawSpriteAlpha(screen, sprites.PlayerGlow[p.CarIndex], p.X, p.Y, 0.5)
	drawSprite(screen, sprites.PlayerCars[p.CarIndex], p.X, p.Y)
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
