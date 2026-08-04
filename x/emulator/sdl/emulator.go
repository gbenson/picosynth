package sdl

import (
	"gbenson.net/go/picosynth/internal/hw/machine"
	encoders "gbenson.net/go/picosynth/x/emulator"
)

type Emulator struct{}

// SetPin implements [emulator.PinSetter].
func (e *Emulator) SetPin(p machine.Pin, value bool) {
	pins[p].Set(value)
}

// GetPin implements [emulator.PinGetter].
func (e *Emulator) GetPin(p machine.Pin) bool {
	return pins[p].Get()
}

// GetADC implements [emulator.ADCGetter].
func (e *Emulator) GetADC(a machine.ADC) uint16 {
	return pins[a.Pin].GetADC()
}

// EncoderPosition implements [emulator.EncoderPositioner].
func (e *Emulator) EncoderPosition(enc encoders.Encoder) int {
	return encoder.Position()
}
