package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

var pixel *ebiten.Image

// renderScaleGlobal is set each frame from Game.renderScale
// so all Draw helpers can convert logical to render coordinates.
var renderScaleGlobal float64 = 1.0

// fontFace is the game-wide monospace font (7x13 pixels).
var fontFace = text.NewGoXFace(basicfont.Face7x13)

var colorWhite = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}

func init() {
	pixel = ebiten.NewImage(1, 1)
	pixel.Fill(color.White)
}

// DrawRect draws a filled rectangle by scaling a single white pixel.
func DrawRect(screen *ebiten.Image, x, y, w, h float64, clr color.RGBA) {
	rs := renderScaleGlobal
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w*rs, h*rs)
	op.GeoM.Translate(x*rs, y*rs)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(pixel, op)
}

// DrawText draws white text at logical coordinates.
func DrawText(screen *ebiten.Image, msg string, x, y int) {
	DrawTextColor(screen, msg, x, y, colorWhite)
}

// DrawTextColor draws text at logical coordinates with the given color.
func DrawTextColor(screen *ebiten.Image, msg string, x, y int, clr color.RGBA) {
	rs := renderScaleGlobal
	op := &text.DrawOptions{}
	op.GeoM.Scale(rs, rs)
	op.GeoM.Translate(math.Round(float64(x)*rs), math.Round(float64(y)*rs))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, msg, fontFace, op)
}

// DrawTextScaled draws text with an extra scale multiplier (for titles, bounce).
func DrawTextScaled(screen *ebiten.Image, msg string, x, y int, scale float64, clr color.RGBA) {
	rs := renderScaleGlobal
	op := &text.DrawOptions{}
	op.GeoM.Scale(rs*scale, rs*scale)
	op.GeoM.Translate(math.Round(float64(x)*rs), math.Round(float64(y)*rs))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, msg, fontFace, op)
}

// MeasureText returns the width of a string in logical pixels.
func MeasureText(msg string) float64 {
	return text.Advance(msg, fontFace)
}

// drawRectRaw draws a rectangle WITHOUT renderScale (for init-time image building).
func drawRectRaw(dst *ebiten.Image, x, y, w, h float64, clr color.RGBA) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	dst.DrawImage(pixel, op)
}
