package main

import (
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
	scrollSpeed float64
	tickCount   int
	fuel        float64

	nitroCharges    int
	nitroActive     bool
	nitroTimer      int
	nitroSpawnTimer int

	zone       ZoneSystem
	decor      DecorSystem
	scoreState ScoreState
	particles  ParticleSystem
	braking    bool

	audio          *AudioSystem
	sprites        *SpriteCache
	offscreen      *ebiten.Image
	renderScale    float64 // ratio of render buffer to logical size
	shakeTimer     int
	shakeOffsetX   float64
	shakeOffsetY   float64
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
	oilSpawnTimer      int
	coinSpawnTimer     int
	coinLineCount      int
	coinLineCollected  int
	gameOverSelection  int
	newUnlocks         []string
	isNewHighScore     bool
	busted             bool
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
		scrollSpeed:     BaseScrollSpeed,
		fuel:            FuelMax,
		fuelSpawnTimer:  randRange(10*TPS, 20*TPS),
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
	return g
}

func (g *Game) Update() error {
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
	}
	return nil
}

// --- Menu (stub for now) ---

func (g *Game) updateMenu() {
	g.menu.Update()

	if IsUpMenuPressed() {
		g.menu.Selection--
		if g.menu.Selection < 0 {
			g.menu.Selection = 1
		}
	}
	if IsDownMenuPressed() {
		g.menu.Selection++
		if g.menu.Selection > 1 {
			g.menu.Selection = 0
		}
	}
	if IsRestartPressed() {
		switch g.menu.Selection {
		case 0: // PLAY
			g.reset()
		case 1: // GARAGE
			g.state = StateGarage
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
	DebugPrintScaled(dst, "P A U S E D", cx, cy)

	items := []string{"RESUME", "RESTART", "QUIT"}
	for i, item := range items {
		marker := "  "
		if i == g.pauseSelection {
			marker = "> "
		}
		DebugPrintScaled(dst, marker+item, cx, cy+30+i*18)
	}
}

// --- Game Over ---

func (g *Game) updateGameOver() {
	g.particles.Update()
	if g.shakeTimer > 0 {
		g.shakeTimer--
		g.shakeOffsetX = (rand.Float64()*2 - 1) * 3
		g.shakeOffsetY = (rand.Float64()*2 - 1) * 3
	} else {
		g.shakeOffsetX = 0
		g.shakeOffsetY = 0
	}

	if IsLeftMenuPressed() || IsRightMenuPressed() {
		g.gameOverSelection = 1 - g.gameOverSelection
	}
	if IsRestartPressed() {
		if g.gameOverSelection == 0 {
			g.reset()
		} else {
			g.state = StateMenu
		}
	}
}

// --- Garage (stub for now) ---

func (g *Game) updateGarage() {
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
	if IsEscPressed() {
		g.state = StateMenu
	}
}

func (g *Game) drawGarage(dst *ebiten.Image) {
	DrawGarage(dst, g.garageSelection, &g.save, g.sprites)
}

// --- Playing ---

func (g *Game) updatePlaying() {
	if IsEscPressed() {
		g.pauseSelection = 0
		g.state = StatePaused
		return
	}

	g.tickCount++
	g.score++
	g.zone.Update()

	// Fuel consumption.
	g.fuel -= g.fuelConsumption
	if g.fuel <= 0 {
		g.fuel = 0
		g.enterGameOver()
		return
	}

	// Increase base speed.
	g.scrollSpeed += SpeedIncrement * g.activeCar.SpeedMod
	if g.scrollSpeed > MaxScrollSpeed {
		g.scrollSpeed = MaxScrollSpeed
	}

	g.audio.UpdateEngineSpeed(g.scrollSpeed)

	// Apply speed modifier from vertical input.
	vertical := GetVerticalInput()
	speedMod := 1.0
	if vertical < 0 {
		speedMod = SpeedBoostFactor
	} else if vertical > 0 {
		speedMod = SpeedBrakeFactor
	}
	effectiveSpeed := g.scrollSpeed * speedMod

	g.road.Update(effectiveSpeed)
	g.decor.Update(effectiveSpeed, g.zone.CurrentZone, g.zone.ActivePalette)
	g.player.Update()

	// Rain particles.
	if g.zone.CurrentZone == ZoneRain {
		g.particles.EmitRain(ScreenWidth)
	}

	// Spawn traffic.
	g.spawnTimer--
	if g.spawnTimer <= 0 {
		if car := SpawnTraffic(g.traffic, effectiveSpeed, g.tickCount); car != nil {
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
		if item := SpawnItem(ItemFuel, g.items, g.traffic); item != nil {
			g.items = append(g.items, item)
		}
		g.fuelSpawnTimer = randRange(10*TPS, 20*TPS)
	}

	// Spawn nitro pickups.
	g.nitroSpawnTimer--
	if g.nitroSpawnTimer <= 0 {
		if item := SpawnItem(ItemNitro, g.items, g.traffic); item != nil {
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

	// Nitro activation.
	if IsNitroPressed() && g.nitroCharges > 0 && !g.nitroActive {
		g.nitroCharges--
		g.nitroActive = true
		g.nitroTimer = NitroDuration
		g.audio.PlayNitro()
	}
	if g.nitroActive {
		g.nitroTimer--
		if g.nitroTimer <= 0 {
			g.nitroActive = false
		}
		effectiveSpeed *= 2.0
	}

	var overtakeScore int
	g.traffic, overtakeScore = UpdateTraffic(g.traffic, effectiveSpeed, g.player.X)
	g.score += overtakeScore

	g.items = UpdateItems(g.items, effectiveSpeed)

	// Pick up items.
	picked, remaining := CheckPlayerItemCollision(&g.player, g.items)
	g.items = remaining
	for _, it := range picked {
		switch it.Type {
		case ItemFuel:
			g.fuel = min(g.fuel+FuelCanBonus, FuelMax)
			g.score += 25
			g.audio.PlayPickup()
		case ItemNitro:
			if g.nitroCharges < g.activeCar.MaxNitro {
				g.nitroCharges++
			}
			g.score += 15
			g.audio.PlayPickup()
		case ItemOil:
			g.player.ApplyOilSpin()
			g.shakeTimer = 5
		case ItemCoin:
			g.score += 100
			g.coinLineCollected++
			g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts, FloatingText{
				X: it.X - 15, Y: it.Y, Text: "+100", TTL: 40, MaxTTL: 40,
				Color: colorCoin,
			})
			g.audio.PlayPickup()
			if g.coinLineCollected == g.coinLineCount && g.coinLineCount > 0 {
				bonus := 100 * g.coinLineCount
				g.score += bonus
				g.scoreState.FloatingTexts = append(g.scoreState.FloatingTexts, FloatingText{
					X: g.player.X - 50, Y: g.player.Y - 40,
					Text: "PERFECT LINE! x2", TTL: 60, MaxTTL: 60,
					Color: colorCoin,
				})
			}
		}
	}

	// Near-miss checks.
	for _, car := range g.traffic {
		bonus := CheckNearMiss(&g.player, car, &g.scoreState, g.nearMissThreshold)
		if bonus > 0 {
			g.score += bonus
			g.nearMissCount++
			g.particles.EmitFlash(g.player.X, g.player.Y)
			g.audio.PlayWoosh()
			if g.scoreState.ComboMultiplier > g.peakCombo {
				g.peakCombo = g.scoreState.ComboMultiplier
				g.audio.PlayCombo()
			}
		}
	}
	UpdateScoreState(&g.scoreState)

	// Brake trail particles.
	g.braking = vertical > 0
	if g.braking && g.tickCount%3 == 0 {
		g.particles.EmitBrakeTrails(g.player.X, g.player.Y+g.player.Height/2)
	}

	// Nitro flame particles.
	if g.nitroActive && g.tickCount%2 == 0 {
		g.particles.EmitNitroFlame(g.player.X, g.player.Y+g.player.Height/2)
	}

	g.particles.Update()

	// Screen shake.
	if g.shakeTimer > 0 {
		g.shakeTimer--
		g.shakeOffsetX = (rand.Float64()*2 - 1) * 3
		g.shakeOffsetY = (rand.Float64()*2 - 1) * 3
	} else {
		g.shakeOffsetX = 0
		g.shakeOffsetY = 0
	}

	// Check collisions (nitro grants invulnerability).
	if cr := CheckPlayerTrafficCollision(&g.player, g.traffic); cr.Hit && !g.nitroActive {
		if g.ghostShieldActive {
			g.ghostShieldActive = false
			g.particles.EmitFlash(g.player.X, g.player.Y)
			g.shakeTimer = 5
		} else {
			g.particles.EmitSparks(cr.HitX, cr.HitY, 10)
			g.shakeTimer = 10
			g.scoreState.ComboMultiplier = 1
			g.scoreState.ComboTimer = 0
			g.busted = cr.CarType == CarTypePolice
			g.enterGameOver()
		}
	}
}

func (g *Game) drawPlaying(dst *ebiten.Image) {
	g.road.Draw(dst, g.zone.ActivePalette, g.zone.CurrentZone)
	g.decor.Draw(dst, g.zone.ActivePalette, g.sprites)

	for _, it := range g.items {
		it.Draw(dst, g.sprites)
	}

	for _, car := range g.traffic {
		car.Draw(dst, g.sprites)
	}

	g.player.Draw(dst, g.sprites)
	g.particles.Draw(dst)

	if g.zone.CurrentZone == ZoneRain {
		DrawFogOverlay(dst)
	}

	DrawHUD(dst, HUDData{
		Score:           g.score,
		ScrollSpeed:     g.scrollSpeed,
		Fuel:            g.fuel,
		NitroCharges:    g.nitroCharges,
		NitroActive:     g.nitroActive,
		ComboMultiplier: g.scoreState.ComboMultiplier,
		ComboTimer:      g.scoreState.ComboTimer,
	})
	DrawFloatingTexts(dst, g.scoreState.FloatingTexts)
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
		DrawGameOver(dst, GameOverData{
			Score:          g.score,
			HighScore:      g.save.HighScore,
			IsNewHighScore: g.isNewHighScore,
			NearMisses:     g.nearMissCount,
			BestCombo:      g.peakCombo,
			Distance:       float64(g.tickCount) / float64(TPS) * 0.05,
			ZoneName:       g.zone.ZoneName(),
			Selection:      g.gameOverSelection,
			NewUnlocks:     g.newUnlocks,
			TotalScore:     g.save.TotalScore,
			Busted:         g.busted,
		})
	case StateGarage:
		g.drawGarage(dst)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.shakeOffsetX*rs, g.shakeOffsetY*rs)
	screen.DrawImage(g.offscreen, op)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	scale := ebiten.Monitor().DeviceScaleFactor()
	renderW := int(float64(outsideWidth) * scale)
	renderH := int(float64(outsideHeight) * scale)
	g.renderScale = float64(renderW) / float64(ScreenWidth)
	return renderW, renderH
}

func (g *Game) enterGameOver() {
	g.state = StateGameOver
	g.gameOverSelection = 0
	g.audio.PlayCrash()
	g.audio.StopEngine()

	g.isNewHighScore = g.score > g.save.HighScore
	g.save.GamesPlayed++
	g.save.TotalScore += g.score
	if g.isNewHighScore {
		g.save.HighScore = g.score
	}
	if g.peakCombo > g.save.BestCombo {
		g.save.BestCombo = g.peakCombo
	}
	if g.nearMissCount > g.save.BestNearMisses {
		g.save.BestNearMisses = g.nearMissCount
	}
	g.newUnlocks = g.save.CheckUnlocks()
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
	g.road = NewRoad()
	g.traffic = g.traffic[:0]
	g.items = g.items[:0]
	g.score = 0
	g.scrollSpeed = BaseScrollSpeed
	g.tickCount = 0
	g.spawnTimer = 0
	g.spawnInterval = TrafficSpawnRate
	g.fuel = FuelMax
	g.fuelSpawnTimer = randRange(10*TPS, 20*TPS)
	g.nitroCharges = 0
	g.nitroActive = false
	g.nitroTimer = 0
	g.nitroSpawnTimer = randRange(30*TPS, 60*TPS)
	g.zone = NewZoneSystem()
	g.decor = NewDecorSystem()
	g.scoreState = NewScoreState()
	g.particles = NewParticleSystem()
	g.braking = false
	g.shakeTimer = 0
	g.shakeOffsetX = 0
	g.shakeOffsetY = 0
	g.pauseSelection = 0
	g.peakCombo = 0
	g.nearMissCount = 0
	g.busted = false
	g.oilSpawnTimer = randRange(15*TPS, 25*TPS)
	g.coinSpawnTimer = randRange(8*TPS, 15*TPS)
	g.coinLineCount = 0
	g.coinLineCollected = 0
}

func randRange(minVal, maxVal int) int {
	return minVal + rand.IntN(maxVal-minVal+1)
}
