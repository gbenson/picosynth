package sdl

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"

	"gbenson.net/go/microfont"
)

type Draw struct {
	r *sdl.Renderer
}

func (d *Draw) Box(x, y, w, h int32, c sdl.Color) {
	gfx.BoxColor(d.r, x, y, x+w-1, y+h-1, c)
}

func (d *Draw) Disk(x, y, r int32, c sdl.Color) {
	gfx.AACircleColor(d.r, x, y, r, c)
	gfx.FilledCircleColor(d.r, x, y, r, c)
}

func (d *Draw) RoundedBox(x, y, w, h, r int32, c sdl.Color) {
	x0 := x
	y0 := y
	x1 := x0 + w - 1
	y1 := y0 + h - 1

	cx0 := x0 + r
	cy0 := y0 + r
	cx1 := x1 - r
	cy1 := y1 - r

	d.Disk(cx0, cy0, r, c)
	d.Disk(cx1, cy0, r, c)
	d.Disk(cx0, cy1, r, c)
	d.Disk(cx1, cy1, r, c)

	by0 := cy0 + r
	by1 := cy1 - r

	gfx.BoxColor(d.r, cx0, y0, cx1, by0, c)
	gfx.BoxColor(d.r, x0, cy0, x1, cy1, c)
	gfx.BoxColor(d.r, cx0, by1, cx1, y1, c)
}

func (d *Draw) Text(x, y int32, s string, c sdl.Color) {
	face := microfont.Face04B08
	ascent := face.Metrics().Ascent.Round()

	fd := font.Drawer{
		Face: face,
		Dot:  fixed.P(int(x), int(y)+ascent),
		Src:  &image.Uniform{C: color.NRGBA(c)},
		Dst:  &scalerImage{&drawImage{d}, int(x), int(y), 3},
	}
	fd.DrawString(s)
}
