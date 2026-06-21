package keyboard

import "gbenson.net/go/picosynth/internal/hw"

type (
	KeySwitchMatrix = hw.KeySwitchMatrix
	InputPin        = hw.InputPin
	OutputPin       = hw.OutputPin
)

func Open() (KeySwitchMatrix, error) {
	return open()
}
