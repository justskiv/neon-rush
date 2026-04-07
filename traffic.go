package main

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// TrafficCar is an NPC vehicle on the road.
type TrafficCar struct {
	X, Y          float64
	Width, Height float64
	Color         color.RGBA
	Speed         float64 // fixed at spawn time
	CarType       CarType

	NearMissChecked bool
	ColorVariant    int // index into sprite color variants

	// Sport lane-change fields.
	LaneChangeTimer int
	TargetX         float64
	ChangingLane    bool

	// Police chase fields.
	ChaseTimer int

	// Tick counter for visual effects (siren, headlights).
	TickAge int
}

// SpawnTraffic creates a new NPC on a random free lane.
func SpawnTraffic(existing []*TrafficCar, scrollSpeed float64, tickCount int) *TrafficCar {
	ct, def := chooseCarType(tickCount)

	start := rand.IntN(LaneCount)
	for attempt := range LaneCount {
		lane := (start + attempt) % LaneCount
		cx := float64(RoadLeft) + float64(lane)*LaneWidth + LaneWidth/2

		if !canSpawnSafely(existing, cx, def.Height) {
			continue
		}

		car := &TrafficCar{
			X:       cx,
			Y:       -def.Height / 2,
			Width:   def.Width,
			Height:  def.Height,
			CarType: ct,
		}

		car.Speed = scrollSpeed * def.SpeedRatio
		switch ct {
		case CarTypeSedan:
			car.ColorVariant = rand.IntN(len(SedanColors))
			car.Color = SedanColors[car.ColorVariant]
		case CarTypeTruck:
			car.ColorVariant = rand.IntN(len(TruckColors))
			car.Color = TruckColors[car.ColorVariant]
		case CarTypeSport:
			car.ColorVariant = rand.IntN(len(SportColors))
			car.Color = SportColors[car.ColorVariant]
			car.LaneChangeTimer = 60 + rand.IntN(61)
		case CarTypePolice:
			car.Color = color.RGBA{0x22, 0x22, 0x44, 0xFF}
			car.ChaseTimer = 30
		case CarTypeOncoming:
			car.Color = color.RGBA{0x33, 0x33, 0x33, 0xFF}
		}

		return car
	}
	return nil
}

func chooseCarType(tickCount int) (CarType, CarDef) {
	switch {
	case tickCount < 30*TPS:
		return CarTypeSedan, SedanDef
	case tickCount < 60*TPS:
		if rand.IntN(100) < 30 {
			return CarTypeTruck, TruckDef
		}
		return CarTypeSedan, SedanDef
	case tickCount < 90*TPS:
		r := rand.IntN(100)
		switch {
		case r < 25:
			return CarTypeTruck, TruckDef
		case r < 50:
			return CarTypeSport, SportDef
		default:
			return CarTypeSedan, SedanDef
		}
	case tickCount < 120*TPS:
		r := rand.IntN(100)
		switch {
		case r < 15:
			return CarTypePolice, PoliceDef
		case r < 35:
			return CarTypeTruck, TruckDef
		case r < 55:
			return CarTypeSport, SportDef
		default:
			return CarTypeSedan, SedanDef
		}
	default: // 120s+
		r := rand.IntN(100)
		switch {
		case r < 10:
			return CarTypeOncoming, OncomingDef
		case r < 25:
			return CarTypePolice, PoliceDef
		case r < 40:
			return CarTypeTruck, TruckDef
		case r < 60:
			return CarTypeSport, SportDef
		default:
			return CarTypeSedan, SedanDef
		}
	}
}

func isLaneBlocked(cars []*TrafficCar, cx, threshold float64) bool {
	for _, c := range cars {
		if math.Abs(c.X-cx) < LaneWidth/2 && c.Y < threshold {
			return true
		}
	}
	return false
}

// canSpawnSafely checks that spawning won't create an impassable wall.
func canSpawnSafely(cars []*TrafficCar, cx, carHeight float64) bool {
	// Proximity check.
	if isLaneBlocked(cars, cx, carHeight*2) {
		return false
	}
	// Wall check: ensure no Y-line would have all lanes occupied.
	for _, c := range cars {
		if c.Y > float64(ScreenHeight)/2 || c.Y < -carHeight {
			continue
		}
		if lanesOccupiedAtY(cars, c.Y, carHeight) >= LaneCount-1 {
			return false
		}
	}
	return true
}

func lanesOccupiedAtY(cars []*TrafficCar, y, margin float64) int {
	var lanes [LaneCount]bool
	for _, c := range cars {
		if math.Abs(c.Y-y) < margin {
			lane := int((c.X - float64(RoadLeft)) / LaneWidth)
			if lane >= 0 && lane < LaneCount {
				lanes[lane] = true
			}
		}
	}
	count := 0
	for _, b := range lanes {
		if b {
			count++
		}
	}
	return count
}

