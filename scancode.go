package picosynth

type Scancode uint

const (
	// KO0 ×
	KeyF3  Scancode = iota // KI0
	KeyGb3                 // KI1
	KeyG3                  // KI2
	KeyAb3                 // KI3
	KeyA3                  // KI4
	KeyBb3                 // KI5
	KeyB3                  // KI6
	KeyC4                  // KI7

	// KO1 ×
	KeyDb4 // KI0
	KeyD4  // KI1
	KeyEb4 // KI2
	KeyE4  // KI3
	KeyF4  // KI4
	KeyGb4 // KI5
	KeyG4  // KI6
	KeyAb4 // KI7

	// KO2 ×
	KeyA4  // KI0
	KeyBb4 // KI1
	KeyB4  // KI2
	KeyC5  // KI3
	KeyDb5 // KI4
	KeyD5  // KI5
	KeyEb5 // KI6
	KeyE5  // KI7

	// KO3 ×
	KeyF5  // KI0
	KeyGb5 // KI1
	KeyG5  // KI2
	KeyAb5 // KI3
	KeyA5  // KI4
	KeyBb5 // KI5
	KeyB5  // KI6
	KeyC6  // KI7

	// KO4 ×
	ButtonKeyboard // KI0
	ButtonWind     // KI1
	ButtonString   // KI2
	ButtonSynth    // KI3
	ButtonSE       // KI4
	ButtonTempoUp  // KI5
	ButtonVolumeUp // KI6
	ButtonToneEdit // KI7

	// KO5 ×
	undefKO5KI0      // KI0
	undefKO5KI1      // KI1
	ButtonRhythm     // KI2
	ButtonAccomp     // KI3
	ButtonFunny      // KI4
	ButtonStop       // KI5
	ButtonTempoDown  // KI6
	ButtonVolumeDown // KI7

	// KO6 ×
	undefKO6KI0 // KI0
	undefKO6KI1 // KI1
	undefKO6KI2 // KI2
	undefKO6KI3 // KI3
	undefKO6KI4 // KI4
	undefKO6KI5 // KI5
	undefKO6KI6 // KI6
	ButtonSong  // KI7
)

// Note returns the MIDI note of the key encoded by this scancode,
// or NoNote if this scancode does not encode a musical note.
func (sc Scancode) Note() Note {
	if sc > KeyC6 {
		return NoNote
	}
	return Note(sc-KeyC4) + noteC4
}
