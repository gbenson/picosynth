package picosynth

import (
	"machine"
	"time"

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
	// MaxLatency.
	bufferFrames := int(SampleRate * MaxLatency / time.Second)
	buffer := make([]uint16, bufferFrames)

	// sine wave data
	var sine []int16 = []int16{
		6392, 12539, 18204, 23169, 27244, 30272, 32137, 32767, 32137,
		30272, 27244, 23169, 18204, 12539, 6392, 0, -6393, -12540,
		-18205, -23170, -27245, -30273, -32138, -32767, -32138, -30273, -27245,
		-23170, -18205, -12540, -6393, -1,
	}
	for i := range buffer {
		buffer[i] = uint16(sine[i%len(sine)])
	}

	for {
		for i := 0; i < 50; i++ {
			if _, err := i2s.WriteMono(buffer); err != nil {
				return err
			}
		}

		time.Sleep(time.Millisecond * 500)
	}
}
