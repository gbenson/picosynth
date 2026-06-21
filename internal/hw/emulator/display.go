package emulator

import (
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

type Display struct {
	mu     sync.Mutex
	width  int32
	height int32
	buf    []byte
}

var display Display

func (d *Display) SetBuffer(width, height int32, buf []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.width = width
	d.height = height
	d.buf = buf
}

func (d *Display) Render(r *sdl.Renderer) {
	d.mu.Lock()
	defer d.mu.Unlock()

	const x = 0 // XXX
	const y = 0 // XXX

	for p := int32(0); p < d.height/8; p++ {
		x0 := p * d.width
		y0 := p * 8
		for i := int32(0); i < d.width; i++ {
			vv := d.buf[x0+i]
			for j := range int32(8) {
				var v uint8
				if vv&(1<<j) != 0 {
					v = 255
				}
				r.SetDrawColor(v, v, v, 255)
				r.DrawPoint(x+i, y+y0+j)
			}
		}
	}
}