// breakWalls nudges NPC cars apart when they form impassable lines.
// Applies a one-time speed reduction (flagged to prevent repeated decay).
func breakWalls(cars []*TrafficCar) {
	for _, a := range cars {
		neighbors := 0
		for _, b := range cars {
			if a == b {
				continue
			}
			if math.Abs(a.Y-b.Y) < PlayerHeight {
				neighbors++
			}
		}
		if neighbors >= LaneCount-1 && a.Speed > 0.5 {
			a.Speed = 0.5 // set to low fixed speed instead of repeated multiply
			break
		}
	}
}

// UpdateTraffic moves all NPC cars and removes off-screen ones.
// Returns updated slice and overtake score.
func UpdateTraffic(cars []*TrafficCar, scrollSpeed float64, playerX float64) ([]*TrafficCar, int) {
	breakWalls(cars)

	overtakeScore := 0
	n := 0
	for _, c := range cars {
		c.TickAge++

		switch c.CarType {
		case CarTypeOncoming:
			// Oncoming cars move DOWN the screen fast (towards the player).
			c.Y += scrollSpeed + c.Speed
		default:
			c.Y += scrollSpeed - c.Speed
		}

		switch c.CarType {
		case CarTypeSport:
			updateSportLaneChange(c)
		case CarTypePolice:
			updatePoliceChase(c, playerX)
		}

		// Remove off-screen cars.
		if c.Y-c.Height/2 > ScreenHeight || c.Y+c.Height/2 < -200 {
			if c.CarType != CarTypeOncoming { // don't score for oncoming
				overtakeScore += 10
			}
			continue
		}
		cars[n] = c
		n++
	}
	for i := n; i < len(cars); i++ {
		cars[i] = nil
	}
	return cars[:n], overtakeScore
}

func updatePoliceChase(c *TrafficCar, playerX float64) {
	c.ChaseTimer--
	if c.ChaseTimer <= 0 {
		c.ChaseTimer = 30
		// Steer towards player's X.
		if c.X < playerX-2 {
			c.X += 1.5
		} else if c.X > playerX+2 {
			c.X -= 1.5
		}
		// Clamp to road.
		halfW := c.Width / 2
		if c.X-halfW < RoadLeft {
			c.X = float64(RoadLeft) + halfW
		}
		if c.X+halfW > RoadRight {
			c.X = float64(RoadRight) - halfW
		}
	}
}

func updateSportLaneChange(c *TrafficCar) {
	if c.ChangingLane {
		diff := c.TargetX - c.X
		if math.Abs(diff) < 1.5 {
			c.X = c.TargetX
			c.ChangingLane = false
			c.LaneChangeTimer = 60 + rand.IntN(61)
		} else if diff > 0 {
			c.X += 1.2
		} else {
			c.X -= 1.2
		}
		return
	}

	c.LaneChangeTimer--
	if c.LaneChangeTimer <= 0 {
		currentLane := int((c.X - float64(RoadLeft)) / LaneWidth)
		targetLane := currentLane
		if currentLane <= 0 {
			targetLane = 1
		} else if currentLane >= LaneCount-1 {
			targetLane = LaneCount - 2
		} else if rand.IntN(2) == 0 {
			targetLane = currentLane - 1
		} else {
			targetLane = currentLane + 1
		}
		c.TargetX = float64(RoadLeft) + float64(targetLane)*LaneWidth + LaneWidth/2
		c.ChangingLane = true
	}
}

func (c *TrafficCar) Draw(screen *ebiten.Image, sprites *SpriteCache) {
	switch c.CarType {
	case CarTypeSedan:
		drawSprite(screen, sprites.TrafficSedan[c.ColorVariant], c.X, c.Y)

	case CarTypeTruck:
		drawSprite(screen, sprites.TrafficTruck[c.ColorVariant], c.X, c.Y)

	case CarTypeSport:
		drawSprite(screen, sprites.TrafficSport[c.ColorVariant], c.X, c.Y)
		// Blinking turn signal when changing lane.
		if c.ChangingLane && c.TickAge%10 < 5 {
			sigX := c.X - c.Width/2 - 1
			if c.TargetX > c.X {
				sigX = c.X + c.Width/2 - 2
			}
			drawSprite(screen, sprites.TurnSignal, sigX, c.Y)
		}

	case CarTypePolice:
		drawSprite(screen, sprites.TrafficPolice, c.X, c.Y)
		// Animated siren on top.
		sirenFrame := (c.TickAge / 8) % 2
		drawSprite(screen, sprites.PoliceSiren[sirenFrame], c.X, c.Y-c.Height/2+5)

	case CarTypeOncoming:
		// Headlight cone under the car.
		alpha := float32(0.6)
		if c.TickAge%30 < 5 {
			alpha = 0.3
		}
		drawSpriteAlpha(screen, sprites.HeadlightCone, c.X, c.Y+c.Height/2+18, alpha)
		drawSprite(screen, sprites.TrafficOncoming, c.X, c.Y)
	}
}

func (c *TrafficCar) Bounds() Rect {
	return NewRect(c.X, c.Y, c.Width*0.8, c.Height*0.9)
}
