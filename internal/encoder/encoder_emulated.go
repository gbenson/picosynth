//go:build !pico

package encoder

import emulator "gbenson.net/go/picosynth/x/emulator/sdl"

var open = emulator.OpenEncoder
