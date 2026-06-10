package picosynth

type KeyEvent uint

// NoEvent is returned by [KeyScanner.Poll] when there are no events.
const NoEvent KeyEvent = 0

const (
	keyEventCodeMask = (1 << 8) - 1
	isKeyEvent       = 1 << 8 // so keyEvent(0, false) != NoEvent
	isKeyDownEvent   = 1 << 9
)

func keyEvent(sc Scancode, down bool) KeyEvent {
	e := KeyEvent(sc)
	if down {
		e |= isKeyDownEvent
	}
	return e | isKeyEvent
}

// Scancode identifies the key or button that changed state.
func (e KeyEvent) Scancode() Scancode {
	return Scancode(e) & keyEventCodeMask
}

// Down reports whether the stage change was from up to down.
func (e KeyEvent) Down() bool {
	return (e & isKeyDownEvent) != 0
}
