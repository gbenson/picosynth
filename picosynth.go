package picosynth

import (
	"sync/atomic"
	"time"

	"gbenson.net/go/picosynth/internal/adc"
	"gbenson.net/go/picosynth/internal/audio"
	"gbenson.net/go/picosynth/internal/counts"
	"gbenson.net/go/picosynth/internal/display"
	"gbenson.net/go/picosynth/internal/encoder"
	"gbenson.net/go/picosynth/internal/hw/machine"
	"gbenson.net/go/picosynth/internal/keyboard"
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
	Engine Engine
	ui     UI

	encoder *encoder.Encoder
	pots    []machine.ADC

	keyscanner KeyScanner
	keytracker KeyTracker
	midiport   MIDIPort

	octave int

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

// Run is the main entry point of the firmware.
func Run() error {
	var ps Picosynth
	if err := ps.Init(); err != nil {
		return err
	}
	return ps.Run()
}

// Init initializes a [Picosynth].
func (ps *Picosynth) Init() error {
	if ps.ui != nil {
		panic("already initialized")
	}
	ps.ui = &psUI{ps}

	if err := ps.Engine.Init(); err != nil {
		return err
	}

	enc, err := encoder.Open(0)
	if err != nil {
		return err
	}
	ps.encoder = enc

	pots := make([]machine.ADC, len(potRegisters))
	for i := range pots {
		pot, err := adc.Open(i)
		if err != nil {
			return err
		}
		pots[i] = pot
	}
	ps.pots = pots

	ps.keytracker.init()

	if rx := keyboard.MIDIIn; rx != machine.NoPin {
		if err := ps.midiport.Open(rx-1, rx); err != nil {
			return err
		}
	}

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
	p.OnInit(ps.ui, &ps.Engine.Memory)
	ps.pages = append(ps.pages, p)

	if ps.defaultPage == nil {
		ps.defaultPage = p
		ps.focus(p)
	}
}

func (ps *Picosynth) Run() error {
	out, err := audio.Open(SampleRate)
	if err != nil {
		return err
	}
	defer out.Close()

	const numWorkers = 5 // display, keyscanner, midiport, filler, player
	wm := newWorkerManager(numWorkers)
	wm.Start(&ps.keyscanner)
	wm.Start(&psDisplay{ps})

	if ps.midiport.IsOpen() {
		wm.Start(&ps.midiport)
	}

	db := newDoubleBuffer[int16](BufferFrames, ps.fill, out.WriteMono)

	wm.Start(db.Filler)
	wm.Start(db.Player)

	return wm.Wait()
}

// fill generates samples into the supplied buffer.
func (ps *Picosynth) fill(buf []int16) error {
	currentStep := ps.currentStep.Add(1)
	counts.FillBuffer.Add(1)

	engine := &ps.Engine
	mem := &engine.Memory

	var activity bool
	for m := ps.midiport.Poll(); m != NoMessage; m = ps.midiport.Poll() {
		switch m.Type() {
		case NoteOn:
			ps.keytracker.Receive(m.Note(), true)
		case NoteOff:
			ps.keytracker.Receive(m.Note(), false)
		default:
			lastMIDI.Store(uint32(m))
			ps.ui.InvalidateDisplay()
		}
		activity = true
	}
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
			mem.StorePitch(r, p)
		default:
			mem.StoreSignal(r, Signal(v)<<15) // 0xffff -> 0x7fff8000
		}
		//activity = true
	}

	if activity {
		ps.lastActivityStep.Store(currentStep)
	}

	ps.keytracker.Step()
	mem.StorePitch(VoicePitch, ps.keytracker.Note.Pitch())

	engine.ampEnv.Gate = ps.keytracker.Gate // XXX pass this in a register
	return engine.Fill(buf)
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
		ps.SetVolume(ps.Engine.volume + 1)
	case ButtonVolumeDown:
		ps.SetVolume(ps.Engine.volume - 1)
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
	ps.Engine.volume = max(MinVolume, min(MaxVolume, v))
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

type psDisplay struct {
	ps *Picosynth
}

// Name implements [worker].
func (psd *psDisplay) Name() string {
	return "display"
}

// Run implements [worker].
func (psd *psDisplay) Run() func() error {
	return psd.run
}

func (psd *psDisplay) run() error {
	ps := psd.ps

	d, err := display.Open()
	if err != nil {
		return err
	}

	for _ = range time.Tick(FrameRate) {
		counts.DisplayTick.Add(1)
		now := ps.currentStep.Load()

		// if now-ps.lastActivityStep.Load() > activityTimeout {
		// 	// user inactive
		// 	needBlank := ps.screenBlanked.Swap(true)

		// 	if needBlank {
		// 		d.Sleep()
		// 	}

		// } else {
		// 	// user recently active
		// 	needUnblank := ps.screenBlanked.Swap(false)
		// 	needRefresh := !ps.screenCurrent.Swap(true)

		// 	if needUnblank || needRefresh {
		ps.CurrentPage().Render(d, now)
		// 	}
		// }
	}
	return nil
}
