package audio

import (
	"unsafe"

	"gbenson.net/go/picosynth/internal/counts"
	"gbenson.net/go/picosynth/internal/hw/machine"
	"gbenson.net/go/picosynth/internal/hw/pio"
	"gbenson.net/go/picosynth/internal/hw/pio/piolib"
)

const (
	I2S_DATA_PIN = machine.GP9
	I2S_BCLK_PIN = machine.GP10
)

type Device struct {
	sm  pio.StateMachine
	i2s *piolib.I2S
}

// Open opens an audio output device with the specified sample rate.
func Open(sampleRate int) (*Device, error) {
	sm, err := pio.PIO0.ClaimStateMachine()
	if err != nil {
		return nil, err
	}

	i2s, err := piolib.NewI2S(sm, I2S_DATA_PIN, I2S_BCLK_PIN)
	if err != nil {
		defer sm.Unclaim()
		return nil, err
	}

	if err = i2s.SetSampleFrequency(uint32(sampleRate)); err != nil {
		defer sm.Unclaim()
		return nil, err
	}

	return &Device{sm: sm, i2s: i2s}, nil
}

// WriteMono writes a mono audio buffer to the audio output device.
func (d *Device) WriteMono(buf []int16) error {
	ptr := unsafe.Pointer(unsafe.SliceData(buf))
	data := unsafe.Slice((*uint16)(ptr), len(buf))

	counts.WriteMono.Add(1)
	_, err := d.i2s.WriteMono(data)
	counts.WriteMono.Add(1)
	return err
}

// Close implements [io.Closer].
func (d *Device) Close() error {
	d.sm.Unclaim()
	return nil
}
