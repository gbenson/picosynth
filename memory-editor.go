package picosynth

import (
	"sync/atomic"

	"gbenson.net/go/picosynth/internal/display"
	"gbenson.net/go/picosynth/internal/ui"
)

// The code in this file is kind of a mess, it was originally the only
// way to edit anything, before the UI was written, and the more of the
// UI gets written the more this becomes a semi-hidden debug-tool/last-
// resort that's been mangled to still work with the newer UI methods.
// That's why it's total spaghetti!

type MemoryEditor struct {
	// Register and byte cursors
	register int
	selected int // byte 0..3, or -1 for none

	value atomic.Uint32
}

// OnInit implements [Page].
func (me *MemoryEditor) OnInit(ui *Picosynth) {
	// Advance to the first named editable register.
	me.navigate(ui, 0)

	// Be selecting a register, not editing a byte.
	me.selected = -1
}

// OnFocus implements [Page].
func (me *MemoryEditor) OnFocus(ui *Picosynth) {
	me.value.Store(ui.mem.load(me.register))
}

// OnButtonPress implements [Page].
func (me *MemoryEditor) OnButtonPress(ui *Picosynth, sc Scancode, longpress bool) bool {
	const Hotkey = ButtonToneEdit

	if longpress {
		return false
	} else if !ui.PageHasFocus(me) {
		// maybe take focus
		return sc == Hotkey
	} else if ui.ScreenBlanked() {
		// eat the keypress
		return true
	} else if sc == Hotkey {
		// switch to visualizer
		ui.YieldFocus()
		return true
	}

	switch sc {
	case ButtonKeyboard:
		me.onDecreaseMulti(ui)
	case ButtonWind:
		me.onIncreaseMulti(ui)
	case ButtonString:
		me.onCycle(ui)
	case ButtonSynth:
		me.onDecrease(ui)
	case ButtonSE:
		me.onIncrease(ui)
	default:
		return false
	}

	return true
}

// OnEncoderMove implements [Page].
func (me *MemoryEditor) OnEncoderMove(ui *Picosynth, delta int) {
	me.onChange(ui, delta, false)
}

func (me *MemoryEditor) onCycle(ui *Picosynth) {
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
	me.invalidateDisplay(ui)
}

func (me *MemoryEditor) onDecrease(ui *Picosynth) {
	me.onChange(ui, -1, false)
}

func (me *MemoryEditor) onIncrease(ui *Picosynth) {
	me.onChange(ui, +1, false)
}

func (me *MemoryEditor) onDecreaseMulti(ui *Picosynth) {
	me.onChange(ui, -1, true)
}

func (me *MemoryEditor) onIncreaseMulti(ui *Picosynth) {
	me.onChange(ui, +1, true)
}

func (me *MemoryEditor) onChange(ui *Picosynth, step int, multi bool) {
	if me.selected < 0 {
		// moving through registers
		if multi {
			step *= 10
		}
		me.navigate(ui, step)
	} else {
		// adjusting a value
		mask := uint32(15)
		if multi {
			step <<= 4
			mask <<= 4
		}
		me.adjust(ui.mem, uint32(step), mask)
	}
	me.invalidateDisplay(ui)
}

// nagivate moves up and down the list of named editable registers by
// the given number amount.
func (me *MemoryEditor) navigate(ui *Picosynth, steps int) {
	step := 1
	if steps < 0 {
		step = -1
		steps *= -1
	}

	for _ = range NumRegisters {
		r := ui.mem.Register(me.register)
		if r.Assigned() && r.Editable() {
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
func (me *MemoryEditor) adjust(mem *Memory, delta, mask uint32) {
	shift := (3 - me.selected) * 8
	delta <<= shift
	mask <<= shift

	r := mem.Register(me.register)
	v := r.load()
	mem.Store(me.register, (v & ^mask)|((v+delta)&mask))
}

func (me *MemoryEditor) invalidateDisplay(ui *Picosynth) {
	me.OnFocus(ui) // update me.value
	ui.InvalidateDisplay()
}

// Render implements [Page].
func (me *MemoryEditor) Render(d *display.Display, now uint32) {
	d.Clear()

	n := me.register
	if n < MatrixCellsBase {
		ui.RenderHexAt(d, 117, 0, 8, uint8(n))
	} else {
		d.TextAt(117, 0, 16, string([]byte{'1' + byte(n&3)}))
		d.TextAt(113, 0, 8, "+")
	}

	r := Register{nil, me.register}
	ui.RenderRegisterName(d, r.Name())
	ui.RenderHexValue(d, me.value.Load())

	if i := int32(me.selected); i >= 0 {
		d.Box(35*i+1, 30, 22, 2)
	}

	d.Sync()
}
