package main

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// ItemType identifies a pickup type.
type ItemType int

const (
	ItemFuel ItemType = iota
	ItemNitro
	ItemOil
	ItemCoin
)

// Item is a pickup object on the road.
type Item struct {
	X, Y          float64
	Width, Height float64
	Type          ItemType
	Color         color.RGBA
}

var (
	colorFuel  = color.RGBA{0x00, 0xDD, 0x00, 0xFF}
	colorNitro = color.RGBA{0xFF, 0xDD, 0x00, 0xFF}
	colorOil   = color.RGBA{0x44, 0x33, 0x22, 0x99}
	colorCoin  = color.RGBA{0xFF, 0xCC, 0x00, 0xFF}
)

// SpawnItem creates a pickup on a random free lane.
func SpawnItem(itemType ItemType, existing []*Item, traffic []*TrafficCar) *Item {
	var w, h float64
	var clr color.RGBA

	switch itemType {
	case ItemFuel:
		w, h, clr = 12, 16, colorFuel
	case ItemNitro:
		w, h, clr = 14, 14, colorNitro
	case ItemOil:
		w, h, clr = 20, 14, colorOil
	case ItemCoin:
		w, h, clr = 10, 10, colorCoin
	}

	start := rand.IntN(LaneCount)
	for attempt := range LaneCount {
		lane := (start + attempt) % LaneCount
		cx := float64(RoadLeft) + float64(lane)*LaneWidth + LaneWidth/2

		if isItemLaneBlocked(existing, traffic, cx) {
			continue
		}

		return &Item{
			X: cx, Y: -h / 2,
			Width: w, Height: h,
			Type: itemType, Color: clr,
		}
	}
	return nil
}

func isItemLaneBlocked(items []*Item, traffic []*TrafficCar, cx float64) bool {
	for _, it := range items {
		if math.Abs(it.X-cx) < LaneWidth/2 && it.Y < 100 {
			return true
		}
	}
	for _, c := range traffic {
		if math.Abs(c.X-cx) < LaneWidth/2 && c.Y < 100 {
			return true
		}
	}
	return false
}

// UpdateItems moves items down and removes off-screen ones in-place.
func UpdateItems(items []*Item, scrollSpeed float64) []*Item {
	n := 0
	for _, it := range items {
		it.Y += scrollSpeed
		if it.Y-it.Height/2 <= ScreenHeight {
			items[n] = it
			n++
		}
	}
	for i := n; i < len(items); i++ {
		items[i] = nil
	}
	return items[:n]
}

// CheckPlayerItemCollision splits items into picked-up and remaining.
func CheckPlayerItemCollision(p *Player, items []*Item) ([]*Item, []*Item) {
	pb := p.Bounds()
	var picked []*Item
	n := 0
	for _, it := range items {
		ib := NewRect(it.X, it.Y, it.Width, it.Height)
		if CheckCollision(pb, ib) {
			picked = append(picked, it)
		} else {
			items[n] = it
			n++
		}
	}
	for i := n; i < len(items); i++ {
		items[i] = nil
	}
	return picked, items[:n]
}

func (it *Item) Draw(screen *ebiten.Image) {
	switch it.Type {
	case ItemFuel:
		DrawRect(screen, it.X-it.Width/2, it.Y-it.Height/2, it.Width, it.Height, it.Color)
		// "F" letter indicator — small dark rect in center.
		DrawRect(screen, it.X-2, it.Y-3, 4, 6, color.RGBA{0x00, 0x66, 0x00, 0xFF})
	case ItemNitro:
		// Draw as a diamond (rotated pixel).
		s := it.Width * 0.7
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(s, s)
		op.GeoM.Translate(-s/2, -s/2)
		op.GeoM.Rotate(math.Pi / 4)
		op.GeoM.Translate(it.X, it.Y)
		op.ColorScale.ScaleWithColor(it.Color)
		screen.DrawImage(pixel, op)
	case ItemOil:
		// Dark semi-transparent blob.
		DrawRect(screen, it.X-it.Width/2, it.Y-it.Height/2, it.Width, it.Height, it.Color)
		// Slightly offset smaller blob for organic look.
		DrawRect(screen, it.X-it.Width/3, it.Y-it.Height/3, it.Width*0.6, it.Height*0.6,
			color.RGBA{0x33, 0x22, 0x11, 0x77})
	case ItemCoin:
		// Golden square with shimmer.
		DrawRect(screen, it.X-it.Width/2, it.Y-it.Height/2, it.Width, it.Height, it.Color)
		// Bright center.
		DrawRect(screen, it.X-2, it.Y-2, 4, 4, color.RGBA{0xFF, 0xFF, 0x88, 0xFF})
	}
}

// SpawnCoinLine spawns a line of 3-5 coins on one lane.
func SpawnCoinLine(existing []*Item, traffic []*TrafficCar) []*Item {
	count := 3 + rand.IntN(3) // 3-5
	start := rand.IntN(LaneCount)
	for attempt := range LaneCount {
		lane := (start + attempt) % LaneCount
		cx := float64(RoadLeft) + float64(lane)*LaneWidth + LaneWidth/2
		if isItemLaneBlocked(existing, traffic, cx) {
			continue
		}
		coins := make([]*Item, 0, count)
		for i := range count {
			coins = append(coins, &Item{
				X: cx, Y: -10 - float64(i)*30,
				Width: 10, Height: 10,
				Type: ItemCoin, Color: colorCoin,
			})
		}
		return coins
	}
	return nil
}
