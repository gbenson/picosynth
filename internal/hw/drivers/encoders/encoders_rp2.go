//go:build rp2040 || rp2350

package encoders

import (
	"tinygo.org/x/drivers/encoders"

	"gbenson.net/go/picosynth/x/machine"
)

type (
	QuadratureConfig = encoders.QuadratureConfig
	QuadratureDevice = encoders.QuadratureDevice
)

// NewQuadratureViaInterrupt returns a rotary encoder device that uses GPIO
// interrupts and a lookup table to keep track of quadrature state changes.
func NewQuadratureViaInterrupt(pinA, pinB machine.Pin) *QuadratureDevice {
	return encoders.NewQuadratureViaInterrupt(pinA, pinB)
}
