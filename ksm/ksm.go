// Package ksm defines the key/switch matrix.
package ksm

import "gbenson.net/go/picosynth/x/machine"

// Scancode identifies a key in the key/switch matrix.
type Scancode uint

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
