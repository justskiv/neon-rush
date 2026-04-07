package main

import (
	"image/color"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// DecorType identifies a decoration element type.
type DecorType int

const (
	DecorBuilding DecorType = iota
	DecorLampPost
	DecorBush
	DecorTunnelWall
	DecorMountain
	DecorCactus
	DecorBillboard
)

// DecorObject is a single decorative element alongside the road.
type DecorObject struct {
	X, Y            float64
	Width, Height   float64
	Type            DecorType
	Color           color.RGBA
	WindowLit       [6]bool  // for buildings: which windows are lit
	SpriteIdx       int      // index into sprite arrays (buildings)
	ParallaxFactor  float64  // scroll speed multiplier (0.6 for far decor, 1.0 for road-edge)
}

// DecorSystem manages scrolling decorations on both sides of the road.
type DecorSystem struct {
	Left       []DecorObject
	Right      []DecorObject
	SpawnTimer int
}

func NewDecorSystem() DecorSystem {
	return DecorSystem{
		Left:  make([]DecorObject, 0, 32),
		Right: make([]DecorObject, 0, 32),
	}
}

func (ds *DecorSystem) Update(scrollSpeed float64, zoneID ZoneID, palette ZonePalette) {
	// Move all decor down.
	ds.Left = updateDecorSide(ds.Left, scrollSpeed)
	ds.Right = updateDecorSide(ds.Right, scrollSpeed)

	// Spawn new decorations at top.
	ds.SpawnTimer--
	if ds.SpawnTimer <= 0 {
		ds.spawnForZone(zoneID, palette)
		ds.SpawnTimer = 30 + rand.IntN(30) // every 0.5-1 second
	}
}

func updateDecorSide(objs []DecorObject, scrollSpeed float64) []DecorObject {
	n := 0
	for i := range objs {
		pf := objs[i].ParallaxFactor
		if pf == 0 {
			pf = 1.0
		}
		objs[i].Y += scrollSpeed * pf
		if objs[i].Y-objs[i].Height <= ScreenHeight {
			objs[n] = objs[i]
			n++
		}
	}
	return objs[:n]
}

func (ds *DecorSystem) spawnForZone(zoneID ZoneID, palette ZonePalette) {
	switch zoneID {
	case ZoneNightCity:
		ds.spawnBuilding(true, palette)
		ds.spawnBuilding(false, palette)
	case ZoneHighway:
		ds.spawnLampPost(true, palette)
		ds.spawnLampPost(false, palette)
		if rand.IntN(3) == 0 {
			ds.spawnBush(true)
			ds.spawnBush(false)
		}
		if rand.IntN(4) == 0 {
			ds.spawnBillboard(rand.IntN(2) == 0)
		}
	case ZoneRain:
		// Minimal decor in rain.
		if rand.IntN(3) == 0 {
			ds.spawnLampPost(true, palette)
		}
	case ZoneNeonTunnel:
		ds.spawnTunnelWall(true, palette)
		ds.spawnTunnelWall(false, palette)
	case ZoneSunsetChaos:
		if rand.IntN(2) == 0 {
			ds.spawnCactus(true)
			ds.spawnCactus(false)
		}
		if rand.IntN(5) == 0 {
			ds.spawnBillboard(rand.IntN(2) == 0)
		}
	}
}

func (ds *DecorSystem) spawnBuilding(left bool, palette ZonePalette) {
	w := 20.0 + rand.Float64()*25
	h := 40.0 + rand.Float64()*80
	x := 5.0
	if !left {
		x = float64(RoadRight) + curbWidth + 5
	}
	obj := DecorObject{
		X: x, Y: -h,
		Width: w, Height: h,
		Type:           DecorBuilding,
		Color:          darken(palette.Background, 0.6),
		SpriteIdx:      rand.IntN(NumBuildingVariants),
		ParallaxFactor: DecorParallaxFactor,
	}
	for i := range obj.WindowLit {
		obj.WindowLit[i] = rand.IntN(3) != 0
	}
	if left {
		ds.Left = append(ds.Left, obj)
	} else {
		ds.Right = append(ds.Right, obj)
	}
}

func (ds *DecorSystem) spawnLampPost(left bool, palette ZonePalette) {
	x := 20.0
	if !left {
		x = float64(RoadRight) + curbWidth + 15
	}
	obj := DecorObject{
		X: x, Y: -60,
		Width: 2, Height: 55,
		Type:           DecorLampPost,
		Color:          color.RGBA{0x88, 0x88, 0x88, 0xFF},
		ParallaxFactor: DecorParallaxFactor,
	}
	if left {
		ds.Left = append(ds.Left, obj)
	} else {
		ds.Right = append(ds.Right, obj)
	}
}

func (ds *DecorSystem) spawnBush(left bool) {
	x := 10.0 + rand.Float64()*20
	if !left {
		x = float64(RoadRight) + curbWidth + 10 + rand.Float64()*20
	}
	obj := DecorObject{
		X: x, Y: -12,
		Width: 10 + rand.Float64()*8, Height: 8 + rand.Float64()*6,
		Type:           DecorBush,
		Color:          color.RGBA{0x22, 0x66, 0x22, 0xFF},
		ParallaxFactor: DecorParallaxFactor,
	}
	if left {
		ds.Left = append(ds.Left, obj)
	} else {
		ds.Right = append(ds.Right, obj)
	}
}

func (ds *DecorSystem) spawnTunnelWall(left bool, palette ZonePalette) {
	x := 0.0
	w := float64(RoadLeft) - curbWidth
	if !left {
		x = float64(RoadRight) + curbWidth
		w = float64(ScreenWidth) - x
	}
	clr := palette.NeonAccent
	if rand.IntN(2) == 0 && palette.NeonAccent2.A > 0 {
		clr = palette.NeonAccent2
	}
	obj := DecorObject{
		X: x, Y: -6,
		Width: w, Height: 3,
		Type:  DecorTunnelWall,
		Color: color.RGBA{clr.R, clr.G, clr.B, 0x80},
	}
	if left {
		ds.Left = append(ds.Left, obj)
	} else {
		ds.Right = append(ds.Right, obj)
	}
}

func (ds *DecorSystem) spawnCactus(left bool) {
	x := 15.0 + rand.Float64()*15
	if !left {
		x = float64(RoadRight) + curbWidth + 15 + rand.Float64()*15
	}
	obj := DecorObject{
		X: x, Y: -30,
		Width: 4, Height: 25 + rand.Float64()*10,
		Type:  DecorCactus,
		Color: color.RGBA{0x22, 0x55, 0x22, 0xFF},
	}
	if left {
		ds.Left = append(ds.Left, obj)
	} else {
		ds.Right = append(ds.Right, obj)
	}
}

func (ds *DecorSystem) spawnBillboard(left bool) {
	x := 2.0
	if !left {
		x = float64(RoadRight) + curbWidth + 4
	}
	colors := []color.RGBA{
		{0xFF, 0x44, 0x44, 0xFF},
		{0x44, 0xFF, 0x44, 0xFF},
		{0x44, 0x44, 0xFF, 0xFF},
		{0xFF, 0xAA, 0x00, 0xFF},
		{0xFF, 0x00, 0xFF, 0xFF},
	}
	obj := DecorObject{
		X: x, Y: -22,
		Width: 30, Height: 18,
		Type:           DecorBillboard,
		Color:          colors[rand.IntN(len(colors))],
		ParallaxFactor: DecorParallaxFactor,
	}
	if left {
		ds.Left = append(ds.Left, obj)
	} else {
		ds.Right = append(ds.Right, obj)
	}
}

func (ds *DecorSystem) Draw(screen *ebiten.Image, palette ZonePalette, sprites *SpriteCache, offsetFn func(float64) float64) {
	drawDecorSide(screen, ds.Left, palette, sprites, offsetFn)
	drawDecorSide(screen, ds.Right, palette, sprites, offsetFn)
}

func drawDecorSide(screen *ebiten.Image, objs []DecorObject, palette ZonePalette, sprites *SpriteCache, offsetFn func(float64) float64) {
	for _, obj := range objs {
		ox := obj.X + offsetFn(obj.Y)*0.5 // parallax: decor shifts less than road
		switch obj.Type {
		case DecorBuilding:
			rs := renderScaleGlobal
			bldg := sprites.Buildings[obj.SpriteIdx]
			if bldg != nil {
				op := &ebiten.DrawImageOptions{}
				bw, bh := bldg.Bounds().Dx(), bldg.Bounds().Dy()
				sx := obj.Width * rs / float64(bw)
				sy := obj.Height * rs / float64(bh)
				op.GeoM.Scale(sx, sy)
				op.GeoM.Translate(ox*rs, obj.Y*rs)
				op.ColorScale.ScaleWithColor(obj.Color)
				screen.DrawImage(bldg, op)
			}
			winSize := 3.0
			cols := int(obj.Width / 8)
			rows := int(obj.Height / 12)
			for r := range rows {
				for c := range cols {
					if c*rows+r < len(obj.WindowLit) && obj.WindowLit[c*rows+r] {
						wx := ox + 4 + float64(c)*8
						wy := obj.Y + 4 + float64(r)*12
						DrawRect(screen, wx, wy, winSize, winSize, palette.NeonAccent)
					}
				}
			}

		case DecorLampPost:
			drawSprite(screen, sprites.LampPost, ox+1, obj.Y+obj.Height/2)

		case DecorBush:
			DrawRect(screen, ox, obj.Y, obj.Width, obj.Height, obj.Color)

		case DecorTunnelWall:
			DrawRect(screen, ox, obj.Y, obj.Width, obj.Height, obj.Color)

		case DecorCactus:
			DrawRect(screen, ox, obj.Y, obj.Width, obj.Height, obj.Color)
			armY := obj.Y + obj.Height*0.3
			DrawRect(screen, ox-4, armY, 4, 3, obj.Color)
			DrawRect(screen, ox+obj.Width, armY+8, 4, 3, obj.Color)

		case DecorBillboard:
			DrawRect(screen, ox+obj.Width/2-1, obj.Y+obj.Height, 2, 10,
				color.RGBA{0x66, 0x66, 0x66, 0xFF})
			DrawRect(screen, ox, obj.Y, obj.Width, obj.Height, obj.Color)
			DrawRect(screen, ox, obj.Y, obj.Width, 2,
				color.RGBA{0xFF, 0xFF, 0xFF, 0x88})
		}
	}
}

// DrawFogOverlay renders a semi-transparent fog layer for the Rain zone.
func DrawFogOverlay(screen *ebiten.Image) {
	DrawRect(screen, 0, 0, ScreenWidth, ScreenHeight,
		color.RGBA{0x0D, 0x1A, 0x1A, 0x30})
}

func darken(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}
