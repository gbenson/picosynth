package emulator

import "github.com/veandco/go-sdl2/sdl"

var (
	Red    = Color(216, 48, 32)
	Orange = Color(204, 100, 4)
	Yellow = Color(200, 148, 0)
	Green  = Color(0, 130, 44)
	Purple = Color(148, 48, 180)
	Black  = Gray(0)
	White  = Gray(255)
	Pink   = Color(255, 20, 147)
)

func Gray(v uint8) sdl.Color {
	return Color(v, v, v)
}

func Color(r, g, b uint8) sdl.Color {
	return sdl.Color{R: r, G: g, B: b, A: 255}
}
