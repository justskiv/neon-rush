package main

import "image/color"

const (
	ZoneDurationTicks   = 1800 // 30 seconds at 60 TPS
	ZoneTransitionTicks = 180  // 3 seconds transition
	ZoneCount           = 5
)

// ZoneID identifies a zone/biome.
type ZoneID int

const (
	ZoneNightCity   ZoneID = iota
	ZoneHighway
	ZoneRain
	ZoneNeonTunnel
	ZoneSunsetChaos
)

// ZonePalette holds colors defining a zone's visual identity.
type ZonePalette struct {
	Background  color.RGBA
	Asphalt     color.RGBA
	DashLine    color.RGBA
	CurbPrimary color.RGBA
	NeonAccent  color.RGBA
	NeonAccent2 color.RGBA // only used by Neon Tunnel
}

var zonePalettes = [ZoneCount]ZonePalette{
	{ // Night City
		Background:  color.RGBA{0x0D, 0x0D, 0x1A, 0xFF},
		Asphalt:     color.RGBA{0x1A, 0x1A, 0x2E, 0xFF},
		DashLine:    color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
		CurbPrimary: color.RGBA{0xFF, 0xCC, 0x00, 0xFF},
		NeonAccent:  color.RGBA{0x00, 0xAA, 0xFF, 0xFF},
	},
	{ // Highway
		Background:  color.RGBA{0x1A, 0x1A, 0x0D, 0xFF},
		Asphalt:     color.RGBA{0x2E, 0x2E, 0x1A, 0xFF},
		DashLine:    color.RGBA{0xFF, 0xCC, 0x00, 0xFF},
		CurbPrimary: color.RGBA{0xFF, 0xAA, 0x00, 0xFF},
		NeonAccent:  color.RGBA{0xFF, 0xAA, 0x00, 0xFF},
	},
	{ // Rain
		Background:  color.RGBA{0x0D, 0x1A, 0x1A, 0xFF},
		Asphalt:     color.RGBA{0x1A, 0x2E, 0x2E, 0xFF},
		DashLine:    color.RGBA{0x99, 0x99, 0x99, 0xFF},
		CurbPrimary: color.RGBA{0x66, 0x66, 0x66, 0xFF},
		NeonAccent:  color.RGBA{0x44, 0xDD, 0xFF, 0xFF},
	},
	{ // Neon Tunnel
		Background:  color.RGBA{0x1A, 0x0D, 0x2E, 0xFF},
		Asphalt:     color.RGBA{0x2E, 0x1A, 0x44, 0xFF},
		DashLine:    color.RGBA{0xFF, 0x00, 0xFF, 0xFF},
		CurbPrimary: color.RGBA{0xFF, 0x00, 0xFF, 0xFF},
		NeonAccent:  color.RGBA{0xFF, 0x00, 0xFF, 0xFF},
		NeonAccent2: color.RGBA{0x00, 0xFF, 0x66, 0xFF},
	},
	{ // Sunset Chaos
		Background:  color.RGBA{0x2E, 0x0D, 0x0D, 0xFF},
		Asphalt:     color.RGBA{0x44, 0x1A, 0x1A, 0xFF},
		DashLine:    color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
		CurbPrimary: color.RGBA{0xFF, 0x44, 0x44, 0xFF},
		NeonAccent:  color.RGBA{0xFF, 0x44, 0x44, 0xFF},
	},
}

var zoneNames = [ZoneCount]string{
	"NIGHT CITY", "HIGHWAY", "RAIN", "NEON TUNNEL", "SUNSET CHAOS",
}

// ZoneSystem manages zone transitions and palette interpolation.
type ZoneSystem struct {
	CurrentZone     ZoneID
	ZoneTick        int
	Transitioning   bool
	TransitionT     float64
	ActivePalette   ZonePalette
	ZonesReached    int  // total zone transitions survived
	JustTransitioned bool // set for one tick on zone change
}

func NewZoneSystem() ZoneSystem {
	return ZoneSystem{
		ActivePalette: zonePalettes[0],
		ZonesReached:  1, // starting zone counts
	}
}

func (zs *ZoneSystem) Update() {
	zs.ZoneTick++
	zs.JustTransitioned = false

	transStart := ZoneDurationTicks - ZoneTransitionTicks
	if zs.ZoneTick >= transStart && zs.ZoneTick < ZoneDurationTicks {
		zs.Transitioning = true
		zs.TransitionT = float64(zs.ZoneTick-transStart) / float64(ZoneTransitionTicks)
		next := zs.nextZone()
		zs.ActivePalette = lerpPalette(zonePalettes[zs.CurrentZone], zonePalettes[next], zs.TransitionT)
	} else {
		zs.Transitioning = false
		zs.ActivePalette = zonePalettes[zs.CurrentZone]
	}

	if zs.ZoneTick >= ZoneDurationTicks {
		zs.CurrentZone = zs.nextZone()
		zs.ZoneTick = 0
		zs.ZonesReached++
		zs.JustTransitioned = true
	}
}

func (zs *ZoneSystem) nextZone() ZoneID {
	return ZoneID((int(zs.CurrentZone) + 1) % ZoneCount)
}

func (zs *ZoneSystem) ZoneName() string {
	return zoneNames[zs.CurrentZone]
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: uint8(float64(a.A) + (float64(b.A)-float64(a.A))*t),
	}
}

func lerpPalette(a, b ZonePalette, t float64) ZonePalette {
	return ZonePalette{
		Background:  lerpColor(a.Background, b.Background, t),
		Asphalt:     lerpColor(a.Asphalt, b.Asphalt, t),
		DashLine:    lerpColor(a.DashLine, b.DashLine, t),
		CurbPrimary: lerpColor(a.CurbPrimary, b.CurbPrimary, t),
		NeonAccent:  lerpColor(a.NeonAccent, b.NeonAccent, t),
		NeonAccent2: lerpColor(a.NeonAccent2, b.NeonAccent2, t),
	}
}
