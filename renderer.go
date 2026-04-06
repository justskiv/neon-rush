package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var pixel *ebiten.Image

func init() {
	pixel = ebiten.NewImage(1, 1)
	pixel.Fill(color.White)
}

// DrawRect draws a filled rectangle by scaling a single white pixel.
func DrawRect(screen *ebiten.Image, x, y, w, h float64, clr color.RGBA) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(pixel, op)
}
