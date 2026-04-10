package main

import (
	"image/color"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
)

// GameState represents the current state of the game.
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateGarage
	StateSettings
)

// Game implements ebiten.Game interface.
type Game struct {
	state GameState

	player Player
	road   Road

	traffic       []*TrafficCar
	spawnTimer    int
	spawnInterval int

	items          []*Item
	fuelSpawnTimer int

	score       int
	speed       float64
	tickCount   int
	fuel        float64

	nitroCharges    int
	nitroActive     bool
	nitroTimer      int
	nitroGrace      int // invulnerability ticks after nitro ends
	nitroSpawnTimer int

	zone       ZoneSystem
	decor      DecorSystem
	scoreState ScoreState
	particles  ParticleSystem
	speedLines   SpeedLineSystem
	braking      bool
	accelerating bool
	offsetFn     func(float64) float64 // per-row curve offset

	audio          *AudioSystem
	sprites        *SpriteCache
	offscreen      *ebiten.Image
	renderScale    float64 // ratio of render buffer to logical size
	shake          ScreenShake
	chromatic      ChromaticAberration
	freezeTimer    int
	pauseSelection int
	menu           MenuState

	save              SaveData
	peakCombo         int
	nearMissCount     int
	activeCar         PlayerCarDef
	nearMissThreshold float64
	fuelConsumption   float64
	ghostShieldActive  bool
	garageSelection    int
	garageSection      int // 0=cars, 1=trails
	trailSelection     int
	oilSpawnTimer      int
	coinSpawnTimer     int
	coinLineCount      int
	coinLineCollected  int
	gameOverSelection  int
	newUnlocks         []string
	isNewHighScore     bool
	topSpeed           float64
	busted              bool
	fuelEmpty           bool
	lastChanceAvailable bool
	lastChanceActive    bool
	lastChanceTimer     int
	repairSpawnTimer    int
	dailyMode           bool
	dailyChallenge      DailyChallenge
	milestones          MilestoneSystem
	displaySpeed        float64
	scoreFlashTimer     int
	prevScore           int
	transition          Transition
	settingsSelection   int
	ghostRecording      GhostRecording
}

// NewGame creates and initializes a new game instance.
func NewGame() *Game {
	g := &Game{
		state:           StateMenu,
		menu:            NewMenuState(),
		road:            NewRoad(),
		player:          NewPlayer(),
		traffic:         make([]*TrafficCar, 0, 32),
		items:           make([]*Item, 0, 16),
		spawnInterval:   TrafficSpawnRate,
		speed:           MinSpeed,
		fuel:            FuelMax,
		fuelSpawnTimer:  randRange(5*TPS, 8*TPS),
		nitroSpawnTimer: randRange(30*TPS, 60*TPS),
		zone:            NewZoneSystem(),
		decor:           NewDecorSystem(),
		scoreState:      NewScoreState(),
		particles:       NewParticleSystem(),
		audio:           NewAudioSystem(),
		sprites:         NewSpriteCache(),
		renderScale:     1.0,
		save:            LoadSave(),
	}
	g.audio.Muted = g.save.SoundOff
	InitVignette()
	return g
}

func (g *Game) Update() error {
	if g.transition.Active() {
		g.transition.Update()
		return nil
	}
	if g.freezeTimer > 0 {
		g.freezeTimer--
		return nil
	}
	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StatePlaying:
		g.updatePlaying()
	case StatePaused:
		g.updatePaused()
	case StateGameOver:
		g.updateGameOver()
	case StateGarage:
		g.updateGarage()
	case StateSettings:
		g.updateSettings()
	}
	return nil
}

// --- Menu (stub for now) ---

func (g *Game) updateMenu() {
	g.menu.Update()

	if IsUpMenuPressed() {
		g.menu.Selection--
		if g.menu.Selection < 0 {
			g.menu.Selection = 3
		}
	}
	if IsDownMenuPressed() {
		g.menu.Selection++
		if g.menu.Selection > 3 {
			g.menu.Selection = 0
		}
	}
	if IsRestartPressed() {
		switch g.menu.Selection {
		case 0: // PLAY
			g.transition.Start(17, func() {
				g.dailyMode = false
				g.reset()
			})
		case 1: // DAILY
			if g.save.DailyDone != TodayDateStr() {
				g.transition.Start(17, func() {
					g.dailyMode = true
					g.dailyChallenge = TodayChallenge()
					g.reset()
				})
			}
		case 2: // GARAGE
			g.state = StateGarage
		case 3: // SETTINGS
			g.state = StateSettings
		}
	}
}

func (g *Game) drawMenu(dst *ebiten.Image) {
	DrawMenu(dst, &g.menu, &g.save)
}

// --- Pause ---

func (g *Game) updatePaused() {
	if IsUpMenuPressed() {
		g.pauseSelection--
		if g.pauseSelection < 0 {
			g.pauseSelection = 2
		}
	}
	if IsDownMenuPressed() {
		g.pauseSelection++
		if g.pauseSelection > 2 {
			g.pauseSelection = 0
		}
	}
	if IsRestartPressed() {
		switch g.pauseSelection {
		case 0: // RESUME
			g.state = StatePlaying
		case 1: // RESTART
			g.reset()
		case 2: // QUIT
			g.state = StateMenu
		}
	}
	if IsEscPressed() {
		g.state = StatePlaying
	}
}

