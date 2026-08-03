package picosynth

import "gbenson.net/go/picosynth/internal/display"

type Page interface {
	// OnInit is called exactly once, when the page is added to the
	// UI, before any other event handlers are called.
	OnInit(ui *Picosynth)

	// OnFocus is called whenever the page becomes focussed.
	OnFocus(ui *Picosynth)

	// OnButtonPress is called whenever a button is released after
	// having been held down.  Returns true if the event has been
	// handled, false if processing should continue.
	OnButtonPress(ui *Picosynth, sc Scancode, longpress bool) bool

	// OnEncoderMove is called whenever the rotary encoder is moved.
	OnEncoderMove(ui *Picosynth, delta int)

	// Render renders the screen.  It's called at FrameRate whenever
	// the page is visible, with fills being the number of times the
	// audio buffer has been filled since startup.
	Render(d *display.Display, fills uint32)
}
