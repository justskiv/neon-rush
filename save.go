package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

var carUnlockThresholds = [5]int{0, 5000, 20000, 50000, 100000}

// SaveData holds persistent player progress.
type SaveData struct {
	TotalScore     int   `json:"total_score"`
	HighScore      int   `json:"high_score"`
	GamesPlayed    int   `json:"games_played"`
	UnlockedCars   []int `json:"unlocked_cars"`
	SelectedCar    int   `json:"selected_car"`
	BestCombo      int   `json:"best_combo"`
	BestNearMisses int   `json:"best_near_misses"`
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
		return SaveData{UnlockedCars: []int{0}}
	}
	var s SaveData
	if err := json.Unmarshal(data, &s); err != nil {
		return SaveData{UnlockedCars: []int{0}}
	}
	if len(s.UnlockedCars) == 0 {
		s.UnlockedCars = []int{0}
	}
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
