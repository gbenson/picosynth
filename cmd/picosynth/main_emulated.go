//go:build !baremetal

package main

import (
	"fmt"
	"os"

	"gbenson.net/go/picosynth"
	"gbenson.net/go/picosynth/x/emulator"
	"gbenson.net/go/picosynth/x/emulator/sdl"
)

func main() {
	emulator.Install(&sdl.Emulator{})

	if err := picosynth.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
