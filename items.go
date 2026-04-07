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
	ItemRepair
)

// Item is a pickup object on the road.
type Item struct {
	X, Y          float64
	Width, Height float64
	Type          ItemType
	Color         color.RGBA
	TickAge       int
}

var (
	colorFuel  = color.RGBA{0x00, 0xDD, 0x00, 0xFF}
	colorNitro = color.RGBA{0xFF, 0xDD, 0x00, 0xFF}
	colorOil   = color.RGBA{0x44, 0x33, 0x22, 0x99}
	colorCoin   = color.RGBA{0xFF, 0xCC, 0x00, 0xFF}
	colorRepair = color.RGBA{0x00, 0xDD, 0xFF, 0xFF}
)

// itemDef returns the size and color for an item type.
func itemDef(t ItemType) (w, h float64, clr color.RGBA) {
	switch t {
	case ItemFuel:
		return 22, 26, colorFuel
	case ItemNitro:
		return 24, 24, colorNitro
	case ItemOil:
		return 20, 14, colorOil
	case ItemCoin:
		return 18, 18, colorCoin
	case ItemRepair:
		return 22, 22, colorRepair
	}
	return 10, 10, colorCoin
}

// xToLane converts an X coordinate to a lane index (0-based).
func xToLane(x float64) int {
	lane := int((x - float64(RoadLeft)) / LaneWidth)
	return max(0, min(lane, LaneCount-1))
}

// laneToX converts a lane index to the lane center X.
func laneToX(lane int) float64 {
	return float64(RoadLeft) + float64(lane)*LaneWidth + LaneWidth/2
}

// lanesNearPlayer returns lanes ordered by distance from the player's lane.
func lanesNearPlayer(playerLane int) []int {
	lanes := make([]int, 0, LaneCount)
	lanes = append(lanes, playerLane)
	for offset := 1; offset < LaneCount; offset++ {
		if l := playerLane - offset; l >= 0 {
			lanes = append(lanes, l)
		}
		if l := playerLane + offset; l < LaneCount {
			lanes = append(lanes, l)
		}
	}
	return lanes
}

// SpawnItem creates a pickup preferring the player's lane or adjacent ones.
func SpawnItem(itemType ItemType, existing []*Item, traffic []*TrafficCar, playerX ...float64) *Item {
	w, h, clr := itemDef(itemType)

	pLane := LaneCount / 2
	if len(playerX) > 0 {
		pLane = xToLane(playerX[0])
	}

	for _, lane := range lanesNearPlayer(pLane) {
		cx := laneToX(lane)
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
		if math.Abs(it.X-cx) < LaneWidth/2 && it.Y < 200 {
			return true
		}
	}
	for _, c := range traffic {
		if math.Abs(c.X-cx) < LaneWidth/2 && c.Y < 300 {
			return true
		}
	}
	return false
}

// UpdateItems moves items down, applies magnetism, and removes off-screen ones.
func UpdateItems(items []*Item, scrollSpeed, playerX, playerY float64, offsetFn func(float64) float64) []*Item {
	n := 0
	for _, it := range items {
		it.TickAge++
		it.Y += scrollSpeed

		// Magnetism: pull pickups toward player when close (not oil).
		// Item X is in road-space; compare to player using per-Y offset.
		if it.Type != ItemOil {
			dx := playerX - (it.X + offsetFn(it.Y))
			dy := playerY - it.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 40 && dist > 1 {
				it.X += dx / dist * 2.5
				it.Y += dy / dist * 1.5
			}
		}
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
// offsetFn returns per-Y offset to shift item X into screen space.
func CheckPlayerItemCollision(p *Player, items []*Item, offsetFn func(float64) float64) ([]*Item, []*Item) {
	pb := p.Bounds()
	var picked []*Item
	n := 0
	for _, it := range items {
		ib := NewRect(it.X+offsetFn(it.Y), it.Y, it.Width, it.Height)
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

func (it *Item) Draw(screen *ebiten.Image, sprites *SpriteCache, offsetFn func(float64) float64) {
	dx := it.X + offsetFn(it.Y)
	switch it.Type {
	case ItemFuel:
		glowAlpha := float32(0.3 + 0.2*math.Sin(float64(it.TickAge)*0.1))
		drawSpriteAlpha(screen, sprites.FuelGlow, dx, it.Y, glowAlpha)
		drawSprite(screen, sprites.FuelCan, dx, it.Y)
	case ItemNitro:
		glowAlpha := float32(0.3 + 0.2*math.Sin(float64(it.TickAge)*0.12))
		drawSpriteAlpha(screen, sprites.NitroGlow, dx, it.Y, glowAlpha)
		img := sprites.NitroItem
		rs := renderScaleGlobal
		op := &ebiten.DrawImageOptions{}
		s := rs / SpriteScale
		op.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
		op.GeoM.Scale(s, s)
		op.GeoM.Rotate(math.Sin(float64(it.TickAge)*0.1) * 0.09)
		op.GeoM.Translate(dx*rs, it.Y*rs)
		op.Filter = ebiten.FilterLinear
		screen.DrawImage(img, op)
	case ItemOil:
		drawSprite(screen, sprites.OilSpill, dx, it.Y)
	case ItemCoin:
		frame := (it.TickAge / 8) % 4
		drawSprite(screen, sprites.Coin[frame], dx, it.Y)
	case ItemRepair:
		glowAlpha := float32(0.3 + 0.25*math.Sin(float64(it.TickAge)*0.15))
		drawSpriteAlpha(screen, sprites.RepairGlow, dx, it.Y, glowAlpha)
		drawSprite(screen, sprites.RepairKit, dx, it.Y)
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
				Width: 18, Height: 18,
				Type: ItemCoin, Color: colorCoin,
			})
		}
		return coins
	}
	return nil
}
