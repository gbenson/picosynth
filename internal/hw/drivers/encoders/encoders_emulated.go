//go:build !baremetal

package encoders

import (
	encoders "gbenson.net/go/picosynth/x/emulator"
	"gbenson.net/go/picosynth/x/machine"
)

type (
	QuadratureDevice = encoders.Encoder
	QuadratureConfig = encoders.EncoderConfig
)

// NewQuadratureViaInterrupt returns a rotary encoder device that uses GPIO
// interrupts and a lookup table to keep track of quadrature state changes.
func NewQuadratureViaInterrupt(pinA, pinB machine.Pin) *QuadratureDevice {
	return encoders.NewQuadratureViaInterrupt(pinA, pinB)
}
