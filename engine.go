package picosynth

import (
	"math/bits"
	"time"

	"gbenson.net/go/picosynth/internal/adc"
	"gbenson.net/go/picosynth/internal/audio"
	"gbenson.net/go/picosynth/internal/display"
)

const (
	SampleRate = 48000
	MaxLatency = 10 * time.Millisecond

	// Split MaxLatency between scanning keys and playing audio.
	// Currently maxLatency is *half* of MaxLatency because:
	//  - The KeyScanner loop takes the first part, the limiting
	//    case being a key or button changing state the instant
	//    after being scanned, and a full loop has to complete
	//    before a KeyEvent will be emitted.
	//  - The time it takes to play the audio buffer takes up
	//    the second part, the limiting case being that the first
	//    sample generated into every buffer has to wait for the
	//    entire other buffer to play before it will be output.
	maxLatency = MaxLatency / 2

	// maxBufferFrames is the maximum number of audio frames we can
	// buffer without exceeding maxLatency.
	maxBufferFrames = uint(SampleRate * maxLatency / time.Second)

	MinVolume     = 0
	MaxVolume     = 10
	InitialVolume = 7

	MinOctave     = -3
	MaxOctave     = 3
	InitialOctave = -1
)

var (
	// BufferFrames is the size of buffers passed to [Engine.Fill].
	// The SDL emulator requires this to be a power of two.
	BufferFrames = 1 << (bits.Len(maxBufferFrames) - 1)

	// FillRate is the rate at which [Engine.Fill] is called.
	//
	// By extension this is the rate at which things that happen
	// once per fill happen: scanning keys and buttons, reading
	// potentiometer values, etc.
	FillRate = SampleRate / BufferFrames
)

var potRegisters = []int{
	Filt1Cutoff,
	Filt1Resonance,
}

type Engine struct {
	Memory Memory

	pots    []adc.ADC
	display display.Display
	editor  MemoryEditor

	ks KeyScanner
	kt KeyTracker

	octave int

	lfo1 BasicOscillator

	ampEnv Envelope

	osc1 BasicOscillator
	osc2 BasicOscillator

	filt1 ChamberlinFilter

	volume int
}

func (ps *Engine) init() {
	ps.editor.init(&ps.Memory, &ps.display)

	ps.kt.init()

	ps.setOctave(InitialOctave)
	ps.setVolume(InitialVolume)

	ps.Reset()
}

// Reset restores all synthesis parameters to their initial states.
func (ps *Engine) Reset() {
	ps.Memory.Matrix().Reset()

	// Connect each audio oscillator's pitch input to the voice pitch.
	ps.Connect(ModulatedOsc1Pitch, VoicePitch, MaxSignal)
	ps.Connect(ModulatedOsc2Pitch, VoicePitch, MaxSignal)

	// XXX above is general reset, below is a specific patch
	// XXX (this also sets some things that aren't in the matrix... yet?)
	ps.lfo1.Frequency = 10 * Hz
	ps.lfo1.Shaper = SineShaper
	ps.Connect(ModulatedOsc1Pitch, LFO1Output, MaxSignal>>9)

	ps.osc1.Shaper = TriSawShaper
	ps.osc1.Shape = MaxSignal // rising saw
	ps.store(Osc1Level, MaxSignal)

	ps.osc2.Shaper = SineShaper
	ps.osc2.Shape = 0                      // zero phase shift
	ps.store(Osc2Pitch, -1*Signal(Octave)) // XXX make pitch signed!

	ps.storeFilterMode(Filt1Mode, FilterChamberlinLowPass)
}

// Connect adds or updates a connection in the modulation matrix.
func (ps *Engine) Connect(dst, src int, gain Signal) {
	ps.Memory.Matrix().Connect(dst, src, gain)
}

// Run is the main entry point of the firmware.
func (ps *Engine) Run() error {
	ps.init()

	out, err := audio.Open(SampleRate)
	if err != nil {
		return err
	}
	defer out.Close()

	ps.pots = make([]adc.ADC, len(potRegisters))
	for i, _ := range ps.pots {
		k, err := adc.Open(i)
		if err != nil {
			return err
		}
		ps.pots[i] = k
	}

	const numWorkers = 4 // display, keyscanner, filler, player
	wm := newWorkerManager(numWorkers)
	wm.Start(&ps.display)
	wm.Start(&ps.ks)

	db := newDoubleBuffer[int16](BufferFrames, ps.Fill, out.WriteMono)

	wm.Start(db.Filler)
	wm.Start(db.Player)

	return wm.Wait()
}

