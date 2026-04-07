package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SpriteCache holds all pre-rendered vector sprites.
// Sprites are generated at SpriteScale×logical resolution for smooth curves.
type SpriteCache struct {
	PlayerCars [5]*ebiten.Image
	PlayerGlow [5]*ebiten.Image

	TrafficSedan    [4]*ebiten.Image
	TrafficTruck    [3]*ebiten.Image
	TrafficSport    [3]*ebiten.Image
	TrafficPolice   *ebiten.Image
	PoliceSiren     [2]*ebiten.Image
	TrafficOncoming *ebiten.Image
	HeadlightCone   *ebiten.Image
	TurnSignal      *ebiten.Image

	FuelCan   *ebiten.Image
	FuelGlow  *ebiten.Image
	NitroItem *ebiten.Image
	NitroGlow *ebiten.Image
	Coin      [4]*ebiten.Image
	OilSpill  *ebiten.Image

	RepairKit  *ebiten.Image
	RepairGlow *ebiten.Image

	Buildings [NumBuildingVariants]*ebiten.Image
	LampPost  *ebiten.Image
}

// SpriteScale is the supersampling factor, set dynamically at init.
var SpriteScale = DefaultSpriteScale

// S is SpriteScale as float32 for path coordinates.
var S float32

func NewSpriteCache() *SpriteCache {
	// Compute sprite scale from the actual monitor so sprites are
	// always rendered above the native screen resolution.
	dsf := ebiten.Monitor().DeviceScaleFactor()
	mw, mh := ebiten.Monitor().Size()
	winScale := min(float64(mh)*0.85/float64(ScreenHeight), float64(mw)*0.85/float64(ScreenWidth))
	expectedRenderScale := winScale * dsf
	// Add 50% headroom so sprites are clearly downscaled, never upscaled.
	SpriteScale = max(DefaultSpriteScale, math.Ceil(expectedRenderScale*1.5))
	S = float32(SpriteScale)

	sc := &SpriteCache{}
	sc.generatePlayerCars()
	sc.generateTrafficSprites()
	sc.generateItemSprites()
	sc.generateDecorSprites()
	return sc
}

func drawPathOpts(clr color.RGBA) *vector.DrawPathOptions {
	opts := &vector.DrawPathOptions{AntiAlias: true}
	opts.ColorScale.ScaleWithColor(clr)
	return opts
}

