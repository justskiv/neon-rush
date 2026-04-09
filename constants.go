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

	// Speed (px/tick).
	BaseScrollSpeed = 3.0  // base NPC speed reference
	MaxScrollSpeed  = 13.0 // max possible (FURY)

	// Player speed control.
	MinSpeed     = 2.0
	Deceleration = 0.03

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


	// Default sprite scale (overridden dynamically in NewSpriteCache).
	DefaultSpriteScale = 4.0

	// Sprite sizes at logical resolution.
	PlayerSpriteW = 36
	PlayerSpriteH = 58
	PlayerGlowPad = 4

	// Road curve rendering.
	CurveViewDistance = 400.0 // how far ahead top of screen "sees"

	// Decor parallax.
	DecorParallaxFactor = 0.6
	NumBuildingVariants = 12

	// Freeze frame durations (ticks).
	FreezeFrameNearMiss  = 2
	FreezeFrameCollision = 6

	// Last Chance.
	LastChanceDuration  = 90  // 1.5 seconds
	LastChanceSpeedMult = 0.2 // 20% speed during slow-mo

	// Drift.
	DriftMinSpeed       = 5.0
	DriftLateralMult    = 2.0
	DriftInertiaMax     = 18
	DriftSpeedDrag      = 0.997
	DriftOverheatMax    = 180  // 3s to overheat
	DriftOverheatForced = 30   // loss of control ticks
	DriftCooldownMult   = 2.0
	DriftRotationMax    = 0.28 // ~16 degrees
	DriftScorePerTick   = 8
	DriftNearMissMult   = 3.0
	DriftDurationTier2  = 60  // 1s → ×2
	DriftDurationTier3  = 120 // 2s → ×3
	DriftDangerMult     = 5   // shoulder drift multiplier
)
