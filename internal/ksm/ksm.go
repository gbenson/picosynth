package ksm

import "gbenson.net/go/picosynth/internal/hw/machine"

// Configure configures the key/switch matrix.
func Configure() error {
	for _, pin := range Outputs {
		pin.Configure(
			machine.PinConfig{
				Mode: machine.PinOutput,
			},
		)
	}

	for _, pin := range Inputs {
		pin.Configure(
			machine.PinConfig{
				Mode: machine.PinInputPulldown,
			},
		)
	}

	return nil
}