// drawSprite draws a cached sprite centered at (cx, cy) in logical
// coordinates, automatically scaling for the current render resolution.
func drawSprite(screen, img *ebiten.Image, cx, cy float64) {
	rs := renderScaleGlobal
	op := &ebiten.DrawImageOptions{}
	s := rs / SpriteScale
	w := float64(img.Bounds().Dx()) * s
	h := float64(img.Bounds().Dy()) * s
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(cx*rs-w/2, cy*rs-h/2)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

// drawSpriteRotated draws a sprite centered at (cx,cy) with rotation in radians.
func drawSpriteRotated(screen, img *ebiten.Image, cx, cy, rotation float64) {
	rs := renderScaleGlobal
	op := &ebiten.DrawImageOptions{}
	s := rs / SpriteScale
	iw := float64(img.Bounds().Dx())
	ih := float64(img.Bounds().Dy())
	op.GeoM.Translate(-iw/2, -ih/2)
	op.GeoM.Rotate(rotation)
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(cx*rs, cy*rs)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

// drawSpriteTinted draws a sprite with RGB color scaling (for damage tint etc).
func drawSpriteTinted(screen, img *ebiten.Image, cx, cy float64, r, g, b float32) {
	rs := renderScaleGlobal
	op := &ebiten.DrawImageOptions{}
	s := rs / SpriteScale
	w := float64(img.Bounds().Dx()) * s
	h := float64(img.Bounds().Dy()) * s
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(cx*rs-w/2, cy*rs-h/2)
	op.ColorScale.Scale(r, g, b, 1)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

func drawSpriteAlpha(screen, img *ebiten.Image, cx, cy float64, alpha float32) {
	rs := renderScaleGlobal
	op := &ebiten.DrawImageOptions{}
	s := rs / SpriteScale
	w := float64(img.Bounds().Dx()) * s
	h := float64(img.Bounds().Dy()) * s
	op.GeoM.Scale(s, s)
	op.GeoM.Translate(cx*rs-w/2, cy*rs-h/2)
	op.ColorScale.ScaleAlpha(alpha)
	op.Filter = ebiten.FilterLinear
	screen.DrawImage(img, op)
}

// si creates a scaled int for image dimensions.
func si(v int) int { return int(float32(v) * S) }

// --- Player car sprites ---

func (sc *SpriteCache) generatePlayerCars() {
	for i, car := range PlayerCars {
		sc.PlayerCars[i] = generatePlayerCarBody(car.Color)
		sc.PlayerGlow[i] = generatePlayerCarGlow(car.Color)
	}
}

func generatePlayerCarBody(clr color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(si(PlayerSpriteW), si(PlayerSpriteH))

	wheelClr := color.RGBA{0x1E, 0x1E, 0x28, 0xFF}
	drawRoundedRect(img, 3*S, 10*S, 4*S, 9*S, wheelClr)
	drawRoundedRect(img, 29*S, 10*S, 4*S, 9*S, wheelClr)
	drawRoundedRect(img, 3*S, 38*S, 4*S, 9*S, wheelClr)
	drawRoundedRect(img, 29*S, 38*S, 4*S, 9*S, wheelClr)

	body := &vector.Path{}
	body.MoveTo(18*S, 2*S)
	body.CubicTo(24*S, 2*S, 30*S, 10*S, 30*S, 16*S)
	body.LineTo(30*S, 38*S)
	body.CubicTo(30*S, 44*S, 28*S, 48*S, 26*S, 50*S)
	body.LineTo(10*S, 50*S)
	body.CubicTo(8*S, 48*S, 6*S, 44*S, 6*S, 38*S)
	body.LineTo(6*S, 16*S)
	body.CubicTo(6*S, 10*S, 12*S, 2*S, 18*S, 2*S)
	body.Close()
	vector.FillPath(img, body, &vector.FillOptions{}, drawPathOpts(clr))
	vector.StrokePath(img, body, &vector.StrokeOptions{Width: 1.0 * S}, drawPathOpts(lighten(clr, 40)))

	ws := &vector.Path{}
	ws.MoveTo(12*S, 14*S)
	ws.CubicTo(13*S, 12*S, 23*S, 12*S, 24*S, 14*S)
	ws.LineTo(25*S, 20*S)
	ws.LineTo(11*S, 20*S)
	ws.Close()
	vector.FillPath(img, ws, &vector.FillOptions{}, drawPathOpts(color.RGBA{0x8C, 0xDC, 0xFF, 0x78}))

	rw := &vector.Path{}
	rw.MoveTo(13*S, 32*S)
	rw.LineTo(23*S, 32*S)
	rw.LineTo(24*S, 37*S)
	rw.LineTo(12*S, 37*S)
	rw.Close()
	vector.FillPath(img, rw, &vector.FillOptions{}, drawPathOpts(color.RGBA{0x8C, 0xDC, 0xFF, 0x50}))

	vector.StrokeLine(img, 18*S, 20*S, 18*S, 32*S, 1.0*S, lighten(clr, 20), true)

	vector.FillCircle(img, 10*S, 48*S, 4*S, color.RGBA{0xFF, 0x28, 0x28, 0x3C}, true)
	vector.FillCircle(img, 26*S, 48*S, 4*S, color.RGBA{0xFF, 0x28, 0x28, 0x3C}, true)
	vector.FillCircle(img, 10*S, 48*S, 2.5*S, color.RGBA{0xFF, 0x28, 0x28, 0xDC}, true)
	vector.FillCircle(img, 26*S, 48*S, 2.5*S, color.RGBA{0xFF, 0x28, 0x28, 0xDC}, true)

	vector.FillCircle(img, 12*S, 6*S, 2*S, color.RGBA{0xFF, 0xF0, 0x96, 0xC8}, true)
	vector.FillCircle(img, 24*S, 6*S, 2*S, color.RGBA{0xFF, 0xF0, 0x96, 0xC8}, true)

	return img
}

func generatePlayerCarGlow(clr color.RGBA) *ebiten.Image {
	pad := float32(PlayerGlowPad) * S
	img := ebiten.NewImage(si(PlayerSpriteW+PlayerGlowPad*2), si(PlayerSpriteH+PlayerGlowPad*2))

	body := &vector.Path{}
	body.MoveTo(18*S+pad, 0)
	body.CubicTo(26*S+pad, 0, 32*S+pad, 10*S, 32*S+pad, 16*S)
	body.LineTo(32*S+pad, 40*S)
	body.CubicTo(32*S+pad, 46*S, 30*S+pad, 52*S, 28*S+pad, 54*S)
	body.LineTo(8*S+pad, 54*S)
	body.CubicTo(6*S+pad, 52*S, 4*S+pad, 46*S, 4*S+pad, 40*S)
	body.LineTo(4*S+pad, 16*S)
	body.CubicTo(4*S+pad, 10*S, 10*S+pad, 0, 18*S+pad, 0)
	body.Close()

	glowClr := color.RGBA{clr.R, clr.G, clr.B, 0x30}
	vector.FillPath(img, body, &vector.FillOptions{}, drawPathOpts(glowClr))
	return img
}

// --- Drawing helpers ---

func drawRoundedRect(img *ebiten.Image, x, y, w, h float32, clr color.RGBA) {
	r := min(w, h) * 0.25
	p := &vector.Path{}
	p.MoveTo(x+r, y)
	p.LineTo(x+w-r, y)
	p.CubicTo(x+w, y, x+w, y+r, x+w, y+r)
	p.LineTo(x+w, y+h-r)
	p.CubicTo(x+w, y+h, x+w-r, y+h, x+w-r, y+h)
	p.LineTo(x+r, y+h)
	p.CubicTo(x, y+h, x, y+h-r, x, y+h-r)
	p.LineTo(x, y+r)
	p.CubicTo(x, y, x+r, y, x+r, y)
	p.Close()
	vector.FillPath(img, p, &vector.FillOptions{}, drawPathOpts(clr))
}

func lighten(c color.RGBA, amount uint8) color.RGBA {
	r := min(int(c.R)+int(amount), 255)
	g := min(int(c.G)+int(amount), 255)
	b := min(int(c.B)+int(amount), 255)
	return color.RGBA{uint8(r), uint8(g), uint8(b), c.A}
}

func darkenBy(c color.RGBA, amount uint8) color.RGBA {
	r := max(int(c.R)-int(amount), 0)
	g := max(int(c.G)-int(amount), 0)
	b := max(int(c.B)-int(amount), 0)
	return color.RGBA{uint8(r), uint8(g), uint8(b), c.A}
}

// --- NPC traffic sprites ---

func (sc *SpriteCache) generateTrafficSprites() {
	for i, clr := range SedanColors {
		sc.TrafficSedan[i] = generateSedanSprite(clr)
	}
	for i, clr := range TruckColors {
		sc.TrafficTruck[i] = generateTruckSprite(clr)
	}
	for i, clr := range SportColors {
		sc.TrafficSport[i] = generateSportSprite(clr)
	}
	sc.TrafficPolice = generatePoliceSprite()
	sc.PoliceSiren[0] = generateSirenFrame(true)
	sc.PoliceSiren[1] = generateSirenFrame(false)
	sc.TrafficOncoming = generateOncomingSprite()
	sc.HeadlightCone = generateHeadlightCone()
	sc.TurnSignal = generateTurnSignal()
}

func generateSedanSprite(clr color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(si(32), si(49))

	wheelClr := color.RGBA{0x1E, 0x1E, 0x28, 0xFF}
	drawRoundedRect(img, 1*S, 8*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 27*S, 8*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 1*S, 32*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 27*S, 32*S, 4*S, 8*S, wheelClr)

	body := &vector.Path{}
	body.MoveTo(16*S, 3*S)
	body.CubicTo(20*S, 2*S, 24*S, 3*S, 26*S, 6*S)
	body.LineTo(27*S, 14*S)
	body.LineTo(27*S, 36*S)
	body.CubicTo(27*S, 40*S, 25*S, 43*S, 22*S, 43*S)
	body.LineTo(10*S, 43*S)
	body.CubicTo(7*S, 43*S, 5*S, 40*S, 5*S, 36*S)
	body.LineTo(5*S, 14*S)
	body.LineTo(6*S, 6*S)
	body.CubicTo(8*S, 3*S, 12*S, 2*S, 16*S, 3*S)
	body.Close()
	vector.FillPath(img, body, &vector.FillOptions{}, drawPathOpts(clr))
	vector.StrokePath(img, body, &vector.StrokeOptions{Width: 0.8 * S}, drawPathOpts(lighten(clr, 30)))

	vector.FillPath(img, rectPath(10*S, 12*S, 12*S, 6*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0x8C, 0xDC, 0xFF, 0x60}))
	vector.FillPath(img, rectPath(11*S, 30*S, 10*S, 4*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0x8C, 0xDC, 0xFF, 0x40}))

	vector.FillCircle(img, 9*S, 42*S, 2*S, color.RGBA{0xFF, 0x88, 0x00, 0xCC}, true)
	vector.FillCircle(img, 23*S, 42*S, 2*S, color.RGBA{0xFF, 0x88, 0x00, 0xCC}, true)
	return img
}

func generateTruckSprite(clr color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(si(38), si(70))

	wheelClr := color.RGBA{0x1E, 0x1E, 0x28, 0xFF}
	drawRoundedRect(img, 1*S, 6*S, 5*S, 10*S, wheelClr)
	drawRoundedRect(img, 32*S, 6*S, 5*S, 10*S, wheelClr)
	drawRoundedRect(img, 1*S, 48*S, 5*S, 10*S, wheelClr)
	drawRoundedRect(img, 32*S, 48*S, 5*S, 10*S, wheelClr)
	drawRoundedRect(img, 1*S, 56*S, 5*S, 10*S, wheelClr)
	drawRoundedRect(img, 32*S, 56*S, 5*S, 10*S, wheelClr)

	cabClr := lighten(clr, 15)
	cab := &vector.Path{}
	cab.MoveTo(7*S, 0)
	cab.LineTo(31*S, 0)
	cab.CubicTo(33*S, 0, 34*S, 2*S, 34*S, 4*S)
	cab.LineTo(34*S, 22*S)
	cab.LineTo(4*S, 22*S)
	cab.LineTo(4*S, 4*S)
	cab.CubicTo(4*S, 2*S, 5*S, 0, 7*S, 0)
	cab.Close()
	vector.FillPath(img, cab, &vector.FillOptions{}, drawPathOpts(cabClr))
	vector.FillPath(img, rectPath(8*S, 4*S, 22*S, 10*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0x8C, 0xDC, 0xFF, 0x60}))

	cargo := &vector.Path{}
	cargo.MoveTo(3*S, 22*S)
	cargo.LineTo(35*S, 22*S)
	cargo.LineTo(35*S, 66*S)
	cargo.CubicTo(35*S, 68*S, 33*S, 69*S, 31*S, 69*S)
	cargo.LineTo(7*S, 69*S)
	cargo.CubicTo(5*S, 69*S, 3*S, 68*S, 3*S, 66*S)
	cargo.Close()
	vector.FillPath(img, cargo, &vector.FillOptions{}, drawPathOpts(clr))
	vector.StrokePath(img, cargo, &vector.StrokeOptions{Width: 0.8 * S}, drawPathOpts(lighten(clr, 20)))

	plankClr := lighten(clr, 10)
	vector.StrokeLine(img, 5*S, 35*S, 33*S, 35*S, 1*S, plankClr, true)
	vector.StrokeLine(img, 5*S, 48*S, 33*S, 48*S, 1*S, plankClr, true)
	vector.StrokeLine(img, 5*S, 60*S, 33*S, 60*S, 1*S, plankClr, true)

	vector.FillPath(img, rectPath(5*S, 65*S, 6*S, 3*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0xFF, 0x22, 0x22, 0xCC}))
	vector.FillPath(img, rectPath(27*S, 65*S, 6*S, 3*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0xFF, 0x22, 0x22, 0xCC}))
	return img
}

