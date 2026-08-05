//go:build !pico

package keyboard

import emulator "gbenson.net/go/picosynth/x/emulator/sdl"

func init() {
	m := emulator.NewKeySwitchMatrix()
	Matrix.Inputs = m.Inputs()
	Matrix.Outputs = m.Outputs()
}
