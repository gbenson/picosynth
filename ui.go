package picosynth

import (
	"sync/atomic"
	"time"

	"gbenson.net/go/picosynth/internal/adc"
	"gbenson.net/go/picosynth/internal/display"
	"gbenson.net/go/picosynth/internal/encoder"
)

const (
	MinVolume     = 0
	MaxVolume     = 10
	InitialVolume = 7

	MinOctave     = -3
	MaxOctave     = 3
	InitialOctave = -1

	// LongPressTimeout is the length of time a button must be held
	// before a press turns into a long press.
	LongPressTimeout = 500 * time.Millisecond

	// FrameRate is the rate at which the display is updated.
	FrameRate = 10

	// ActivityTimeout is the length of time without user activity
	// before the user will be considered inactive.
	ActivityTimeout = 30 * time.Second
)

var (
	// longPressTimeout is the number of UI steps a button must be
	// held before a press turns into a long press.
	longPressTimeout = uiStepCount(LongPressTimeout)

	// activityTimeout is the number of UI steps without user activity
	// before the user will be considered inactive.
	activityTimeout = uiStepCount(ActivityTimeout)
)

// uiStepCount returns a duration as a number of UI steps.
func uiStepCount(d time.Duration) uint32 {
	return uint32(int64(d) * int64(TickRate) / int64(time.Second))
}

var potRegisters = []int{
	Filt1Cutoff,
	Filt1Resonance,
}

type UI struct {
	mem *Memory

	encoder *encoder.Encoder
	pots    []adc.ADC

	keyscanner KeyScanner
	keytracker KeyTracker

	octave int
	volume int

	// currentStep is the step number of the current step.  It's
	// incremented at the start of every [UI.Step], which at the
	// current TickRate of 375Hz gives us 132.5 days without wrapping.
	currentStep atomic.Uint32

	// buttonDownStep is the step numbers of the last steps each
	// button transitioned from up to down.
	buttonDownStep [numScancodes]uint32

	// lastActivityStep is the step number of the last step where
	// user activity was observed.
	lastActivityStep atomic.Uint32

	screenBlanked atomic.Bool // screensaver active
	screenCurrent atomic.Bool // no redraw required

	visualizer  Visualizer
	currentPage atomic.Pointer[Page]
	pages       []Page
	defaultPage Page
}

func (ui *UI) init(m *Memory) error {
	ui.mem = m

	enc, err := encoder.Open(0)
	if err != nil {
		return err
	}
	ui.encoder = enc

	pots := make([]adc.ADC, len(potRegisters))
	for i := range pots {
		pot, err := adc.Open(i)
		if err != nil {
			return err
		}
		pots[i] = pot
	}
	ui.pots = pots

	ui.keytracker.init()

	ui.setOctave(InitialOctave)
	ui.setVolume(InitialVolume)

	ui.AddPage(&ui.visualizer)
	ui.AddPage(&MemoryEditor{})
	for _, pg := range ParameterGroups {
		ui.AddPage(NewParameterGroupPage(pg))
	}

	if ui.CurrentPage() == nil {
		panic("no pages")
	}

	return nil
}

func (ui *UI) AddPage(p Page) {
	p.OnInit(ui)
	ui.pages = append(ui.pages, p)

	if ui.defaultPage == nil {
		ui.defaultPage = p
		ui.focus(p)
	}
}

func (ui *UI) Step() {
	currentStep := ui.currentStep.Add(1)

	var activity bool
	for e := ui.keyscanner.Poll(); e != NoEvent; e = ui.keyscanner.Poll() {
		sc := e.Scancode()
		if note := sc.Note(); note.IsValid() {
			ui.keytracker.Receive(note, e.Down())
			activity = true
		} else if e.Down() {
			ui.onButtonDown(sc, currentStep)
			// don't register activity, or, if the screen is blanked,
			// it'll unblank on the buttonDown so the eat-the-key
			// login in various Page OnButtonPress handles will never
			// be entered: the screen *won't* be blank on the buttonUp.
		} else {
			ui.onButtonUp(sc, currentStep)
			activity = true
		}
	}

	if v := ui.encoder.Read(); v != 0 {
		ui.onEncoderMove(v)
		activity = true
	}

	for i := range ui.pots {
		v := ui.pots[i].Get()
		r := potRegisters[i]
		switch r {
		case Filt1Cutoff:
			p := Pitch(v)
			p *= (MaxAudiblePitch - MinAudiblePitch) >> 16
			p += MinAudiblePitch
			ui.mem.StorePitch(r, p)
		default:
			ui.mem.StoreSignal(r, Signal(v)<<15) // 0xffff -> 0x7fff8000
		}
		//activity = true
	}

	if activity {
		ui.lastActivityStep.Store(currentStep)
	}

	ui.keytracker.Step()
	ui.mem.StorePitch(VoicePitch, ui.keytracker.Note.Pitch())
}

