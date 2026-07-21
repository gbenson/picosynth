//go:build pico

package keyboard

import "machine"

func init() {
	// Pins we energize.
	Matrix.Outputs = []machine.Pin{
		machine.GP16, // KO0
		machine.GP17, // KO1
		machine.GP18, // KO2
		machine.GP19, // KO3
		machine.GP20, // KO4
		machine.GP21, // KO5
		machine.GP22, // KO6
	}
	for _, pin := range Matrix.Outputs {
		pin.Configure(
			machine.PinConfig{
				Mode: machine.PinOutput,
			},
		)
	}

	// Pins we read.
	Matrix.Inputs = []machine.Pin{
		machine.GP4,  // KI0
		machine.GP5,  // KI1
		machine.GP6,  // KI2
		machine.GP7,  // KI3
		machine.GP12, // KI4
		machine.GP13, // KI5
		machine.GP14, // KI6
		machine.GP15, // KI7
	}
	for _, pin := range Matrix.Inputs {
		pin.Configure(
			machine.PinConfig{
				Mode: machine.PinInputPulldown,
			},
		)
	}
}
