package sdl

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/veandco/go-sdl2/sdl"
)

type drawImage struct {
	draw *Draw
}

// ColorModel implements [image.Image].
func (im *drawImage) ColorModel() color.Model {
	panic("drawImage.ColorModel: not implemented")
}

// Bounds implements [image.Image].
func (im *drawImage) Bounds() image.Rectangle {
	r := im.draw.r.GetViewport()
	return image.Rect(int(r.X), int(r.Y), int(r.X+r.W), int(r.Y+r.H))
}

// At implements [image.Image].
func (im *drawImage) At(x, y int) color.Color {
	return color.NRGBA{0, 200, 255, 255} // random cyan
}

// Set implements [draw.Image].
func (im *drawImage) Set(x, y int, c color.Color) {
	r, g, b, a := c.RGBA()
	k := sdl.Color{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}

	im.draw.Box(int32(x), int32(y), 1, 1, k)
}

type scalerImage struct {
	dst         draw.Image
	x, y, scale int
}

// ColorModel implements [image.Image].
func (im *scalerImage) ColorModel() color.Model {
	panic("scalerImage.ColorModel: not implemented")
}

// Bounds implements [image.Image].
func (im *scalerImage) Bounds() image.Rectangle {
	return im.dst.Bounds() // XXX wrong but who cares?
}

// At implements [image.Image].
func (im *scalerImage) At(x, y int) color.Color {
	return im.dst.At(x, y) // XXX also wrong but also who cares?
}

// Set implements [draw.Image].
func (im *scalerImage) Set(x, y int, c color.Color) {
	x0 := (x-im.x)*im.scale + im.x
	y0 := (y-im.y)*im.scale + im.y
	for y := range im.scale {
		for x := range im.scale {
			im.dst.Set(x0+x, y0+y, c)
		}
	}
}
