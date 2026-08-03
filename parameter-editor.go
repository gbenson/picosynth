package picosynth

import (
	"sync/atomic"

	"gbenson.net/go/picosynth/internal/display"
	"gbenson.net/go/picosynth/internal/ui"
)

type Parameter = ui.Parameter

type ParameterSpec interface {
	Parameter() Parameter
}

type ParameterGroup struct {
	Name       string
	Hotkey     Scancode
	LongPress  bool
	Parameters []ParameterSpec
}

type ParameterGroupPage struct {
	ParameterGroup

	ui       UI
	mem      *Memory
	params   []Parameter
	selected atomic.Uint32
}

// NewParameterGroupPage creates and initializes a new [ParameterGroupPage].
func NewParameterGroupPage(pg ParameterGroup) *ParameterGroupPage {
	return &ParameterGroupPage{ParameterGroup: pg}
}

// SelectedParameter returns the currently selected parameter.
func (pg *ParameterGroupPage) SelectedParameter() Parameter {
	return pg.params[pg.selected.Load()]
}

// OnInit implements [Page].
func (pg *ParameterGroupPage) OnInit(ui UI, mem *Memory) {
	if len(pg.Parameters) < 1 {
		panic("no parameters")
	}

	pg.ui = ui
	pg.mem = mem

	pg.params = make([]Parameter, len(pg.Parameters))
	for i, ps := range pg.Parameters {
		pg.params[i] = ps.Parameter()
	}
}

// OnFocus implements [Page].
func (pg *ParameterGroupPage) OnFocus() {
	pg.SelectedParameter().Focus(pg.mem)
}

// OnButtonPress implements [Page].
func (pg *ParameterGroupPage) OnButtonPress(sc Scancode, longpress bool) bool {
	if sc != pg.Hotkey {
		return false
	}
	ui := pg.ui

	if !ui.PageHasFocus(pg) {
		// only the exact press grants focus to an unfocused page.
		if longpress != pg.LongPress {
			return false
		}
		return true

	} else if ui.ScreenBlanked() {
		// a short press can wake the display for a long press page,
		// but a long press can't wake the display for a short press
		// page. return false and let the long-press page for this
		// hotkey take focus and wake the screen.
		return !longpress || pg.LongPress

	} else if longpress {
		// a long press on a short-press page means switch to the
		// long-press page.	we return false so the long-press page
		// for this hotkey will take the focus.
		if !pg.LongPress {
			return false
		}

		// a long press on a long press page switches to the short-
		// press page for the same hotkey. We fake a short press to
		// make the switch.
		ui.YieldFocus()
		if impl, ok := ui.(*psUI); ok {
			// XXX how to do this nicely?
			impl.ps.onButtonPress(sc, false)
		}
		return true

	} else {
		// we're focussed, the screen isn't alseep, and we got a short
		// press: step through parameters. NB I know this isn't atomic,
		// pg.selected is atomic because it's shared between the filler
		// goroutine (us, we sets it) and the UI display refresh one
		// (which only reads it.)  The only place it's stored is here.
		selected := pg.selected.Load()
		selected++
		if selected >= uint32(len(pg.Parameters)) {
			selected = 0
		}
		pg.selected.Store(selected)

		pg.OnFocus()
		ui.InvalidateDisplay()

		return true
	}
}

// OnEncoderMove implements [Page].
func (pg *ParameterGroupPage) OnEncoderMove(delta int) {
	pg.SelectedParameter().Adjust(pg.mem, int32(delta))
	pg.ui.InvalidateDisplay()
}

// Render implements [Page].
func (pg *ParameterGroupPage) Render(d *display.Display, now uint32) {
	pg.SelectedParameter().Render(d)
}