func (g *Game) drawPaused(dst *ebiten.Image) {
	DrawRect(dst, 0, 0, ScreenWidth, ScreenHeight, colorOverlay)

	cx := ScreenWidth/2 - 40
	cy := ScreenHeight/2 - 40
	DrawText(dst, "P A U S E D", cx, cy)

	items := []string{"RESUME", "RESTART", "QUIT"}
	for i, item := range items {
		marker := "  "
		if i == g.pauseSelection {
			marker = "> "
		}
		DrawText(dst, marker+item, cx, cy+30+i*18)
	}
}

// --- Game Over ---

func (g *Game) updateGameOver() {
	g.particles.Update()

	if IsLeftMenuPressed() || IsRightMenuPressed() {
		g.gameOverSelection = 1 - g.gameOverSelection
	}
	if IsRestartPressed() {
		if g.gameOverSelection == 0 {
			g.transition.Start(25, func() { g.reset() })
		} else {
			g.transition.Start(17, func() { g.state = StateMenu })
		}
	}
}

// --- Garage (stub for now) ---

func (g *Game) updateGarage() {
	if IsUpMenuPressed() || IsDownMenuPressed() {
		g.garageSection = 1 - g.garageSection
	}
	if g.garageSection == 0 {
		// Car browsing.
		if IsLeftMenuPressed() {
			g.garageSelection--
			if g.garageSelection < 0 {
				g.garageSelection = len(PlayerCars) - 1
			}
		}
		if IsRightMenuPressed() {
			g.garageSelection++
			if g.garageSelection >= len(PlayerCars) {
				g.garageSelection = 0
			}
		}
		if IsRestartPressed() && g.save.IsCarUnlocked(g.garageSelection) {
			g.save.SelectedCar = g.garageSelection
			g.save.Save()
		}
	} else {
		// Trail browsing.
		if IsLeftMenuPressed() {
			g.trailSelection--
			if g.trailSelection < 0 {
				g.trailSelection = len(TrailDefs) - 1
			}
		}
		if IsRightMenuPressed() {
			g.trailSelection++
			if g.trailSelection >= len(TrailDefs) {
				g.trailSelection = 0
			}
		}
		if IsRestartPressed() && g.save.IsTrailUnlocked(g.trailSelection) {
			g.save.SelectedTrail = g.trailSelection
			g.save.Save()
		}
	}
	if IsEscPressed() {
		g.state = StateMenu
	}
}

func (g *Game) drawGarage(dst *ebiten.Image) {
	DrawGarage(dst, g.garageSelection, g.garageSection, g.trailSelection, &g.save, g.sprites)
}

// --- Settings ---

func (g *Game) updateSettings() {
	if IsUpMenuPressed() {
		g.settingsSelection--
		if g.settingsSelection < 0 {
			g.settingsSelection = 2
		}
	}
	if IsDownMenuPressed() {
		g.settingsSelection++
		if g.settingsSelection > 2 {
			g.settingsSelection = 0
		}
	}
	if IsLeftMenuPressed() || IsRightMenuPressed() || IsRestartPressed() {
		switch g.settingsSelection {
		case 0:
			g.save.SoundOff = !g.save.SoundOff
			g.audio.Muted = g.save.SoundOff
		case 1:
			g.save.ShakeOff = !g.save.ShakeOff
		case 2:
			g.save.GhostOff = !g.save.GhostOff
		}
		g.save.Save()
	}
	if IsEscPressed() {
		g.state = StateMenu
	}
}

func (g *Game) drawSettings(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0x0D, 0x0D, 0x1A, 0xFF})
	DrawText(dst, "S E T T I N G S", ScreenWidth/2-52, 60)

	type settingItem struct {
		label string
		on    bool
	}
	items := []settingItem{
		{"Sound", !g.save.SoundOff},
		{"Screen Shake", !g.save.ShakeOff},
		{"Ghost Car", !g.save.GhostOff},
	}
	for i, it := range items {
		y := 140 + i*30
		clr := colorUnselected
		prefix := "  "
		if i == g.settingsSelection {
			clr = colorSelected
			prefix = "> "
		}
		val := "OFF"
		if it.on {
			val = "ON"
		}
		DrawTextColor(dst, prefix+it.label, 100, y, clr)
		DrawTextColor(dst, val, 280, y, clr)
	}
	DrawTextColor(dst, "Left/Right toggle   Esc back", 60, ScreenHeight-30,
		color.RGBA{0x66, 0x66, 0x66, 0xFF})
}

// --- Playing ---

func (g *Game) dailyName() string {
	if g.dailyMode {
		return g.dailyChallenge.Name
	}
	return ""
}

func driftDurationMult(ticks int) int {
	switch {
	case ticks >= DriftDurationTier3:
		return 3
	case ticks >= DriftDurationTier2:
		return 2
	default:
		return 1
	}
}