func generateSportSprite(clr color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(si(30), si(46))

	wheelClr := color.RGBA{0x1E, 0x1E, 0x28, 0xFF}
	drawRoundedRect(img, 0, 8*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 26*S, 8*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 0, 30*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 26*S, 30*S, 4*S, 8*S, wheelClr)

	body := &vector.Path{}
	body.MoveTo(15*S, 0)
	body.CubicTo(20*S, 0, 26*S, 6*S, 26*S, 12*S)
	body.LineTo(26*S, 34*S)
	body.CubicTo(26*S, 40*S, 22*S, 42*S, 18*S, 42*S)
	body.LineTo(12*S, 42*S)
	body.CubicTo(8*S, 42*S, 4*S, 40*S, 4*S, 34*S)
	body.LineTo(4*S, 12*S)
	body.CubicTo(4*S, 6*S, 10*S, 0, 15*S, 0)
	body.Close()
	vector.FillPath(img, body, &vector.FillOptions{}, drawPathOpts(clr))
	vector.StrokePath(img, body, &vector.StrokeOptions{Width: 0.8 * S}, drawPathOpts(lighten(clr, 30)))

	vector.FillPath(img, rectPath(9*S, 10*S, 12*S, 6*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0x8C, 0xDC, 0xFF, 0x60}))

	spoilerClr := darkenBy(clr, 40)
	vector.StrokeLine(img, 7*S, 40*S, 23*S, 40*S, 2*S, spoilerClr, true)
	drawRoundedRect(img, 7*S, 38*S, 2*S, 3*S, spoilerClr)
	drawRoundedRect(img, 21*S, 38*S, 2*S, 3*S, spoilerClr)

	vector.FillCircle(img, 8*S, 41*S, 2*S, color.RGBA{0xFF, 0x22, 0x22, 0xCC}, true)
	vector.FillCircle(img, 22*S, 41*S, 2*S, color.RGBA{0xFF, 0x22, 0x22, 0xCC}, true)
	return img
}

