package emulator

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/veandco/go-sdl2/sdl"
)

var started atomic.Bool

func ensureStarted() error {
	if started.Swap(true) {
		return nil // already started
	}

	if err := onceInitSDL(); err != nil {
		return err
	}

	const scale = 2
	const width = 128 // XXX
	const height = 32 // XXX
	window, err := sdl.CreateWindow(
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

	renderer, err := sdl.CreateRenderer(
		window,
		-1,
		sdl.RENDERER_ACCELERATED,
	)
	if err != nil {
		return err
	}

	if err = renderer.SetScale(scale, scale); err != nil {
		return err
	}

	go func() {
		const TargetFrameRate = 20
		const TargetFrameTicks = 1000 / TargetFrameRate // ms/frame

		for {
			startTime := sdl.GetTicks64()

			for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
				handleEvent(e)
			}

			//keyboard.Render(renderer)
			display.Render(renderer)
			renderer.Present()

			loopTicks := uint32(sdl.GetTicks64() - startTime)
			if loopTicks >= TargetFrameTicks {
				continue
			}
			sdl.Delay(TargetFrameTicks - loopTicks)
		}
	}()

	return nil
}

// onceInitSDL ensures [sdl.Init] is called exactly once.
// It is safe to be called concurrently.
var onceInitSDL = sync.OnceValue(func() error {
	return sdl.Init(sdl.INIT_EVERYTHING)
})

func handleEvent(event sdl.Event) {
	switch event.(type) {
	case *sdl.QuitEvent:
		os.Exit(0)
	}
}
