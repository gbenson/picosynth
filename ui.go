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

	if v := ui.encoder.Read(); v < 0 {
		println("decrease")
		activity = true
	} else if v > 0 {
		println("increase")
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

	if activity {
		ui.display.KeepAlive()
	}

	ui.kt.Step()
	ui.mem.StorePitch(VoicePitch, ui.kt.Note.Pitch())
}

func (ui *UI) onButtonDown(sc Scancode) {
	ui.buttonDownStep[sc] = ui.currentStep
}

func (ui *UI) onButtonUp(sc Scancode) {
	holdTime := ui.currentStep - ui.buttonDownStep[sc]
	if holdTime < longPressTimeout {
		ui.onPress(sc)
	} else {
		ui.onLongPress(sc)
	}
}

// onPress is called when a button is released after having been held
// for less than LongPressTimeout.
func (ui *UI) onPress(sc Scancode) {
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
		ui.editor.onButton(sc)
	}
}

// onLongPress is called when a button is released after having been
// held for LongPressTimeout or longer.
func (ui *UI) onLongPress(sc Scancode) {
	println("long press", sc, "ignored")
}

func (ui *UI) setOctave(v int) {
	ui.octave = max(MinOctave, min(MaxOctave, v))
	ui.kt.Transpose = ui.octave * 12
}

func (ui *UI) setVolume(v int) {
	ui.volume = max(MinVolume, min(MaxVolume, v))
}