func (g *Game) updatePlaying() {
	if IsEscPressed() {
		g.pauseSelection = 0
		g.state = StatePaused
		return
	}

	g.tickCount++
	if !(g.dailyMode && g.dailyChallenge.ID == ChallengeSerpentine) {
		g.score++
	}
	g.zone.Update()

	// Zone transition announcements.
	if g.zone.JustTransitioned {
		zoneName := g.zone.ZoneName()
		if g.zone.ZonesReached > g.save.BestZone {
			g.score += 500
			g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts,
				FloatingText{
					X: ScreenWidth/2 - 70, Y: ScreenHeight/2 - 60,
					Text: "NEW ZONE: " + zoneName,
					TTL: 120, MaxTTL: 120,
					Color: color.RGBA{0xFF, 0xFF, 0xFF, 0xFF},
					VY: -0.3, ScaleStart: 3.5, ScaleEnd: 2.0,
					ScaleTicks: 15,
				})
		} else {
			g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts,
				FloatingText{
					X: ScreenWidth/2 - 40, Y: ScreenHeight/2 - 50,
					Text: zoneName, TTL: 60, MaxTTL: 60,
					Color: color.RGBA{0xCC, 0xCC, 0xCC, 0xFF},
					VY: -0.5, ScaleStart: 2.0, ScaleEnd: 1.5,
					ScaleTicks: 10,
				})
		}
	}

	// Last Chance timer.
	if g.lastChanceActive {
		g.lastChanceTimer--
		if g.lastChanceTimer <= 0 {
			g.lastChanceActive = false
		}
		g.player.Blink = (g.lastChanceTimer/4)%2 == 0
	} else {
		g.player.Blink = false
	}

	// Fuel consumption.
	g.fuel -= g.fuelConsumption
	if g.fuel <= 0 {
		g.fuel = 0
		g.fuelEmpty = true
		g.enterGameOver()
		return
	}

	// Player speed control: gas / brake / coast.
	vertical := GetVerticalInput()
	// Daily modifiers on vertical input.
	if g.dailyMode {
		switch g.dailyChallenge.ID {
		case ChallengeNoBrakes:
			if vertical > 0 {
				vertical = 0
			}
		case ChallengeHeavyFoot:
			vertical = -1
		}
	}
	switch {
	case vertical < 0: // gas
		g.speed += g.activeCar.Acceleration
	case vertical > 0: // brake
		g.speed -= g.activeCar.BrakeForce
	default: // coast
		g.speed -= Deceleration
	}
	if g.speed < MinSpeed {
		g.speed = MinSpeed
	}
	if g.speed > g.activeCar.MaxSpeed {
		g.speed = g.activeCar.MaxSpeed
	}
	if g.speed > g.topSpeed {
		g.topSpeed = g.speed
	}

	g.accelerating = vertical < 0
	g.braking = vertical > 0

	g.audio.UpdateEngineSpeed(g.speed)

	effectiveSpeed := g.speed
	if g.lastChanceActive {
		effectiveSpeed *= LastChanceSpeedMult
	}

	g.road.Update(effectiveSpeed, g.tickCount)
	// Turn warning sound.
	if g.road.Curve != nil && g.road.Curve.JustEntered {
		g.audio.PlayTurnWarn()
	}
	if g.road.Curve != nil {
		speed := effectiveSpeed
		curve := g.road.Curve
		g.offsetFn = func(y float64) float64 {
			return curve.ScreenOffset(y, speed)
		}
	} else {
		g.offsetFn = func(float64) float64 { return 0 }
	}
	playerOff := g.offsetFn(PlayerStartY)
	g.decor.Update(effectiveSpeed, g.zone.CurrentZone, g.zone.ActivePalette)
	mirror := g.dailyMode && g.dailyChallenge.ID == ChallengeMirror
	g.player.Update(playerOff, mirror)

	// Drift: Shift + direction at speed > 5.
	if !g.lastChanceActive {
		hInput := GetHorizontalInput()
		if mirror {
			hInput = -hInput
		}
		if g.player.Drift.Update(&g.player, hInput, IsDriftPressed(), g.speed) {
			g.speed *= DriftSpeedDrag
		}
	} else {
		// Force-cancel drift during Last Chance.
		g.player.Drift.Active = false
		g.player.Drift.Rotation *= 0.85
	}

	// Drift scoring and effects.
	if g.player.Drift.Active {
		driftMult := driftDurationMult(g.player.Drift.DriftTicks)
		// DANGER DRIFT: shoulder drift = ×5 multiplier.
		if g.player.IsOnShoulder(playerOff) {
			driftMult = DriftDangerMult
			if g.tickCount%30 == 0 {
				g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts,
					FloatingText{
						X: g.player.X - 25, Y: g.player.Y - 45,
						Text: "DANGER!", TTL: 30, MaxTTL: 30,
						Color: color.RGBA{0xFF, 0x00, 0x00, 0xFF},
						VY: -1.8, ScaleStart: 3.0, ScaleEnd: 1.8,
						ScaleTicks: 8,
					})
			}
		}

		g.score += DriftScorePerTick * driftMult * g.scoreState.ComboMultiplier
		g.scoreState.ComboTimer = ComboDecayTicks

		// Smoke from wheels every 3 ticks.
		if g.tickCount%3 == 0 {
			g.particles.EmitDriftSmoke(g.player.X,
				g.player.Y+g.player.Height/2,
				g.player.Drift.Direction)
		}

		// Duration tier announcements.
		if g.player.Drift.DriftTicks == DriftDurationTier2 {
			g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts,
				FloatingText{
					X: g.player.X - 30, Y: g.player.Y - 35,
					Text: "DRIFT x2", TTL: 50, MaxTTL: 50,
					Color: color.RGBA{0xFF, 0x00, 0xFF, 0xFF},
					VY: -1.2, ScaleStart: 2.5, ScaleEnd: 1.5,
					ScaleTicks: 10,
				})
		}
		if g.player.Drift.DriftTicks == DriftDurationTier3 {
			g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts,
				FloatingText{
					X: g.player.X - 30, Y: g.player.Y - 35,
					Text: "DRIFT x3", TTL: 50, MaxTTL: 50,
					Color: color.RGBA{0xFF, 0x00, 0xFF, 0xFF},
					VY: -1.2, ScaleStart: 3.0, ScaleEnd: 1.8,
					ScaleTicks: 10,
				})
		}
	}

	// Shoulder: penalty when driving off the road surface.
	if g.player.IsOnShoulder(playerOff) {
		g.milestones.CleanTimer = 0
		g.speed *= 0.95
		if g.tickCount%4 == 0 {
			g.particles.EmitSparks(g.player.X, g.player.Y+g.player.Height/2, 2)
		}
		if g.tickCount%15 == 0 {
			g.audio.PlayScrape()
		}
		if g.shake.intensity < 1 {
			g.shake = ScreenShake{intensity: 1, decay: 0.9, frequency: 2}
		}
	}

	// Barrier collision: hitting the guardrail triggers crash cascade.
	if g.player.IsAtBarrier(playerOff) && !g.nitroActive && g.nitroGrace <= 0 && !g.lastChanceActive {
		if g.lastChanceAvailable {
			g.lastChanceAvailable = false
			g.lastChanceActive = true
			g.lastChanceTimer = LastChanceDuration
			g.player.Damaged = true
			g.freezeTimer = FreezeFrameCollision
			g.shake = ShakeCollision()
			g.particles.EmitCollisionBurst(g.player.X, g.player.Y, g.activeCar.Color)
			g.audio.PlayCrash()
			g.milestones.CleanTimer = 0
			g.scoreState.ComboMultiplier = 1
			g.scoreState.ComboTimer = 0
		} else if !g.ghostShieldActive {
			g.particles.EmitCollisionBurst(g.player.X, g.player.Y, g.activeCar.Color)
			g.shake = ShakeCollision()
			g.freezeTimer = FreezeFrameCollision
			g.milestones.CleanTimer = 0
			g.scoreState.ComboMultiplier = 1
			g.scoreState.ComboTimer = 0
			g.enterGameOver()
			return
		}
	}

	g.speedLines.Update(effectiveSpeed, MaxScrollSpeed, g.nitroActive)

	// Rain particles.
	if g.zone.CurrentZone == ZoneRain {
		g.particles.EmitRain(ScreenWidth)
	}

	// Spawn traffic.
	g.spawnTimer--
	if g.spawnTimer <= 0 {
		if car := SpawnTraffic(g.traffic, g.tickCount); car != nil {
			g.traffic = append(g.traffic, car)
		}
		base := max(MinSpawnRate, TrafficSpawnRate-g.tickCount/TPS)
		density := len(g.traffic)
		g.spawnInterval = base + density*8
		g.spawnTimer = g.spawnInterval
	}

	// Spawn fuel canisters.
	g.fuelSpawnTimer--
	if g.fuelSpawnTimer <= 0 {
		if item := SpawnItem(ItemFuel, g.items, g.traffic, g.player.X); item != nil {
			g.items = append(g.items, item)
		}
		g.fuelSpawnTimer = randRange(7*TPS, 13*TPS)
	}

	// Spawn nitro pickups.
	g.nitroSpawnTimer--
	if g.nitroSpawnTimer <= 0 {
		if item := SpawnItem(ItemNitro, g.items, g.traffic, g.player.X); item != nil {
			g.items = append(g.items, item)
		}
		g.nitroSpawnTimer = randRange(30*TPS, 60*TPS)
	}

	// Spawn oil spills (after 90 seconds).
	if g.tickCount >= 90*TPS {
		g.oilSpawnTimer--
		if g.oilSpawnTimer <= 0 {
			if item := SpawnItem(ItemOil, g.items, g.traffic); item != nil {
				g.items = append(g.items, item)
			}
			g.oilSpawnTimer = randRange(15*TPS, 25*TPS)
		}
	}

	// Spawn coin lines.
	g.coinSpawnTimer--
	if g.coinSpawnTimer <= 0 {
		if coins := SpawnCoinLine(g.items, g.traffic); coins != nil {
			g.coinLineCount = len(coins)
			g.coinLineCollected = 0
			g.items = append(g.items, coins...)
		}
		g.coinSpawnTimer = randRange(8*TPS, 15*TPS)
	}

	// Spawn repair kit (only when damaged, rare).
	if g.player.Damaged && !g.lastChanceActive {
		g.repairSpawnTimer--
		if g.repairSpawnTimer <= 0 {
			if item := SpawnItem(ItemRepair, g.items, g.traffic, g.player.X); item != nil {
				g.items = append(g.items, item)
			}
			g.repairSpawnTimer = randRange(20*TPS, 40*TPS)
		}
	}

	// Nitro activation.
	if IsNitroPressed() && g.nitroCharges > 0 && !g.nitroActive {
		g.nitroCharges--
		g.nitroActive = true
		g.nitroTimer = NitroDuration
		g.audio.PlayNitro()
		g.shake = ShakeNitroStart()
	}
	if g.nitroActive {
		g.nitroTimer--
		if g.nitroTimer <= 0 {
			g.nitroActive = false
			g.nitroGrace = 30 // 0.5s grace period after nitro ends
		}
		effectiveSpeed *= 2.0
	}
	if g.nitroGrace > 0 {
		g.nitroGrace--
	}

	var overtakeScore int
	g.traffic, overtakeScore = UpdateTraffic(g.traffic, effectiveSpeed, g.player.X-playerOff)
	g.score += overtakeScore

	g.items = UpdateItems(g.items, effectiveSpeed, g.player.X, g.player.Y, g.offsetFn)

	// Pick up items.
	picked, remaining := CheckPlayerItemCollision(&g.player, g.items, g.offsetFn)
	g.items = remaining
	for _, it := range picked {
		switch it.Type {
		case ItemFuel:
			g.fuel = min(g.fuel+FuelCanBonus, FuelMax)
			g.score += 25
			g.audio.PlayPickup()
			g.particles.EmitFuelPickup(it.X, it.Y)
		case ItemNitro:
			if g.nitroCharges < g.activeCar.MaxNitro {
				g.nitroCharges++
			}
			g.score += 15
			g.audio.PlayPickup()
		case ItemOil:
			g.player.ApplyOilSpin()
			g.shake = ShakeNearMiss()
		case ItemCoin:
			g.score += 100
			g.coinLineCollected++
			g.particles.EmitCoinPickup(it.X, it.Y)
			g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts, FloatingText{
				X: it.X - 15, Y: it.Y, Text: "+100", TTL: 40, MaxTTL: 40,
				Color: colorCoin, VY: -1.5, ScaleStart: 1.5, ScaleEnd: 1.2, ScaleTicks: 8,
			})
			g.audio.PlayPickup()
			if g.coinLineCollected == g.coinLineCount && g.coinLineCount > 0 {
				bonus := 100 * g.coinLineCount
				g.score += bonus
				g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts, FloatingText{
					X: g.player.X - 50, Y: g.player.Y - 40,
					Text: "PERFECT LINE! x2", TTL: 60, MaxTTL: 60,
					Color: colorCoin, VY: -1, ScaleStart: 2.5, ScaleEnd: 1.2, ScaleTicks: 12,
				})
			}
		case ItemRepair:
			g.lastChanceAvailable = true
			g.player.Damaged = false
			g.player.RepairGlowTimer = 20
			g.audio.PlayRepair()
			g.particles.EmitRepairBurst(it.X, it.Y)
			// Brief cyan screen flash.
			DrawRect(g.offscreen, 0, 0, ScreenWidth, ScreenHeight,
				color.RGBA{0x00, 0xFF, 0xDD, 0x23})
			g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts, FloatingText{
				X: it.X - 25, Y: it.Y, Text: "REPAIRED!", TTL: 60, MaxTTL: 60,
				Color: colorRepair, VY: -1.5, ScaleStart: 2.5, ScaleEnd: 1.2, ScaleTicks: 12,
			})
		}
	}

	// Near-miss checks.
	for _, car := range g.traffic {
		res := CheckNearMiss(&g.player, car, &g.scoreState, g.nearMissThreshold, effectiveSpeed, g.offsetFn)
		if res.Bonus > 0 {
			if g.player.Drift.Active {
				res.Bonus = int(float64(res.Bonus) * DriftNearMissMult)
				g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts,
					FloatingText{
						X: res.X - 20, Y: res.Y - 25,
						Text: "STYLE!", TTL: 45, MaxTTL: 45,
						Color: color.RGBA{0x00, 0xFF, 0xFF, 0xFF},
						VY: -2.0, ScaleStart: 3.0, ScaleEnd: 1.5,
						ScaleTicks: 10,
					})
			}
			g.score += res.Bonus
			g.nearMissCount++
			g.particles.EmitFlash(res.X, res.Y, car.X+g.offsetFn(car.Y), res.Tier)
			g.audio.PlayWoosh(res.Tier)
			g.chromatic = NewChromaticAberration(res.Tier)
			if g.scoreState.ComboMultiplier > g.peakCombo {
				g.peakCombo = g.scoreState.ComboMultiplier
				g.audio.PlayCombo()
			}
		}
	}
	UpdateScoreState(&g.scoreState)

	// Milestone checks.
	g.milestones.CleanTimer++
	if achieved, id := g.milestones.Check(
		g.topSpeed, g.activeCar.MaxSpeed,
		g.scoreState.ComboMultiplier,
		g.zone.ZonesReached, g.tickCount,
	); achieved {
		switch id {
		case MilestoneSpeedDemon:
			g.nitroCharges++
		case MilestoneUntouchable:
			g.milestones.ScoreMultTimer = 10 * TPS
		case MilestoneComboKing:
			g.score += milestoneDefs[id].Bonus
		case MilestoneMarathoner:
			g.fuel = FuelMax
		case MilestoneZoneSurfer:
			g.milestones.ZoneSurferX2 = true
		}
		g.audio.PlayArpeggio()
	}
	g.milestones.Update()

	// Apply milestone score multiplier.
	if mult := g.milestones.ScoreMultiplier(); mult > 1 {
		g.score += (mult - 1) // extra +1 per tick for each active x2
	}

	// Exhaust particles when accelerating.
	if g.accelerating && g.tickCount%3 == 0 {
		trailClr := TrailColor(g.save.SelectedTrail, g.tickCount)
		g.particles.EmitExhaust(g.player.X, g.player.Y+g.player.Height/2,
			g.speed/g.activeCar.MaxSpeed, trailClr)
	}

	// Brake trail particles.
	if g.braking && g.tickCount%3 == 0 {
		g.particles.EmitBrakeTrails(g.player.X, g.player.Y+g.player.Height/2)
	}

	// Ghost recording.
	g.ghostRecording.Record(g.player.X, g.tickCount)

	// Damage sparks: occasional sparks from rear of damaged car.
	if g.player.Damaged && g.tickCount%8 == 0 {
		g.particles.EmitSparks(
			g.player.X+(rand.Float64()*12-6),
			g.player.Y+g.player.Height/2, 2)
	}

	// Nitro flame particles.
	if g.nitroActive && g.tickCount%2 == 0 {
		g.particles.EmitNitroFlame(g.player.X, g.player.Y+g.player.Height/2)
	}

	g.particles.Update()

	// Check collisions: nitro/grace → ghost shield → Last Chance → game over.
	if cr := CheckPlayerTrafficCollision(&g.player, g.traffic, g.offsetFn); cr.Hit && !g.nitroActive && g.nitroGrace <= 0 && !g.lastChanceActive {
		if g.ghostShieldActive {
			g.ghostShieldActive = false
			g.particles.EmitFlash(g.player.X, g.player.Y, cr.HitX, TierClose)
			g.shake = ShakeNearMiss()
		} else if g.lastChanceAvailable {
			// Last Chance: slow-mo instead of death.
			g.lastChanceAvailable = false
			g.lastChanceActive = true
			g.lastChanceTimer = LastChanceDuration
			g.player.Damaged = true
			g.freezeTimer = FreezeFrameCollision
			g.shake = ShakeCollision()
			g.particles.EmitCollisionBurst(cr.HitX, cr.HitY, g.activeCar.Color)
			g.audio.PlayCrash()
			g.milestones.CleanTimer = 0
			g.scoreState.ComboMultiplier = 1
			g.scoreState.ComboTimer = 0
		} else {
			g.particles.EmitCollisionBurst(cr.HitX, cr.HitY, g.activeCar.Color)
			g.shake = ShakeCollision()
			g.freezeTimer = FreezeFrameCollision
			g.milestones.CleanTimer = 0
			g.scoreState.ComboMultiplier = 1
			g.scoreState.ComboTimer = 0
			g.busted = cr.CarType == CarTypePolice
			g.enterGameOver()
		}
	}
}

