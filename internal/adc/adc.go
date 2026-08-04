package adc

import (
	"sync/atomic"

	"gbenson.net/go/picosynth/internal/hw/machine"
)

var pins = []machine.Pin{
	machine.ADC0,
	machine.ADC1,
	machine.ADC2,
	machine.ADC3,
}

var initialized atomic.Bool

// Open opens the specified ADC.
func Open(n int) (machine.ADC, error) {
	if !initialized.Swap(true) {
		machine.InitADC()
	}

	if n < 0 || n >= len(pins) {
		panic("invalid ADC")
	}

	adc := machine.ADC{Pin: pins[n]}
	return adc, adc.Configure(machine.ADCConfig{})
}
