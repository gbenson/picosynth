//go:build pico

package encoder

import (
	"machine"

	"gbenson.net/go/picosynth/internal/pio"
	"gbenson.net/go/picosynth/internal/piolib"
)

var pinss = [][]machine.Pin{
	[]machine.Pin{machine.GP0, machine.GP1},
}

func open(n int) (*piolib.QuadratureDevice, error) {
	if n < 0 || n >= len(pinss) {
		panic("invalid encoder")
	}
	pins := pinss[n]

	// I2S is on PIO0, idk if they'd clash but there's two PIOs so...
	sm, err := pio.PIO1.ClaimStateMachine()
	if err != nil {
		return nil, err
	}

	// We set PinMode machine.PinInputPullup here even though
	// NewQuadratureDevice is going to set PinMode machine.PinPIOx.
	// Why?  Because machine.PinMode combines function and pull, and
	// while there isn't a PinMode to setFunc(PIOx) _and_ set a pull,
	// setting PinMode machine.PinPIOx won't change whatever pull is
	// configured, so we set machine.PinInputPullup to get the pull
	// up we want, then let NewQuadratureDevice set the function it
	// wants.
	for _, pin := range pins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	qd, err := piolib.NewQuadratureDevice(sm, pins[0], pins[1])
	if err != nil {
		defer sm.Unclaim()
		return nil, err
	}

	return qd, nil
}
