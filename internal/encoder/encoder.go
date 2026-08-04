package encoder

import (
	"gbenson.net/go/picosynth/internal/hw/drivers/encoders"
	"gbenson.net/go/picosynth/internal/hw/machine"
)

type Encoder struct {
	device  *encoders.QuadratureDevice
	lastPos int
}

var encoderPins = [][]machine.Pin{
	[]machine.Pin{machine.GP0, machine.GP1},
}

// Open opens the specified rotary encoder.
func Open(n int) (*Encoder, error) {
	if n < 0 || n >= len(encoderPins) {
		panic("invalid encoder")
	}

	pins := encoderPins[n]
	d := encoders.NewQuadratureViaInterrupt(pins[0], pins[1])
	config := encoders.QuadratureConfig{Precision: 4}
	if err := d.Configure(config); err != nil {
		return nil, err
	}

	e := &Encoder{device: d}
	e.Read()

	return e, nil
}

func (e *Encoder) Read() int {
	pos := e.device.Position()
	delta := pos - e.lastPos
	if delta != 0 {
		e.lastPos = pos
	}
	return delta
}
