package main

import "math/rand/v2"

// DriftState tracks the player's drift (powerslide) state.
type DriftState struct {
	Active       bool
	Direction    float64 // -1 left, +1 right
	HeatLevel    float64 // 0..1
	Overheated   bool
	ForcedTimer  int     // loss-of-control countdown
	InertiaTimer int     // post-drift momentum countdown
	DriftTicks   int     // ticks of current drift session
	Rotation     float64 // visual rotation in radians
}

// Update advances drift state. Returns true if speed drag should be applied.
func (d *DriftState) Update(p *Player, hInput float64, driftPressed bool, speed float64) bool {
	// Forced loss of control after overheat.
	if d.ForcedTimer > 0 {
		d.ForcedTimer--
		p.LateralVelocity += (rand.Float64()*2 - 1) * 0.8
		d.HeatLevel -= (1.0 / DriftOverheatMax) * DriftCooldownMult
		if d.HeatLevel < 0 {
			d.HeatLevel = 0
		}
		if d.ForcedTimer <= 0 {
			d.Overheated = false
		}
		d.Rotation *= 0.9
		return false
	}

	// Active drift.
	if d.Active {
		released := !driftPressed || hInput == 0 || speed < DriftMinSpeed
		overheated := d.HeatLevel >= 1.0

		if released || overheated {
			d.Active = false
			if overheated {
				d.Overheated = true
				d.ForcedTimer = DriftOverheatForced
				d.HeatLevel = 1.0
			} else {
				d.InertiaTimer = DriftInertiaMax
			}
			return false
		}

		// Heat up.
		d.HeatLevel += 1.0 / DriftOverheatMax
		d.DriftTicks++

		// Lateral boost: add extra velocity beyond normal input.
		p.LateralVelocity += d.Direction * p.Acceleration * (DriftLateralMult - 1.0)

		// Visual rotation proportional to speed.
		speedNorm := speed / MaxScrollSpeed
		d.Rotation = d.Direction * DriftRotationMax * speedNorm

		return true // apply speed drag
	}

	// Post-drift inertia.
	if d.InertiaTimer > 0 {
		frac := float64(d.InertiaTimer) / float64(DriftInertiaMax)
		p.LateralVelocity += d.Direction * p.Acceleration * DriftLateralMult * frac * 0.5
		d.Rotation *= 0.85
		d.InertiaTimer--
		return false
	}

	// Cooldown when not drifting.
	if d.HeatLevel > 0 {
		d.HeatLevel -= (1.0 / DriftOverheatMax) * DriftCooldownMult
		if d.HeatLevel < 0 {
			d.HeatLevel = 0
		}
	}

	// Decay rotation when idle.
	if d.Rotation != 0 {
		d.Rotation *= 0.85
		if d.Rotation > -0.01 && d.Rotation < 0.01 {
			d.Rotation = 0
		}
	}

	// Activation: Shift + direction + enough speed + not overheated.
	if driftPressed && hInput != 0 && speed >= DriftMinSpeed && !d.Overheated && p.SpinTimer <= 0 {
		d.Active = true
		d.Direction = hInput
		d.DriftTicks = 0
		return false
	}

	return false
}
