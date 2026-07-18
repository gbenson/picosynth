package picosynth

import (
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
)

// longPressTimeout is the number of UI steps a button must be held
// before a press turns into a long press.  The right shifts cancel
// out, but prevent overflow on 32-bit platforms.
var longPressTimeout = int(LongPressTimeout>>8) * FillRate / int(time.Second>>8)

var potRegisters = []int{
	Filt1Cutoff,
	Filt1Resonance,
}

type UI struct {
	mem *Memory

	encoder *encoder.Encoder
	pots    []adc.ADC
	display display.Display

	keyscanner KeyScanner
	keytracker KeyTracker

	octave int
	volume int

	// currentStep is the step number of the current step.
	// It's incremented at the start of every [UI.Step].
	currentStep int

	// buttonDownStep holds the step numbers at the last
	// time each button transitioned from up to down.
	buttonDownStep [numScancodes]int

	// lastEventStep holds the step number when the last
	// UIEvent was sent.
	lastEventStep int

	// filler->ui communication
	events chan UIEvent

	// ui->filler communication
	storeTarget int
	storeValue  chan uint32
	valueStored chan bool

	currentPage Page
	pages       []Page
}

func (ui *UI) init(m *Memory) error {
	ui.mem = m

	ui.storeValue = make(chan uint32)
	ui.valueStored = make(chan bool)

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

	ui.AddPage(&MemoryEditor{})
	for _, pg := range ParameterGroups {
		ui.AddPage(NewParameterGroupPage(pg))
	}

	return nil
}

func (ui *UI) AddPage(p Page) {
	p.OnInit(ui)
	ui.pages = append(ui.pages, p)
}

func (ui *UI) Step() {
	ui.currentStep++

	var activity bool
	for e := ui.keyscanner.Poll(); e != NoEvent; e = ui.keyscanner.Poll() {
		sc := e.Scancode()
		if note := sc.Note(); note.IsValid() {
			ui.keytracker.Receive(note, e.Down())
		} else if e.Down() {
			ui.onButtonDown(sc)
		} else {
			ui.onButtonUp(sc)
		}
		activity = true
	}

	if v := ui.encoder.Read(); v != 0 {
		ui.sendEvent(EncoderMovedEvent(v))
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
	}

	if activity && ui.lastEventStep != ui.currentStep {
		ui.sendEvent(GenericActivityEvent) // inhibit screensaver
	}

	select {
	case v := <-ui.storeValue:
		ui.mem.Store(ui.storeTarget, v)
		ui.valueStored <- true
	default:
	}

	ui.keytracker.Step()
	ui.mem.StorePitch(VoicePitch, ui.keytracker.Note.Pitch())
}

func (ui *UI) onButtonDown(sc Scancode) {
	ui.buttonDownStep[sc] = ui.currentStep
}

func (ui *UI) onButtonUp(sc Scancode) {
	holdTime := ui.currentStep - ui.buttonDownStep[sc]
	if holdTime > longPressTimeout {
		ui.sendEvent(LongPressEvent(sc))
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
		ui.sendEvent(ButtonPressEvent(sc))
	}
}

func (ui *UI) setOctave(v int) {
	ui.octave = max(MinOctave, min(MaxOctave, v))
	ui.keytracker.Transpose = ui.octave * 12
}

func (ui *UI) setVolume(v int) {
	ui.volume = max(MinVolume, min(MaxVolume, v))
}

func (ui *UI) sendEvent(e UIEvent) {
	// non-blocking to prevent slow UI glitching audio
	select {
	case ui.events <- e:
	default:
		println(e.String(), "dropped")
	}
	ui.lastEventStep = ui.currentStep
}

// Display implements [ui.Engine].
func (ui *UI) Display() *display.Display {
	return &ui.display
}

// Load implements [ui.Engine].
func (ui *UI) Load(n int) uint32 {
	// ui.Engine.Load absolutely should not step!
	return ui.mem.load(n)
}

// Store implements [ui.Engine].
func (ui *UI) Store(n int, v uint32) {
	ui.storeTarget = n
	ui.storeValue <- v
	<-ui.valueStored
}

// Name implements [worker].
func (ui *UI) Name() string {
	return "ui"
}

// Run implements [worker].
func (ui *UI) Run() func() error {
	return ui.run
}

func (ui *UI) run() error {
	if ui.events != nil {
		panic("already started")
	}

	ui.events = make(chan UIEvent, 8) // small buffer
	for e := range ui.events {
		switch e.Type() {
		case ButtonPressEventType:
			ui.onButton(e.Scancode(), false)
		case LongPressEventType:
			ui.onButton(e.Scancode(), true)
		case EncoderMovedEventType:
			ui.onEncoder(e.Delta())
		case GenericActivityEvent:
			ui.display.KeepAlive()
		default:
			println(e.String(), "unhandled")
		}
	}

	return nil
}

func (ui *UI) onButton(sc Scancode, longpress bool) {
	if p := ui.currentPage; p != nil {
		if p.OnButton(sc, longpress) {
			return // handled
		}
	}

	for _, p := range ui.pages {
		if p == ui.currentPage {
			continue
		}
		if p.OnButton(sc, longpress) {
			ui.currentPage = p
			return
		}
	}

	if longpress {
		println("button", sc, "(long press) ignored")
	} else {
		println("button", sc, "(short press) ignored")
	}
}

func (ui *UI) onEncoder(delta int) {
	if p := ui.currentPage; p != nil {
		p.OnEncoder(delta)
	}
}

func (ui *UI) HasFocus(p Page) bool {
	return p == ui.currentPage
}