// Fill generates samples into the supplied buffer.
func (ps *Engine) Fill(buf []int16) error {
	var activity bool
	for e := ps.ks.Poll(); e != NoEvent; e = ps.ks.Poll() {
		sc := e.Scancode()
		if note := sc.Note(); note.IsValid() {
			ps.kt.Receive(note, e.Down())
		} else if !e.Down() {
			ps.onButton(sc)
		}
		activity = true
	}
	for i, _ := range ps.pots {
		v := ps.pots[i].Get()
		r := potRegisters[i]
		switch r {
		case Filt1Cutoff:
			p := Pitch(v)
			p *= (MaxAudiblePitch - MinAudiblePitch) >> 16
			p += MinAudiblePitch
			ps.storePitch(r, p)
		default:
			ps.store(r, Signal(v)<<15) // 0xffff -> 0x7fff8000
		}
	}
	if activity {
		ps.display.KeepAlive()
	}

	// Final shift converts 32-bit to 16 and applies ps.volume.
	var finalShift int
	if v := ps.volume; v <= MinVolume {
		finalShift = 32 // silence
	} else {
		finalShift = (16 + (MaxVolume - v)) & 0x1f
	}

	ps.kt.Step()
	ps.storePitch(VoicePitch, ps.kt.Note.Pitch())
	ps.ampEnv.Gate = ps.kt.Gate

	for i := range buf {
		ps.Memory.Step()

		ps.lfo1.Step()
		ps.store(LFO1Output, ps.lfo1.Output)

		osc1Pitch := ps.loadPitch(ModulatedOsc1Pitch)
		ps.osc1.Frequency = osc1Pitch.Frequency()
		ps.osc1.Step()

		osc2Pitch := ps.loadPitch(ModulatedOsc2Pitch)
		ps.osc2.Frequency = osc2Pitch.Frequency()
		ps.osc2.Step()

		premix := ps.osc1.Output.Mul(ps.load(Osc1Level))
		premix += ps.osc2.Output.Mul(ps.load(Osc2Level))

		filt1Cutoff := ps.loadPitch(ModulatedFilt1Cutoff)

		ps.filt1.Input = premix
		ps.filt1.Frequency = filt1Cutoff.Frequency()
		ps.filt1.Resonance = ps.load(Filt1Resonance)
		ps.filt1.Step()

		var output Signal
		switch ps.loadFilterMode(Filt1Mode) {
		default:
			output = premix // bypass
		case FilterChamberlinLowPass:
			output = ps.filt1.Lout << 1 // XXX ???
		case FilterChamberlinHighPass:
			output = ps.filt1.Hout << 1 // XXX ???
		case FilterChamberlinBandPass:
			output = ps.filt1.Bout << 1 // XXX ???
		case FilterChamberlinNotch:
			output = ps.filt1.Nout << 1 // XXX ???
		}

		ps.ampEnv.Step()
		output = output.Mul(ps.ampEnv.Level)

		// Convert 32-bit to 16 and apply ps.volume.
		buf[i] = int16(output >> finalShift)
	}

	return nil
}

// load returns the contents of register n as a Signal.
func (ps *Engine) load(n int) Signal {
	r := ps.Memory.Register(n)
	return Signal(r.Load())
}

// store replaces the contents of register n with v.
func (ps *Engine) store(n int, v Signal) {
	r := ps.Memory.Register(n)
	r.Store(uint32(v))
}

// loadPitch returns the contents of register n as a Pitch.
func (ps *Engine) loadPitch(n int) Pitch {
	r := ps.Memory.Register(n)
	return Pitch(r.Load())
}

// storePitch replaces the contents of register n with p.
func (ps *Engine) storePitch(n int, p Pitch) {
	r := ps.Memory.Register(n)
	r.Store(uint32(p))
}

// loadFilterMode returns the contents of register n as a FilterMode.
func (ps *Engine) loadFilterMode(n int) FilterMode {
	r := ps.Memory.Register(n)
	return FilterMode(r.Load())
}

// storeFilterMode replaces the contents of register n with p.
func (ps *Engine) storeFilterMode(n int, m FilterMode) {
	r := ps.Memory.Register(n)
	r.Store(uint32(m))
}

func (ps *Engine) onButton(sc Scancode) {
	switch sc {
	case ButtonVolumeUp:
		ps.setVolume(ps.volume + 1)
	case ButtonVolumeDown:
		ps.setVolume(ps.volume - 1)
	case ButtonTempoUp:
		ps.setOctave(ps.octave + 1)
	case ButtonTempoDown:
		ps.setOctave(ps.octave - 1)
	default:
		ps.editor.onButton(sc)
	}
}

func (ps *Engine) setOctave(v int) {
	ps.octave = max(MinOctave, min(MaxOctave, v))
	ps.kt.Transpose = ps.octave * 12
}

func (ps *Engine) setVolume(v int) {
	ps.volume = max(MinVolume, min(MaxVolume, v))
}
