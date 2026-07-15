package picosynth

type MemoryEditor struct {
	ui *UI

	// Startup management
	started bool

	// Register and byte cursors
	register int
	selected int // byte 0..3, or -1 for none
}

// OnInit implements [Page].
func (me *MemoryEditor) OnInit(ui *UI) {
	me.ui = ui

	// Advance to the first named editable register.
	me.navigate(0)

	// Be selecting a register, not editing a byte.
	me.selected = -1
}

// OnButton implements [Page].
func (me *MemoryEditor) OnButton(sc Scancode, longPress bool) bool {
	if longPress {
		return false
	}

	var handler func()
	switch sc {
	case ButtonKeyboard:
		handler = me.onDecreaseMulti
	case ButtonWind:
		handler = me.onIncreaseMulti
	case ButtonString:
		handler = me.onCycle
	case ButtonSynth:
		handler = me.onDecrease
	case ButtonSE:
		handler = me.onIncrease
	default:
		return false
	}

	if !me.started {
		me.started = true
	} else if !me.ui.display.Sleeping() {
		handler()
	}

	me.redraw()

	return true
}

// OnEncoder implements [Page].
func (me *MemoryEditor) OnEncoder(delta int) {
	if !me.started {
		return
	}

	me.onChange(delta, false)
	me.redraw()
}

func (me *MemoryEditor) onCycle() {
	if me.selected < 0 {
		// switch from moving through registers to editing the first byte
		me.selected = 0
	} else {
		me.selected++
		me.selected &= 3
		if me.selected == 0 {
			me.selected -= 1
		}
	}
}

func (me *MemoryEditor) onDecrease() {
	me.onChange(-1, false)
}

func (me *MemoryEditor) onIncrease() {
	me.onChange(+1, false)
}

func (me *MemoryEditor) onDecreaseMulti() {
	me.onChange(-1, true)
}

func (me *MemoryEditor) onIncreaseMulti() {
	me.onChange(+1, true)
}

func (me *MemoryEditor) onChange(step int, multi bool) {
	if me.selected < 0 {
		// moving through registers
		if multi {
			step *= 10
		}
		me.navigate(step)
	} else {
		// adjusting a value
		mask := uint32(15)
		if multi {
			step <<= 4
			mask <<= 4
		}
		me.adjust(uint32(step), mask)
	}
}

// nagivate moves up and down the list of named editable registers by
// the given number amount.
func (me *MemoryEditor) navigate(steps int) {
	step := 1
	if steps < 0 {
		step = -1
		steps *= -1
	}

	for _ = range NumRegisters {
		r := me.Register()
		if r.Name() != "" && r.Editable() {
			if steps == 0 {
				return
			}
			steps--
		}

		me.register += step
		me.register &= NumRegisters - 1 // hope its a power of two!
	}

	panic("no named editable registers")
}

// adjust changes the selected byte by the given amount.
func (me *MemoryEditor) adjust(delta, mask uint32) {
	shift := (3 - me.selected) * 8
	delta <<= shift
	mask <<= shift

	r := me.Register()
	v := r.load()
	me.ui.Store(me.register, (v & ^mask)|((v+delta)&mask))
}

func (me *MemoryEditor) Register() Register {
	return me.ui.mem.Register(me.register)
}

func (me *MemoryEditor) redraw() {
	d := &me.ui.display
	d.Clear()

	n := me.register
	if n < MatrixCellsBase {
		me.hexAt(117, 0, 8, uint8(n))
	} else {
		d.TextAt(117, 0, 16, string([]byte{'1' + byte(n&3)}))
		d.TextAt(113, 0, 8, "+")
	}

	r := me.Register()
	d.TextAt(0, 0, 16, r.Name())

	v := r.load()
	for i := range 4 {
		me.hexAt(int32(35*i+1), 16, 16, uint8((v>>((3-i)*8))&255))
	}

	if i := int32(me.selected); i >= 0 {
		d.Box(35*i+1, 30, 22, 2)
	}

	d.Sync()
}

func (me *MemoryEditor) hexAt(x, y, h int32, v uint8) {
	d := &me.ui.display
	const digits = "0123456789abcdef"

	d.TextAt(x, y, h, string(digits[v>>4]))
	d.TextAt(x+(h/8)*6, y, h, string(digits[v&15]))
}
