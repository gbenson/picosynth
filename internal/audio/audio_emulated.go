//go:build !pico

package audio

import "gbenson.net/go/picosynth/internal/hw/emulator"

var open = emulator.OpenAudio
