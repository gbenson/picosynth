package keyboard

import "gbenson.net/go/picosynth/internal/hw"

type (
	KeySwitchMatrix = hw.KeySwitchMatrix
	Row             = hw.Row
	Column          = hw.Column
)

func Open() (KeySwitchMatrix, error) {
	return open()
}
