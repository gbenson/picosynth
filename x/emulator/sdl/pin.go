package sdl

import (
	"slices"
	"sync/atomic"

	kbd "gbenson.net/go/picosynth/internal/keyboard"
	machine "gbenson.net/go/picosynth/x/emulator"
)

type Pin struct {
	pin   machine.Pin
	value atomic.Bool
}

var pins [256]Pin

func init() {
	for i := range pins {
		pins[i].pin = machine.Pin(i)
	}
}

func (p *Pin) Set(value bool) {
	p.value.Store(value)
}

func (p *Pin) Get() bool {
	ksm := kbd.Matrix()
	col := slices.Index(ksm.Inputs, p.pin)
	if col < 0 {
		panic("invalid column")
	}

	keyboard.mu.Lock()
	defer keyboard.mu.Unlock()

	for row, held := range keyboard.matrix[col] {
		if held && pins[ksm.Outputs[row]].value.Load() {
			return true
		}
	}
	return false
}