func (g *Game) drawPlaying(dst *ebiten.Image) {
	g.road.Draw(dst, g.zone.ActivePalette, g.zone.CurrentZone, g.offsetFn)
	g.decor.Draw(dst, g.zone.ActivePalette, g.sprites, g.offsetFn)

	for _, it := range g.items {
		it.Draw(dst, g.sprites, g.offsetFn)
	}

	for _, car := range g.traffic {
		car.Draw(dst, g.sprites, g.offsetFn)
	}

	// Ghost car from best run.
	if !g.save.GhostOff && len(g.save.GhostData) > 0 {
		idx := g.tickCount / ghostSampleInterval
		if idx >= 0 && idx < len(g.save.GhostData) {
			gx := g.save.GhostData[idx].X
			drawSpriteAlpha(dst, g.sprites.PlayerCars[0], gx, PlayerStartY, 0.15)
		}
	}

	g.player.Draw(dst, g.sprites, g.tickCount, g.braking)
	g.particles.Draw(dst)
	g.speedLines.Draw(dst)

	if g.zone.CurrentZone == ZoneRain {
		DrawFogOverlay(dst)
	}

	DrawVignette(dst, g.speed)

	// Damaged state overlay.
	if g.lastChanceActive {
		DrawRect(dst, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0xFF, 0x00, 0x00, 40})
		if (g.lastChanceTimer/8)%2 == 0 {
			DrawText(dst, "HULL DAMAGED", ScreenWidth/2-40, ScreenHeight/2-35)
			DrawText(dst, "Next hit is fatal", ScreenWidth/2-55, ScreenHeight/2-15)
		}
	}

	// Update display speed (smooth lerp).
	g.displaySpeed += (g.speed - g.displaySpeed) * 0.1
	// Score flash.
	if g.score != g.prevScore {
		g.scoreFlashTimer = 6
		g.prevScore = g.score
	}
	if g.scoreFlashTimer > 0 {
		g.scoreFlashTimer--
	}

	DrawHUD(dst, HUDData{
		Score:           g.score,
		HighScore:       g.save.HighScore,
		DisplaySpeed:    g.displaySpeed,
		Fuel:            g.fuel,
		NitroCharges:    g.nitroCharges,
		NitroActive:     g.nitroActive,
		ComboMultiplier: g.scoreState.ComboMultiplier,
		ComboTimer:      g.scoreState.ComboTimer,
		Damaged:         g.player.Damaged,
		RepairFlash:     g.player.RepairGlowTimer,
		TickCount:       g.tickCount,
		ScoreFlash:      g.scoreFlashTimer,
		DriftActive:     g.player.Drift.Active,
		DriftHeat:       g.player.Drift.HeatLevel,
		DriftOverheat:   g.player.Drift.Overheated,
		NeonAccent:      g.zone.ActivePalette.NeonAccent,
		DailyName:       g.dailyName(),
	})
	DrawFloatingTexts(dst, g.scoreState.FloatingTexts)

	// Milestone gold bar.
	if g.milestones.ShowTimer > 0 {
		t := g.milestones.ShowTimer
		maxW := float64(ScreenWidth) * 0.6
		var barW float64
		switch {
		case t > 128: // ease-in (first 8 ticks)
			progress := float64(136-t) / 8.0
			barW = maxW * progress * progress
		case t < 8: // ease-out (last 8 ticks)
			progress := float64(t) / 8.0
			barW = maxW * progress * progress
		default:
			barW = maxW
		}
		barX := (float64(ScreenWidth) - barW) / 2
		barY := 48.0
		alpha := uint8(220)
		if t < 8 {
			alpha = uint8(220 * t / 8)
		}
		DrawRect(dst, barX, barY, barW, 16,
			color.RGBA{0xFF, 0xD7, 0x00, alpha})
		if barW > 60 {
			DrawText(dst, g.milestones.ShowName,
				int(float64(ScreenWidth)/2)-len(g.milestones.ShowName)*3,
				int(barY)+2)
		}
	}
}

