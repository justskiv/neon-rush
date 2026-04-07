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

// Color palettes for NPC sprites.
var SedanColors = [4]color.RGBA{
	{0x88, 0x88, 0xAA, 0xFF},
	{0xAA, 0xAA, 0xCC, 0xFF},
	{0x88, 0x77, 0x66, 0xFF},
	{0x66, 0x66, 0x88, 0xFF},
}

var SportColors = [3]color.RGBA{
	{0xFF, 0x33, 0x33, 0xFF}, // red
	{0xFF, 0xCC, 0x00, 0xFF}, // yellow
	{0xFF, 0x66, 0x00, 0xFF}, // orange
}

var TruckColors = [3]color.RGBA{
	{0x33, 0x44, 0x44, 0xFF},
	{0x44, 0x33, 0x44, 0xFF},
	{0x44, 0x44, 0x33, 0xFF},
}

func sedanColor() color.RGBA {
	return SedanColors[rand.IntN(len(SedanColors))]
}

func sportColor() color.RGBA {
	return SportColors[rand.IntN(len(SportColors))]
}

func truckColor() color.RGBA {
	return TruckColors[rand.IntN(len(TruckColors))]
}
