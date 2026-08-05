package sdl

import "github.com/veandco/go-sdl2/sdl"

type SoftKey struct {
	Legend string
	Offset int
	Color  sdl.Color
	Wide   bool
	Name   string
	Code   sdl.Keycode
	KI, KO int
}
