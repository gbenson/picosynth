package emulator

import "github.com/veandco/go-sdl2/sdl"

type SoftKey struct {
	Legend string
	Offset int
	Color  Color
	Wide   bool
	Code   sdl.Keycode
	KI, KO int
}