func generatePoliceSprite() *ebiten.Image {
	img := ebiten.NewImage(si(32), si(52))

	wheelClr := color.RGBA{0x1E, 0x1E, 0x28, 0xFF}
	drawRoundedRect(img, 2*S, 10*S, 4*S, 9*S, wheelClr)
	drawRoundedRect(img, 26*S, 10*S, 4*S, 9*S, wheelClr)
	drawRoundedRect(img, 2*S, 38*S, 4*S, 9*S, wheelClr)
	drawRoundedRect(img, 26*S, 38*S, 4*S, 9*S, wheelClr)

	topClr := color.RGBA{0xDD, 0xDD, 0xEE, 0xFF}
	botClr := color.RGBA{0x1A, 0x1A, 0x2E, 0xFF}

	bot := &vector.Path{}
	bot.MoveTo(6*S, 26*S)
	bot.LineTo(26*S, 26*S)
	bot.LineTo(27*S, 38*S)
	bot.CubicTo(27*S, 44*S, 25*S, 48*S, 22*S, 48*S)
	bot.LineTo(10*S, 48*S)
	bot.CubicTo(7*S, 48*S, 5*S, 44*S, 5*S, 38*S)
	bot.Close()
	vector.FillPath(img, bot, &vector.FillOptions{}, drawPathOpts(botClr))

	top := &vector.Path{}
	top.MoveTo(16*S, 3*S)
	top.CubicTo(22*S, 2*S, 26*S, 6*S, 27*S, 10*S)
	top.LineTo(27*S, 26*S)
	top.LineTo(5*S, 26*S)
	top.LineTo(5*S, 10*S)
	top.CubicTo(6*S, 6*S, 10*S, 2*S, 16*S, 3*S)
	top.Close()
	vector.FillPath(img, top, &vector.FillOptions{}, drawPathOpts(topClr))

	vector.StrokeLine(img, 5*S, 26*S, 27*S, 26*S, 1.5*S, color.RGBA{0xCC, 0xAA, 0x44, 0xFF}, true)
	vector.FillPath(img, rectPath(10*S, 10*S, 12*S, 7*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0x8C, 0xDC, 0xFF, 0x60}))
	return img
}

