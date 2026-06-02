package picosynth

import (
	"machine"
	"math/bits"
	"time"

	"gbenson.net/go/picosynth/internal/dbuf"
	"github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
)

const (
	SampleRate = 48000
	MaxLatency = 10 * time.Millisecond

	i2sDataPin  = machine.GPIO9  // physical pin 12
	i2sClockPin = machine.GPIO10 // physical pin 14
)

type Engine struct {
}

// Run is the main entry point of the firmware.
func (ps *Engine) Run() error {
	sm, err := pio.PIO0.ClaimStateMachine()
	if err != nil {
		return err
	}
	defer sm.Unclaim()

	i2s, err := piolib.NewI2S(sm, i2sDataPin, i2sClockPin)
	if err != nil {
		return err
	}

	if err := i2s.SetSampleFrequency(SampleRate); err != nil {
		return err
	}

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
		Func: func(buf []uint16) error {
			_, err := i2s.WriteMono(buf)
			return err
		},
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