func (ui *UI) onButtonDown(sc Scancode, currentStep uint32) {
	ui.buttonDownStep[sc] = currentStep
}

func (ui *UI) onButtonUp(sc Scancode, currentStep uint32) {
	holdTime := currentStep - ui.buttonDownStep[sc]
	if holdTime > longPressTimeout {
		ui.onButtonPress(sc, true)
		return
	}

	switch sc {
	case ButtonVolumeUp:
		ui.setVolume(ui.volume + 1)
	case ButtonVolumeDown:
		ui.setVolume(ui.volume - 1)
	case ButtonTempoUp:
		ui.setOctave(ui.octave + 1)
	case ButtonTempoDown:
		ui.setOctave(ui.octave - 1)
	default:
		ui.onButtonPress(sc, false)
	}
}

func (ui *UI) setOctave(v int) {
	ui.octave = max(MinOctave, min(MaxOctave, v))
	ui.keytracker.Transpose = ui.octave * 12
}

func (ui *UI) setVolume(v int) {
	ui.volume = max(MinVolume, min(MaxVolume, v))
}

func (ui *UI) onButtonPress(sc Scancode, longpress bool) {
	currentPage := ui.CurrentPage()

	if currentPage.OnButtonPress(ui, sc, longpress) {
		return // handled
	}

	for _, p := range ui.pages {
		if p == currentPage {
			continue // currentPage had its turn
		}
		if p.OnButtonPress(ui, sc, longpress) {
			ui.focus(p)
			return
		}
	}

	if longpress {
		println("button", sc, "(long press) ignored")
	} else {
		println("button", sc, "(short press) ignored")
	}
}

func (ui *UI) onEncoderMove(delta int) {
	ui.CurrentPage().OnEncoderMove(ui, delta)
}

// CurrentPage returns the currently displayed page.
func (ui *UI) CurrentPage() Page {
	return *ui.currentPage.Load()
}

func (ui *UI) HasFocus(p Page) bool {
	return ui.CurrentPage() == p
}

func (ui *UI) YieldFocus() {
	ui.focus(ui.defaultPage)
}

func (ui *UI) focus(p Page) {
	if p == nil {
		panic("nil page")
	}
	ui.currentPage.Store(&p)
	p.OnFocus(ui)
	ui.InvalidateDisplay()
}

// ScreenBlanked reports whether the screen is blanked.
func (ui *UI) ScreenBlanked() bool {
	return ui.screenBlanked.Load()
}

// InvalidateDisplay requests a redraw at the next refresh.
func (ui *UI) InvalidateDisplay() {
	ui.screenCurrent.Store(false)
}

// Name implements [worker].
func (ui *UI) Name() string {
	return "display"
}

// Run implements [worker].
func (ui *UI) Run() func() error {
	return ui.run
}

func (ui *UI) run() error {
	d, err := display.Open()
	if err != nil {
		return err
	}

	for _ = range time.Tick(FrameRate) {
		now := ui.currentStep.Load()

		if now-ui.lastActivityStep.Load() > activityTimeout {
			// user inactive
			needBlank := ui.screenBlanked.Swap(true)

			if needBlank {
				d.Sleep()
			}

		} else {
			// user recently active
			needUnblank := ui.screenBlanked.Swap(false)
			needRefresh := !ui.screenCurrent.Swap(true)

			if needUnblank || needRefresh {
				ui.CurrentPage().Render(d, now)
			}
		}
	}
	return nil
}
