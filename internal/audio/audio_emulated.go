//go:build !pico

package audio

import emulator "gbenson.net/go/picosynth/x/emulator/sdl"

var open = emulator.OpenAudio
