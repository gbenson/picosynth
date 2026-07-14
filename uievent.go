package picosynth

import "gbenson.net/go/picosynth/internal/fmt"

type UIEvent uint

const (
	// GenericActivityEvent is transmitted to indicate user activity
	// other than those described by more specific types below.  Its
	// main purpose is to prevent the screensaver from activating.
	GenericActivityEvent UIEvent = iota

	uiEventDataBits UIEvent = 24
	uiEventDataMask UIEvent = (1 << uiEventDataBits) - 1
	uiEventTypeMask UIEvent = ^uiEventDataMask

	ButtonPressEventType UIEvent = 1 << (uiEventDataBits + iota)
	LongPressEventType
	EncoderMovedEventType
)

// Type returns a type of the event.
func (e UIEvent) Type() UIEvent {
	return e & uiEventTypeMask
}

// Scancode returns the scancode of ButtonPressEvent and LongPressEvent.
func (e UIEvent) Scancode() Scancode {
	return Scancode(e & uiEventDataMask)
}

// Scancode returns the delta of EncoderMovedEvent.
func (e UIEvent) Delta() int {
	return int(int16(e & uiEventDataMask))
}

// ButtonPressEvent is transmitted whenever a button is released after
// having been held down for less than LongPressTimeout.
func ButtonPressEvent(sc Scancode) UIEvent {
	return ButtonPressEventType | UIEvent(sc)
}

// LongPressEvent is transmitted whenever a button is released after
// having been held down for LongPressTimeout or longer.
func LongPressEvent(sc Scancode) UIEvent {
	return LongPressEventType | UIEvent(sc)
}

// EncoderMovedEvent is transmitted whenever the rotary encoder is moved.
func EncoderMovedEvent(delta int) UIEvent {
	return EncoderMovedEventType | (UIEvent(delta) & uiEventDataMask)
}

// String implements [fmt.Stringer].
func (e UIEvent) String() string {
	return fmt.Hex32Stringer("UIEvent", e)
}
