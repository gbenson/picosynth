//go:build pico

package piolib

import (
	"gbenson.net/go/picosynth/internal/hw/machine"
	"gbenson.net/go/picosynth/internal/hw/pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
)

type (
	I2S = piolib.I2S
)

// NewI2S creates a new I2S peripheral using the given PIO state machine.
func NewI2S(sm pio.StateMachine, data, clockAndNext machine.Pin) (*I2S, error) {
	return piolib.NewI2S(sm, data, clockAndNext)
}
