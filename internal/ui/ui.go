package ui

import "gbenson.net/go/picosynth/internal/display"

type Display = display.Display

type System interface {
	// Display returns the surface for rendering visuals.
	Display() *Display

	// Load returns the contents of register n.
	Load(n int) uint32

	// Store replaces the contents of register n with v.
	Store(n int, v uint32)
}
