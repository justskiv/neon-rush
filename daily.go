package main

import "time"

// ChallengeID identifies a daily challenge variant.
type ChallengeID int

const (
	ChallengeNoBrakes   ChallengeID = iota // brake key disabled
	ChallengeHeavyFoot                     // always accelerating
	ChallengeMirror                        // left/right swapped
	ChallengeFuelCrisis                    // 40% fuel, 1.5× drain
	ChallengeSerpentine                    // only near-miss points
)

// DailyChallenge describes a daily challenge.
type DailyChallenge struct {
	ID   ChallengeID
	Name string
}

var challenges = [5]DailyChallenge{
	{ChallengeNoBrakes, "NO BRAKES"},
	{ChallengeHeavyFoot, "HEAVY FOOT"},
	{ChallengeMirror, "MIRROR"},
	{ChallengeFuelCrisis, "FUEL CRISIS"},
	{ChallengeSerpentine, "SERPENTINE"},
}

// TodayChallenge returns the daily challenge seeded by today's date.
func TodayChallenge() DailyChallenge {
	now := time.Now()
	seed := now.Year()*10000 + int(now.Month())*100 + now.Day()
	return challenges[seed%len(challenges)]
}

// TodayDateStr returns today's date as "YYYY-MM-DD".
func TodayDateStr() string {
	return time.Now().Format("2006-01-02")
}
