package main

import (
	"math"
	"math/rand/v2"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const sampleRate = 48000

// AudioSystem manages all game sounds.
type AudioSystem struct {
	context      *audio.Context
	engine       *EngineSound
	enginePlayer *audio.Player

	sfxWoosh      [4][]byte // per tier: Near, Close, VeryClose, Insane
	sfxCrash      []byte
	sfxPickup     []byte
	sfxNitro      []byte
	sfxCombo      []byte
}

// NewAudioSystem creates the audio context and generates all sound effects.
func NewAudioSystem() *AudioSystem {
	ctx := audio.NewContext(sampleRate)

	eng := &EngineSound{
		baseFreq1:   80,
		baseFreq2:   120,
		sampleRate:  sampleRate,
		freqModBits: math.Float64bits(1.0),
	}

	engPlayer, _ := ctx.NewPlayerF32(eng)
	engPlayer.SetVolume(0.12)

	return &AudioSystem{
		context:      ctx,
		engine:       eng,
		enginePlayer: engPlayer,
		sfxWoosh: [4][]byte{
			generateWooshTier(0.15, 8.0, []float64{880}, 0.15, 0.35, 8000),
			generateWooshTier(0.20, 6.0, []float64{1047}, 0.20, 0.40, 8000),
			generateWooshTier(0.25, 5.0, []float64{1047, 1319}, 0.25, 0.40, 10000),
			generateWooshTier(0.30, 4.0, []float64{1047, 1319, 1568}, 0.20, 0.40, 10000),
		},
		sfxCrash: generateCrash(),
		sfxPickup:    generatePickup(),
		sfxNitro:     generateNitroSFX(),
		sfxCombo:     generateCombo(),
	}
}

func (a *AudioSystem) StartEngine() {
	a.enginePlayer.Play()
}

func (a *AudioSystem) StopEngine() {
	a.enginePlayer.Pause()
}

func (a *AudioSystem) UpdateEngineSpeed(scrollSpeed float64) {
	a.engine.SetSpeedMod(scrollSpeed)
}

func (a *AudioSystem) PlayWoosh(tier NearMissTier) {
	idx := int(tier) - 1 // TierNear=1 → index 0
	if idx < 0 || idx >= len(a.sfxWoosh) {
		idx = 0
	}
	vol := 0.20 + float64(idx)*0.05 // 0.20, 0.25, 0.30, 0.35
	a.playSFX(a.sfxWoosh[idx], vol)
}
func (a *AudioSystem) PlayCrash()  { a.playSFX(a.sfxCrash, 0.35) }
func (a *AudioSystem) PlayPickup() { a.playSFX(a.sfxPickup, 0.30) }
func (a *AudioSystem) PlayNitro()  { a.playSFX(a.sfxNitro, 0.25) }
func (a *AudioSystem) PlayCombo()  { a.playSFX(a.sfxCombo, 0.20) }

func (a *AudioSystem) playSFX(buf []byte, volume float64) {
	p := a.context.NewPlayerFromBytes(buf)
	p.SetVolume(volume)
	p.Play()
}

// EngineSound generates a continuous sawtooth drone, implementing io.Reader
// for Ebitengine's F32 audio player (IEEE 754 float32 LE, stereo, 48kHz).
type EngineSound struct {
	phase1, phase2  float64
	baseFreq1       float64
	baseFreq2       float64
	freqModBits     uint64 // atomic
	sampleRate      int
}

func (e *EngineSound) SetSpeedMod(scrollSpeed float64) {
	mod := 1.0 + (scrollSpeed-BaseScrollSpeed)/(MaxScrollSpeed-BaseScrollSpeed)*1.5
	if mod < 0.5 {
		mod = 0.5
	}
	atomic.StoreUint64(&e.freqModBits, math.Float64bits(mod))
}

func (e *EngineSound) freqMod() float64 {
	return math.Float64frombits(atomic.LoadUint64(&e.freqModBits))
}

// Read fills buf with F32 stereo PCM (4 bytes per sample, 2 channels = 8 bytes per frame).
func (e *EngineSound) Read(buf []byte) (int, error) {
	mod := e.freqMod()
	sr := float64(e.sampleRate)
	frameSize := 8 // 2 channels * 4 bytes (float32)
	numFrames := len(buf) / frameSize

	for i := range numFrames {
		s1 := sawtooth(e.phase1)
		s2 := sawtooth(e.phase2)
		sample := float32((s1*0.5 + s2*0.5) * 0.25)

		e.phase1 += e.baseFreq1 * mod / sr
		e.phase2 += e.baseFreq2 * mod / sr
		if e.phase1 >= 1.0 {
			e.phase1 -= 1.0
		}
		if e.phase2 >= 1.0 {
			e.phase2 -= 1.0
		}

		bits := math.Float32bits(sample)
		off := i * frameSize
		// Left channel.
		buf[off] = byte(bits)
		buf[off+1] = byte(bits >> 8)
		buf[off+2] = byte(bits >> 16)
		buf[off+3] = byte(bits >> 24)
		// Right channel (same).
		buf[off+4] = buf[off]
		buf[off+5] = buf[off+1]
		buf[off+6] = buf[off+2]
		buf[off+7] = buf[off+3]
	}
	return numFrames * frameSize, nil
}

func sawtooth(phase float64) float64 {
	return 2.0*phase - 1.0
}

// --- SFX generators (PCM signed 16-bit LE, stereo, 48kHz → []byte) ---

// generateWooshTier creates a near-miss swoosh with tier-specific parameters.
// freqs: tonal ping frequencies, pingAmp: amplitude per tone,
// noiseAmp: filtered noise amplitude, startCutoff: initial LP cutoff Hz.
func generateWooshTier(dur, decay float64, freqs []float64, pingAmp, noiseAmp, startCutoff float64) []byte {
	n := int(dur * sampleRate)
	buf := make([]byte, n*4)
	var lpState float64

	for i := range n {
		t := float64(i) / float64(n)
		env := math.Exp(-t*decay) * math.Sin(t*math.Pi)

		cutoff := 400.0 + (startCutoff-400.0)*(1.0-t)*(1.0-t)
		rc := 1.0 / (2.0 * math.Pi * cutoff)
		a := 1.0 / (1.0 + rc*sampleRate)

		noise := rand.Float64()*2 - 1
		lpState += a * (noise - lpState)

		// Tonal pings (chord for higher tiers).
		var ping float64
		for _, freq := range freqs {
			ping += math.Sin(2*math.Pi*freq*float64(i)/sampleRate) * pingAmp
		}
		ping *= math.Exp(-t * 16.0)

		sample := (lpState*noiseAmp + ping) * env
		writeSample16(buf, i, sample)
	}
	return buf
}

func generateCrash() []byte {
	dur := 0.5
	n := int(dur * sampleRate)
	buf := make([]byte, n*4)
	for i := range n {
		t := float64(i) / float64(n)
		env := math.Exp(-t * 6)
		noise := (rand.Float64()*2 - 1) * env * 0.5
		// Low thud for first 0.1s.
		thud := 0.0
		if t < 0.1 {
			thud = math.Sin(2*math.Pi*60*t) * (1.0 - t/0.1) * 0.6
		}
		writeSample16(buf, i, noise+thud)
	}
	return buf
}

func generatePickup() []byte {
	dur := 0.15
	n := int(dur * sampleRate)
	buf := make([]byte, n*4)
	mid := n / 2
	for i := range n {
		t := float64(i) / float64(sampleRate)
		freq := 523.0 // C5
		if i >= mid {
			freq = 659.0 // E5
		}
		env := 1.0 - float64(i%mid)/float64(mid)
		sample := math.Sin(2*math.Pi*freq*t) * env * 0.4
		writeSample16(buf, i, sample)
	}
	return buf
}

func generateNitroSFX() []byte {
	dur := 0.6
	n := int(dur * sampleRate)
	buf := make([]byte, n*4)
	for i := range n {
		t := float64(i) / float64(sampleRate)
		progress := float64(i) / float64(n)
		freq := 200.0 * math.Pow(10, progress) // 200 → 2000 Hz
		env := math.Min(1.0, progress*4) * (1.0 - progress*0.5)
		sweep := math.Sin(2*math.Pi*freq*t) * 0.3
		noise := (rand.Float64()*2 - 1) * 0.15
		writeSample16(buf, i, (sweep+noise)*env)
	}
	return buf
}

func generateCombo() []byte {
	dur := 0.1
	n := int(dur * sampleRate)
	buf := make([]byte, n*4)
	for i := range n {
		t := float64(i) / float64(sampleRate)
		progress := float64(i) / float64(n)
		freq := 400.0 + 400.0*progress // 400→800 Hz
		env := 1.0 - progress
		sample := math.Sin(2*math.Pi*freq*t) * env * 0.35
		writeSample16(buf, i, sample)
	}
	return buf
}

func writeSample16(buf []byte, i int, sample float64) {
	val := int16(clampF(-1, 1, sample) * 32767)
	off := i * 4
	buf[off] = byte(val)
	buf[off+1] = byte(val >> 8)
	buf[off+2] = byte(val)
	buf[off+3] = byte(val >> 8)
}

func clampF(lo, hi, v float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
