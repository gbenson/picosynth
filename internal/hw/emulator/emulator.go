package emulator

import (
	"sync"

	"github.com/veandco/go-sdl2/sdl"
)

// onceInitSDL ensures [sdl.Init] is called exactly once.
// It is safe to be called concurrently.
var onceInitSDL = sync.OnceValue(func() error {
	return sdl.Init(sdl.INIT_EVERYTHING)
})

var window *sdl.Window
var renderer *sdl.Renderer

func initRenderer(width, height int32) error {
	if err := onceInitSDL(); err != nil {
		return err
	}

	const scale = 2
	w, err := sdl.CreateWindow(
		"picosynth",
		sdl.WINDOWPOS_UNDEFINED, // X
		sdl.WINDOWPOS_UNDEFINED, // Y
		width*scale,
		height*scale,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return err
	}

	r, err := sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return err
	}

	if err = r.SetScale(scale, scale); err != nil {
		return err
	}

	window = w
	renderer = r

	return nil
}

func renderPixels(width, height int32, b []byte) error {
	if renderer == nil {
		if err := initRenderer(width, height); err != nil {
			return err
		}
	}

	r := renderer
	for p := int32(0); p < height/8; p++ {
		x0 := p * width
		y0 := p * 8
		for x := int32(0); x < width; x++ {
			vv := b[x0+x]
			for j := range int32(8) {
				var v uint8
				if vv&(1<<j) != 0 {
					v = 255
				}
				r.SetDrawColor(v, v, v, 255)
				r.DrawPoint(x, y0+j)
			}
		}
	}

	r.Present()

	return nil
}
