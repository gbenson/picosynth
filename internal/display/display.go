package display

import (
	"sync/atomic"
	"time"

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
	cmds   chan<- Command
	serial atomic.Int32
}

// Name implements [worker].
func (d *Display) Name() string {
	return "display"
}

// Run implements [worker].
func (d *Display) Run() func() error {
	return d.run
}

func (d *Display) run() error {
	if d.cmds != nil {
		panic("already started")
	}

	cmds := make(chan Command)
	defer func() {
		// Delay closing the channel if we stop, to allow the error
		// that stopped us has time to be reported. Closing without
		// waiting when something's trying to send a command means our
		// close will panic the sender and the recovered panic will
		// mask the our (underlying!).  Stopping without closing means
		// eventual deadlock.
		go func() {
			defer close(cmds)
			time.Sleep(time.Second)
		}()
	}()
	d.cmds = cmds

	d.buf = d.buffer[:]
	d.page2 = d.buf[Width : Width*2]

	if bus, err := openBus(); err != nil {
		return err
	} else {
		d.bus = bus
	}

	d.device = ssd1306.NewI2C(d.bus)
	d.device.Configure(ssd1306.Config{
		Width:   Width,
		Height:  Height,
		Address: Address,
	})

	go func() {
		n := d.Serial()
		d.TextIfSerial(n+0, "hello")
		time.Sleep(time.Second * 4)
		d.TextIfSerial(n+1, "this")
		time.Sleep(time.Second / 2)
		d.TextIfSerial(n+2, "-= is =-")
		time.Sleep(time.Second / 2)
		d.TextIfSerial(n+3, "picosynth")
		time.Sleep(time.Second * 3)
		d.TextIfSerial(n+4, "")
	}()

	for cmd := range cmds {
		if err := d.do(cmd); err != nil {
			return err
		}
	}

	return nil
}

// Text displays the given text, expanded to fill the entire screen.
func (d *Display) Text(s string) {
	d.Do(NewTextCommand(s))
}

// Serial returns the serial number the next received command will
// be assigned.
func (d *Display) Serial() int32 {
	return d.serial.Load()
}

// TextIfSerial is like [Text], except the command is ignored unless
// [Serial] matches the supplied value when the command is executed.
// This allows scheduling interruptable sequences of commands that
// stop like magic if the user presses a button or whatever.  The
// "Hello, this is Picosynth" startup message is implemented using
// this method.
func (d *Display) TextIfSerial(n int32, s string) {
	d.Do(NewTextIfSerialCommand(n, s))
}

// Do issues a command to the display.
func (d *Display) Do(cmd Command) {
	d.cmds <- cmd
}

// do executes received commands.
func (d *Display) do(cmd Command) error {
	serial := d.serial.Add(1) - 1
	if cmd[0] == '\x1B' { // showIfSerial
		// XXX replace this mess with [Command.Decode]
		runes := []rune(cmd)
		if int32(runes[1]) != serial {
			return nil
		}
		cmd = Command(runes[2:]) // fall through into regular show command
	}

	d.renderFullscreen(string(cmd))
	if err := d.device.SetBuffer(d.buf); err != nil {
		return err
	}
	return d.device.Display()
}

// renderFullscreen renders the given text, expanded to fill the
// entire screen.  This method is the expected use-case for the
// display, for real-time feedback while playing.  It should operate
// without allocation.
func (d *Display) renderFullscreen(s string) {
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
		u := verticals[v]
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

var verticals [32]uint32

func init() {
	for i := range verticals {
		v := i
		var result uint32
		for _ = range 6 {
			result <<= 6
			if (v & 32) != 0 {
				result |= 0x3f
			}
			v <<= 1
		}
		verticals[i] = result
	}
}
