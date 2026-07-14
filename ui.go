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
// before a press turns into a long press.
var longPressTimeout = int(LongPressTimeout) * FillRate / int(time.Second)

var potRegisters = []int{
	Filt1Cutoff,
	Filt1Resonance,
}

type UI struct {
	mem *Memory

	encoder *encoder.Encoder
	pots    []adc.ADC
	display display.Display

	editor MemoryEditor

	ks KeyScanner
	kt KeyTracker

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

	ui.editor.init(ui.mem, &ui.display)

	ui.kt.init()

	ui.setOctave(InitialOctave)
	ui.setVolume(InitialVolume)

	return nil
}

func (ui *UI) Step() {
	ui.currentStep++

	var activity bool
	for e := ui.ks.Poll(); e != NoEvent; e = ui.ks.Poll() {
		sc := e.Scancode()
		if note := sc.Note(); note.IsValid() {
			ui.kt.Receive(note, e.Down())
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

	ui.kt.Step()
	ui.mem.StorePitch(VoicePitch, ui.kt.Note.Pitch())
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
	ui.kt.Transpose = ui.octave * 12
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

// Store replaces the contents of register n with v.
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

func (ui *UI) onButton(sc Scancode, longPress bool) {
	if longPress {
		println("button", sc, "(long press) ignored")
	} else {
		ui.editor.onButton(sc)
	}
}

func (ui *UI) onEncoder(delta int) {
	println("encoder moved:", delta, "(unhandled)")
}
