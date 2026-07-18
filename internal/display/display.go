package display

import (
	"sync/atomic"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ssd1306"

	"gbenson.net/go/picosynth/internal/microfont"
)

const (
	Width   = 128
	Height  = 32
	Address = ssd1306.Address_128_32

	bufsiz = Width * Height / 8
)

type Display struct {
	bus    drivers.I2C
	device *ssd1306.Device
	buffer [bufsiz]byte
	buf    []byte
	page2  []byte // scratch space

	sleeping atomic.Bool
}

// Open opens the display.
func Open() (*Display, error) {
	d := &Display{}

	d.buf = d.buffer[:]
	d.page2 = d.buf[Width : Width*2]

	if bus, err := openBus(); err != nil {
		return nil, err
	} else {
		d.bus = bus
	}

	d.device = ssd1306.NewI2C(d.bus)
	d.device.Configure(ssd1306.Config{
		Width:   Width,
		Height:  Height,
		Address: Address,
	})

	return d, nil
}

// Sleeping reports whether the display is sleeping.
func (d *Display) Sleeping() bool {
	return d.sleeping.Load()
}

// Sleep turns off the display. Sending any other immediate-effect
// command afterward turns it back on again.  Sleep has immediate
// effect, it does not require a call to [Sync].
func (d *Display) Sleep() {
	if !d.sleeping.Swap(true) {
		d.device.Command(ssd1306.DISPLAYOFF)
	}
}

// Sync updates the display with any changes since its last call.
func (d *Display) Sync() {
	if d.sleeping.Swap(false) {
		d.device.Command(ssd1306.DISPLAYON)
	}
	if err := d.sync(); err != nil {
		println("d.sync:error:", err)
	}
}

// sync updates the display with any changes since its last call.
func (d *Display) sync() error {
	if err := d.device.SetBuffer(d.buf); err != nil {
		return err
	}
	return d.device.Display()
}

// Clear clears the buffer.  Clear is asynchronous, requiring a call
// to [Sync] to have visible effect.
func (d *Display) Clear() {
	dst := d.buf
	for i := range dst {
		dst[i] = 0
	}
}

// Box draws a filled rectangle.  Box is asynchronous, requiring a
// call to [Sync] to have visible effect.
func (d *Display) Box(x, y, w, h int32) {
	dst := d.buf

	maskBit := uint8(1 << (y & 7))
	var mask uint8
	for _ = range h {
		mask |= maskBit
		maskBit <<= 1
	}

	rowStart := (y/8)*Width + x
	rowLimit := rowStart + w
	for i := rowStart; i < rowLimit; i++ {
		dst[i] |= mask
	}
}

// Text displays the given text, expanded to fill the entire screen.
// Text is asynchronous, requiring a call to [Sync] to have visible
// effect.
func (d *Display) Text(s string) {
	// pass 1: unpack glyphs into the first page of the buffer.
	dst := d.buf
	width := microfont.Render(dst, s)
	if width == 0 {
		for i := range dst {
			dst[i] = 0
		}
		return
	}

	// pass 2: horizontal expansion, using the second page of the
	// buffer as scratch space.
	src := dst
	dst = d.page2

	// bresenham
	dx := Width // of display
	dy := width // of rendered text
	D := 2*dy - dx
	y := 0

	// hack (that I think it makes it follow the left edges of the
	// "pixel" rather than its center).
	D *= 2

	for x := range Width {
		dst[x] = src[y]
		if D > 0 {
			y++
			D += 2 * (dy - dx)
		} else {
			D += 2 * dy
		}
	}

	// pass 3: vertical expansion.
	src = dst
	dst = d.buf
	for i, v := range src {
		u := verticals32[v]
		dst[i] = uint8(u & 255)

		u >>= 8
		i += Width
		dst[i] = uint8(u & 255)

		u >>= 8
		i += Width
		dst[i] = uint8(u & 255)

		u >>= 8
		i += Width
		dst[i] = uint8(u & 255)
	}
}

// TextAt displays the given text with its top left corner at x, y,
// roughly h pixels high. TextAt is asynchronous, requiring a call
// to [Sync] to have visible effect.
func (d *Display) TextAt(x, y, h int32, s string) {
	dst := d.buf[(y/8)*Width+x:]
	width := microfont.Render(dst, s)
	if h < 9 {
		for i := range width {
			dst[i] <<= 1
		}
		return
	}

	for sx := width - 1; sx >= 0; sx-- {
		dx := sx * 2
		u := verticals16[dst[sx]]
		v := uint8(u & 255)
		dst[dx] = v
		dst[dx+1] = v

		u >>= 8
		v = uint8(u & 255)
		dx += Width
		dst[dx] = v
		dst[dx+1] = v
	}
}

var verticals16 [32]uint32
var verticals32 [32]uint32

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
	for i := range verticals32 {
		v := i
		var result uint32
		for _ = range 6 {
			result <<= 6
			if (v & 32) != 0 {
				result |= 0x3f
			}
			v <<= 1
		}
		verticals32[i] = result
	}
}