// --- Draw dispatcher ---

func (g *Game) Draw(screen *ebiten.Image) {
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Recreate offscreen when render resolution changes.
	if g.offscreen == nil || g.offscreen.Bounds().Dx() != sw || g.offscreen.Bounds().Dy() != sh {
		g.offscreen = ebiten.NewImage(sw, sh)
	}

	// Publish render scale for all Draw helpers.
	renderScaleGlobal = g.renderScale

	g.offscreen.Clear()
	dst := g.offscreen

	// Scale all drawing from logical coordinates to render resolution.
	rs := g.renderScale

	switch g.state {
	case StateMenu:
		g.drawMenu(dst)
	case StatePlaying:
		g.drawPlaying(dst)
	case StatePaused:
		g.drawPlaying(dst)
		g.drawPaused(dst)
	case StateGameOver:
		g.drawPlaying(dst)
		nextName, nextThreshold := g.save.NextUnlock()
		var nextPct float64
		if nextThreshold > 0 {
			nextPct = float64(g.save.TotalScore) / float64(nextThreshold)
			if nextPct > 1 {
				nextPct = 1
			}
		}
		DrawGameOver(dst, GameOverData{
			Score:          g.score,
			HighScore:      g.save.HighScore,
			IsNewHighScore: g.isNewHighScore,
			NearMisses:     g.nearMissCount,
			BestNearMisses: g.save.BestNearMisses,
			BestCombo:      g.peakCombo,
			SaveBestCombo:  g.save.BestCombo,
			TopSpeed:       g.topSpeed,
			BestTopSpeed:   g.save.BestTopSpeed,
			Distance:       float64(g.tickCount) / float64(TPS) * 0.05,
			ZoneName:       g.zone.ZoneName(),
			ZonesReached:   g.zone.ZonesReached,
			BestZone:       g.save.BestZone,
			Selection:      g.gameOverSelection,
			NewUnlocks:     g.newUnlocks,
			TotalScore:     g.save.TotalScore,
			Busted:         g.busted,
			FuelEmpty:      g.fuelEmpty,
			NextUnlockName: nextName,
			NextUnlockPct:  nextPct,
		})
	case StateGarage:
		g.drawGarage(dst)
	case StateSettings:
		g.drawSettings(dst)
	}

	var ox, oy float64
	if !g.save.ShakeOff {
		ox, oy = g.shake.Update()
	} else {
		g.shake = ScreenShake{} // suppress
	}
	g.chromatic.Update()
	caOffset := g.chromatic.Offset() * rs

	if caOffset < 0.5 {
		// Normal single blit.
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(ox*rs, oy*rs)
		screen.DrawImage(g.offscreen, op)
	} else {
		// RGB channel separation: 3 additive passes.
		opR := &ebiten.DrawImageOptions{}
		opR.GeoM.Translate(ox*rs+caOffset, oy*rs)
		opR.ColorScale.Scale(1, 0, 0, 1)
		opR.Blend = ebiten.BlendLighter
		screen.DrawImage(g.offscreen, opR)

		opG := &ebiten.DrawImageOptions{}
		opG.GeoM.Translate(ox*rs, oy*rs)
		opG.ColorScale.Scale(0, 1, 0, 1)
		opG.Blend = ebiten.BlendLighter
		screen.DrawImage(g.offscreen, opG)

		opB := &ebiten.DrawImageOptions{}
		opB.GeoM.Translate(ox*rs-caOffset, oy*rs)
		opB.ColorScale.Scale(0, 0, 1, 1)
		opB.Blend = ebiten.BlendLighter
		screen.DrawImage(g.offscreen, opB)
	}

	// Fade transition overlay (drawn directly on screen, after post-processing).
	g.transition.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	dsf := ebiten.Monitor().DeviceScaleFactor()
	// Fit logical 400×600 into the available area, preserving aspect ratio.
	fitScale := min(
		float64(outsideWidth)/float64(ScreenWidth),
		float64(outsideHeight)/float64(ScreenHeight),
	)
	renderW := int(float64(ScreenWidth) * fitScale * dsf)
	renderH := int(float64(ScreenHeight) * fitScale * dsf)
	g.renderScale = float64(renderW) / float64(ScreenWidth)
	return renderW, renderH
}

