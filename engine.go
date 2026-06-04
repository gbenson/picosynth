package picosynth

import (
	"math/bits"
	"time"

	"gbenson.net/go/picosynth/internal/audio"
	"gbenson.net/go/picosynth/internal/dbuf"
)

const (
	SampleRate = 48000
	MaxLatency = 10 * time.Millisecond

	MinVolume     = 0
	MaxVolume     = 10
	InitialVolume = 7
)

type Engine struct {
	note Note
	osc1 PhaseAccumulator

	volume int
}

func (ps *Engine) init() {
	ps.volume = InitialVolume

	go func() {
		for {
			for _, ps.note = range []Note{48, 52, 55, 59, 60, 59, 55, 52} {
				time.Sleep(180 * time.Millisecond)
			}
		}
	}()
}

// Run is the main entry point of the firmware.
func (ps *Engine) Run() error {
	ps.init()

	out, err := audio.Open(SampleRate)
	if err != nil {
		return err
	}
	defer out.Close()

	// Calculate how many frames we can buffer without exceeding
	// MaxLatency, round down to a power of two, then allocate a
	// buffer *twice* that size for double buffering.
	bufferFrames := int(SampleRate * MaxLatency / time.Second)
	bufferFrames = 1 << (bits.Len(uint(bufferFrames)) - 1)
	buffers := make([]uint16, bufferFrames*2)

	// player plays A then waits for B
	// filler fills B then waits for A
	fillMe := make(chan []uint16, 2)
	playMe := make(chan []uint16)
	errors := make(chan error, 2)

	filler := dbuf.Worker{
		Name: "filler",
		InC:  fillMe,
		OutC: playMe,
		ErrC: errors,
		Func: ps.Fill,
	}

	player := dbuf.Worker{
		Name: "player",
		InC:  playMe,
		OutC: fillMe,
		ErrC: errors,
		Func: out.WriteMono,
	}

	fillMe <- buffers[:bufferFrames]
	fillMe <- buffers[bufferFrames:]

	filler.Start()
	player.Start()

	return <-errors
}

// Fill generates samples into the supplied buffer.
func (ps *Engine) Fill(buf []uint16) error {
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

		// Convert uint32 to int32
		output := int32(ps.osc1.Phase - 0x80000000)

		// Convert int32 to uint16 and apply ps.volume
		buf[i] = uint16(int16(output>>finalShift)) + 0x8000
	}

	return nil
}
