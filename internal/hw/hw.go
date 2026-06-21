// Package hw defines the hardware abstraction layer.
package hw

import "io"

type AudioDevice interface {
	io.Closer
	WriteMono(buf []int16) error
}

type KeySwitchMatrix interface {
	Rows() []Row
	Columns() []Column
}

type Row interface {
	Set(bool)
}

type Column interface {
	Get() bool
}
