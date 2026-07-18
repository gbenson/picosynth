package picosynth

type Page interface {
	// OnInit is called exactly once, when the Page is added to the
	// UI, before any other event handlers are called.
	OnInit(ui *UI)

	// OnButton is called whenever a button is released after having
	// been held down.  Returns true if the event has been handled,
	// false if processing should continue.
	OnButton(sc Scancode, longpress bool) bool

	// OnEncoder is called whenever the rotary encoder is moved.
	OnEncoder(delta int)
}
