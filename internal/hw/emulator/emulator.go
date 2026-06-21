package emulator

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	WindowWidth  = 1280
	WindowHeight = 512
)

var started atomic.Bool

func ensureStarted() error {
	if started.Swap(true) {
		return nil // already started
	}

	if err := onceInitSDL(); err != nil {
		return err
	}

	window, err := sdl.CreateWindow(
		"picosynth",
		sdl.WINDOWPOS_UNDEFINED, // X
		sdl.WINDOWPOS_UNDEFINED, // Y
		WindowWidth,
		WindowHeight,
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

	draw := Draw{renderer}

	go func() {
		const TargetFrameRate = 20
		const TargetFrameTicks = 1000 / TargetFrameRate // ms/frame

		for {
			startTime := sdl.GetTicks64()

			for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
				handleEvent(e)
			}

			keyboard.Render(draw)
			display.Render(draw)
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
	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		key := e.Keysym
		switch {
		case e.Type != sdl.KEYUP:
		case key.Sym != sdl.K_q && key.Sym != sdl.K_c:
		case (key.Mod & sdl.KMOD_CTRL) == 0:
		default:
			Exit(0)
		}
		keyboard.handleEvent(e)
	case *sdl.QuitEvent:
		Exit(0)
	}
}

func Exit(code int) {
	sdl.Quit()
	os.Exit(code)
}
