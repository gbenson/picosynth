//go:build pico

package keyboard

import (
	"time"

	"gbenson.net/go/picosynth/internal/hw/machine"
)

func init() {
	// KO6-KI6 is an empty slot in the Casio SA-5's switch matrix.
	// If connected it indicates GP17 is a MIDI in we should read
	// notes from, instead of using the switch matrix note rows.
	KI6.Configure(
		machine.PinConfig{
			Mode: machine.PinInputPulldown,
		},
	)
	KO6.Configure(
		machine.PinConfig{
			Mode: machine.PinOutput,
		},
	)
	KO6.Set(true)
	time.Sleep(time.Millisecond)

	if !KI6.Get() {
		return
	}

	MIDIIn = machine.GP17
}
