package display

import (
	"sync/atomic"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ssd1306"

	"gbenson.net/go/picosynth/internal/hw/machine"
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
	buf    []byte

	sleeping atomic.Bool
}

// Open opens the display.
func Open() (*Display, error) {
	i2c := machine.I2C1
	err := i2c.Configure(machine.I2CConfig{
		SDA: machine.I2C1_SDA_PIN,
		SCL: machine.I2C1_SCL_PIN,
	})
	if err != nil {
		return nil, err
	}

	d := &Display{bus: i2c}

	d.device = ssd1306.NewI2C(d.bus)
	d.device.Configure(ssd1306.Config{
		Width:   Width,
		Height:  Height,
		Address: Address,
	})

	d.buf = d.device.GetBuffer()

	return d, nil
}

// Sleeping reports whether the display is sleeping.
func (d *Display) Sleeping() bool {
	return d.sleeping.Load()
}

// Sleep turns off the display until the next [Sync].
func (d *Display) Sleep() {
	if !d.sleeping.Swap(true) {
		d.device.Command(ssd1306.DISPLAYOFF)
	}
}

// Sync updates the display with any rendering performed since the
// previous call.  If the display was sleeping, Sync will wake it.
func (d *Display) Sync() {
	if d.sleeping.Swap(false) {
		d.device.Command(ssd1306.DISPLAYON)
	}
	if err := d.device.Display(); err != nil {
		println("d.sync:error:", err)
	}
}

// Clear clears the display buffer.
func (d *Display) Clear() {
	dst := d.buf
	for i := range dst {
		dst[i] = 0
	}
}

// Box draws a filled rectangle.
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
func (d *Display) Text(s string) {
	// pass 1: unpack glyphs into the first page of the buffer.
	dst := d.buf
	width := microfont.Render(dst, s)
	if width == 0 {
		d.Clear()
		return
	}

	// pass 2: horizontal expansion, using the second page of the
	// buffer as scratch space.
	src := dst
	dst = dst[Width : Width*2]

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
// roughly h pixels high.
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
