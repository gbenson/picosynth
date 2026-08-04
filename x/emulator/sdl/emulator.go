package sdl

import (
	"unsafe"

	"gbenson.net/go/picosynth/internal/hw/machine"
	"gbenson.net/go/picosynth/internal/hw/pio/piolib"
	encoders "gbenson.net/go/picosynth/x/emulator"
)

type Emulator struct {
	i2c I2C
	i2s AudioDevice
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

// I2SSetSampleFrequency implements [emulator.I2SSampleFrequencySetter].
func (e *Emulator) I2SSetSampleFrequency(i2s *piolib.I2S, freq uint32) error {
	e.i2s.sampleRate = int(freq)
	return nil
}

// I2SWriteMono implements [emulator.I2SMonoWriter],
func (e *Emulator) I2SWriteMono(i2s *piolib.I2S, buf []uint16) (int, error) {
	ptr := unsafe.Pointer(unsafe.SliceData(buf))
	return len(buf), e.i2s.WriteMono(unsafe.Slice((*int16)(ptr), len(buf)))
}