func generateSirenFrame(redLeft bool) *ebiten.Image {
	img := ebiten.NewImage(si(14), si(6))
	left := color.RGBA{0xFF, 0x00, 0x00, 0xFF}
	right := color.RGBA{0x00, 0x00, 0xFF, 0xFF}
	if !redLeft {
		left, right = right, left
	}
	vector.FillPath(img, rectPath(0, 0, 7*S, 6*S), &vector.FillOptions{}, drawPathOpts(left))
	vector.FillPath(img, rectPath(7*S, 0, 7*S, 6*S), &vector.FillOptions{}, drawPathOpts(right))
	return img
}

func generateOncomingSprite() *ebiten.Image {
	img := ebiten.NewImage(si(30), si(48))
	bodyClr := color.RGBA{0x33, 0x33, 0x3E, 0xFF}

	wheelClr := color.RGBA{0x1E, 0x1E, 0x28, 0xFF}
	drawRoundedRect(img, 0, 8*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 26*S, 8*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 0, 32*S, 4*S, 8*S, wheelClr)
	drawRoundedRect(img, 26*S, 32*S, 4*S, 8*S, wheelClr)

	body := &vector.Path{}
	body.MoveTo(15*S, 46*S)
	body.CubicTo(20*S, 46*S, 26*S, 40*S, 26*S, 34*S)
	body.LineTo(26*S, 12*S)
	body.CubicTo(26*S, 6*S, 22*S, 4*S, 18*S, 4*S)
	body.LineTo(12*S, 4*S)
	body.CubicTo(8*S, 4*S, 4*S, 6*S, 4*S, 12*S)
	body.LineTo(4*S, 34*S)
	body.CubicTo(4*S, 40*S, 10*S, 46*S, 15*S, 46*S)
	body.Close()
	vector.FillPath(img, body, &vector.FillOptions{}, drawPathOpts(bodyClr))
	vector.StrokePath(img, body, &vector.StrokeOptions{Width: 0.8 * S}, drawPathOpts(lighten(bodyClr, 20)))

	vector.FillPath(img, rectPath(9*S, 34*S, 12*S, 6*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0x8C, 0xDC, 0xFF, 0x50}))

	vector.FillCircle(img, 9*S, 44*S, 3*S, color.RGBA{0xFF, 0xF0, 0x64, 0xFF}, true)
	vector.FillCircle(img, 21*S, 44*S, 3*S, color.RGBA{0xFF, 0xF0, 0x64, 0xFF}, true)
	return img
}

