package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// GetHorizontalInput returns -1, 0, or 1 based on pressed keys.
func GetHorizontalInput() float64 {
	left := ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	right := ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)

	switch {
	case left && right:
		return 0
	case left:
		return -1
	case right:
		return 1
	default:
		return 0
	}
}

// GetVerticalInput returns -1 (up/accelerate), 0, or +1 (down/brake).
func GetVerticalInput() float64 {
	up := ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW)
	down := ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS)

	switch {
	case up && down:
		return 0
	case up:
		return -1
	case down:
		return 1
	default:
		return 0
	}
}

// IsNitroPressed returns true on the frame Space is first pressed.
func IsNitroPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

// IsRestartPressed returns true on the frame Enter is first pressed.
func IsRestartPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

// IsEscPressed returns true on the frame Escape is first pressed.
func IsEscPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEscape)
}

// IsUpMenuPressed returns true on the frame Up/W is first pressed (menu navigation).
func IsUpMenuPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW)
}

// IsDownMenuPressed returns true on the frame Down/S is first pressed (menu navigation).
func IsDownMenuPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS)
}

// IsLeftMenuPressed returns true on the frame Left/A is first pressed (garage navigation).
func IsLeftMenuPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA)
}

// IsRightMenuPressed returns true on the frame Right/D is first pressed (garage navigation).
func IsRightMenuPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyD)
}
