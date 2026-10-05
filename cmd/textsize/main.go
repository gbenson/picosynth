package main

import (
	"fmt"
	"os"
	"strings"

	"gbenson.net/go/picosynth/internal/display"
	"gbenson.net/go/picosynth/internal/microfont"
)

func main() {
	s := strings.Join(os.Args[1:], " ")
	fmt.Printf("%q\n", s)
	buf := make([]byte, display.Width)
	width := microfont.Render(buf, s)
	fmt.Printf("%d px unscaled, %d px doubled\n", width, width*2)
}
