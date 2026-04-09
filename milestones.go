package main

// MilestoneID identifies a milestone type.
type MilestoneID int

const (
	MilestoneSpeedDemon  MilestoneID = iota // reach 90% maxSpeed
	MilestoneUntouchable                     // 30s without collision/shoulder
	MilestoneComboKing                       // combo x10
	MilestoneMarathoner                      // survive 100s
	MilestoneZoneSurfer                      // reach 5 zones
)

const milestoneCount = 5

var milestoneDefs = [milestoneCount]struct {
	Name  string
	Bonus int
}{
	{"SPEED DEMON", 0},    // +1 nitro (applied in game.go)
	{"UNTOUCHABLE", 0},    // x2 score for 10s (applied in game.go)
	{"COMBO KING", 1000},  // instant points
	{"MARATHONER", 0},     // full fuel (applied in game.go)
	{"ZONE SURFER", 0},    // x2 permanent (applied in game.go)
}

// MilestoneSystem tracks in-session milestone achievements.
type MilestoneSystem struct {
	Achieved       [milestoneCount]bool
	CleanTimer     int    // ticks without collision/shoulder
	ShowTimer      int    // ticks remaining for gold bar
	ShowName       string // milestone name being displayed
	ScoreMultTimer int    // ticks remaining for UNTOUCHABLE x2
	ZoneSurferX2   bool   // permanent x2 from ZONE SURFER
}

// Check tests all milestone conditions and returns the first newly
// achieved milestone. Returns false if nothing new was achieved.
func (ms *MilestoneSystem) Check(
	topSpeed, maxSpeed float64,
	combo, zones, ticks int,
) (bool, MilestoneID) {
	checks := [milestoneCount]bool{
		topSpeed >= maxSpeed*0.9 && maxSpeed > 0,
		ms.CleanTimer >= 30*TPS,
		combo >= ComboMultiplierMax,
		ticks >= 100*TPS,
		zones >= 5,
	}
	for i, cond := range checks {
		if cond && !ms.Achieved[i] {
			ms.Achieved[i] = true
			ms.ShowTimer = 136 // 8 ease-in + 120 hold + 8 ease-out
			ms.ShowName = milestoneDefs[i].Name
			return true, MilestoneID(i)
		}
	}
	return false, 0
}

// Update decrements display and effect timers.
func (ms *MilestoneSystem) Update() {
	if ms.ShowTimer > 0 {
		ms.ShowTimer--
	}
	if ms.ScoreMultTimer > 0 {
		ms.ScoreMultTimer--
	}
}

// ScoreMultiplier returns the current score multiplier from milestones.
func (ms *MilestoneSystem) ScoreMultiplier() int {
	mult := 1
	if ms.ScoreMultTimer > 0 {
		mult *= 2
	}
	if ms.ZoneSurferX2 {
		mult *= 2
	}
	return mult
}
