package picosynth

import "gbenson.net/go/picosynth/internal/audio"

type Engine struct {
	mem Memory
	ui  UI

	lfo1 BasicOscillator

	ampEnv Envelope

	osc1 BasicOscillator
	osc2 BasicOscillator

	filt1 ChamberlinFilter
}

func (ps *Engine) init() error {
	ps.Reset()
	return ps.ui.init(&ps.mem)
}

// Reset restores all synthesis parameters to their initial states.
func (ps *Engine) Reset() {
	mem := &ps.mem
	mx := mem.Matrix()

	mx.Reset()

	// Connect each audio oscillator's pitch input to the voice pitch.
	mx.Connect(ModulatedOsc1Pitch, VoicePitch, MaxSignal)
	mx.Connect(ModulatedOsc2Pitch, VoicePitch, MaxSignal)

	// XXX above is general reset, below is a specific patch
	// XXX (this also sets some things that aren't in the matrix... yet?)
	ps.lfo1.Frequency = 10 * Hz
	ps.lfo1.Shaper = SineShaper
	mx.Connect(ModulatedOsc1Pitch, LFO1Output, MaxSignal>>9)

	ps.osc1.Shaper = TriSawShaper
	ps.osc1.Shape = MaxSignal // rising saw
	mem.StoreSignal(Osc1Level, MaxSignal)

	ps.osc2.Shaper = SineShaper
	ps.osc2.Shape = 0                             // zero phase shift
	mem.StoreSignal(Osc2Pitch, -1*Signal(Octave)) // XXX make pitch signed!

	mem.Store(Filt1Mode, uint32(FilterChamberlinLowPass))
}

// Run is the main entry point of the firmware.
func (ps *Engine) Run() error {
	if err := ps.init(); err != nil {
		return err
	}

	out, err := audio.Open(SampleRate)
	if err != nil {
		return err
	}
	defer out.Close()

	const numWorkers = 3 // filler, player, ui
	wm := newWorkerManager(numWorkers)
	wm.Start(&ps.ui)

	db := newDoubleBuffer[int16](BufferFrames, ps.Fill, out.WriteMono)

	wm.Start(db.Filler)
	wm.Start(db.Player)

	return wm.Wait()
}

// Fill generates samples into the supplied buffer.
func (ps *Engine) Fill(buf []int16) error {
	ps.ui.Step()
	ps.ampEnv.Gate = ps.ui.keytracker.Gate

	// Final shift converts 32-bit to 16 and applies ps.volume.
	var finalShift int
	if v := ps.ui.volume; v <= MinVolume {
		finalShift = 32 // silence
	} else {
		finalShift = (16 + (MaxVolume - v)) & 0x1f
	}

	for i := range buf {
		if i&15 == 0 {
			// This yields the processor when i==0: this is expected.
			// When len(buf) == 32, for example, we yield as follows:
			//  - per-Fill operations ([UI.Step] etc)
			//  - yield processor
			//  - generate 16 audio frames
			//  - yield processor
			//  - generate 16 audio frames
			//  - return (no yield here)
			gosched()
		}
		ps.mem.Step()

		ps.lfo1.Step()
		ps.mem.StoreSignal(LFO1Output, ps.lfo1.Output)

		osc1Pitch := ps.mem.LoadPitch(ModulatedOsc1Pitch)
		ps.osc1.Frequency = osc1Pitch.Frequency()
		ps.osc1.Step()

		osc2Pitch := ps.mem.LoadPitch(ModulatedOsc2Pitch)
		ps.osc2.Frequency = osc2Pitch.Frequency()
		ps.osc2.Step()

		premix := ps.osc1.Output.Mul(ps.mem.LoadSignal(ModulatedOsc1Level))
		premix += ps.osc2.Output.Mul(ps.mem.LoadSignal(ModulatedOsc2Level))

		filt1Cutoff := ps.mem.LoadPitch(ModulatedFilt1Cutoff)

		ps.filt1.Input = premix
		ps.filt1.Frequency = filt1Cutoff.Frequency()
		ps.filt1.Resonance = ps.mem.LoadSignal(ModulatedFilt1Resonance)
		ps.filt1.Step()

		var output Signal
		switch FilterMode(ps.mem.Load(Filt1Mode)) {
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
