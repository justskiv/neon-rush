package main

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- Screen Shake ---

// ScreenShake provides sinusoidal screen shake with exponential decay.
type ScreenShake struct {
	intensity float64
	decay     float64
	frequency float64
	time      float64
}

func ShakeCollision() ScreenShake  { return ScreenShake{8, 0.88, 1.5, 0} }
func ShakeNearMiss() ScreenShake   { return ScreenShake{2, 0.75, 3.0, 0} }
func ShakeNitroStart() ScreenShake { return ScreenShake{4, 0.92, 0.8, 0} }

// Update advances the shake and returns the current offset.
func (s *ScreenShake) Update() (float64, float64) {
	if s.intensity < 0.1 {
		return 0, 0
	}
	s.time++
	s.intensity *= s.decay
	ox := s.intensity * math.Sin(s.time*s.frequency)
	oy := s.intensity * math.Cos(s.time*s.frequency*0.7)
	return ox, oy
}

// --- Chromatic Aberration ---

// ChromaticAberration splits the scene into RGB channels with horizontal offset.
// Applied at the final blit stage — not as an overlay.
type ChromaticAberration struct {
	offset float64 // current pixel offset (logical)
	decay  float64
}

var chromaticOffsets = [5]float64{0, 4, 6, 8, 10}

func NewChromaticAberration(tier NearMissTier) ChromaticAberration {
	if tier <= TierNone || int(tier) >= len(chromaticOffsets) {
		return ChromaticAberration{}
	}
	return ChromaticAberration{offset: chromaticOffsets[tier], decay: 0.82}
}

func (ca *ChromaticAberration) Update() {
	if ca.offset < 0.3 {
		ca.offset = 0
		return
	}
	ca.offset *= ca.decay
}

func (ca *ChromaticAberration) Offset() float64 {
	return ca.offset
}

// --- Speed Lines ---

const speedLinePoolSize = 30

// SpeedLine is a single speed-indicator line at screen edges.
type SpeedLine struct {
	Active bool
	X, Y   float64
	Length float64
	Speed  float64
	Alpha  uint8
}

// SpeedLineSystem renders vertical speed lines at screen edges.
type SpeedLineSystem struct {
	Pool [speedLinePoolSize]SpeedLine
}

func (sl *SpeedLineSystem) spawn() *SpeedLine {
	for i := range sl.Pool {
		if !sl.Pool[i].Active {
			return &sl.Pool[i]
		}
	}
	return nil
}

func (sl *SpeedLineSystem) Update(scrollSpeed, maxSpeed float64, nitroActive bool) {
	speedRatio := (scrollSpeed - maxSpeed*0.7) / (maxSpeed * 0.3)
	if speedRatio < 0 && !nitroActive {
		for i := range sl.Pool {
			sl.Pool[i].Active = false
		}
		return
	}

	for i := range sl.Pool {
		l := &sl.Pool[i]
		if !l.Active {
			continue
		}
		l.Y += l.Speed
		if l.Y > ScreenHeight+l.Length {
			l.Active = false
		}
	}

	spawnCount := 2
	if nitroActive {
		spawnCount = 4
	}
	for range spawnCount {
		l := sl.spawn()
		if l == nil {
			break
		}
		if nitroActive {
			l.X = rand.Float64() * ScreenWidth
			l.Length = 50 + rand.Float64()*30
			l.Alpha = 150
		} else {
			if rand.IntN(2) == 0 {
				l.X = rand.Float64() * 80
			} else {
				l.X = 320 + rand.Float64()*80
			}
			l.Length = 15 + rand.Float64()*25
			if speedRatio > 1 {
				speedRatio = 1
			}
			l.Alpha = uint8(speedRatio * 100)
		}
		l.Y = -l.Length
		l.Speed = scrollSpeed * 3
		l.Active = true
	}
}

func (sl *SpeedLineSystem) Draw(screen *ebiten.Image) {
	for i := range sl.Pool {
		l := &sl.Pool[i]
		if !l.Active {
			continue
		}
		DrawRect(screen, l.X, l.Y, 1, l.Length, color.RGBA{0xFF, 0xFF, 0xFF, l.Alpha})
	}
}

// --- Vignette ---

var vignetteImage *ebiten.Image

// InitVignette pre-renders the vignette overlay at logical resolution.
func InitVignette() {
	vignetteImage = ebiten.NewImage(ScreenWidth, ScreenHeight)
	depth := 60
	for d := range depth {
		alpha := uint8(float64(depth-d) / float64(depth) * 80)
		clr := color.RGBA{0, 0, 0, alpha}
		fd := float64(d)
		fw := float64(ScreenWidth) - 2*fd
		fh := float64(ScreenHeight) - 2*fd
		if fw <= 0 || fh <= 0 {
			break
		}
		drawRectRaw(vignetteImage, fd, fd, fw, 1, clr)                  // top
		drawRectRaw(vignetteImage, fd, float64(ScreenHeight)-1-fd, fw, 1, clr) // bottom
		drawRectRaw(vignetteImage, fd, fd, 1, fh, clr)                  // left
		drawRectRaw(vignetteImage, float64(ScreenWidth)-1-fd, fd, 1, fh, clr)  // right
	}
}

// DrawVignette renders the vignette with intensity based on speed.
func DrawVignette(screen *ebiten.Image, scrollSpeed float64) {
	if vignetteImage == nil {
		return
	}
	speedNorm := (scrollSpeed - BaseScrollSpeed) / (MaxScrollSpeed - BaseScrollSpeed)
	if speedNorm < 0 {
		speedNorm = 0
	}
	if speedNorm > 1 {
		speedNorm = 1
	}
	alpha := float32(0.3 + speedNorm*0.7)
	rs := renderScaleGlobal
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(rs, rs)
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(vignetteImage, op)
}
