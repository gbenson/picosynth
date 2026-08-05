package sdl

import (
	"sync"
	"sync/atomic"
)

type Display struct {
	mu      sync.Mutex
	width   int32
	height  int32
	buf     []byte
	blanked atomic.Bool
}

var display Display

func (d *Display) Sleep() {
	d.blanked.Store(true)
}

func (d *Display) Wake() {
	d.blanked.Store(false)
}

func (d *Display) SetBuffer(width, height int32, buf []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.width = width
	d.height = height
	d.buf = buf
}

func (d *Display) Render(draw Draw) {
	d.mu.Lock()
	defer d.mu.Unlock()

	width := d.width
	height := d.height

	margin := width / 8
	pad := margin / 4
	inset := margin + pad

	// top-left corner
	x := WindowWidth - width - inset
	y := inset

	draw.Box(0, 0, WindowWidth, inset*2+height, Gray(30))
	if d.blanked.Load() {
		return
	}
	draw.Box(x-pad, y-pad, width+2*pad, height+2*pad, Black)

	for p := int32(0); p < height/8; p++ {
		x0 := p * width
		y0 := p * 8
		for i := int32(0); i < width; i++ {
			vv := d.buf[x0+i]
			for j := range int32(8) {
				if vv&(1<<j) == 0 {
					continue
				}
				draw.Box(x+i, y+y0+j, 1, 1, White)
			}
		}
	}
}
