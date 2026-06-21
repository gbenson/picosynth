// Package hw defines the hardware abstraction layer.
package hw

import "io"

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