func (g *Game) enterGameOver() {
	g.state = StateGameOver
	g.gameOverSelection = 0
	g.scoreState.FloatingTexts = g.scoreState.FloatingTexts[:0]
	if g.fuelEmpty {
		// Fuel empty — engine sputter, not a crash.
		g.audio.StopEngine()
	} else {
		g.audio.PlayCrashEcho()
		g.audio.StopEngine()
	}

	g.isNewHighScore = g.score > g.save.HighScore
	g.save.GamesPlayed++
	g.save.TotalScore += g.score
	// Always save ghost recording; overwrite only on new record.
	if g.isNewHighScore || len(g.save.GhostData) == 0 {
		g.save.HighScore = max(g.score, g.save.HighScore)
		g.save.GhostData = g.ghostRecording.Frames
	}
	if g.peakCombo > g.save.BestCombo {
		g.save.BestCombo = g.peakCombo
	}
	if g.nearMissCount > g.save.BestNearMisses {
		g.save.BestNearMisses = g.nearMissCount
	}
	if g.topSpeed > g.save.BestTopSpeed {
		g.save.BestTopSpeed = g.topSpeed
	}
	if g.zone.ZonesReached > g.save.BestZone {
		g.save.BestZone = g.zone.ZonesReached
	}
	// Daily challenge reward.
	if g.dailyMode && g.save.DailyDone != TodayDateStr() {
		g.save.DailyDone = TodayDateStr()
		g.save.DailyStars++
		g.save.TotalScore += 5000
	}
	g.dailyMode = false

	g.newUnlocks = g.save.CheckUnlocks()
	g.newUnlocks = append(g.newUnlocks, g.save.CheckTrailUnlocks()...)
	g.save.Save()
}

