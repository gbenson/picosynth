//go:build !pico

package encoder

import "gbenson.net/go/picosynth/internal/hw/emulator"

var open = emulator.OpenEncoder
