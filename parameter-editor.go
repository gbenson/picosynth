package picosynth

import "gbenson.net/go/picosynth/internal/ui"

type ParameterSpec interface {
	Parameter() ui.Parameter
}

type ParameterGroup struct {
	Name       string
	Hotkey     Scancode
	LongPress  bool
	Parameters []ParameterSpec
}

type ParameterGroupPage struct {
	ParameterGroup

	ui       *UI
	params   []ui.Parameter
	selected int
}

// NewParameterGroupPage creates and initializes a new [ParameterGroupPage].
func NewParameterGroupPage(pg ParameterGroup) *ParameterGroupPage {
	return &ParameterGroupPage{ParameterGroup: pg}
}

// OnInit implements [Page].
func (pg *ParameterGroupPage) OnInit(sys *UI) {
	if len(pg.Parameters) < 1 {
		panic("no parameters")
	}

	pg.ui = sys
	pg.params = make([]ui.Parameter, len(pg.Parameters))
	for i, ps := range pg.Parameters {
		pg.params[i] = ps.Parameter()
	}
}

// OnButton implements [Page].
func (pg *ParameterGroupPage) OnButton(sc Scancode, longpress bool) bool {
	if sc != pg.Hotkey {
		return false
	}

	if !pg.ui.HasFocus(pg) {
		// only the exact press grants focus to an unfocused page.
		if longpress != pg.LongPress {
			return false
		}
	} else if pg.ui.display.Sleeping() {
		// a short press can wake the display for a long press page,
		// but a long press can't wake the display for a short press
		// page. return false and let the long-press page for this
		// hotkey take focus and wake the screen.
		if longpress && !pg.LongPress {
			return false
		}
	} else if longpress {
		// a long press on a short-press page means switch to the
		// long-press page.	we return false so the long-press page
		// for this hotkey will take the focus.
		if !pg.LongPress {
			return false
		}

		// a long press on a long press page switches to the short-
		// press page for the same hotkey. ideally :)  we fake a
		// short press to try and switch, but the event might drop,
		// in which case having unset currentPage means we'll enter
		// the short-press page if the user short presses the hotkey
		// (which they likely will if that's what they wanted).
		pg.ui.currentPage = nil
		pg.ui.sendEvent(ButtonPressEvent(sc))
		return true
	} else {
		// we're focussed, the screen isn't alseep, and we got a short press.
		// step through parameters.
		pg.selected++
		if pg.selected >= len(pg.Parameters) {
			pg.selected = 0
		}
	}

	pg.params[pg.selected].Render(pg.ui)
	return true
}

// OnEncoder implements [Page].
func (pg *ParameterGroupPage) OnEncoder(delta int) {
	p := pg.params[pg.selected]
	p.Adjust(pg.ui, int32(delta))
	p.Render(pg.ui)
}
