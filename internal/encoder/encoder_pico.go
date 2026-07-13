//go:build pico

package encoder

import (
	"machine"

	"tinygo.org/x/drivers/encoders"
)

var pinss = [][]machine.Pin{
	[]machine.Pin{machine.GP0, machine.GP1},
}

func open(n int) (*encoders.QuadratureDevice, error) {
	if n < 0 || n >= len(pinss) {
		panic("invalid encoder")
	}

	pins := pinss[n]
	qd := encoders.NewQuadratureViaInterrupt(pins[0], pins[1])

	qd.Configure(encoders.QuadratureConfig{
		Precision: 4,
	})

	return qd, nil
}
