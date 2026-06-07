package picosynth

// PhaseAccumulator outputs a rising sawtooth wave of the specified frequency.
//
// See https://en.wikipedia.org/wiki/Numerically_controlled_oscillator#Phase_accumulator.
type PhaseAccumulator struct {
	Frequency Frequency
	Phase     Signal
}

func (pa *PhaseAccumulator) Step() {
	pa.Phase += Signal(pa.Frequency)
}
