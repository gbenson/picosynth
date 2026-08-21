package picosynth

import (
	"sync/atomic"
	"time"

	"gbenson.net/go/picosynth/internal/display"
	"gbenson.net/go/picosynth/internal/ui"
)

type Visualizer struct {
	ui   UI
	text string
}

var lastMIDI atomic.Uint32 // XXX remove

// OnInit implements [Page].
func (v *Visualizer) OnInit(ui UI, mem *Memory) {
	v.ui = ui

	v.text = "hello"
	go func() {
		time.Sleep(time.Second * 4)
		if lastMIDI.Load() != 0 {
			return
		}
		v.text = "this"
		ui.InvalidateDisplay()
		time.Sleep(time.Second / 2)
		if lastMIDI.Load() != 0 {
			return
		}
		v.text = "-= is =-"
		ui.InvalidateDisplay()
		time.Sleep(time.Second / 2)
		if lastMIDI.Load() != 0 {
			return
		}
		v.text = "picosynth"
		ui.InvalidateDisplay()
		time.Sleep(time.Second * 3)
		if lastMIDI.Load() != 0 {
			return
		}
		v.text = ""
		ui.InvalidateDisplay()
	}()
}

// OnFocus implements [Page].
func (v *Visualizer) OnFocus() {
}

// OnButtonPress implements [Page].
func (v *Visualizer) OnButtonPress(sc Scancode, longpress bool) bool {
	const Hotkey = ButtonToneEdit

	if longpress {
		return false
	} else if sc != Hotkey {
		return false
	} else if !v.ui.PageHasFocus(v) {
		return true // take focus
	} else if v.ui.ScreenBlanked() {
		return true // eat the keypress
	} else {
		return false // switch to memory editor
	}
}

// OnEncoderMove implements [Page].
func (v *Visualizer) OnEncoderMove(delta int) {
}

// Render implements [Page].
func (v *Visualizer) Render(d *display.Display, now uint32) {
	m := MIDIMessage(lastMIDI.Load())
	if m != NoMessage {
		d.Clear()
		ui.RenderRegisterName(d, "MIDI")
		ui.RenderHexAt(d, 1, 16, 16, uint8(m))
		ui.RenderHexAt(d, 36, 16, 16, uint8(m>>8))
		ui.RenderHexAt(d, 71, 16, 16, uint8(m>>16))

		d.Sync()
		return
	}

	if v.text == "" {
		d.Clear()
	} else {
		d.Text(v.text)
	}
	d.Sync()
}
