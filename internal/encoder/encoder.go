package encoder

import "gbenson.net/go/picosynth/internal/hw"

type Encoder struct {
	device  hw.QuadratureDevice
	lastPos int
}

// Open opens the specified rotary encoder.
func Open(n int) (*Encoder, error) {
	qd, err := open(n)
	if err != nil {
		return nil, err
	}
	e := &Encoder{device: qd}
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