func (g *Game) reset() {
	g.state = StatePlaying
	g.audio.StartEngine()
	g.activeCar = PlayerCars[g.save.SelectedCar]
	g.player = newPlayerFromCar(g.activeCar)
	g.nearMissThreshold = NearMissThreshold
	g.fuelConsumption = FuelConsumption
	g.ghostShieldActive = false
	switch g.activeCar.Special {
	case SpecialWiderNearMiss:
		g.nearMissThreshold = 12.0
	case SpecialFuelSaver:
		g.fuelConsumption = FuelConsumption * 0.8
	case SpecialGhostShield:
		g.ghostShieldActive = true
	}
	g.road = NewRoadWithCurve()
	g.offsetFn = func(float64) float64 { return 0 }
	g.traffic = g.traffic[:0]
	g.items = g.items[:0]
	g.score = 0
	g.speed = MinSpeed
	g.tickCount = 0
	g.spawnTimer = 0
	g.spawnInterval = TrafficSpawnRate
	g.fuel = FuelMax
	g.fuelSpawnTimer = randRange(7*TPS, 13*TPS)
	g.nitroCharges = 0
	g.nitroActive = false
	g.nitroTimer = 0
	g.nitroSpawnTimer = randRange(30*TPS, 60*TPS)
	g.zone = NewZoneSystem()
	g.decor = NewDecorSystem()
	g.scoreState = NewScoreState()
	g.particles = NewParticleSystem()
	g.braking = false
	g.accelerating = false
	g.shake = ScreenShake{}
	g.chromatic = ChromaticAberration{}
	g.freezeTimer = 0
	g.speedLines = SpeedLineSystem{}
	g.pauseSelection = 0
	g.peakCombo = 0
	g.nearMissCount = 0
	g.topSpeed = 0
	g.busted = false
	g.fuelEmpty = false
	g.lastChanceAvailable = true
	g.lastChanceActive = false
	g.lastChanceTimer = 0
	g.repairSpawnTimer = randRange(20*TPS, 40*TPS)
	g.nitroGrace = 0
	g.oilSpawnTimer = randRange(15*TPS, 25*TPS)
	g.coinSpawnTimer = randRange(8*TPS, 15*TPS)
	g.coinLineCount = 0
	g.coinLineCollected = 0
	g.ghostRecording = GhostRecording{}

	// Daily challenge modifiers.
	if g.dailyMode {
		switch g.dailyChallenge.ID {
		case ChallengeHeavyFoot:
			g.speed = 6.0
		case ChallengeFuelCrisis:
			g.fuel = FuelMax * 0.4
			g.fuelConsumption *= 1.5
		}
	}
}

func randRange(minVal, maxVal int) int {
	return minVal + rand.IntN(maxVal-minVal+1)
}
