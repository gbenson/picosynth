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

	BlankTime = 30 * time.Second
)

type Display struct {
	bus     drivers.I2C
	device  *ssd1306.Device
	buffer  [bufsiz]byte
	buf     []byte
	page2   []byte // scratch space
	cmds    chan<- Command
	serial  atomic.Int32
	blanker *time.Timer
	blanked atomic.Bool
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

	d.blanker = time.AfterFunc(BlankTime, d.Sleep)

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

// Clear clears the buffer.  Clear is asynchronous, requiring a call
// to [Sync] to have visible effect.
func (d *Display) Clear() {
	d.Do(ClearCommand)
}

// Line draws a line from x1, y1 to x2, y2.  Line is asynchronous,
// requiring a call to [Sync] to have visible effect.
func (d *Display) Line(x1, y1, x2, y2 int32) {
	d.Do(NewLineCommand(x1, y1, x2, y2))
}

// Sleeping reports whether the display is sleeping.
func (d *Display) Sleeping() bool {
	return d.blanked.Load()
}

// Sleep turns off the display. Sending any other command afterward
// turns it back on again.  Sleep has immediate effect, it does not
// require a call to [Sync].
func (d *Display) Sleep() {
	d.Do(SleepCommand)
}

// KeepAlive unblanks the screen as necessary, then resets the screensaver
// timer so the display won't sleep again for the maximum interval.
func (d *Display) KeepAlive() {
	d.Do(KeepAliveCommand)
}

// Sync updates the display with any changes since its last call.
func (d *Display) Sync() {
	d.Do(SyncCommand)
}

// Text displays the given text, expanded to fill the entire screen.
// Text has immediate effect, it does not require a call to [Sync].
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

// TextAt displays the given text with its top left corner at x, y,
// roughly h pixels high. TextAt is asynchronous, requiring a call
// to [Sync] to have visible effect.
func (d *Display) TextAt(x, y, h int32, s string) {
	d.Do(NewTextAtCommand(x, y, h, s))
}

// Do issues a command to the display.
func (d *Display) Do(cmd Command) {
	d.cmds <- cmd
}

// do executes received commands.
func (d *Display) do(cmd Command) error {
	if cmd == SleepCommand {
		if !d.blanked.Swap(true) {
			d.device.Command(ssd1306.DISPLAYOFF)
		}
		return nil
	} else if d.blanked.Swap(false) {
		d.device.Command(ssd1306.DISPLAYON)
	}
	d.blanker.Reset(BlankTime)
	if cmd == KeepAliveCommand {
		return nil
	}

	// SleepCommand and KeepAliveCommand not incrementing d.serial is
	// intentional, the former so sequences can have gaps longer than
	// BlankTime without being cut off every time, and the latter so
	// sequences aren't broken by playing notes or changing the volume
	// or tempo.
	serial := d.serial.Add(1) - 1

	// XXX replace this mess with [Command.Decode] or something...
	switch cmd[0] {
	case ClearCommand[0]:
		d.clear()
		return nil

	case SyncCommand[0]:
		return d.sync()

	case '\x01': // Line
		runes := []rune(cmd)
		d.line(
			r2i(runes[1]),
			r2i(runes[2]),
			r2i(runes[3]),
			r2i(runes[4]),
		)
		return nil
	case '\x02': // TextAt
		runes := []rune(cmd)
		d.renderTextAt(
			r2i(runes[1]),
			r2i(runes[2]),
			r2i(runes[3]),
			string(runes[4:]),
		)
		return nil

	case '\x1B': // TextIfSerial
		runes := []rune(cmd)
		if r2i(runes[1]) != serial {
			return nil
		}
		cmd = Command(runes[2:]) // fall through into regular show command
	}

	d.renderFullscreen(string(cmd))
	return d.sync()
}

// clear clears the buffer.
func (d *Display) clear() {
	dst := d.buf
	for i := range dst {
		dst[i] = 0
	}
}

// Line draws a line from x1, y1 to x2, y2.
func (d *Display) line(x1, y1, x2, y2 int32) {
	if y2 != y1 {
		println("line", x1, y1, x2, y2, "not implemented")
		return
	}

	// horizontal line
	dst := d.buf
	mask := uint8(1 << (y1 & 7))
	pageStart := (y1 / 8) * Width
	limit := pageStart + x2
	for i := pageStart + x1; i < limit; i++ {
		dst[i] |= mask
	}
}

// sync updates the display with any changes since its last call.
func (d *Display) sync() error {
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

// renderTextAt renders the given text, with its top left corner at
// x, y, roughly h pixels high.
func (d *Display) renderTextAt(x, y, h int32, s string) {
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
