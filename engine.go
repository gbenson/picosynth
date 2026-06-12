package picosynth

import (
	"math/bits"
	"time"

	"gbenson.net/go/picosynth/internal/audio"
)

const (
	SampleRate = 48000
	MaxLatency = 10 * time.Millisecond

	MinVolume     = 0
	MaxVolume     = 10
	InitialVolume = 7
)

type Engine struct {
	ks KeyScanner

	note Note
	osc1 PhaseAccumulator

	volume int
}

func (ps *Engine) init() {
	ps.volume = InitialVolume
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
	// MaxLatency, then round down to a power of two.
	bufferFrames := int(SampleRate * MaxLatency / time.Second)
	bufferFrames = 1 << (bits.Len(uint(bufferFrames)) - 1)

	db := newDoubleBuffer[uint16](bufferFrames, ps.Fill, out.WriteMono)

	wm.Start(db.Filler)
	wm.Start(db.Player)

	return wm.Wait()
}

// Fill generates samples into the supplied buffer.
func (ps *Engine) Fill(buf []uint16) error {
	for e := ps.ks.Poll(); e != NoEvent; e = ps.ks.Poll() {
		if e.Down() {
			println(e.Scancode(), "down")
		} else {
			println(e.Scancode(), "up")
		}
	}

	// Final shift converts uint32 to uint16 and applies ps.volume.
	var finalShift int
	if v := ps.volume; v <= MinVolume {
		finalShift = 32 // silence
	} else {
		finalShift = (16 + (MaxVolume - v)) & 0x1f
	}

	for i := range buf {
		ps.osc1.Frequency = ps.note.Pitch().Frequency()
		ps.osc1.Step()

		output := ps.osc1.Phase

		// Convert int32 to uint16 and apply ps.volume
		buf[i] = uint16(int16(output>>finalShift)) + 0x8000
	}

	return nil
}
