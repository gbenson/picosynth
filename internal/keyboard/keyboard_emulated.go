//go:build !pico

package keyboard

import "gbenson.net/go/picosynth/internal/hw/emulator"

func init() {
	m := emulator.NewKeySwitchMatrix()
	Matrix.Inputs = m.Inputs()
	Matrix.Outputs = m.Outputs()
}
