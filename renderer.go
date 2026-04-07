package main

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var pixel *ebiten.Image

// renderScaleGlobal is set each frame from Game.renderScale
// so all Draw helpers can convert logical to render coordinates.
var renderScaleGlobal float64 = 1.0

func init() {
	pixel = ebiten.NewImage(1, 1)
	pixel.Fill(color.White)
}

// DrawRect draws a filled rectangle by scaling a single white pixel.
// Coordinates are in logical space and automatically scaled to
// the render buffer.
func DrawRect(screen *ebiten.Image, x, y, w, h float64, clr color.RGBA) {
	rs := renderScaleGlobal
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w*rs, h*rs)
	op.GeoM.Translate(x*rs, y*rs)
	op.ColorScale.ScaleWithColor(clr)
	screen.DrawImage(pixel, op)
}

// debugTextBuf is a reusable buffer for scaled debug text rendering.
var debugTextBuf *ebiten.Image

// DebugPrintScaled draws debug text at logical coordinates, scaling
// for the current render resolution. Uses integer scaling + nearest
// filter to keep the bitmap font crisp on any display.
func DebugPrintScaled(screen *ebiten.Image, msg string, x, y int) {
	rs := renderScaleGlobal
	if rs <= 1.0 {
		ebitenutil.DebugPrintAt(screen, msg, x, y)
		return
	}
	// Approximate text bounds: 6px per char width, 16px height.
	tw := len(msg)*6 + 2
	th := 16

	// Grow the shared buffer if needed.
	if debugTextBuf == nil || debugTextBuf.Bounds().Dx() < tw || debugTextBuf.Bounds().Dy() < th {
		newW := tw + 100 // add slack to avoid frequent resizes
		if debugTextBuf != nil && debugTextBuf.Bounds().Dx() > newW {
			newW = debugTextBuf.Bounds().Dx()
		}
		debugTextBuf = ebiten.NewImage(newW, th)
	}
	debugTextBuf.Clear()
	ebitenutil.DebugPrintAt(debugTextBuf, msg, 0, 0)

	sub := debugTextBuf.SubImage(image.Rect(0, 0, tw, th)).(*ebiten.Image)

	// Integer scale for pixel-perfect bitmap font.
	intScale := math.Max(1, math.Round(rs))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(intScale, intScale)
	op.GeoM.Translate(math.Round(float64(x)*rs), math.Round(float64(y)*rs))
	screen.DrawImage(sub, op)
}

// DebugPrintScaledSize draws debug text with an additional scale multiplier for bounce effects.
func DebugPrintScaledSize(screen *ebiten.Image, msg string, x, y int, extraScale float64) {
	rs := renderScaleGlobal
	tw := len(msg)*6 + 2
	th := 16

	if debugTextBuf == nil || debugTextBuf.Bounds().Dx() < tw || debugTextBuf.Bounds().Dy() < th {
		newW := tw + 100
		if debugTextBuf != nil && debugTextBuf.Bounds().Dx() > newW {
			newW = debugTextBuf.Bounds().Dx()
		}
		debugTextBuf = ebiten.NewImage(newW, th)
	}
	debugTextBuf.Clear()
	ebitenutil.DebugPrintAt(debugTextBuf, msg, 0, 0)
	sub := debugTextBuf.SubImage(image.Rect(0, 0, tw, th)).(*ebiten.Image)

	intScale := math.Max(1, math.Round(rs))
	totalScale := intScale * extraScale
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(totalScale, totalScale)
	op.GeoM.Translate(math.Round(float64(x)*rs), math.Round(float64(y)*rs))
	screen.DrawImage(sub, op)
}

// drawRectRaw draws a rectangle WITHOUT renderScale (for init-time image building).
func drawRectRaw(dst *ebiten.Image, x, y, w, h float64, clr color.RGBA) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w, h)
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	dst.DrawImage(pixel, op)
}