func generateHeadlightCone() *ebiten.Image {
	img := ebiten.NewImage(si(30), si(40))
	cone := &vector.Path{}
	cone.MoveTo(8*S, 0)
	cone.LineTo(0, 40*S)
	cone.CubicTo(7*S, 36*S, 23*S, 36*S, 30*S, 40*S)
	cone.LineTo(22*S, 0)
	cone.Close()
	vector.FillPath(img, cone, &vector.FillOptions{}, drawPathOpts(color.RGBA{0xFF, 0xF0, 0x64, 0x19}))
	return img
}

func generateTurnSignal() *ebiten.Image {
	img := ebiten.NewImage(si(6), si(6))
	vector.FillCircle(img, 3*S, 3*S, 2.5*S, color.RGBA{0xFF, 0xCC, 0x00, 0xDD}, true)
	return img
}

// --- Item sprites ---

func (sc *SpriteCache) generateItemSprites() {
	sc.FuelCan = generateFuelCanSprite()
	sc.FuelGlow = generateGlowCircle(18, 22, color.RGBA{0x22, 0xCC, 0x44, 0x30})
	sc.NitroItem = generateNitroSprite()
	sc.NitroGlow = generateGlowCircle(20, 20, color.RGBA{0xFF, 0xD7, 0x00, 0x30})
	sc.OilSpill = generateOilSpillSprite()
	sc.generateCoinFrames()
	sc.RepairKit = generateRepairSprite()
	sc.RepairGlow = generateGlowCircle(18, 18, color.RGBA{0x00, 0xDD, 0xFF, 0x30})
}

func generateFuelCanSprite() *ebiten.Image {
	img := ebiten.NewImage(si(18), si(22))

	body := &vector.Path{}
	body.MoveTo(5*S, 3*S)
	body.CubicTo(5*S, 1*S, 7*S, 0, 9*S, 0)
	body.CubicTo(11*S, 0, 13*S, 1*S, 13*S, 3*S)
	body.LineTo(13*S, 17*S)
	body.CubicTo(13*S, 19*S, 11*S, 20*S, 9*S, 20*S)
	body.CubicTo(7*S, 20*S, 5*S, 19*S, 5*S, 17*S)
	body.Close()
	vector.FillPath(img, body, &vector.FillOptions{}, drawPathOpts(color.RGBA{0x22, 0xCC, 0x44, 0xFF}))
	vector.StrokePath(img, body, &vector.StrokeOptions{Width: 0.8 * S}, drawPathOpts(color.RGBA{0x33, 0xDD, 0x55, 0xFF}))

	vector.FillPath(img, rectPath(7*S, 0, 4*S, 3*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0x88, 0x88, 0x88, 0xFF}))

	vector.StrokeLine(img, 9*S, 6*S, 9*S, 16*S, 2*S, color.RGBA{0xFF, 0xFF, 0xFF, 0xDD}, true)
	vector.StrokeLine(img, 6*S, 11*S, 12*S, 11*S, 2*S, color.RGBA{0xFF, 0xFF, 0xFF, 0xDD}, true)
	return img
}

func generateNitroSprite() *ebiten.Image {
	img := ebiten.NewImage(si(20), si(20))

	diamond := &vector.Path{}
	diamond.MoveTo(10*S, 1*S)
	diamond.LineTo(19*S, 10*S)
	diamond.LineTo(10*S, 19*S)
	diamond.LineTo(1*S, 10*S)
	diamond.Close()
	vector.FillPath(img, diamond, &vector.FillOptions{}, drawPathOpts(color.RGBA{0xFF, 0xD7, 0x00, 0xFF}))
	vector.StrokePath(img, diamond, &vector.StrokeOptions{Width: 0.8 * S}, drawPathOpts(color.RGBA{0xFF, 0xEE, 0x44, 0xFF}))

	bolt := &vector.Path{}
	bolt.MoveTo(11*S, 3*S)
	bolt.LineTo(7*S, 9*S)
	bolt.LineTo(10*S, 9*S)
	bolt.LineTo(8*S, 17*S)
	bolt.LineTo(13*S, 10*S)
	bolt.LineTo(10*S, 10*S)
	bolt.Close()
	vector.FillPath(img, bolt, &vector.FillOptions{}, drawPathOpts(color.RGBA{0xCC, 0x66, 0x00, 0xFF}))
	return img
}

