//go:build !rp2350

package ksm

import "gbenson.net/go/picosynth/internal/hw/machine"

// configurePin configures a pin.
func configurePin(p machine.Pin, c machine.PinConfig) {
	p.Configure(c)
}

// Read reads the pin value.
func Read(p machine.Pin) bool {
	return p.Get()
}
