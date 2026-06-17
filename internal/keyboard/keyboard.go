package keyboard

import "gbenson.net/go/picosynth/internal/hw"

type (
	Keyboard = hw.Keyboard
	Row      = hw.Row
	Column   = hw.Column
)

func New() Keyboard {
	return newKeyboard()
}
