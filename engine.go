package picosynth

import (
	"math/bits"
	"time"

	"gbenson.net/go/picosynth/internal/audio"
	"gbenson.net/go/picosynth/internal/display"
)

const (
	SampleRate = 48000
	MaxLatency = 10 * time.Millisecond

	// Split MaxLatency between scanning keys and playing audio.
	// Currently maxLatency is *half* of MaxLatency because:
	//  - The KeyScanner loop takes the first part, the limiting
	//    case being a key or button changing state the instant
	//    after being scanned, and a full loop has to complete
	//    before a KeyEvent will be emitted.
	//  - The time it takes to play the audio buffer takes up
	//    the second part, the limiting case being that the first
	//    sample generated into every buffer has to wait for the
	//    entire other buffer to play before it will be output.
	maxLatency = MaxLatency / 2

	MinVolume     = 0
	MaxVolume     = 10
	InitialVolume = 7

	MinOctave     = -3
	MaxOctave     = 3
	InitialOctave = -1
)

type Engine struct {
	display display.Display

	ks KeyScanner
	kt KeyTracker

	octave int

	lfo1 BasicOscillator

	ampEnv Envelope

	osc1 BasicOscillator

	volume int
}

func (ps *Engine) init() {
	ps.kt.init()

	ps.lfo1.Frequency = 10 * Hz
	ps.lfo1.Shaper = SineShaper

	ps.osc1.Shaper = TriSawShaper
	ps.osc1.Shape = MaxSignal // rising saw

	ps.setOctave(InitialOctave)
	ps.setVolume(InitialVolume)
}

// Run is the main entry point of the firmware.
func (ps *Engine) Run() error {
	ps.init()

	out, err := audio.Open(SampleRate)
	if err != nil {
		return err
	}
	defer out.Close()

	const numWorkers = 4 // display, keyscanner, filler, player
	wm := newWorkerManager(numWorkers)
	wm.Start(&ps.display)
	wm.Start(&ps.ks)

	// Calculate how many frames we can buffer without exceeding
	// maxLatency, then round down to a power of two.
	bufferFrames := int(SampleRate * maxLatency / time.Second)
	bufferFrames = 1 << (bits.Len(uint(bufferFrames)) - 1)

	db := newDoubleBuffer[int16](bufferFrames, ps.Fill, out.WriteMono)

	wm.Start(db.Filler)
	wm.Start(db.Player)

	return wm.Wait()
}

// Fill generates samples into the supplied buffer.
func (ps *Engine) Fill(buf []int16) error {
	for e := ps.ks.Poll(); e != NoEvent; e = ps.ks.Poll() {
		sc := e.Scancode()
		if note := sc.Note(); note.IsValid() {
			ps.kt.Receive(note, e.Down())
		} else if !e.Down() {
			ps.onButton(sc)
		}
	}

	// Final shift converts 32-bit to 16 and applies ps.volume.
	var finalShift int
	if v := ps.volume; v <= MinVolume {
		finalShift = 32 // silence
	} else {
		finalShift = (16 + (MaxVolume - v)) & 0x1f
	}

	ps.kt.Step()
	pitch := ps.kt.Note.Pitch()
	ps.ampEnv.Gate = ps.kt.Gate

	for i := range buf {
		ps.lfo1.Step()

		osc1Pitch := Pitch(Signal(pitch) + (ps.lfo1.Output >> 9))
		ps.osc1.Frequency = osc1Pitch.Frequency()
		ps.osc1.Step()

		output := ps.osc1.Output

		ps.ampEnv.Step()
		output = output.Mul(ps.ampEnv.Level)

		// Convert 32-bit to 16 and apply ps.volume.
		buf[i] = int16(output >> finalShift)
	}

	return nil
}

func (ps *Engine) onButton(sc Scancode) {
	switch sc {
	case ButtonVolumeUp:
		ps.setVolume(ps.volume + 1)
	case ButtonVolumeDown:
		ps.setVolume(ps.volume - 1)
	case ButtonTempoUp:
		ps.setOctave(ps.octave + 1)
	case ButtonTempoDown:
		ps.setOctave(ps.octave - 1)
	default:
		println("button", sc, "pressed")
	}
}

func (ps *Engine) setOctave(v int) {
	ps.octave = max(MinOctave, min(MaxOctave, v))
	ps.kt.Transpose = ps.octave * 12
}

func (ps *Engine) setVolume(v int) {
	ps.volume = max(MinVolume, min(MaxVolume, v))
}
