// Package machine exposes a subset of the TinyGo package of the same name.
//
// machine exists to support building Picosynth on both TinyGo and
// regular Go.  On microcontroller targets its contents are direct
// imports from the TinyGo package of the same name: the resulting
// binary should be byte-for-byte identical whichever is imported.
// On non-microcontroller targets its contents are sourced from
// [gbenson.net/go/picosynth/x/emulator], a configurable simulation
// layer for emulation and testcases.
package machine

//go:generate bash make-machine-emulated.sh
