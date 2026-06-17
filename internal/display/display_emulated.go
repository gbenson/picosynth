//go:build !pico

package display

import "gbenson.net/go/picosynth/internal/hw/emulator"

var openBus = emulator.OpenDisplayBus
