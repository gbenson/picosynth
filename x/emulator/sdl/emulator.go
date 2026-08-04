package sdl

import (
	"gbenson.net/go/picosynth/internal/hw/machine"
	encoders "gbenson.net/go/picosynth/x/emulator"
)

type Emulator struct {
	i2c I2C
}

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

// I2CTx implements [emulator.I2CTxer].
func (e *Emulator) I2CTx(i2c machine.I2C, addr uint16, w, r []byte) error {
	return e.i2c.Tx(addr, w, r)
}
