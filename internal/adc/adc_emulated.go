//go:build !pico

package adc

import emulator "gbenson.net/go/picosynth/x/emulator/sdl"

var open = emulator.OpenADC
