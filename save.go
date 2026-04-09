package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

var carUnlockThresholds = [5]int{0, 5000, 20000, 50000, 100000}

var trailUnlockThresholds = [7]int{0, 2000, 6000, 15000, 30000, 60000, 120000}
var trailNames = [7]string{
	"DEFAULT", "TOXIC", "ICE", "ROYAL", "SOLAR", "VOID", "RAINBOW",
}

// SaveData holds persistent player progress.
type SaveData struct {
	TotalScore     int     `json:"total_score"`
	HighScore      int     `json:"high_score"`
	GamesPlayed    int     `json:"games_played"`
	UnlockedCars   []int   `json:"unlocked_cars"`
	SelectedCar    int     `json:"selected_car"`
	BestCombo      int     `json:"best_combo"`
	BestNearMisses int     `json:"best_near_misses"`
	BestTopSpeed   float64 `json:"best_top_speed"`
	BestZone       int     `json:"best_zone"`
	UnlockedTrails []int   `json:"unlocked_trails"`
	SelectedTrail  int     `json:"selected_trail"`
	DailyDone      string  `json:"daily_done"`
	DailyStars     int     `json:"daily_stars"`
}

func savePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".neonrush", "save.json")
}

// LoadSave reads save data from disk, returning defaults if not found.
func LoadSave() SaveData {
	data, err := os.ReadFile(savePath())
	if err != nil {
		return SaveData{UnlockedCars: []int{0}, UnlockedTrails: []int{0}}
	}
	var s SaveData
	if err := json.Unmarshal(data, &s); err != nil {
		return SaveData{UnlockedCars: []int{0}, UnlockedTrails: []int{0}}
	}
	if len(s.UnlockedCars) == 0 {
		s.UnlockedCars = []int{0}
	}
	if len(s.UnlockedTrails) == 0 {
		s.UnlockedTrails = []int{0}
	}
	// Silently unlock anything already earned (no notification spam).
	s.CheckUnlocks()
	s.CheckTrailUnlocks()
	return s
}

// Save writes save data to disk.
func (s *SaveData) Save() {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(savePath())
	_ = os.MkdirAll(dir, 0755)
	_ = os.WriteFile(savePath(), data, 0644)
}

// CheckUnlocks returns names of newly unlocked cars based on TotalScore.
func (s *SaveData) CheckUnlocks() []string {
	var newUnlocks []string
	for i, threshold := range carUnlockThresholds {
		if s.TotalScore >= threshold && !s.IsCarUnlocked(i) {
			s.UnlockedCars = append(s.UnlockedCars, i)
			newUnlocks = append(newUnlocks, carNames[i])
		}
	}
	return newUnlocks
}

// IsCarUnlocked checks if a car index is unlocked.
func (s *SaveData) IsCarUnlocked(idx int) bool {
	return slices.Contains(s.UnlockedCars, idx)
}

var carNames = [5]string{"STARTER", "SWIFT", "FURY", "PHANTOM", "GHOST"}

// IsTrailUnlocked checks if a trail index is unlocked.
func (s *SaveData) IsTrailUnlocked(idx int) bool {
	return slices.Contains(s.UnlockedTrails, idx)
}

// CheckTrailUnlocks returns names of newly unlocked trails.
func (s *SaveData) CheckTrailUnlocks() []string {
	var newUnlocks []string
	for i, threshold := range trailUnlockThresholds {
		if s.TotalScore >= threshold && !s.IsTrailUnlocked(i) {
			s.UnlockedTrails = append(s.UnlockedTrails, i)
			newUnlocks = append(newUnlocks, trailNames[i]+" TRAIL")
		}
	}
	return newUnlocks
}

// NextUnlock returns the name and threshold of the next locked
// car or trail, whichever has the lower threshold.
func (s *SaveData) NextUnlock() (name string, threshold int) {
	best := -1
	for i, t := range carUnlockThresholds {
		if t > s.TotalScore && !s.IsCarUnlocked(i) {
			if best < 0 || t < best {
				best = t
				name = carNames[i]
			}
		}
	}
	for i, t := range trailUnlockThresholds {
		if t > s.TotalScore && !s.IsTrailUnlocked(i) {
			if best < 0 || t < best {
				best = t
				name = trailNames[i] + " TRAIL"
			}
		}
	}
	if best > 0 {
		threshold = best
	}
	return
}
