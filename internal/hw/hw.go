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

type KeySwitchMatrix interface {
	Outputs() []OutputPin
	Inputs() []InputPin
}

type OutputPin interface {
	Set(bool)
}

type InputPin interface {
	Get() bool
}

type QuadratureDevice interface {
	Position() int
}
