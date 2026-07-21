// Package hw defines the hardware abstraction layer.
package hw

import "io"

type ADC interface {
	Get() uint16
}

type AudioDevice interface {
	io.Closer
	WriteMono(buf []int16) error
}

type KeySwitchMatrix struct {
	Inputs, Outputs []Pin
}

type QuadratureDevice interface {
	Position() int
}
