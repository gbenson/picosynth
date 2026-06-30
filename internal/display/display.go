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

	PipelineSize = 32
)

type Display struct {
	bus    drivers.I2C
	device *ssd1306.Device
	buffer [bufsiz]byte
	buf    []byte
	page2  []byte // scratch space

	// Instruction buffer.
	instbuf [PipelineSize]Instruction

	todo chan command // queued instructions
	free chan command // currently unused instruction buffer entries

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
	if d.todo != nil {
		panic("already started")
	}

	d.todo = make(chan command, PipelineSize)
	d.free = make(chan command, PipelineSize)

	defer func() {
		// Delay closing the channel if we stop, to allow the error
		// that stopped us has time to be reported. Closing without
		// waiting when something's trying to send a command means our
		// close will panic the sender and the recovered panic will
		// mask our (underlying!) error.  Stopping without closing means
		// eventual deadlock.
		go func() {
			defer close(d.todo)
			time.Sleep(time.Second)
		}()
	}()

	// Prime the free instructions queue.
	for i := range d.instbuf {
		d.free <- command(i)
	}

	// Start counting serial numbers at 1, so 0 can be used to
	// indicate the absence of a serial number.
	d.serial.Store(1)

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

	maxErrors := 20
	for cmd := range d.todo {
		err := d.do(cmd)
		if cmd >= 0 {
			d.free <- cmd
		}
		if err == nil || maxErrors < 1 {
			continue
		}

		const prefix = "Display.do:"
		println(prefix, err)
		maxErrors--
		if maxErrors > 0 {
			continue
		}

		println(prefix, "maximum number of errors reached")
	}

	return nil
}

// acquire returns the next available instruction buffer entry.
func (d *Display) acquire(op Opcode) (*Instruction, command) {
	index := <-d.free
	cmd := &d.instbuf[index]
	cmd.Opcode = op
	cmd.IfSerial = 0
	return cmd, index
}

// Clear clears the buffer.  Clear is asynchronous, requiring a call
// to [Sync] to have visible effect.
func (d *Display) Clear() {
	d.todo <- command(CmdClear)
}

// Box draws a filled rectangle.  Box is asynchronous, requiring a
// call to [Sync] to have visible effect.
func (d *Display) Box(x, y, w, h int32) {
	cmd, op := d.acquire(CmdBox)
	cmd.X = x
	cmd.Y = y
	cmd.W = w
	cmd.H = h
	d.todo <- op
}

// Sleeping reports whether the display is sleeping.
func (d *Display) Sleeping() bool {
	return d.blanked.Load()
}

// Sleep turns off the display. Sending any other command afterward
// turns it back on again.  Sleep has immediate effect, it does not
// require a call to [Sync].
func (d *Display) Sleep() {
	d.todo <- command(CmdSleep)
}

// KeepAlive unblanks the screen as necessary, then resets the screensaver
// timer so the display won't sleep again for the maximum interval.
func (d *Display) KeepAlive() {
	d.todo <- command(CmdWake)
}

// Sync updates the display with any changes since its last call.
func (d *Display) Sync() {
	d.todo <- command(CmdSync)
}

// Text displays the given text, expanded to fill the entire screen.
// Text has immediate effect, it does not require a call to [Sync].
func (d *Display) Text(s string) {
	d.TextIfSerial(0, s)
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
	cmd, op := d.acquire(CmdText)
	cmd.W = -1 // stretch
	cmd.H = -1 // stretch
	cmd.IfSerial = n
	cmd.String = s
	d.todo <- op
}

// TextAt displays the given text with its top left corner at x, y,
// roughly h pixels high. TextAt is asynchronous, requiring a call
// to [Sync] to have visible effect.
func (d *Display) TextAt(x, y, h int32, s string) {
	cmd, op := d.acquire(CmdText)
	cmd.X = x
	cmd.Y = y
	cmd.W = 0 // whatever it ends up as
	cmd.H = h
	cmd.String = s
	d.todo <- op
}

// do executes received commands.
func (d *Display) do(index command) error {
	var cmd *Instruction
	var op Opcode

	if index < 0 {
		op = Opcode(index)
	} else {
		cmd = &d.instbuf[index]
		op = cmd.Opcode
	}

	if op == CmdSleep {
		if !d.blanked.Swap(true) {
			d.device.Command(ssd1306.DISPLAYOFF)
		}
		return nil
	} else if d.blanked.Swap(false) {
		d.device.Command(ssd1306.DISPLAYON)
	}
	d.blanker.Reset(BlankTime)
	if op == CmdWake {
		return nil
	}

	// SleepCommand and WakeCommand not incrementing d.serial is
	// intentional, the former so sequences can have gaps longer than
	// BlankTime without being cut off every time, and the latter so
	// sequences aren't broken by playing notes or changing the volume
	// or tempo.
	serial := d.serial.Add(1) - 1
	if cmd != nil && cmd.IfSerial != 0 && cmd.IfSerial != serial {
		return nil
	}

	switch op {
	case CmdClear:
		d.clear()
		return nil

	case CmdBox:
		d.box(cmd.X, cmd.Y, cmd.W, cmd.H)
		return nil

	case CmdSync:
		return d.sync()

	case CmdText:
		if cmd.W < 0 && cmd.H < 0 {
			d.renderFullscreen(cmd.String)
			return d.sync()
		}
		d.renderTextAt(cmd.X, cmd.Y, cmd.H, cmd.String)
		return nil

	default:
		println("Display.do:", op, "not implemented")
		return nil
	}
}

// clear clears the buffer.
func (d *Display) clear() {
	dst := d.buf
	for i := range dst {
		dst[i] = 0
	}
}

// box draws a filled rectangle.
func (d *Display) box(x, y, w, h int32) {
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
