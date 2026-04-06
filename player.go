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
}

func NewPlayer() Player {
	return newPlayerFromCar(PlayerCars[0])
}

func newPlayerFromCar(car PlayerCarDef) Player {
	return Player{
		X:            PlayerStartX,
		Y:            PlayerStartY,
		Width:        PlayerWidth,
		Height:       PlayerHeight,
		Color:        car.Color,
		MoveSpeed:    PlayerMoveSpeed * car.ManeuverMod,
		Acceleration: PlayerAcceleration * car.ManeuverMod,
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

func (p *Player) Draw(screen *ebiten.Image) {
	DrawRect(screen, p.X-p.Width/2, p.Y-p.Height/2, p.Width, p.Height, p.Color)
}

// ApplyOilSpin causes the player to lose control for 60 ticks.
func (p *Player) ApplyOilSpin() {
	p.SpinTimer = 60
	p.SpinForce = (rand.Float64()*6 - 3) * 0.05 // gentle random drift
}

func (p *Player) Bounds() Rect {
	return NewRect(p.X, p.Y, p.Width, p.Height)
}
