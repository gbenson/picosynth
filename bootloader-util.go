package picosynth

import (
	"gbenson.net/go/picosynth/internal/display"
	"gbenson.net/go/picosynth/internal/hw/machine"
)

type BootloaderUtil struct {
	ui        UI
	moved     bool
	yesNo     bool
	confirmed bool
}

// OnInit implements [Page].
func (p *BootloaderUtil) OnInit(ui UI, mem *Memory) {
	p.ui = ui
}

// OnFocus implements [Page].
func (p *BootloaderUtil) OnFocus() {
	p.moved = false
	p.yesNo = false
}

// OnButtonPress implements [Page].
func (p *BootloaderUtil) OnButtonPress(sc Scancode, longpress bool) bool {
	const Hotkey = ButtonToneEdit

	if !p.ui.PageHasFocus(p) {
		// maybe take focus
		return sc == Hotkey && longpress
	} else if sc == Hotkey && !longpress && p.yesNo {
		p.confirmed = true
		p.ui.InvalidateDisplay()
	} else {
		p.ui.YieldFocus()
	}
	return true
}

// OnEncoderMove implements [Page].
func (p *BootloaderUtil) OnEncoderMove(delta int) {
	if !p.moved {
		p.moved = true
	} else if delta&1 == 0 {
		return
	} else {
		p.yesNo = !p.yesNo
	}

	p.ui.InvalidateDisplay()
}

// Render implements [Page].
func (p *BootloaderUtil) Render(d *display.Display, now uint32) {
	if p.confirmed {
		d.Sleep()
		machine.EnterBootloader()
	} else if !p.moved {
		d.Clear()
		d.TextAt(0, 0, 16, "Update")
		d.TextAt(0, 16, 16, "firmware?")
	} else if p.yesNo {
		d.Text("Yes")
	} else {
		d.Text("No")
	}
	d.Sync()
}
