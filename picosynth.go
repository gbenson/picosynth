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
	return uint32(int64(d) * int64(FillRate) / int64(time.Second))
}

var potRegisters = []int{
	Filt1Cutoff,
	Filt1Resonance,
}

// Picosynth is a replacement brain for Casio SA-5 keyboards.
type Picosynth struct {
	ui  UI
	mem *Memory

	encoder *encoder.Encoder
	pots    []adc.ADC

	keyscanner KeyScanner
	keytracker KeyTracker

	octave int
	volume int

	// currentStep is the step number of the current step.  It's
	// incremented at the start of every [UI.Step], which at the
	// current FillRate of 375Hz gives us 132.5 days without wrapping.
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

func (ps *Picosynth) init(m *Memory) error {
	ps.ui = &psUI{ps}
	ps.mem = m

	enc, err := encoder.Open(0)
	if err != nil {
		return err
	}
	ps.encoder = enc

	pots := make([]adc.ADC, len(potRegisters))
	for i := range pots {
		pot, err := adc.Open(i)
		if err != nil {
			return err
		}
		pots[i] = pot
	}
	ps.pots = pots

	ps.keytracker.init()

	ps.SetOctave(InitialOctave)
	ps.SetVolume(InitialVolume)

	ps.AddPage(&ps.visualizer)
	ps.AddPage(&MemoryEditor{})
	for _, pg := range ParameterGroups {
		ps.AddPage(NewParameterGroupPage(pg))
	}

	if ps.CurrentPage() == nil {
		panic("no pages")
	}

	return nil
}

func (ps *Picosynth) AddPage(p Page) {
	p.OnInit(ps.ui, ps.mem)
	ps.pages = append(ps.pages, p)

	if ps.defaultPage == nil {
		ps.defaultPage = p
		ps.focus(p)
	}
}

func (ps *Picosynth) Step() {
	currentStep := ps.currentStep.Add(1)

	var activity bool
	for e := ps.keyscanner.Poll(); e != NoEvent; e = ps.keyscanner.Poll() {
		sc := e.Scancode()
		if note := sc.Note(); note.IsValid() {
			ps.keytracker.Receive(note, e.Down())
			activity = true
		} else if e.Down() {
			ps.onButtonDown(sc, currentStep)
			// don't register activity, or, if the screen is blanked,
			// it'll unblank on the buttonDown so the eat-the-key
			// login in various Page OnButtonPress handles will never
			// be entered: the screen *won't* be blank on the buttonUp.
		} else {
			ps.onButtonUp(sc, currentStep)
			activity = true
		}
	}

	if v := ps.encoder.Read(); v != 0 {
		ps.onEncoderMove(v)
		activity = true
	}

	for i := range ps.pots {
		v := ps.pots[i].Get()
		r := potRegisters[i]
		switch r {
		case Filt1Cutoff:
			p := Pitch(v)
			p *= (MaxAudiblePitch - MinAudiblePitch) >> 16
			p += MinAudiblePitch
			ps.mem.StorePitch(r, p)
		default:
			ps.mem.StoreSignal(r, Signal(v)<<15) // 0xffff -> 0x7fff8000
		}
		//activity = true
	}

	if activity {
		ps.lastActivityStep.Store(currentStep)
	}

	ps.keytracker.Step()
	ps.mem.StorePitch(VoicePitch, ps.keytracker.Note.Pitch())
}

func (ps *Picosynth) onButtonDown(sc Scancode, currentStep uint32) {
	ps.buttonDownStep[sc] = currentStep
}

func (ps *Picosynth) onButtonUp(sc Scancode, currentStep uint32) {
	holdTime := currentStep - ps.buttonDownStep[sc]
	if holdTime > longPressTimeout {
		ps.onButtonPress(sc, true)
		return
	}

	switch sc {
	case ButtonVolumeUp:
		ps.SetVolume(ps.volume + 1)
	case ButtonVolumeDown:
		ps.SetVolume(ps.volume - 1)
	case ButtonTempoUp:
		ps.SetOctave(ps.octave + 1)
	case ButtonTempoDown:
		ps.SetOctave(ps.octave - 1)
	default:
		ps.onButtonPress(sc, false)
	}
}

func (ps *Picosynth) SetOctave(v int) {
	ps.octave = max(MinOctave, min(MaxOctave, v))
	ps.keytracker.Transpose = ps.octave * 12
}

func (ps *Picosynth) SetVolume(v int) {
	ps.volume = max(MinVolume, min(MaxVolume, v))
}

func (ps *Picosynth) onButtonPress(sc Scancode, longpress bool) {
	currentPage := ps.CurrentPage()

	if currentPage.OnButtonPress(sc, longpress) {
		return // handled
	}

	for _, p := range ps.pages {
		if p == currentPage {
			continue // currentPage had its turn
		}
		if p.OnButtonPress(sc, longpress) {
			ps.focus(p)
			return
		}
	}

	if longpress {
		println("button", sc, "(long press) ignored")
	} else {
		println("button", sc, "(short press) ignored")
	}
}

func (ps *Picosynth) onEncoderMove(delta int) {
	ps.CurrentPage().OnEncoderMove(delta)
}

// CurrentPage returns the currently displayed page.
func (ps *Picosynth) CurrentPage() Page {
	return *ps.currentPage.Load()
}

// PageHasFocus reports whether p has focus.
func (ps *Picosynth) PageHasFocus(p Page) bool {
	return ps.CurrentPage() == p
}

// ScreenBlanked reports whether the screen is blanked.
func (ps *Picosynth) ScreenBlanked() bool {
	return ps.screenBlanked.Load()
}

type psUI struct {
	ps *Picosynth
}

// PageHasFocus implements [UI].
func (ui *psUI) PageHasFocus(p Page) bool {
	return ui.ps.PageHasFocus(p)
}

// YieldFocus implements [UI].
func (ui *psUI) YieldFocus() {
	ui.ps.focus(ui.ps.defaultPage)
}

func (ps *Picosynth) focus(p Page) {
	if p == nil {
		panic("nil page")
	}
	ps.currentPage.Store(&p)
	p.OnFocus()
	ps.ui.InvalidateDisplay()
}

// InvalidateDisplay implements [UI].
func (ui *psUI) InvalidateDisplay() {
	ui.ps.screenCurrent.Store(false)
}

// ScreenBlanked implements [UI].
func (ui *psUI) ScreenBlanked() bool {
	return ui.ps.ScreenBlanked()
}

// Name implements [worker].
func (ui *Picosynth) Name() string {
	return "display"
}

// Run implements [worker].
func (ui *Picosynth) Run() func() error {
	return ui.run
}

func (ui *Picosynth) run() error {
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
