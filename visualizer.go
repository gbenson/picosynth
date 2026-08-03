package picosynth

import (
	"time"

	"gbenson.net/go/picosynth/internal/display"
)

type Visualizer struct {
	ui   UI
	text string
}

// OnInit implements [Page].
func (v *Visualizer) OnInit(ui UI, mem *Memory) {
	v.ui = ui

	v.text = "hello"
	go func() {
		time.Sleep(time.Second * 4)
		v.text = "this"
		ui.InvalidateDisplay()
		time.Sleep(time.Second / 2)
		v.text = "-= is =-"
		ui.InvalidateDisplay()
		time.Sleep(time.Second / 2)
		v.text = "picosynth"
		ui.InvalidateDisplay()
		time.Sleep(time.Second * 3)
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
	if v.text == "" {
		d.Clear()
	} else {
		d.Text(v.text)
	}
	d.Sync()
}