func generateOilSpillSprite() *ebiten.Image {
	img := ebiten.NewImage(si(26), si(26))
	base := color.RGBA{0x28, 0x1E, 0x14, 0x96}
	vector.FillCircle(img, 13*S, 13*S, 9*S, base, true)
	vector.FillCircle(img, 8*S, 11*S, 6*S, base, true)
	vector.FillCircle(img, 17*S, 15*S, 5*S, base, true)
	vector.FillCircle(img, 12*S, 8*S, 4*S, base, true)

	vector.FillCircle(img, 11*S, 11*S, 2.5*S, color.RGBA{0x50, 0x3C, 0x28, 0x50}, true)
	vector.FillCircle(img, 15*S, 13*S, 2*S, color.RGBA{0x3C, 0x32, 0x1E, 0x3C}, true)
	return img
}

func (sc *SpriteCache) generateCoinFrames() {
	sz := si(14)

	f0 := ebiten.NewImage(sz, sz)
	vector.FillCircle(f0, 7*S, 7*S, 6*S, color.RGBA{0xFF, 0xD7, 0x00, 0xFF}, true)
	vector.StrokeLine(f0, 7*S, 1*S, 7*S, 13*S, 0.5*S, color.RGBA{0xCC, 0x99, 0x00, 0xFF}, true)
	vector.FillCircle(f0, 7*S, 7*S, 2.5*S, color.RGBA{0xFF, 0xEE, 0x88, 0xFF}, true)
	sc.Coin[0] = f0

	f1 := ebiten.NewImage(sz, sz)
	ellipse := &vector.Path{}
	cx, cy := 7*S, 7*S
	rx, ry := 3.5*S, 6*S
	ellipse.MoveTo(cx, cy-ry)
	ellipse.CubicTo(cx+rx*0.55, cy-ry, cx+rx, cy-ry*0.55, cx+rx, cy)
	ellipse.CubicTo(cx+rx, cy+ry*0.55, cx+rx*0.55, cy+ry, cx, cy+ry)
	ellipse.CubicTo(cx-rx*0.55, cy+ry, cx-rx, cy+ry*0.55, cx-rx, cy)
	ellipse.CubicTo(cx-rx, cy-ry*0.55, cx-rx*0.55, cy-ry, cx, cy-ry)
	ellipse.Close()
	vector.FillPath(f1, ellipse, &vector.FillOptions{}, drawPathOpts(color.RGBA{0xFF, 0xD7, 0x00, 0xFF}))
	vector.StrokePath(f1, ellipse, &vector.StrokeOptions{Width: 0.5 * S}, drawPathOpts(color.RGBA{0xCC, 0x99, 0x00, 0xFF}))
	sc.Coin[1] = f1

	f2 := ebiten.NewImage(sz, sz)
	vector.FillPath(f2, rectPath(6*S, 1*S, 2*S, 12*S), &vector.FillOptions{},
		drawPathOpts(color.RGBA{0xCC, 0x99, 0x00, 0xFF}))
	sc.Coin[2] = f2

	sc.Coin[3] = f1
}

func generateRepairSprite() *ebiten.Image {
	img := ebiten.NewImage(si(16), si(16))
	clr := color.RGBA{0x00, 0xDD, 0xFF, 0xFF}
	// Cross: vertical bar.
	vector.FillPath(img, rectPath(6*S, 1*S, 4*S, 14*S), &vector.FillOptions{},
		drawPathOpts(clr))
	// Cross: horizontal bar.
	vector.FillPath(img, rectPath(1*S, 6*S, 14*S, 4*S), &vector.FillOptions{},
		drawPathOpts(clr))
	// Bright center dot.
	vector.FillCircle(img, 8*S, 8*S, 2*S, color.RGBA{0xAA, 0xFF, 0xFF, 0xFF}, true)
	return img
}

func generateGlowCircle(w, h int, clr color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(si(w+8), si(h+8))
	cx := float32(si(w+8)) / 2
	cy := float32(si(h+8)) / 2
	r := float32(max(w, h)) / 2 * S
	vector.FillCircle(img, cx, cy, r+2*S, clr, true)
	return img
}

// --- Decor sprites ---

