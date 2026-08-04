// Package hw defines the hardware abstraction layer.
//
// hw exists to support building Picosynth on both TinyGo and regular Go.
// Its subpackages are shims.  On microcontroller targets their contents
// are direct imports from the real TinyGo packages being abstracted:
// the resulting binary should be byte-for-byte identical whichever is
// imported.  On non-microcontroller targets their contents are sourced
// from [gbenson.net/go/picosynth/x/emulator], a configurable simulation
// layer for emulation and testcases.
package hw

import "io"

type AudioDevice interface {
	io.Closer
	WriteMono(buf []int16) error
}
