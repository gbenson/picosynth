package picosynth

import (
	"sync/atomic"

	"gbenson.net/go/picosynth/internal/counts"
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
	////// go func() {
	////// 	time.Sleep(time.Second * 4)
	////// 	if lastMIDI.Load() != 0 {
	////// 		return
	////// 	}
	////// 	v.text = "this"
	////// 	ui.InvalidateDisplay()
	////// 	time.Sleep(time.Second / 2)
	////// 	if lastMIDI.Load() != 0 {
	////// 		return
	////// 	}
	////// 	v.text = "-= is =-"
	////// 	ui.InvalidateDisplay()
	////// 	time.Sleep(time.Second / 2)
	////// 	if lastMIDI.Load() != 0 {
	////// 		return
	////// 	}
	////// 	v.text = "picosynth"
	////// 	ui.InvalidateDisplay()
	////// 	time.Sleep(time.Second * 3)
	////// 	if lastMIDI.Load() != 0 {
	////// 		return
	////// 	}
	////// 	v.text = ""
	////// 	ui.InvalidateDisplay()
	////// }()
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

func renderHex(d *display.Display, y int32, v uint32) {
	for i := range 4 {
		ui.RenderHexAt(d, int32(17*i+1), y, 8, uint8((v>>((3-i)*8))&255))
	}
}

// Render implements [Page].
func (v *Visualizer) Render(d *display.Display, now uint32) {
	d.Clear()
	renderHex(d, 0, counts.WriteMono.Load())
	renderHex(d, 8, counts.FillBuffer.Load())
	renderHex(d, 16, counts.DisplayTick.Load())
	renderHex(d, 24, counts.ScanRow.Load())
	d.Sync()
}
