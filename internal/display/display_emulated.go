//go:build !pico

package display

import emulator "gbenson.net/go/picosynth/x/emulator/sdl"

var openBus = emulator.OpenDisplayBus
