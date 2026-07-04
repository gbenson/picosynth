//go:build !pico

package adc

import "gbenson.net/go/picosynth/internal/hw/emulator"

var open = emulator.OpenADC
