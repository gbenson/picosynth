//go:build pico

package keyboard

import "machine"

var colPins = []machine.Pin{
	machine.GPIO4,  // KI0
	machine.GPIO5,  // KI1
	machine.GPIO6,  // KI2
	machine.GPIO7,  // KI3
	machine.GPIO12, // KI4
	machine.GPIO13, // KI5
	machine.GPIO14, // KI6
	machine.GPIO15, // KI7
}

var rowPins = []machine.Pin{
	machine.GPIO16, // KO0
	machine.GPIO17, // KO1
	machine.GPIO18, // KO2
	machine.GPIO19, // KO3
	machine.GPIO20, // KO4
	machine.GPIO21, // KO5
	machine.GPIO22, // KO6
}

type keyboard struct {
	rows []Row
	cols []Column
}

func newKeyboard() Keyboard {
	kb := &keyboard{
		make([]Row, len(rowPins)),
		make([]Column, len(colPins)),
	}
	for i, pin := range rowPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		kb.rows[i] = pin
	}
	for j, pin := range colPins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
		kb.cols[j] = pin
	}
	return kb
}

func (kb *keyboard) Rows() []Row {
	return kb.rows
}

func (kb *keyboard) Columns() []Column {
	return kb.cols
}
