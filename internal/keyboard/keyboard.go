package keyboard

import "gbenson.net/go/picosynth/internal/hw/machine"

type SwitchMatrix struct {
	Inputs, Outputs []machine.Pin
}

// Configure configures the key/switch matrix.
func Configure() error {
	ksm := Matrix()

	for _, pin := range ksm.Outputs {
		pin.Configure(
			machine.PinConfig{
				Mode: machine.PinOutput,
			},
		)
	}

	for _, pin := range ksm.Inputs {
		pin.Configure(
			machine.PinConfig{
				Mode: machine.PinInputPulldown,
			},
		)
	}

	return nil
}