func (sc *SpriteCache) generateDecorSprites() {
	widths := [3]int{22, 32, 42}
	idx := 0
	for _, bw := range widths {
		for roofStyle := range 4 {
			bh := 60 + roofStyle*20
			sc.Buildings[idx] = generateBuildingSprite(bw, bh, roofStyle)
			idx++
		}
	}
	sc.LampPost = generateLampPostSprite()
}

func generateBuildingSprite(bw, bh, roofStyle int) *ebiten.Image {
	w, h := bw, bh+15
	img := ebiten.NewImage(si(w), si(h))

	baseY := float32(si(h - bh))
	fw := float32(si(bw))
	fh := float32(si(h))

	bodyClr := color.RGBA{0xCC, 0xCC, 0xCC, 0xFF}
	body := &vector.Path{}
	body.MoveTo(0, baseY)
	body.LineTo(fw, baseY)
	body.LineTo(fw, fh)
	body.LineTo(0, fh)
	body.Close()
	vector.FillPath(img, body, &vector.FillOptions{}, drawPathOpts(bodyClr))

	vector.StrokeLine(img, 0, baseY, 0, fh, 1*S, color.RGBA{0xEE, 0xEE, 0xEE, 0xFF}, true)

	switch roofStyle {
	case 0:
		vector.StrokeLine(img, 0, baseY, fw, baseY, 1.5*S, color.RGBA{0xDD, 0xDD, 0xDD, 0xFF}, true)
	case 1:
		spire := &vector.Path{}
		spire.MoveTo(fw/2, 0)
		spire.LineTo(fw/2+6*S, baseY)
		spire.LineTo(fw/2-6*S, baseY)
		spire.Close()
		vector.FillPath(img, spire, &vector.FillOptions{}, drawPathOpts(color.RGBA{0xBB, 0xBB, 0xBB, 0xFF}))
	case 2:
		stepW := fw * 0.5
		stepH := 10 * S
		sx := (fw - stepW) / 2
		vector.FillPath(img, rectPath(sx, baseY-stepH, stepW, stepH), &vector.FillOptions{},
			drawPathOpts(color.RGBA{0xBB, 0xBB, 0xBB, 0xFF}))
	case 3:
		cx := fw / 2
		vector.StrokeLine(img, cx, 0, cx, baseY, 1*S, color.RGBA{0xAA, 0xAA, 0xAA, 0xFF}, true)
		vector.FillCircle(img, cx, 1*S, 2*S, color.RGBA{0xFF, 0x44, 0x44, 0xFF}, true)
	}

	winClr := color.RGBA{0xFF, 0xFF, 0xFF, 0x88}
	cols := bw / 8
	rows := bh / 12
	for r := range rows {
		for c := range cols {
			wx := float32(3+c*8) * S
			wy := baseY + float32(4+r*12)*S
			vector.FillPath(img, rectPath(wx, wy, 3*S, 3*S), &vector.FillOptions{}, drawPathOpts(winClr))
		}
	}
	return img
}

func generateLampPostSprite() *ebiten.Image {
	img := ebiten.NewImage(si(20), si(60))

	poleX := 5 * S
	poleClr := color.RGBA{0x88, 0x88, 0x88, 0xFF}
	vector.StrokeLine(img, poleX, 20*S, poleX, 60*S, 2*S, poleClr, true)

	neck := &vector.Path{}
	neck.MoveTo(poleX, 20*S)
	neck.CubicTo(poleX, 14*S, poleX+8*S, 12*S, poleX+12*S, 14*S)
	vector.StrokePath(img, neck, &vector.StrokeOptions{Width: 1.5 * S}, drawPathOpts(poleClr))

	lampX, lampY := poleX+12*S, 14*S
	vector.FillCircle(img, lampX, lampY, 3*S, color.RGBA{0xFF, 0xDC, 0x64, 0xB4}, true)

	cone := &vector.Path{}
	cone.MoveTo(lampX, lampY+2*S)
	cone.LineTo(lampX-6*S, lampY+22*S)
	cone.LineTo(lampX+6*S, lampY+22*S)
	cone.Close()
	vector.FillPath(img, cone, &vector.FillOptions{}, drawPathOpts(color.RGBA{0xFF, 0xDC, 0x64, 0x14}))
	return img
}

// rectPath creates a simple rectangle Path.
func rectPath(x, y, w, h float32) *vector.Path {
	p := &vector.Path{}
	p.MoveTo(x, y)
	p.LineTo(x+w, y)
	p.LineTo(x+w, y+h)
	p.LineTo(x, y+h)
	p.Close()
	return p
}
