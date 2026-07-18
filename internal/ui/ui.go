package ui

import "gbenson.net/go/picosynth/internal/display"

type Display = display.Display

type Memory interface {
	// Load returns the contents of register n.
	Load(n int) uint32

	// Store replaces the contents of register n with v.
	Store(n int, v uint32)
}
