package emulator

import (
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

//go:generate python3 make-keyboard-layout.py

type Keyboard struct {
	mu     sync.Mutex
	matrix [8][7]bool
	keymap Keymap
}

type Keymap map[sdl.Keycode]*SoftKey

var keyboard Keyboard

func (kb *Keyboard) matrixSize() (outs, ins int) {
	return len(kb.matrix[0]), len(kb.matrix)
}

func (kb *Keyboard) ensureKeymap() {
	if kb.keymap != nil {
		return // already built
	}

	kb.keymap = make(Keymap)
	for _, row := range keyboardLayout {
		for _, key := range row {
			kb.keymap[key.Code] = &key
		}
	}
}

func (kb *Keyboard) handleEvent(e *sdl.KeyboardEvent) {
	if e.Repeat != 0 {
		return
	}

	kb.mu.Lock()
	defer kb.mu.Unlock()
	kb.ensureKeymap()

	key := kb.keymap[e.Keysym.Sym]
	if key == nil {
		return
	}

	var held bool
	switch e.Type {
	case sdl.KEYDOWN:
		held = true
	case sdl.KEYUP:
	default:
		return
	}

	kb.matrix[key.KI][key.KO] = held
}
