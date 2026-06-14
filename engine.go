package picosynth

import (
	"math/bits"
	"time"

	"gbenson.net/go/picosynth/internal/audio"
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
	ks KeyScanner
	kt KeyTracker

	octave int

	osc1 PhaseAccumulator

	volume int
}

func (ps *Engine) init() {
	ps.kt.init()

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

	const numWorkers = 3 // keyscanner, filler, player
	wm := newWorkerManager(numWorkers)
	wm.Start(&ps.ks)

	// Calculate how many frames we can buffer without exceeding
	// maxLatency, then round down to a power of two.
	bufferFrames := int(SampleRate * maxLatency / time.Second)
	bufferFrames = 1 << (bits.Len(uint(bufferFrames)) - 1)

	db := newDoubleBuffer[uint16](bufferFrames, ps.Fill, out.WriteMono)

	wm.Start(db.Filler)
	wm.Start(db.Player)

	return wm.Wait()
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
	println("octave", ps.octave)
	ps.kt.Transpose = ps.octave * 12
}

func (ps *Engine) setVolume(v int) {
	ps.volume = max(MinVolume, min(MaxVolume, v))
	println("volume", ps.volume)
}

// Fill generates samples into the supplied buffer.
func (ps *Engine) Fill(buf []uint16) error {
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

	for i := range buf {
		ps.osc1.Frequency = pitch.Frequency()
		ps.osc1.Step()

		output := ps.osc1.Phase

		// Convert 32-bit to 16 and apply ps.volume.  Note, we cast
		// to uint16 because that's how piolib I2S.WriteMono wants
		// things, but the cast doesn't change the *bits*, the I2S
		// protocol specifies signed (two's complement) audio data
		// and that's what we're writing into the buffer.
		buf[i] = uint16(output >> finalShift)
	}

	return nil
}
