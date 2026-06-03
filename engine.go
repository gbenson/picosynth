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
)

type Engine struct {
}

// Run is the main entry point of the firmware.
func (ps *Engine) Run() error {
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
	for i := range buf {
		buf[i] = uint16(i << 8)
	}
	return nil
}
