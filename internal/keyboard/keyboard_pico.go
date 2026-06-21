//go:build pico

package keyboard

import "machine"

// Pins we energize.
var OutputPins = []machine.Pin{
	machine.GPIO16, // KO0
	machine.GPIO17, // KO1
	machine.GPIO18, // KO2
	machine.GPIO19, // KO3
	machine.GPIO20, // KO4
	machine.GPIO21, // KO5
	machine.GPIO22, // KO6
}

// Pins we read.
var InputPins = []machine.Pin{
	machine.GPIO4,  // KI0
	machine.GPIO5,  // KI1
	machine.GPIO6,  // KI2
	machine.GPIO7,  // KI3
	machine.GPIO12, // KI4
	machine.GPIO13, // KI5
	machine.GPIO14, // KI6
	machine.GPIO15, // KI7
}

type keyboard struct {
	outputs []OutputPin
	inputs  []InputPin
}

func open() (KeySwitchMatrix, error) {
	kb := &keyboard{
		make([]OutputPin, len(OutputPins)),
		make([]InputPin, len(InputPins)),
	}
	for i, pin := range OutputPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		kb.outputs[i] = pin
	}
	for j, pin := range InputPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
		kb.inputs[j] = pin
	}
	return kb, nil
}

func (kb *keyboard) Outputs() []OutputPin {
	return kb.outputs
}

func (kb *keyboard) Inputs() []InputPin {
	return kb.inputs
}
