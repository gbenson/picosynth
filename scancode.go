package picosynth

import "gbenson.net/go/picosynth/ksm"

type Scancode = ksm.Scancode

// noteFromScancode returns the MIDI note of the key encoded by
// this scancode, or NoNote if this scancode does not encode a
// musical note.
func noteFromScancode(sc Scancode) Note {
	if sc > ksm.KeyC6 {
		return NoNote
	}
	return Note(sc-ksm.KeyC4) + noteC4
}
