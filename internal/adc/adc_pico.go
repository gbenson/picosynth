//go:build pico

package adc

import (
	"machine"
	"sync/atomic"
)

var pins = []machine.Pin{
	machine.ADC0, // GP26
	machine.ADC1, // GP27
	machine.ADC2, // GP28
	machine.ADC3, // GP29 (internal)
}

var initialized atomic.Bool

func open(n int) (machine.ADC, error) {
	if !initialized.Swap(true) {
		machine.InitADC()
	}

	if n < 0 || n >= len(pins) {
		panic("invalid ADC")
	}

	adc := machine.ADC{pins[n]}
	adc.Configure(machine.ADCConfig{})

	return adc, nil
}
