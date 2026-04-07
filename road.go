package main

import (
	"image/color"
	"math"
	"math/rand/v2"

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

	// Curve generation: how far ahead to pre-generate segments.
	curveBufferAhead = 5000.0
)

// CurveSegment describes one curved section of road.
type CurveSegment struct {
	StartDist float64
	Length    float64
	Amplitude float64 // positive = right, negative = left
}

// RoadCurve manages procedurally generated road turns.
type RoadCurve struct {
	segments []CurveSegment
	distance float64
	segIdx   int // index of current or next segment for O(1) lookup
	rng      *rand.Rand
	lastTick int // tickCount at last generation call
}

// NewRoadCurve creates a curve system with a random seed.
func NewRoadCurve() *RoadCurve {
	return &RoadCurve{
		rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

// Update advances the distance and generates new segments as needed.
func (rc *RoadCurve) Update(speed float64, tickCount int) {
	rc.distance += speed
	rc.lastTick = tickCount

	// Advance segment index past completed segments.
	for rc.segIdx < len(rc.segments) {
		seg := &rc.segments[rc.segIdx]
		if rc.distance < seg.StartDist+seg.Length {
			break
		}
		rc.segIdx++
	}

	// Prune old segments (keep 2 behind for OffsetAt lookback).
	if rc.segIdx > 2 {
		copy(rc.segments, rc.segments[rc.segIdx-2:])
		rc.segments = rc.segments[:len(rc.segments)-(rc.segIdx-2)]
		rc.segIdx = 2
	}

	rc.ensureSegmentsAhead(tickCount)
}

// CurveDirection returns the sign of the active curve's amplitude.
// Positive = right, negative = left, 0 = straight.
func (rc *RoadCurve) CurveDirection() float64 {
	for i := rc.segIdx; i < len(rc.segments); i++ {
		seg := &rc.segments[i]
		if rc.distance >= seg.StartDist && rc.distance < seg.StartDist+seg.Length {
			if seg.Amplitude > 0 {
				return 1
			}
			return -1
		}
	}
	return 0
}

// ScreenOffset returns the horizontal offset for a given screen Y position.
// At PlayerStartY (bottom, where the player is): offset = current curve displacement.
// At y=0 (top of screen): offset = future curve displacement (shows upcoming bend).
// speedFactor amplifies the offset at higher speeds: [1.0, 2.0].
func (rc *RoadCurve) ScreenOffset(y, speed float64) float64 {
	lookAhead := (PlayerStartY - y) / PlayerStartY * CurveViewDistance
	if lookAhead < 0 {
		lookAhead = 0
	}
	raw := rc.offsetAt(rc.distance + lookAhead)
	speedFactor := 1.0 + (speed-MinSpeed)/(MaxScrollSpeed-MinSpeed)
	return raw * speedFactor
}

func (rc *RoadCurve) offsetAt(dist float64) float64 {
	for i := range rc.segments {
		seg := &rc.segments[i]
		if dist < seg.StartDist {
			return 0 // in a straight gap before this segment
		}
		if dist < seg.StartDist+seg.Length {
			t := (dist - seg.StartDist) / seg.Length
			return seg.Amplitude * math.Sin(t*math.Pi)
		}
	}
	return 0 // past all segments (straight)
}

// ensureSegmentsAhead generates segments so there's always road ahead.
func (rc *RoadCurve) ensureSegmentsAhead(tickCount int) {
	// Determine how far ahead the last segment reaches.
	var lastEnd float64
	if len(rc.segments) > 0 {
		last := &rc.segments[len(rc.segments)-1]
		lastEnd = last.StartDist + last.Length
	}

	for lastEnd < rc.distance+curveBufferAhead {
		gap, seg := rc.generateNext(lastEnd, tickCount)
		lastEnd += gap
		if seg.Length > 0 {
			seg.StartDist = lastEnd
			rc.segments = append(rc.segments, seg)
			lastEnd += seg.Length
		}
	}
}

// generateNext returns (gapLength, segment) based on time-driven difficulty.
func (rc *RoadCurve) generateNext(afterDist float64, tickCount int) (float64, CurveSegment) {
	seconds := float64(tickCount) / float64(TPS)

	// Difficulty tiers.
	var (
		ampMin, ampMax     float64
		lenMin, lenMax     float64
		gapMin, gapMax     float64
		sCurveChance       float64
	)

	switch {
	case seconds < 30:
		// No curves — straight road while player learns gas/brake.
		return 800 + rc.rng.Float64()*400, CurveSegment{}
	case seconds < 60:
		ampMin, ampMax = 8, 14
		lenMin, lenMax = 1200, 1800
		gapMin, gapMax = 1200, 1800
	case seconds < 120:
		ampMin, ampMax = 14, 22
		lenMin, lenMax = 1000, 1500
		gapMin, gapMax = 800, 1200
		sCurveChance = 0.2
	case seconds < 180:
		ampMin, ampMax = 20, 30
		lenMin, lenMax = 800, 1200
		gapMin, gapMax = 500, 800
		sCurveChance = 0.3
	default:
		ampMin, ampMax = 25, 36
		lenMin, lenMax = 600, 1000
		gapMin, gapMax = 400, 600
		sCurveChance = 0.4
	}

	gap := gapMin + rc.rng.Float64()*(gapMax-gapMin)
	length := lenMin + rc.rng.Float64()*(lenMax-lenMin)
	amp := ampMin + rc.rng.Float64()*(ampMax-ampMin)

	// Random direction.
	if rc.rng.Float64() < 0.5 {
		amp = -amp
	}

	// S-curve: if previous segment exists, go opposite direction.
	if sCurveChance > 0 && rc.rng.Float64() < sCurveChance && len(rc.segments) > 0 {
		prev := &rc.segments[len(rc.segments)-1]
		if prev.Amplitude > 0 {
			amp = -math.Abs(amp)
		} else {
			amp = math.Abs(amp)
		}
		gap = 0 // S-curves: back-to-back, sin(Pi)=0 at junction
	}

	return gap, CurveSegment{Length: length, Amplitude: amp}
}

// Road handles the scrolling road surface, lane markings, and curbs.
type Road struct {
	scrollOffset float64
	Curve        *RoadCurve // nil for menu background
}

func NewRoad() Road {
	return Road{}
}

// NewRoadWithCurve creates a road with procedural turns for gameplay.
func NewRoadWithCurve() Road {
	return Road{Curve: NewRoadCurve()}
}

func (r *Road) Update(scrollSpeed float64, tickCount int) {
	r.scrollOffset += scrollSpeed
	r.scrollOffset = math.Mod(r.scrollOffset, dashPeriod*curbPeriod)
	if r.Curve != nil {
		r.Curve.Update(scrollSpeed, tickCount)
	}
}

// Draw renders the road using the given zone palette.
// offsetFn returns the horizontal offset for a given screen Y position.
func (r *Road) Draw(screen *ebiten.Image, palette ZonePalette, zoneID ZoneID, offsetFn func(float64) float64) {
	screen.Fill(palette.Background)

	roadW := float64(RoadRight - RoadLeft)

	// Asphalt: draw as horizontal strips for per-row curve offset.
	const numStrips = 24
	stripH := float64(ScreenHeight) / numStrips
	for s := range numStrips {
		sy := float64(s) * stripH
		off := offsetFn(sy + stripH/2)
		DrawRect(screen, float64(RoadLeft)+off, sy, roadW, stripH+1, palette.Asphalt)
	}

	// Curbs: each segment gets its own offset.
	scrollOff := math.Mod(r.scrollOffset, curbPeriod)
	for y := -curbSegment + scrollOff; y < ScreenHeight+curbSegment; y += curbSegment {
		seg := int(math.Floor((y - scrollOff) / curbSegment))
		clr := palette.CurbPrimary
		if seg%2 != 0 {
			clr = color.RGBA{0x00, 0x00, 0x00, 0xFF}
		}
		off := offsetFn(y + curbSegment/2)
		DrawRect(screen, float64(RoadLeft)-curbWidth+off, y, curbWidth, curbSegment, clr)
		DrawRect(screen, float64(RoadRight)+off, y, curbWidth, curbSegment, clr)
	}

	// Lane dashes: each dash gets offset at its Y position.
	dashOffset := math.Mod(r.scrollOffset, dashPeriod)
	dashIdx := 0
	for i := 1; i < LaneCount; i++ {
		baseX := float64(RoadLeft) + float64(i)*LaneWidth - dashLineWidth/2
		for y := -dashLength + dashOffset; y < ScreenHeight+dashLength; y += dashPeriod {
			off := offsetFn(y + dashLength/2)
			clr := palette.DashLine
			if zoneID == ZoneSunsetChaos && dashIdx%2 == 0 {
				clr = palette.NeonAccent
			}
			DrawRect(screen, baseX+off, y, dashLineWidth, dashLength, clr)
			dashIdx++
		}
	}
}
