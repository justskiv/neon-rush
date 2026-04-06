package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	dashLength    = 20.0
	dashGap       = 30.0
	dashPeriod    = dashLength + dashGap // 50px
	curbWidth     = 4.0
	curbSegment   = 20.0
	curbPeriod    = curbSegment * 2 // 40px
	dashLineWidth = 4.0
)

// Road handles the scrolling road surface, lane markings, and curbs.
type Road struct {
	scrollOffset float64
}

func NewRoad() Road {
	return Road{}
}

func (r *Road) Update(scrollSpeed float64) {
	r.scrollOffset += scrollSpeed
	r.scrollOffset = math.Mod(r.scrollOffset, dashPeriod*curbPeriod)
}

// Draw renders the road using the given zone palette.
func (r *Road) Draw(screen *ebiten.Image, palette ZonePalette, zoneID ZoneID) {
	screen.Fill(palette.Background)

	DrawRect(screen, RoadLeft, 0, RoadRight-RoadLeft, ScreenHeight, palette.Asphalt)

	r.drawCurb(screen, float64(RoadLeft)-curbWidth, palette.CurbPrimary)
	r.drawCurb(screen, float64(RoadRight), palette.CurbPrimary)

	// Lane dashes (3 lines between 4 lanes).
	dashOffset := math.Mod(r.scrollOffset, dashPeriod)
	dashIdx := 0
	for i := 1; i < LaneCount; i++ {
		x := float64(RoadLeft) + float64(i)*LaneWidth - dashLineWidth/2
		for y := -dashLength + dashOffset; y < ScreenHeight+dashLength; y += dashPeriod {
			clr := palette.DashLine
			// Sunset Chaos: alternate red/white dashes.
			if zoneID == ZoneSunsetChaos && dashIdx%2 == 0 {
				clr = palette.NeonAccent
			}
			DrawRect(screen, x, y, dashLineWidth, dashLength, clr)
			dashIdx++
		}
	}
}

func (r *Road) drawCurb(screen *ebiten.Image, x float64, primary color.RGBA) {
	offset := math.Mod(r.scrollOffset, curbPeriod)
	for y := -curbSegment + offset; y < ScreenHeight+curbSegment; y += curbSegment {
		seg := int(math.Floor((y - offset) / curbSegment))
		clr := primary
		if seg%2 != 0 {
			clr = color.RGBA{0x00, 0x00, 0x00, 0xFF}
		}
		DrawRect(screen, x, y, curbWidth, curbSegment, clr)
	}
}
