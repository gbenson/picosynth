package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"periph.io/x/devices/v3/ssd1306/image1bit"

	"gbenson.net/go/picosynth/internal/display"
	"gbenson.net/go/picosynth/internal/microfont"
)

const display_BufSize = display.Width * display.Height / 8 // XXX expose?

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func _main() error {
	img := image1bit.NewVerticalLSB(image.Rect(0, 0, display.Width, display.Height))
	var pages [4][]byte

	for i := range pages {
		start := i * display.Width
		limit := start + display.Width
		pages[i] = img.Pix[start:limit]
	}

	tmp := pages[2]
	width := microfont.Render(tmp, "do not turn off or reset")
	offset := (display.Width - width) / 2
	dst := pages[3]
	for sx := range width {
		dx := sx + offset
		dst[dx] = tmp[sx] << 3
	}

	width = microfont.Render(tmp, "updating")
	offset = (display.Width - width*2) / 2
	dst1 := pages[0]
	dst2 := pages[1]
	for sx := width - 1; sx >= 0; sx-- {
		dx := sx*2 + offset
		u := verticals16[tmp[sx]] >> 2
		v := uint8(u & 255)
		dst1[dx] = v
		dst1[dx+1] = v
		u >>= 8

		v = uint8(u & 255)
		dst2[dx] = v
		dst2[dx+1] = v
	}

	width = microfont.Render(tmp, "firmware")
	offset = (display.Width - width*2) / 2
	dst1 = pages[1]
	dst2 = pages[2]
	for sx := width - 1; sx >= 0; sx-- {
		dx := sx*2 + offset
		u := verticals16[tmp[sx]] << 3
		v := uint8(u & 255)
		dst1[dx] |= v
		dst1[dx+1] |= v
		u >>= 8

		v = uint8(u & 255)
		dst2[dx] = v
		dst2[dx+1] = v
	}

	for sx := range offset {
		tmp[sx] = 0
	}
	for sx := width*2 + offset; sx < len(tmp); sx++ {
		tmp[sx] = 0
	}

	fp, err := os.Create("firmware-update-screen.png")
	if err != nil {
		return err
	}
	defer fp.Close()
	return png.Encode(fp, img)
}

var verticals16 [32]uint32

func init() {
	for i := range verticals16 {
		v := i
		var result uint32
		for _ = range 6 {
			if (v & 32) != 0 {
				result |= 0x03
			}
			v <<= 1
			result <<= 2
		}
		verticals16[i] = result
	}
}
