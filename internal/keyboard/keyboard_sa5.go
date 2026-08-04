package keyboard

import "gbenson.net/go/picosynth/internal/hw/machine"

// GPIO pins connected to the Casio SA-5's keyboard matrix traces.
// In an unmodified SA-5 they connect to pins of the "LSI1" IC.
const (
	// LSI1's "key and switch scan signal output" pins.
	// These are the key/switch matrix pins we energize.
	KO0 = machine.GP16
	KO1 = machine.GP17
	KO2 = machine.GP18
	KO3 = machine.GP19
	KO4 = machine.GP20
	KO5 = machine.GP21
	KO6 = machine.GP22

	// LSI1's "input terminals from keys and switches."
	// These are the key/switch matrix pins we read.
	KI0 = machine.GP4
	KI1 = machine.GP5
	KI2 = machine.GP6
	KI3 = machine.GP7
	KI4 = machine.GP12
	KI5 = machine.GP13
	KI6 = machine.GP14
	KI7 = machine.GP15
)

// Casio SA-5 key/switch matrix.
var switchMatrix = SwitchMatrix{
	Outputs: []machine.Pin{KO0, KO1, KO2, KO3, KO4, KO5, KO6},
	Inputs:  []machine.Pin{KI0, KI1, KI2, KI3, KI4, KI5, KI6, KI7},
}

func Matrix() *SwitchMatrix {
	return &switchMatrix
}
