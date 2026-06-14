//go:build pico

package audio

import (
	"machine"
	"unsafe"

	"github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"
)

const (
	i2sDataPin  = machine.GPIO9  // physical pin 12
	i2sClockPin = machine.GPIO10 // physical pin 14
)

type device struct {
	sm  pio.StateMachine
	i2s *piolib.I2S
}

func open(sampleRate int) (Device, error) {
	sm, err := pio.PIO0.ClaimStateMachine()
	if err != nil {
		return nil, err
	}

	d, err := newDevice(sm, sampleRate)
	if err != nil {
		defer sm.Unclaim()
	}

	return d, err
}

func newDevice(sm pio.StateMachine, sampleRate int) (*device, error) {
	i2s, err := piolib.NewI2S(sm, i2sDataPin, i2sClockPin)
	if err != nil {
		return nil, err
	}

	if err := i2s.SetSampleFrequency(uint32(sampleRate)); err != nil {
		return nil, err
	}

	return &device{sm: sm, i2s: i2s}, nil
}

// WriteMono implements [Device].
func (d *device) WriteMono(buf []int16) error {
	ptr := unsafe.Pointer(unsafe.SliceData(buf))
	data := unsafe.Slice((*uint16)(ptr), len(buf))

	_, err := d.i2s.WriteMono(data)
	return err
}

// Close implements [io.Closer].
func (d *device) Close() error {
	d.sm.Unclaim()
	return nil
}
