package main

const (
	// Screen dimensions (logical).
	ScreenWidth  = 400
	ScreenHeight = 600

	// Road geometry.
	RoadLeft  = 60
	RoadRight = 340
	LaneCount = 4
	LaneWidth = 70 // (RoadRight - RoadLeft) / LaneCount

	// Player car.
	PlayerWidth  = 30
	PlayerHeight = 50
	PlayerStartX = 200 // center of screen
	PlayerStartY = 480 // near bottom

	// Scroll speed (px/tick).
	BaseScrollSpeed = 3.0
	MaxScrollSpeed  = 12.0
	SpeedIncrement  = 0.002

	// Player lateral movement.
	PlayerMoveSpeed    = 4.0
	PlayerAcceleration = 0.5
	PlayerFriction     = 0.85

	// Traffic spawning (ticks).
	TrafficSpawnRate = 60
	MinSpawnRate     = 25

	// Ticks per second.
	TPS = 60

	// Near-miss and combo.
	NearMissThreshold  = 10.0
	NearMissBonus      = 50
	ComboMultiplierMax = 10
	ComboDecayTicks    = 120

	// Fuel.
	FuelMax         = 100.0
	FuelConsumption = 0.02
	FuelCanBonus    = 30.0

	// Nitro.
	NitroChargeMax = 3
	NitroDuration  = 180 // 3 seconds at 60 TPS

	// Speed modifiers.
	SpeedBoostFactor = 1.3 // Up key: +30%
	SpeedBrakeFactor = 0.5 // Down key: -50%

	// Default sprite scale (overridden dynamically in NewSpriteCache).
	DefaultSpriteScale = 4.0

	// Sprite sizes at logical resolution.
	PlayerSpriteW = 36
	PlayerSpriteH = 58
	PlayerGlowPad = 4

	// Decor parallax.
	DecorParallaxFactor = 0.6
	NumBuildingVariants = 12

	// Freeze frame durations (ticks).
	FreezeFrameNearMiss  = 2
	FreezeFrameCollision = 6

	// Last Chance.
	LastChanceDuration  = 90  // 1.5 seconds
	LastChanceSpeedMult = 0.2 // 20% speed during slow-mo
)
