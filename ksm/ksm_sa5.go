package ksm

import "gbenson.net/go/picosynth/x/machine"

// GPIO pins connected to the Casio SA-5's keyboard matrix traces.
// In an unmodified SA-5 they connect to pins of the "LSI1" IC.
const (
	// LSI1's "key and switch scan signal output" pins.
	// These are the key/switch matrix pins we energize.
	KO0 = machine.GP16
	KO1 = machine.GP17
	KO2 = machine.GP18
	KO3 = machine.GP19
	KO4 = machine.GP20
	KO5 = machine.GP21
	KO6 = machine.GP22

	// LSI1's "input terminals from keys and switches."
	// These are the key/switch matrix pins we read.
	KI0 = machine.GP4
	KI1 = machine.GP5
	KI2 = machine.GP6
	KI3 = machine.GP7
	KI4 = machine.GP12
	KI5 = machine.GP13
	KI6 = machine.GP14
	KI7 = machine.GP15
)

// Casio SA-5 key/switch matrix.
var (
	Outputs = []machine.Pin{KO0, KO1, KO2, KO3, KO4, KO5, KO6}
	Inputs  = []machine.Pin{KI0, KI1, KI2, KI3, KI4, KI5, KI6, KI7}
)

// Casio SA-5 scancodes.
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

	NumScancodes
)
