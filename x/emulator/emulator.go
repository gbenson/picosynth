// Package emulator provides a configurable hardware simulation layer
// for emulation and testcases.  It isn't used in bare metal builds.
package emulator

type Emulator any

var emulator Emulator

// Install installs an Emulator.
func Install(e Emulator) {
	emulator = e
}
