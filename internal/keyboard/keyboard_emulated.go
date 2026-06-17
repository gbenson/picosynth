//go:build !pico

package keyboard

import "gbenson.net/go/picosynth/internal/hw/emulator"

var newKeyboard = emulator.NewKeyboard
