package main

import (
	"image/color"
	"math/rand/v2"
)

// CarType identifies the NPC vehicle type.
type CarType int

const (
	CarTypeSedan CarType = iota
	CarTypeTruck
	CarTypeSport
	CarTypePolice
	CarTypeOncoming
)

// CarDef defines visual and physical properties of a car type.
type CarDef struct {
	Width      float64
	Height     float64
	Color      color.RGBA
	SpeedRatio float64 // fraction of scrollSpeed
}

var (
	SedanDef    = CarDef{28, 45, color.RGBA{0xAA, 0xAA, 0xAA, 0xFF}, 0.7}
	TruckDef    = CarDef{34, 65, color.RGBA{0x44, 0x44, 0x55, 0xFF}, 0.6}
	SportDef    = CarDef{26, 42, color.RGBA{0xFF, 0x33, 0x33, 0xFF}, 0.75}
	PoliceDef   = CarDef{28, 48, color.RGBA{0x22, 0x22, 0x44, 0xFF}, 0.75}
	OncomingDef = CarDef{26, 44, color.RGBA{0x33, 0x33, 0x33, 0xFF}, 2.0}
)

func sedanColor() color.RGBA {
	colors := []color.RGBA{
		{0xAA, 0xAA, 0xAA, 0xFF},
		{0xCC, 0xCC, 0xCC, 0xFF},
		{0xDD, 0xDD, 0xBB, 0xFF},
		{0xFF, 0xFF, 0xFF, 0xFF},
	}
	return colors[rand.IntN(len(colors))]
}

func sportColor() color.RGBA {
	colors := []color.RGBA{
		{0xFF, 0x33, 0x33, 0xFF}, // red
		{0xFF, 0xDD, 0x00, 0xFF}, // yellow
	}
	return colors[rand.IntN(len(colors))]
}

func truckColor() color.RGBA {
	colors := []color.RGBA{
		{0x44, 0x44, 0x55, 0xFF},
		{0x55, 0x44, 0x33, 0xFF},
		{0x33, 0x44, 0x44, 0xFF},
	}
	return colors[rand.IntN(len(colors))]
}
