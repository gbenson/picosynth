package ksm

import "gbenson.net/go/picosynth/internal/hw/machine"

// Configure configures the key/switch matrix.
func Configure() error {
	for _, pin := range Outputs {
		configurePin(
			pin,
			machine.PinConfig{
				Mode: machine.PinOutput,
			},
		)
	}

	for _, pin := range Inputs {
		configurePin(
			pin,
			machine.PinConfig{
				Mode: machine.PinInputPulldown,
			},
		)
	}

	return nil
}
