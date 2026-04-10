package main

// GhostFrame stores player X position at a point in time.
type GhostFrame struct {
	X float64 `json:"x"`
}

// GhostRecording holds a sequence of ghost frames.
type GhostRecording struct {
	Frames []GhostFrame
}

const ghostSampleInterval = 6 // record every 6 ticks

// Record adds a frame if this tick is a sample point.
func (gr *GhostRecording) Record(x float64, tick int) {
	if tick%ghostSampleInterval != 0 {
		return
	}
	gr.Frames = append(gr.Frames, GhostFrame{X: x})
}

// PositionAt returns the ghost X position for a given tick.
func (gr *GhostRecording) PositionAt(tick int) (float64, bool) {
	idx := tick / ghostSampleInterval
	if idx < 0 || idx >= len(gr.Frames) {
		return 0, false
	}
	return gr.Frames[idx].X, true
}
