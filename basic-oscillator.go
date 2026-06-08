package picosynth

// A BasicOscillator combines one [PhaseAccumulator] with one [Shaper].
type BasicOscillator struct {
	PhaseAccumulator
	Shaper Shaper
	Shape  Signal
	Output Signal
}

func (osc *BasicOscillator) Step() {
	osc.PhaseAccumulator.Step()
	osc.Output = osc.Shaper(osc.Phase, osc.Shape)
}

// XXX TODO setpitch, combine modulations, modulate shape, detune (coarse,fine)
